package consumer

import (
	"context"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/primary"
	"github.com/rs/zerolog/log"
)

// NotificationConsumer implements the notification consumer service
type NotificationConsumer struct {
	notificationService primary.NotificationServicePort
	queueConnection     domain.QueueConnection
	concurrency         int
}

// NewNotificationConsumer creates a new notification consumer
func NewNotificationConsumer(
	notificationService primary.NotificationServicePort,
	queueConnection domain.QueueConnection,
	concurrency int,
) *NotificationConsumer {
	return &NotificationConsumer{
		notificationService: notificationService,
		queueConnection:     queueConnection,
		concurrency:         concurrency,
	}
}

// Start starts the notification consumer
func (c *NotificationConsumer) Start(ctx context.Context) error {
	log.Info().
		Str("queue", c.queueConnection.QueueName).
		Str("host", c.queueConnection.Host).
		Int("concurrency", c.concurrency).
		Msg("Starting notification consumer")

	return c.notificationService.StartMessageProcessing(ctx, c.queueConnection, c.concurrency)
}

// Stop stops the notification consumer
func (c *NotificationConsumer) Stop() error {
	log.Info().Msg("Stopping notification consumer")
	return c.notificationService.Close()
}