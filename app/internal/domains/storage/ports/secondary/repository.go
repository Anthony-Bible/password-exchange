// Package secondary defines the outbound (secondary) ports for the storage domain.
// These interfaces describe the dependencies the domain requires from external systems,
// such as databases, loggers, and validators. Adapters in
// adapters/secondary/ provide concrete implementations of these ports.
package secondary

import (
	"context"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
)

// MessageRepository defines the secondary port for persistent message storage operations.
// Implementations of this interface are responsible for storing, retrieving, and
// managing the lifecycle of encrypted messages and the reminder bookkeeping that
// surrounds them. The storage domain depends on this interface so it can remain
// agnostic of the underlying database technology.
type MessageRepository interface {
	// InsertMessage persists a new encrypted message and its metadata.
	// The supplied message MUST already contain a unique identifier and the
	// content payload to store. Implementations may set defaults for fields
	// such as expires_at when they are not provided by the caller.
	//
	// Parameters:
	//   - message: The message to persist; must be non-nil
	//
	// Returns:
	//   - nil on successful insert
	//   - An error wrapping domain.ErrDatabaseOperation or domain.ErrDatabaseConnection
	//     if the underlying database call fails
	InsertMessage(message *contracts.Message) error

	// SelectMessageByUniqueID retrieves a single message by its unique identifier
	// without modifying any of its metadata (for example, view counts).
	//
	// Parameters:
	//   - uniqueID: The unique identifier of the message to retrieve
	//
	// Returns:
	//   - A pointer to the retrieved message on success
	//   - domain.ErrMessageNotFound if no row matches the supplied uniqueID
	//   - An error wrapping domain.ErrDatabaseOperation for any other failure
	SelectMessageByUniqueID(uniqueID string) (*contracts.Message, error)

	// IncrementViewCountAndGet atomically increments the view count for a message
	// and returns the updated record. If the view count meets or exceeds the
	// message's MaxViewCount after the increment, implementations should delete
	// the row in the same transaction so the message can no longer be retrieved.
	//
	// Parameters:
	//   - uniqueID: The unique identifier of the message to read and update
	//
	// Returns:
	//   - The updated message with the new view count
	//   - domain.ErrMessageNotFound if no row matches the supplied uniqueID
	//   - An error wrapping domain.ErrDatabaseOperation for any other failure
	IncrementViewCountAndGet(uniqueID string) (*contracts.Message, error)

	// DeleteExpiredMessages removes every message whose expires_at timestamp has
	// already elapsed. The method is intended to be invoked periodically by a
	// background cleanup job.
	//
	// Returns:
	//   - nil on success (even if no rows were deleted)
	//   - An error wrapping domain.ErrDatabaseOperation if the delete fails
	DeleteExpiredMessages() error

	// GetMessage retrieves a message by its unique identifier without incrementing
	// the view count. This is useful for read-only inspection of a message, for
	// example by reminder workers.
	//
	// Parameters:
	//   - uniqueID: The unique identifier of the message to retrieve
	//
	// Returns:
	//   - A pointer to the retrieved message on success
	//   - domain.ErrMessageNotFound if no row matches the supplied uniqueID
	//   - An error wrapping domain.ErrDatabaseOperation for any other failure
	GetMessage(uniqueID string) (*contracts.Message, error)

	// GetUnviewedMessagesForReminders returns messages that are still unviewed and
	// are eligible to receive a reminder email. Implementations should honor all
	// three policy parameters when computing the candidate set.
	//
	// Parameters:
	//   - olderThanHours: Minimum age of a message (in hours) before it is eligible
	//   - maxReminders: Maximum number of reminders that may have been sent already
	//   - reminderIntervalHours: Minimum gap (in hours) between successive reminders
	//
	// Returns:
	//   - A slice of UnviewedMessage records eligible for a reminder (may be empty)
	//   - An error wrapping domain.ErrDatabaseOperation if the query fails
	GetUnviewedMessagesForReminders(olderThanHours, maxReminders, reminderIntervalHours int) ([]*contracts.UnviewedMessage, error)

	// LogReminderSent records that a reminder email has been delivered for a given
	// message and recipient. Implementations should treat this call as idempotent
	// for the (messageID, emailAddress) pair, incrementing an existing row instead
	// of failing on duplicate keys.
	//
	// Parameters:
	//   - messageID: The numeric identifier of the message that was reminded about
	//   - emailAddress: The recipient address that received the reminder
	//
	// Returns:
	//   - nil on success
	//   - An error wrapping domain.ErrDatabaseOperation if the write fails
	LogReminderSent(messageID int, emailAddress string) error

	// GetReminderHistory returns the full reminder history for a single message.
	// The result reflects every recipient that has received one or more reminders
	// for the supplied messageID.
	//
	// Parameters:
	//   - messageID: The numeric identifier of the message whose history to load
	//
	// Returns:
	//   - A slice of ReminderLogEntry records for the message (may be empty)
	//   - An error wrapping domain.ErrDatabaseOperation if the query fails
	GetReminderHistory(messageID int) ([]*contracts.ReminderLogEntry, error)

	// Close releases any resources held by the repository (database handles,
	// connection pools, etc.). After Close returns, the repository instance
	// MUST NOT be used for any further operations.
	//
	// Returns:
	//   - nil if the resources were released successfully
	//   - An error if the underlying close call fails
	Close() error

	// Ping verifies that the underlying datastore is reachable. Implementations
	// SHOULD honour ctx for cancellation and timeout so callers (notably the
	// gRPC health probe loop) cannot wedge waiting on a hung backend.
	//
	// Returns:
	//   - nil if the datastore responded successfully
	//   - domain.ErrRepositoryClosed if Close has already been called
	//   - An error wrapping domain.ErrDatabaseConnection if the ping fails
	Ping(ctx context.Context) error

	// CreateUploadSession persists a new file upload session row. The session
	// MUST carry a unique SessionID; duplicate session_ids return
	// domain.ErrUploadSessionAlreadyExists.
	CreateUploadSession(ctx context.Context, session contracts.UploadSession) error

	// GetUploadSession retrieves an upload session by either its SessionID or
	// FileID. The id argument is matched against both columns so callers on the
	// chunk-upload path (session_id) and the download path (file_id) can share
	// the same call.
	//
	// Returns domain.ErrUploadSessionNotFound when no row matches.
	GetUploadSession(ctx context.Context, id string) (*contracts.UploadSession, error)

	// AddCompletedPart appends (or replaces, if the PartNumber already exists) a
	// successfully uploaded chunk to the session's completed_parts list.
	AddCompletedPart(ctx context.Context, sessionID string, part contracts.UploadSessionPart) error

	// CompleteUploadSession atomically marks the session status as complete and
	// sets encryption_key = NULL so the key is no longer at rest in the database.
	//
	// Returns domain.ErrUploadSessionNotFound when sessionID does not match any row.
	CompleteUploadSession(ctx context.Context, sessionID string) error

	// DeleteUploadSession removes a single session row. Deletion of a
	// non-existent session is treated as a no-op (best-effort contract).
	DeleteUploadSession(ctx context.Context, sessionID string) error

	// DeleteExpiredUploadSessions selects all incomplete sessions whose
	// expires_at is before asOf, returns them for caller-side object-storage
	// cleanup, then deletes them in a single statement. Completed sessions are
	// kept so their associated files remain accessible.
	DeleteExpiredUploadSessions(ctx context.Context, asOf time.Time) ([]contracts.UploadSession, error)
}
