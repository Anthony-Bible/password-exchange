package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
	"github.com/rs/zerolog/log"
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
	MinCheckAfterHours   = 1     // Minimum 1 hour
	MaxCheckAfterHours   = 8760  // Maximum 1 year (365 * 24)
	MinMaxReminders      = 1     // Minimum 1 reminder
	MaxMaxReminders      = 10    // Maximum 10 reminders
	MinReminderInterval  = 1     // Minimum 1 hour between reminders
	MaxReminderInterval  = 720   // Maximum 30 days (30 * 24)

	// Error recovery constants
	MaxRetries           = 3
	BaseRetryDelay       = 100 * time.Millisecond
	MaxRetryDelay        = 5 * time.Second
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

// CircuitBreaker implements the circuit breaker pattern for error recovery
type CircuitBreaker struct {
	failureCount    int
	lastFailureTime time.Time
	state          CircuitBreakerState
}

// StorageRepository defines the interface for storage operations
type StorageRepository interface {
	GetUnviewedMessagesForReminders(ctx context.Context, checkAfterHours, maxReminders int) ([]*UnviewedMessage, error)
	GetReminderHistory(ctx context.Context, messageID int) ([]*ReminderLogEntry, error)
	LogReminderSent(ctx context.Context, messageID int, recipientEmail string) error
}

// ReminderService provides reminder processing operations
type ReminderService struct {
	storageRepo     StorageRepository
	emailSender     NotificationSender
	circuitBreaker  *CircuitBreaker
}

// NewReminderService creates a new reminder service
func NewReminderService(storageRepo StorageRepository, emailSender NotificationSender) *ReminderService {
	return &ReminderService{
		storageRepo: storageRepo,
		emailSender: emailSender,
		circuitBreaker: &CircuitBreaker{
			state: CircuitBreakerClosed,
		},
	}
}

// ProcessReminders finds and processes messages eligible for reminder emails
func (r *ReminderService) ProcessReminders(ctx context.Context, config ReminderConfig) error {
	// Validate configuration
	if err := r.validateReminderConfig(config); err != nil {
		log.Error().
			Err(err).
			Str("operation", "validate_config").
			Int("checkAfterHours", config.CheckAfterHours).
			Int("maxReminders", config.MaxReminders).
			Int("reminderInterval", config.Interval).
			Msg("Invalid reminder configuration")
		return err
	}

	if !config.Enabled {
		log.Info().Bool("enabled", false).Msg("Reminder system is disabled")
		return nil
	}

	log.Info().
		Str("operation", "start_processing").
		Bool("enabled", config.Enabled).
		Int("checkAfterHours", config.CheckAfterHours).
		Int("maxReminders", config.MaxReminders).
		Msg("Starting reminder email processing")

	var messages []*UnviewedMessage
	
	// Get unviewed messages with retry logic
	err := r.retryWithBackoff(ctx, func() error {
		var err error
		messages, err = r.storageRepo.GetUnviewedMessagesForReminders(
			ctx,
			config.CheckAfterHours,
			config.MaxReminders,
		)
		return err
	}, "get_unviewed_messages")
	
	if err != nil {
		return fmt.Errorf("failed to get unviewed messages: %w", err)
	}

	log.Info().Int("count", len(messages)).Msg("Found messages eligible for reminders")

	if len(messages) == 0 {
		log.Info().Msg("No messages found requiring reminders")
		return nil
	}

	// Process each message with individual error recovery
	processedCount := 0
	errorCount := 0
	for _, message := range messages {
		// Create reminder request
		reminderReq := ReminderRequest{
			MessageID:      message.MessageID,
			UniqueID:       message.UniqueID,
			RecipientEmail: message.RecipientEmail,
			DaysOld:        message.DaysOld,
			DecryptionURL:  fmt.Sprintf("https://password.exchange/decrypt/%s", message.UniqueID),
		}

		// Use retry logic for each message processing
		err := r.retryWithBackoff(ctx, func() error {
			return r.ProcessMessageReminder(ctx, reminderReq)
		}, fmt.Sprintf("process_message_%d", message.MessageID))
		
		if err != nil {
			errorCount++
			log.Error().
				Err(err).
				Int("messageID", message.MessageID).
				Str("email", validation.SanitizeEmailForLogging(message.RecipientEmail)).
				Int("daysOld", message.DaysOld).
				Str("operation", "process_reminder").
				Msg("Failed to process reminder for message after all retry attempts")
			continue // Continue processing other messages
		}
		processedCount++
	}

	log.Info().
		Int("totalMessages", len(messages)).
		Int("processedCount", processedCount).
		Int("errorCount", errorCount).
		Msg("Reminder processing completed")

	// Implement graceful degradation: return success if at least some messages were processed
	if processedCount > 0 {
		log.Info().
			Int("processedCount", processedCount).
			Int("errorCount", errorCount).
			Float64("successRate", float64(processedCount)/float64(len(messages))*100).
			Msg("Reminder processing completed with partial success")
		return nil
	}

	// If no messages were processed and we had errors, this indicates a more serious issue
	if errorCount > 0 {
		log.Error().
			Int("errorCount", errorCount).
			Int("totalMessages", len(messages)).
			Msg("Failed to process any reminder messages")
		return fmt.Errorf("failed to process any of %d reminder messages", len(messages))
	}

	return nil
}

// ProcessMessageReminder sends a reminder email for a specific message
func (r *ReminderService) ProcessMessageReminder(ctx context.Context, req ReminderRequest) error {
	// Validate email address before processing
	if err := validation.ValidateEmail(req.RecipientEmail); err != nil {
		return fmt.Errorf("invalid recipient email address for messageID %d: %w", req.MessageID, err)
	}

	log.Info().
		Int("messageID", req.MessageID).
		Str("email", validation.SanitizeEmailForLogging(req.RecipientEmail)).
		Int("daysOld", req.DaysOld).
		Str("operation", "process_message_reminder").
		Msg("Processing reminder for message")

	// Get reminder history to determine reminder count with retry logic
	var history []*ReminderLogEntry
	err := r.retryWithBackoff(ctx, func() error {
		var err error
		history, err = r.storageRepo.GetReminderHistory(ctx, req.MessageID)
		return err
	}, fmt.Sprintf("get_reminder_history_%d", req.MessageID))
	
	if err != nil {
		return fmt.Errorf("failed to get reminder history for messageID %d: %w", req.MessageID, err)
	}

	reminderCount := 0
	if len(history) > 0 {
		reminderCount = history[0].ReminderCount
	}

	// Update reminder number in request
	req.ReminderNumber = reminderCount + 1

	// Send reminder email
	notificationReq := NotificationRequest{
		To:            req.RecipientEmail,
		From:          "server@password.exchange",
		FromName:      "Password Exchange",
		Subject:       fmt.Sprintf("Reminder: You have an unviewed encrypted message (Reminder #%d)", req.ReminderNumber),
		MessageURL:    req.DecryptionURL,
	}

	_, err = r.emailSender.SendNotification(ctx, notificationReq)
	if err != nil {
		return fmt.Errorf("failed to send reminder email for messageID %d: %w", req.MessageID, err)
	}

	// Record that we sent a reminder with retry logic
	err = r.retryWithBackoff(ctx, func() error {
		return r.storageRepo.LogReminderSent(ctx, req.MessageID, req.RecipientEmail)
	}, fmt.Sprintf("log_reminder_sent_%d", req.MessageID))
	
	if err != nil {
		return fmt.Errorf("failed to log reminder sent for messageID %d: %w", req.MessageID, err)
	}

	log.Info().
		Int("messageID", req.MessageID).
		Str("email", validation.SanitizeEmailForLogging(req.RecipientEmail)).
		Int("reminderNumber", req.ReminderNumber).
		Str("operation", "reminder_sent").
		Msg("Reminder email sent successfully")
	return nil
}

// retryWithBackoff executes a function with exponential backoff retry logic
func (r *ReminderService) retryWithBackoff(ctx context.Context, operation func() error, operationName string) error {
	var lastErr error
	
	for attempt := 0; attempt < MaxRetries; attempt++ {
		// Check circuit breaker before attempting operation
		if err := r.circuitBreaker.CanExecute(); err != nil {
			log.Warn().
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
				log.Info().
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

		log.Warn().
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

	log.Error().
		Err(lastErr).
		Str("operation", operationName).
		Int("maxRetries", MaxRetries).
		Msg("Operation failed after all retry attempts")

	return fmt.Errorf("%w: %s failed after %d attempts: %v", ErrMaxRetriesExceeded, operationName, MaxRetries, lastErr)
}

// validateReminderConfig validates all reminder configuration parameters
func (r *ReminderService) validateReminderConfig(cfg ReminderConfig) error {
	if cfg.CheckAfterHours < MinCheckAfterHours || cfg.CheckAfterHours > MaxCheckAfterHours {
		return fmt.Errorf("%w: got %d", ErrInvalidCheckAfterHours, cfg.CheckAfterHours)
	}

	if cfg.MaxReminders < MinMaxReminders || cfg.MaxReminders > MaxMaxReminders {
		return fmt.Errorf("%w: got %d", ErrInvalidMaxReminders, cfg.MaxReminders)
	}

	if cfg.Interval < MinReminderInterval || cfg.Interval > MaxReminderInterval {
		return fmt.Errorf("%w: got %d", ErrInvalidReminderInterval, cfg.Interval)
	}

	return nil
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() error {
	switch cb.state {
	case CircuitBreakerClosed:
		return nil
	case CircuitBreakerOpen:
		if time.Since(cb.lastFailureTime) > CircuitBreakerTimeout {
			cb.state = CircuitBreakerHalfOpen
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
func (cb *CircuitBreaker) RecordSuccess() {
	cb.failureCount = 0
	cb.state = CircuitBreakerClosed
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	if cb.failureCount >= CircuitBreakerThreshold {
		cb.state = CircuitBreakerOpen
		log.Warn().
			Int("failureCount", cb.failureCount).
			Int("threshold", CircuitBreakerThreshold).
			Msg("Circuit breaker opened due to repeated failures")
	}
}