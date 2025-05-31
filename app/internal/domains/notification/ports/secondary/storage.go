package secondary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
)

// StoragePort defines the secondary port for storage operations needed by reminder service
type StoragePort interface {
	// GetUnviewedMessagesForReminders retrieves messages eligible for reminders
	GetUnviewedMessagesForReminders(ctx context.Context, checkAfterHours, maxReminders, reminderIntervalHours int) ([]*domain.UnviewedMessage, error)
	
	// GetReminderHistory retrieves the reminder history for a specific message
	GetReminderHistory(ctx context.Context, messageID int) ([]*domain.ReminderLogEntry, error)
	
	// LogReminderSent records that a reminder was sent for a message
	LogReminderSent(ctx context.Context, messageID int, recipientEmail string) error
}