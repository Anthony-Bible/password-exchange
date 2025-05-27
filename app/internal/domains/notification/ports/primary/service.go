package primary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
)

// NotificationServicePort defines the primary port for notification operations
type NotificationServicePort interface {
	// SendNotification sends a notification using the configured sender
	SendNotification(ctx context.Context, req domain.NotificationRequest) (*domain.NotificationResponse, error)
	
	// StartMessageProcessing starts consuming messages from the queue and processing them
	StartMessageProcessing(ctx context.Context, queueConn domain.QueueConnection, concurrency int) error
	
	// Close closes any open connections
	Close() error
}

// ReminderServicePort defines the primary port for reminder operations
type ReminderServicePort interface {
	// ProcessReminders finds and processes messages eligible for reminder emails
	ProcessReminders(ctx context.Context, config domain.ReminderConfig) error
	
	// ProcessMessageReminder sends a reminder email for a specific message
	ProcessMessageReminder(ctx context.Context, req domain.ReminderRequest) error
}