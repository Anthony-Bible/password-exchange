package domain

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

// NotificationService provides notification operations
type NotificationService struct {
	emailSender      NotificationSender
	queueConsumer    QueueConsumer
	templateRenderer TemplateRenderer
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	emailSender NotificationSender,
	queueConsumer QueueConsumer,
	templateRenderer TemplateRenderer,
) *NotificationService {
	return &NotificationService{
		emailSender:      emailSender,
		queueConsumer:    queueConsumer,
		templateRenderer: templateRenderer,
	}
}

// SendNotification sends a notification using the configured sender
func (s *NotificationService) SendNotification(ctx context.Context, req NotificationRequest) (*NotificationResponse, error) {
	if err := s.validateNotificationRequest(req); err != nil {
		log.Error().Err(err).Msg("Invalid notification request")
		return nil, fmt.Errorf("%w: %v", ErrInvalidNotificationRequest, err)
	}

	response, err := s.emailSender.SendNotification(ctx, req)
	if err != nil {
		log.Error().Err(err).Str("to", req.To).Msg("Failed to send notification")
		return nil, fmt.Errorf("%w: %v", ErrEmailSendFailed, err)
	}

	log.Info().Str("to", req.To).Str("messageId", response.MessageID).Msg("Notification sent successfully")
	return response, nil
}

// StartMessageProcessing starts consuming messages from the queue
func (s *NotificationService) StartMessageProcessing(ctx context.Context, queueConn QueueConnection, concurrency int) error {
	log.Info().Str("queue", queueConn.QueueName).Int("concurrency", concurrency).Msg("Starting message processing")

	err := s.queueConsumer.StartConsuming(ctx, queueConn, s, concurrency)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start consuming messages")
		return fmt.Errorf("%w: %v", ErrQueueConsumeFailed, err)
	}

	return nil
}

// HandleMessage implements the MessageHandler interface
func (s *NotificationService) HandleMessage(ctx context.Context, msg QueueMessage) error {
	log.Debug().Str("to", msg.OtherEmail).Str("from", msg.FirstName).Msg("Processing notification message")

	// Create notification request from queue message
	notificationReq := s.createNotificationRequest(msg)

	// Send the notification
	_, err := s.SendNotification(ctx, notificationReq)
	if err != nil {
		log.Error().Err(err).Str("to", msg.OtherEmail).Msg("Failed to send notification for queue message")
		return err
	}

	log.Debug().Str("to", msg.OtherEmail).Msg("Successfully processed notification message")
	return nil
}

// createNotificationRequest converts a queue message to a notification request
func (s *NotificationService) createNotificationRequest(msg QueueMessage) NotificationRequest {
	subject := fmt.Sprintf("Encrypted Message from Password Exchange from %s", msg.FirstName)

	return NotificationRequest{
		To:             msg.OtherEmail,
		From:           "server@password.exchange",
		FromName:       "Password Exchange",
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

	// Basic email validation
	if !strings.Contains(req.To, "@") || !strings.Contains(req.From, "@") {
		return ErrInvalidEmailAddress
	}

	return nil
}

// Close closes the queue consumer connection
func (s *NotificationService) Close() error {
	return s.queueConsumer.Close()
}