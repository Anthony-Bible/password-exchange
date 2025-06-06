package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
)

// Error constants for reminder processing
var (
	ErrInvalidCheckAfterHours  = errors.New("checkAfterHours must be between 1 and 8760 hours")
	ErrInvalidMaxReminders     = errors.New("maxReminders must be between 1 and 10")
	ErrInvalidReminderInterval = errors.New("reminderInterval must be between 1 and 720 hours")
	ErrCircuitBreakerOpen      = errors.New("circuit breaker is open")
	ErrMaxRetriesExceeded      = errors.New("maximum retries exceeded")
)

// Validation constants for reminder configuration
const (
	MinCheckAfterHours  = 1    // Minimum 1 hour
	MaxCheckAfterHours  = 8760 // Maximum 1 year (365 * 24)
	MinMaxReminders     = 1    // Minimum 1 reminder
	MaxMaxReminders     = 10   // Maximum 10 reminders
	MinReminderInterval = 1    // Minimum 1 hour between reminders
	MaxReminderInterval = 720  // Maximum 30 days (30 * 24)

	// Error recovery constants
	MaxRetries              = 3
	BaseRetryDelay          = 100 * time.Millisecond
	MaxRetryDelay           = 5 * time.Second
	CircuitBreakerThreshold = 5
	CircuitBreakerTimeout   = 30 * time.Second
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern for error recovery.
// It prevents cascading failures by temporarily blocking requests when error rates are high.
// States: Closed (normal), Open (blocking), HalfOpen (testing recovery)
type CircuitBreaker struct {
	failureCount    int                    // Number of consecutive failures
	lastFailureTime time.Time              // Time of the last failure
	state           CircuitBreakerState    // Current state of the circuit breaker
	logger          secondary.LoggerPort   // Logger for recording state transitions
}

// StorageRepository defines the interface for storage operations
type StorageRepository interface {
	GetUnviewedMessagesForReminders(ctx context.Context, checkAfterHours, maxReminders, reminderIntervalHours int) ([]*UnviewedMessage, error)
	GetReminderHistory(ctx context.Context, messageID int) ([]*ReminderLogEntry, error)
	LogReminderSent(ctx context.Context, messageID int, recipientEmail string) error
}

// NotificationPublisher defines the interface for publishing notifications to queue
type NotificationPublisher interface {
	PublishNotification(ctx context.Context, req NotificationRequest) error
}

// ReminderService provides reminder processing operations
type ReminderService struct {
	storageRepo           secondary.StoragePort
	notificationPublisher secondary.NotificationPort
	circuitBreaker        *CircuitBreaker
	logger                secondary.LoggerPort
	config                secondary.ConfigPort
	validation            secondary.ValidationPort
}

// NewReminderService creates a new reminder service
func NewReminderService(storageRepo secondary.StoragePort, notificationPublisher secondary.NotificationPort, logger secondary.LoggerPort, config secondary.ConfigPort, validation secondary.ValidationPort) *ReminderService {
	return &ReminderService{
		storageRepo:           storageRepo,
		notificationPublisher: notificationPublisher,
		circuitBreaker: &CircuitBreaker{
			state:  CircuitBreakerClosed,
			logger: logger,
		},
		logger:     logger,
		config:     config,
		validation: validation,
	}
}

// ProcessReminders finds and processes messages eligible for reminder emails
func (r *ReminderService) ProcessReminders(ctx context.Context, reminderConfig ReminderConfig) error {
	// Check context cancellation early
	if err := ctx.Err(); err != nil {
		return err
	}

	// Validate configuration
	if err := r.validateReminderConfig(reminderConfig); err != nil {
		r.logger.Error().
			Err(err).
			Str("operation", "validate_config").
			Int("checkAfterHours", reminderConfig.CheckAfterHours).
			Int("maxReminders", reminderConfig.MaxReminders).
			Int("reminderInterval", reminderConfig.Interval).
			Msg("Invalid reminder configuration")
		return err
	}

	if !reminderConfig.Enabled {
		r.logger.Info().Bool("enabled", false).Msg("Reminder system is disabled")
		return nil
	}

	r.logger.Info().
		Str("operation", "start_processing").
		Bool("enabled", reminderConfig.Enabled).
		Int("checkAfterHours", reminderConfig.CheckAfterHours).
		Int("maxReminders", reminderConfig.MaxReminders).
		Msg("Starting reminder email processing")

	var messages []*UnviewedMessage

	// Get unviewed messages with retry logic
	err := r.retryWithBackoff(ctx, func() error {
		var err error
		messages, err = r.storageRepo.GetUnviewedMessagesForReminders(
			ctx,
			reminderConfig.CheckAfterHours,
			reminderConfig.MaxReminders,
			reminderConfig.Interval,
		)
		return err
	}, "get_unviewed_messages")

	if err != nil {
		return fmt.Errorf("failed to get unviewed messages: %w", err)
	}

	r.logger.Info().Int("count", len(messages)).Msg("Found messages eligible for reminders")

	if len(messages) == 0 {
		r.logger.Info().Msg("No messages found requiring reminders")
		return nil
	}

	// Process each message with individual error recovery
	// Strategy: Continue processing other messages even if some fail (graceful degradation)
	processedCount := 0
	errorCount := 0
	for _, message := range messages {
		// Create reminder request
		reminderReq := ReminderRequest{
			MessageID:      message.MessageID,
			UniqueID:       message.UniqueID,
			RecipientEmail: message.RecipientEmail,
			DaysOld:        message.DaysOld,
			DecryptionURL:  "", // Empty - template now references original email
		}

		// Use retry logic for each message processing
		err := r.retryWithBackoff(ctx, func() error {
			return r.ProcessMessageReminder(ctx, reminderReq)
		}, fmt.Sprintf("process_message_%d", message.MessageID))

		if err != nil {
			errorCount++
			r.logger.Error().
				Err(err).
				Int("messageID", message.MessageID).
				Str("email", message.RecipientEmail).
				Int("daysOld", message.DaysOld).
				Str("operation", "process_reminder").
				Msg("Failed to process reminder for message after all retry attempts")
			continue // Continue processing other messages
		}
		processedCount++
	}

	r.logger.Info().
		Int("totalMessages", len(messages)).
		Int("processedCount", processedCount).
		Int("errorCount", errorCount).
		Msg("Reminder processing completed")

	// Implement graceful degradation: return success if at least some messages were processed
	// This allows partial success rather than all-or-nothing failure
	if processedCount > 0 {
		r.logger.Info().
			Int("processedCount", processedCount).
			Int("errorCount", errorCount).
			Float64("successRate", float64(processedCount)/float64(len(messages))*100).
			Msg("Reminder processing completed with partial success")
		return nil
	}

	// If no messages were processed and we had errors, this indicates a more serious issue
	if errorCount > 0 {
		r.logger.Error().
			Int("errorCount", errorCount).
			Int("totalMessages", len(messages)).
			Msg("Failed to process any reminder messages")
		return fmt.Errorf("failed to process any of %d reminder messages", len(messages))
	}

	return nil
}

// ProcessMessageReminder sends a reminder email for a specific message
func (r *ReminderService) ProcessMessageReminder(ctx context.Context, reminderRequest ReminderRequest) error {
	// Validate request parameters
	if reminderRequest.MessageID <= 0 {
		return fmt.Errorf("messageID must be greater than 0, got %d", reminderRequest.MessageID)
	}

	if reminderRequest.UniqueID == "" {
		return fmt.Errorf("uniqueID cannot be empty")
	}

	if reminderRequest.DaysOld < 0 {
		return fmt.Errorf("daysOld must be non-negative, got %d", reminderRequest.DaysOld)
	}

	// Validate recipient email address
	if err := r.validation.ValidateEmail(reminderRequest.RecipientEmail); err != nil {
		return fmt.Errorf("invalid recipient email address: %w", err)
	}

	r.logger.Info().
		Int("messageID", reminderRequest.MessageID).
		Str("email", reminderRequest.RecipientEmail).
		Int("daysOld", reminderRequest.DaysOld).
		Str("operation", "process_message_reminder").
		Msg("Processing reminder for message")

	// Get reminder history to determine reminder count with retry logic
	var history []*ReminderLogEntry
	err := r.retryWithBackoff(ctx, func() error {
		var err error
		history, err = r.storageRepo.GetReminderHistory(ctx, reminderRequest.MessageID)
		return err
	}, fmt.Sprintf("get_reminder_history_%d", reminderRequest.MessageID))

	if err != nil {
		return fmt.Errorf("failed to get reminder history for messageID %d: %w", reminderRequest.MessageID, err)
	}

	reminderCount := 0
	if len(history) > 0 {
		reminderCount = history[0].ReminderCount
	}

	// Update reminder number in request
	reminderRequest.ReminderNumber = reminderCount + 1

	// Publish reminder notification to queue
	notificationReq := NotificationRequest{
		To:            reminderRequest.RecipientEmail,
		From:          r.config.GetServerEmail(),
		FromName:      r.config.GetServerName(),
		Subject:       fmt.Sprintf(r.config.GetReminderNotificationSubject(), reminderRequest.ReminderNumber),
		MessageURL:    reminderRequest.DecryptionURL,
		MessageContent: r.config.GetReminderMessageContent(),
	}

	if r.notificationPublisher != nil {
		err = r.notificationPublisher.PublishNotification(ctx, notificationReq)
		if err != nil {
			return fmt.Errorf("failed to publish reminder notification for messageID %d: %w", reminderRequest.MessageID, err)
		}
	} else {
		r.logger.Info().
			Int("messageID", reminderRequest.MessageID).
			Str("recipientEmail", reminderRequest.RecipientEmail).
			Int("reminderNumber", reminderRequest.ReminderNumber).
			Msg("Notification publisher not configured - reminder would be published")
	}

	// Record that we sent a reminder with retry logic
	err = r.retryWithBackoff(ctx, func() error {
		return r.storageRepo.LogReminderSent(ctx, reminderRequest.MessageID, reminderRequest.RecipientEmail)
	}, fmt.Sprintf("log_reminder_sent_%d", reminderRequest.MessageID))

	if err != nil {
		return fmt.Errorf("failed to log reminder sent for messageID %d: %w", reminderRequest.MessageID, err)
	}

	r.logger.Info().
		Int("messageID", reminderRequest.MessageID).
		Str("email", reminderRequest.RecipientEmail).
		Int("reminderNumber", reminderRequest.ReminderNumber).
		Str("operation", "reminder_sent").
		Msg("Reminder email sent successfully")
	return nil
}

// retryWithBackoff executes a function with exponential backoff retry logic.
// Implements resilience pattern: starts with 100ms delay, doubles each retry, max 5s delay.
// Integrates with circuit breaker to prevent repeated attempts when system is failing.
func (r *ReminderService) retryWithBackoff(ctx context.Context, operation func() error, operationName string) error {
	var lastErr error

	for attempt := 0; attempt < MaxRetries; attempt++ {
		// Check circuit breaker before attempting operation
		if err := r.circuitBreaker.CanExecute(); err != nil {
			r.logger.Warn().
				Err(err).
				Str("operation", operationName).
				Int("attempt", attempt+1).
				Msg("Circuit breaker preventing operation execution")
			return err
		}

		// Execute the operation
		err := operation()
		if err == nil {
			// Success - record success and return
			r.circuitBreaker.RecordSuccess()
			if attempt > 0 {
			r.logger.Info().
				Str("operation", operationName).
				Int("successfulAttempt", attempt+1).
				Msg("Operation succeeded after retry")
			}
			return nil
		}

		lastErr = err
		r.circuitBreaker.RecordFailure()

		// Don't retry on last attempt
		if attempt == MaxRetries-1 {
			break
		}

		// Calculate delay with exponential backoff
		delay := BaseRetryDelay * time.Duration(1<<uint(attempt))
		if delay > MaxRetryDelay {
			delay = MaxRetryDelay
		}

		r.logger.Warn().
			Err(err).
			Str("operation", operationName).
			Int("attempt", attempt+1).
			Dur("retryDelay", delay).
			Msg("Operation failed, retrying with backoff")

		// Wait before retry with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	r.logger.Error().
		Err(lastErr).
		Str("operation", operationName).
		Int("maxRetries", MaxRetries).
		Msg("Operation failed after all retry attempts")

	return fmt.Errorf("%w: %s failed after %d attempts: %v", ErrMaxRetriesExceeded, operationName, MaxRetries, lastErr)
}

// validateReminderConfig validates all reminder configuration parameters
func (r *ReminderService) validateReminderConfig(reminderConfig ReminderConfig) error {
	if reminderConfig.CheckAfterHours < MinCheckAfterHours || reminderConfig.CheckAfterHours > MaxCheckAfterHours {
		return fmt.Errorf("%w: got %d", ErrInvalidCheckAfterHours, reminderConfig.CheckAfterHours)
	}

	if reminderConfig.MaxReminders < MinMaxReminders || reminderConfig.MaxReminders > MaxMaxReminders {
		return fmt.Errorf("%w: got %d", ErrInvalidMaxReminders, reminderConfig.MaxReminders)
	}

	if reminderConfig.Interval < MinReminderInterval || reminderConfig.Interval > MaxReminderInterval {
		return fmt.Errorf("%w: got %d", ErrInvalidReminderInterval, reminderConfig.Interval)
	}

	return nil
}

// CanExecute checks if the circuit breaker allows execution
func (circuitBreaker *CircuitBreaker) CanExecute() error {
	switch circuitBreaker.state {
	case CircuitBreakerClosed:
		return nil
	case CircuitBreakerOpen:
		if time.Since(circuitBreaker.lastFailureTime) > CircuitBreakerTimeout {
			circuitBreaker.state = CircuitBreakerHalfOpen
			
			// Log state transition from OPEN to HALF_OPEN
			if circuitBreaker.logger != nil {
				circuitBreaker.logger.Info().
					Str("state", "HALF_OPEN").
					Str("reason", "timeout_expired").
					Dur("timeout", CircuitBreakerTimeout).
					Msg("Circuit breaker transitioned to HALF_OPEN state after timeout")
			}
			return nil
		}
		return ErrCircuitBreakerOpen
	case CircuitBreakerHalfOpen:
		return nil
	default:
		return nil
	}
}

// RecordSuccess records a successful operation
func (circuitBreaker *CircuitBreaker) RecordSuccess() {
	oldState := circuitBreaker.state
	circuitBreaker.failureCount = 0
	circuitBreaker.state = CircuitBreakerClosed
	
	// Log state transition to CLOSED when coming from HALF_OPEN
	if oldState == CircuitBreakerHalfOpen && circuitBreaker.logger != nil {
		circuitBreaker.logger.Info().
			Str("state", "CLOSED").
			Str("reason", "operation_succeeded").
			Msg("Circuit breaker transitioned to CLOSED state after successful operation")
	}
}

// RecordFailure records a failed operation
func (circuitBreaker *CircuitBreaker) RecordFailure() {
	circuitBreaker.failureCount++
	circuitBreaker.lastFailureTime = time.Now()

	if circuitBreaker.failureCount >= CircuitBreakerThreshold {
		circuitBreaker.state = CircuitBreakerOpen
		
		// Log state transition to OPEN when threshold is exceeded
		if circuitBreaker.logger != nil {
			circuitBreaker.logger.Error().
				Str("state", "OPEN").
				Str("reason", "threshold_exceeded").
				Int("failures", CircuitBreakerThreshold).
				Msg("Circuit breaker transitioned to OPEN state due to repeated failures")
		}
	}
}
