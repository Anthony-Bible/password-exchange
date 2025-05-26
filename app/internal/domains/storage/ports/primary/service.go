package primary

import (
	"context"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
)

// StorageServicePort defines the primary interface for storage operations
// This will be implemented by the storage service and used by external adapters
type StorageServicePort interface {
	// StoreMessage stores a new encrypted message
	StoreMessage(ctx context.Context, message *domain.Message) error

	// RetrieveMessage retrieves a message by its unique ID
	RetrieveMessage(ctx context.Context, uniqueID string) (*domain.Message, error)

	// GetMessage retrieves a message by its unique ID without incrementing view count
	GetMessage(ctx context.Context, uniqueID string) (*domain.Message, error)

	// CleanupExpiredMessages removes expired messages from storage
	CleanupExpiredMessages(ctx context.Context) error

	// GetUnviewedMessagesForReminders retrieves messages eligible for reminder emails
	GetUnviewedMessagesForReminders(ctx context.Context, olderThanHours, maxReminders int) ([]*domain.UnviewedMessage, error)

	// LogReminderSent records that a reminder email was sent for a message
	LogReminderSent(ctx context.Context, messageID int, emailAddress string) error

	// GetReminderHistory retrieves the reminder history for a specific message
	GetReminderHistory(ctx context.Context, messageID int) ([]*domain.ReminderLogEntry, error)

	// HealthCheck verifies the storage service is healthy
	HealthCheck(ctx context.Context) error
}
