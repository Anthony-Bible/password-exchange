package storage

import (
	"context"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	storagePorts "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/primary"
)

// GRPCStorageAdapter adapts the storage service gRPC client for reminder operations.
// This adapter implements the hexagonal architecture pattern by translating between
// notification domain entities and storage domain entities via the gRPC interface.
type GRPCStorageAdapter struct {
	storageService storagePorts.StorageServicePort
}

// NewGRPCStorageAdapter creates a new gRPC storage adapter
func NewGRPCStorageAdapter(storageService storagePorts.StorageServicePort) *GRPCStorageAdapter {
	return &GRPCStorageAdapter{
		storageService: storageService,
	}
}

// GetUnviewedMessagesForReminders retrieves messages eligible for reminders
func (a *GRPCStorageAdapter) GetUnviewedMessagesForReminders(ctx context.Context, checkAfterHours, maxReminders, reminderIntervalHours int) ([]*domain.UnviewedMessage, error) {
	storageMessages, err := a.storageService.GetUnviewedMessagesForReminders(ctx, checkAfterHours, maxReminders, reminderIntervalHours)
	if err != nil {
		return nil, err
	}

	// Convert storage entities to notification domain entities
	// This translation maintains domain separation and prevents tight coupling
	messages := make([]*domain.UnviewedMessage, len(storageMessages))
	for i, sm := range storageMessages {
		messages[i] = &domain.UnviewedMessage{
			MessageID:      sm.MessageID,
			UniqueID:       sm.UniqueID,
			RecipientEmail: sm.RecipientEmail,
			DaysOld:        sm.DaysOld,
			Created:        sm.Created,
		}
	}

	return messages, nil
}

// GetReminderHistory retrieves the reminder history for a specific message
func (a *GRPCStorageAdapter) GetReminderHistory(ctx context.Context, messageID int) ([]*domain.ReminderLogEntry, error) {
	storageHistory, err := a.storageService.GetReminderHistory(ctx, messageID)
	if err != nil {
		return nil, err
	}

	// Convert storage entities to notification domain entities
	history := make([]*domain.ReminderLogEntry, len(storageHistory))
	for i, sh := range storageHistory {
		history[i] = &domain.ReminderLogEntry{
			MessageID:      sh.MessageID,
			RecipientEmail: sh.EmailAddress,
			ReminderCount:  sh.ReminderCount,
			SentAt:         sh.LastReminderSent,
		}
	}

	return history, nil
}

// LogReminderSent records that a reminder was sent for a message
func (a *GRPCStorageAdapter) LogReminderSent(ctx context.Context, messageID int, recipientEmail string) error {
	return a.storageService.LogReminderSent(ctx, messageID, recipientEmail)
}