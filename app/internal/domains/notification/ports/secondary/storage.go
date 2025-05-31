package secondary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
)

// StoragePort defines the secondary port for storage operations needed by reminder service
type StoragePort interface {
	// GetUnviewedMessagesForReminders retrieves messages eligible for reminders
	GetUnviewedMessagesForReminders(ctx context.Context, checkAfterHours, maxReminders, reminderIntervalHours int) ([]*contracts.UnviewedMessage, error)
	
	// GetReminderHistory retrieves the reminder history for a specific message
	GetReminderHistory(ctx context.Context, messageID int) ([]*contracts.ReminderLogEntry, error)
	
	// LogReminderSent records that a reminder was sent for a message
	LogReminderSent(ctx context.Context, messageID int, recipientEmail string) error
}