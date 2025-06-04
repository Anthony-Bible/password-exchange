package domain

import (
	"context"
	"fmt"
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
)

// NotificationService provides notification operations
type NotificationService struct {
	emailSender      secondary.EmailPort
	queueConsumer    secondary.QueuePort
	templateRenderer secondary.TemplatePort
	reminderService  *ReminderService
	logger           secondary.LoggerPort
	validation       secondary.ValidationPort
	config           secondary.ConfigPort
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	emailSender secondary.EmailPort,
	queueConsumer secondary.QueuePort,
	templateRenderer secondary.TemplatePort,
	storageRepo secondary.StoragePort,
	notificationPublisher secondary.NotificationPort,
	logger secondary.LoggerPort,
	validation secondary.ValidationPort,
	config secondary.ConfigPort,
) *NotificationService {
	reminderService := NewReminderService(storageRepo, notificationPublisher, logger, config, validation)
	return &NotificationService{
		emailSender:      emailSender,
		queueConsumer:    queueConsumer,
		templateRenderer: templateRenderer,
		reminderService:  reminderService,
		logger:           logger,
		validation:       validation,
		config:           config,
	}
}

// NewNotificationServiceWithReminder creates a new notification service with an existing reminder service
func NewNotificationServiceWithReminder(
	emailSender secondary.EmailPort,
	queueConsumer secondary.QueuePort,
	templateRenderer secondary.TemplatePort,
	reminderService *ReminderService,
	logger secondary.LoggerPort,
	validation secondary.ValidationPort,
	config secondary.ConfigPort,
) *NotificationService {
	return &NotificationService{
		emailSender:      emailSender,
		queueConsumer:    queueConsumer,
		templateRenderer: templateRenderer,
		reminderService:  reminderService,
		logger:           logger,
		validation:       validation,
		config:           config,
	}
}

// SendNotification sends a notification using the configured sender
func (s *NotificationService) SendNotification(ctx context.Context, req NotificationRequest) (*NotificationResponse, error) {
	if err := s.validateNotificationRequest(req); err != nil {
		s.logger.Error().Err(err).Msg("Invalid notification request")
		return nil, fmt.Errorf("%w: %v", ErrInvalidNotificationRequest, err)
	}

	response, err := s.emailSender.SendNotification(ctx, req)
	if err != nil {
		s.logger.Error().Err(err).Str("to", s.validation.SanitizeEmailForLogging(req.To)).Msg("Failed to send notification")
		return nil, fmt.Errorf("%w: %v", ErrEmailSendFailed, err)
	}

	s.logger.Info().Str("to", s.validation.SanitizeEmailForLogging(req.To)).Str("messageId", response.MessageID).Msg("Notification sent successfully")
	return response, nil
}

// StartMessageProcessing starts consuming messages from the queue
func (s *NotificationService) StartMessageProcessing(ctx context.Context, queueConn QueueConnection, concurrency int) error {
	s.logger.Info().Str("queue", queueConn.QueueName).Int("concurrency", concurrency).Msg("Starting message processing")

	err := s.queueConsumer.StartConsuming(ctx, queueConn, s, concurrency)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to start consuming messages")
		return fmt.Errorf("%w: %v", ErrQueueConsumeFailed, err)
	}

	return nil
}

// HandleMessage implements the MessageHandler interface
func (s *NotificationService) HandleMessage(ctx context.Context, msg QueueMessage) error {
	s.logger.Debug().Str("to", s.validation.SanitizeEmailForLogging(msg.OtherEmail)).Str("from", msg.FirstName).Msg("Processing notification message")

	// Create notification request from queue message
	notificationReq := s.createNotificationRequest(msg)

	// Send the notification
	_, err := s.SendNotification(ctx, notificationReq)
	if err != nil {
		s.logger.Error().Err(err).Str("to", s.validation.SanitizeEmailForLogging(msg.OtherEmail)).Msg("Failed to send notification for queue message")
		return err
	}

	s.logger.Debug().Str("to", s.validation.SanitizeEmailForLogging(msg.OtherEmail)).Msg("Successfully processed notification message")
	return nil
}

// createNotificationRequest converts a queue message to a notification request
func (s *NotificationService) createNotificationRequest(msg QueueMessage) NotificationRequest {
	subject := fmt.Sprintf(s.config.GetInitialNotificationSubject(), msg.FirstName)

	return NotificationRequest{
		To:             msg.OtherEmail,
		From:           s.config.GetServerEmail(),
		FromName:       s.config.GetServerName(),
		Subject:        subject,
		MessageContent: msg.Content,
		SenderName:     msg.FirstName,
		RecipientName:  msg.OtherFirstName,
		MessageURL:     msg.URL,
		Hidden:         msg.Hidden,
	}
}

// validateNotificationRequest validates the notification request
func (s *NotificationService) validateNotificationRequest(req NotificationRequest) error {
	if strings.TrimSpace(req.To) == "" {
		return fmt.Errorf("recipient email is required")
	}

	if strings.TrimSpace(req.From) == "" {
		return fmt.Errorf("sender email is required")
	}

	if strings.TrimSpace(req.Subject) == "" {
		return fmt.Errorf("subject is required")
	}

	// Validate email addresses using validation port
	if err := s.validation.ValidateEmail(req.To); err != nil {
		return fmt.Errorf("invalid recipient email: %w", err)
	}
	if err := s.validation.ValidateEmail(req.From); err != nil {
		return fmt.Errorf("invalid sender email: %w", err)
	}

	return nil
}

// ProcessReminders finds and processes messages eligible for reminder emails
func (s *NotificationService) ProcessReminders(ctx context.Context, config ReminderConfig) error {
	return s.reminderService.ProcessReminders(ctx, config)
}

// ProcessMessageReminder sends a reminder email for a specific message
func (s *NotificationService) ProcessMessageReminder(ctx context.Context, req ReminderRequest) error {
	return s.reminderService.ProcessMessageReminder(ctx, req)
}

// Close closes the queue consumer connection
func (s *NotificationService) Close() error {
	return s.queueConsumer.Close()
}