package primary

import (
	"context"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
)

// StorageServicePort defines the primary interface for storage operations
// This will be implemented by the storage service and used by external adapters
type StorageServicePort interface {
	// StoreMessage stores a new encrypted message
	StoreMessage(ctx context.Context, message *contracts.Message) error

	// RetrieveMessage retrieves a message by its unique ID
	RetrieveMessage(ctx context.Context, uniqueID string) (*contracts.Message, error)

	// GetMessage retrieves a message by its unique ID without incrementing view count
	GetMessage(ctx context.Context, uniqueID string) (*contracts.Message, error)

	// CleanupExpiredMessages removes expired messages from storage
	CleanupExpiredMessages(ctx context.Context) error

	// GetUnviewedMessagesForReminders retrieves messages eligible for reminder emails
	GetUnviewedMessagesForReminders(ctx context.Context, olderThanHours, maxReminders, reminderIntervalHours int) ([]*contracts.UnviewedMessage, error)

	// LogReminderSent records that a reminder email was sent for a message
	LogReminderSent(ctx context.Context, messageID int, emailAddress string) error

	// GetReminderHistory retrieves the reminder history for a specific message
	GetReminderHistory(ctx context.Context, messageID int) ([]*contracts.ReminderLogEntry, error)

	// HealthCheck verifies the storage service is healthy
	HealthCheck(ctx context.Context) error

	// CreateUploadSession persists a new file upload session.
	CreateUploadSession(ctx context.Context, session contracts.UploadSession) error

	// GetUploadSession retrieves an upload session by session_id or file_id.
	GetUploadSession(ctx context.Context, id string) (*contracts.UploadSession, error)

	// AddCompletedPart records a successfully uploaded chunk for the session.
	AddCompletedPart(ctx context.Context, sessionID string, part contracts.UploadSessionPart) error

	// CompleteUploadSession marks the session as assembled and clears its
	// encryption key from persistent state.
	CompleteUploadSession(ctx context.Context, sessionID string) error

	// DeleteUploadSession removes a session row (best-effort, no-op if missing).
	DeleteUploadSession(ctx context.Context, sessionID string) error

	// DeleteExpiredUploadSessions sweeps incomplete sessions expiring before asOf
	// and returns the removed sessions for caller-side object-storage cleanup.
	DeleteExpiredUploadSessions(ctx context.Context, asOf time.Time) ([]contracts.UploadSession, error)
}
