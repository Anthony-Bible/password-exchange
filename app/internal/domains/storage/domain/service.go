package domain

import (
	"context"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/secondary"
)

// StorageService implements the primary port and provides business logic for storage operations.
// It depends only on secondary ports so the domain can remain free of infrastructure concerns.
type StorageService struct {
	repository secondary.MessageRepository
	logger     secondary.LoggerPort
	validation secondary.ValidationPort
}

// NewStorageService creates a new storage service with the given dependencies.
// All three secondary ports are required and wired with concrete adapters at
// the composition root.
func NewStorageService(
	repository secondary.MessageRepository,
	logger secondary.LoggerPort,
	validation secondary.ValidationPort,
) *StorageService {
	return &StorageService{
		repository: repository,
		logger:     logger,
		validation: validation,
	}
}

// StoreMessage stores a new encrypted message with validation
func (s *StorageService) StoreMessage(ctx context.Context, message *contracts.Message) error {
	// Business rule validation
	if message == nil {
		s.logger.Warn().Msg("Attempted to store nil message")
		return ErrNilMessage
	}
	if message.Content == "" {
		s.logger.Warn().Msg("Attempted to store message with empty content")
		return ErrEmptyContent
	}
	if message.UniqueID == "" {
		s.logger.Warn().Msg("Attempted to store message with empty unique ID")
		return ErrEmptyUniqueID
	}
	// Storage only guards the structurally invalid case (< 1). The upper bound
	// is a product policy owned by the message domain (message.AbsoluteMaxViewCount),
	// so callers are responsible for enforcing it; duplicating the cap here would
	// silently reject otherwise-valid counts if that policy ever changes.
	if message.MaxViewCount < 1 {
		s.logger.Warn().Int("maxViewCount", message.MaxViewCount).Msg("Attempted to store message with invalid max view count")
		return ErrInvalidMaxViewCount
	}

	// Delegate to repository
	return s.repository.InsertMessage(message)
}

// RetrieveMessage retrieves a message by its unique ID with validation and increments view count
func (s *StorageService) RetrieveMessage(ctx context.Context, uniqueID string) (*contracts.Message, error) {
	// Business rule validation
	if uniqueID == "" {
		s.logger.Warn().Msg("Attempted to retrieve message with empty unique ID")
		return nil, ErrEmptyUniqueID
	}

	// Increment view count and get message atomically
	message, err := s.repository.IncrementViewCountAndGet(uniqueID)
	if err != nil {
		s.logger.Warn().Err(err).Str("uniqueID", uniqueID).Msg("Failed to increment view count and retrieve message")
		return nil, err
	}

	s.logger.Info().Str("uniqueID", uniqueID).Int("viewCount", message.ViewCount).Msg("Message retrieved and view count incremented")
	return message, nil
}

// GetMessage retrieves a message by its unique ID without incrementing view count
func (s *StorageService) GetMessage(ctx context.Context, uniqueID string) (*contracts.Message, error) {
	// Business rule validation
	if uniqueID == "" {
		s.logger.Warn().Msg("Attempted to get message with empty unique ID")
		return nil, ErrEmptyUniqueID
	}

	// Delegate to repository
	message, err := s.repository.GetMessage(uniqueID)
	if err != nil {
		s.logger.Warn().Err(err).Str("uniqueID", uniqueID).Msg("Failed to retrieve message")
		return nil, err
	}

	s.logger.Info().Str("uniqueID", uniqueID).Msg("Message retrieved successfully")
	return message, nil
}

// CleanupExpiredMessages removes expired messages from storage
func (s *StorageService) CleanupExpiredMessages(ctx context.Context) error {
	s.logger.Info().Msg("Starting cleanup of expired messages")
	return s.repository.DeleteExpiredMessages()
}

// GetUnviewedMessagesForReminders retrieves messages eligible for reminder emails
func (s *StorageService) GetUnviewedMessagesForReminders(ctx context.Context, olderThanHours, maxReminders, reminderIntervalHours int) ([]*contracts.UnviewedMessage, error) {
	// Business rule validation
	if olderThanHours < 1 {
		s.logger.Warn().Int("olderThanHours", olderThanHours).Msg("Invalid olderThanHours parameter")
		return nil, ErrInvalidParameter
	}
	if maxReminders < 1 {
		s.logger.Warn().Int("maxReminders", maxReminders).Msg("Invalid maxReminders parameter")
		return nil, ErrInvalidParameter
	}
	if reminderIntervalHours < 1 {
		s.logger.Warn().Int("reminderIntervalHours", reminderIntervalHours).Msg("Invalid reminderIntervalHours parameter")
		return nil, ErrInvalidParameter
	}

	// Delegate to repository
	messages, err := s.repository.GetUnviewedMessagesForReminders(olderThanHours, maxReminders, reminderIntervalHours)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to retrieve unviewed messages for reminders")
		return nil, err
	}

	s.logger.Info().Int("count", len(messages)).Int("olderThanHours", olderThanHours).Int("maxReminders", maxReminders).Int("reminderIntervalHours", reminderIntervalHours).Msg("Retrieved unviewed messages for reminders")
	return messages, nil
}

// LogReminderSent records that a reminder email was sent for a message
func (s *StorageService) LogReminderSent(ctx context.Context, messageID int, emailAddress string) error {
	// Business rule validation
	if messageID < 1 {
		s.logger.Warn().Int("messageID", messageID).Msg("Invalid messageID parameter")
		return ErrInvalidParameter
	}
	if emailAddress == "" {
		s.logger.Warn().Msg("Attempted to log reminder with empty email address")
		return ErrEmptyEmailAddress
	}

	// Delegate to repository
	err := s.repository.LogReminderSent(messageID, emailAddress)
	if err != nil {
		s.logger.Error().Err(err).Int("messageID", messageID).Str("emailAddress", s.validation.SanitizeEmailForLogging(emailAddress)).Msg("Failed to log reminder sent")
		return err
	}

	s.logger.Info().Int("messageID", messageID).Str("emailAddress", s.validation.SanitizeEmailForLogging(emailAddress)).Msg("Reminder sent logged successfully")
	return nil
}

// GetReminderHistory retrieves the reminder history for a specific message
func (s *StorageService) GetReminderHistory(ctx context.Context, messageID int) ([]*contracts.ReminderLogEntry, error) {
	// Business rule validation
	if messageID < 1 {
		s.logger.Warn().Int("messageID", messageID).Msg("Invalid messageID parameter")
		return nil, ErrInvalidParameter
	}

	// Delegate to repository
	history, err := s.repository.GetReminderHistory(messageID)
	if err != nil {
		s.logger.Error().Err(err).Int("messageID", messageID).Msg("Failed to retrieve reminder history")
		return nil, err
	}

	s.logger.Info().Int("messageID", messageID).Int("count", len(history)).Msg("Retrieved reminder history")
	return history, nil
}

// HealthCheck verifies the storage service is healthy by probing the
// underlying repository. The error is returned to the caller (the gRPC
// health probe loop) so it can flip the standard health-service status.
func (s *StorageService) HealthCheck(ctx context.Context) error {
	if err := s.repository.Ping(ctx); err != nil {
		s.logger.Warn().Err(err).Msg("Storage health check failed")
		return err
	}
	s.logger.Debug().Msg("Storage health check succeeded")
	return nil
}

// CreateUploadSession persists a new file upload session with validation.
func (s *StorageService) CreateUploadSession(ctx context.Context, session contracts.UploadSession) error {
	if session.SessionID == "" {
		s.logger.Warn().Msg("Attempted to create upload session with empty session ID")
		return ErrInvalidParameter
	}
	return s.repository.CreateUploadSession(ctx, session)
}

// GetUploadSession retrieves an upload session by session_id or file_id.
func (s *StorageService) GetUploadSession(ctx context.Context, id string) (*contracts.UploadSession, error) {
	if id == "" {
		s.logger.Warn().Msg("Attempted to get upload session with empty ID")
		return nil, ErrInvalidParameter
	}
	return s.repository.GetUploadSession(ctx, id)
}

// AddCompletedPart records a successfully uploaded chunk for the session.
func (s *StorageService) AddCompletedPart(ctx context.Context, sessionID string, part contracts.UploadSessionPart) error {
	if sessionID == "" {
		s.logger.Warn().Msg("Attempted to add completed part with empty session ID")
		return ErrInvalidParameter
	}
	return s.repository.AddCompletedPart(ctx, sessionID, part)
}

// CompleteUploadSession marks the session assembled and clears its encryption
// key from persistent state.
func (s *StorageService) CompleteUploadSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		s.logger.Warn().Msg("Attempted to complete upload session with empty session ID")
		return ErrInvalidParameter
	}
	return s.repository.CompleteUploadSession(ctx, sessionID)
}

// DeleteUploadSession removes a session row; non-existent sessions are a no-op.
func (s *StorageService) DeleteUploadSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		s.logger.Warn().Msg("Attempted to delete upload session with empty session ID")
		return ErrInvalidParameter
	}
	return s.repository.DeleteUploadSession(ctx, sessionID)
}

// DeleteExpiredUploadSessions sweeps incomplete sessions expiring before asOf
// and returns them for caller-side object-storage cleanup.
func (s *StorageService) DeleteExpiredUploadSessions(ctx context.Context, asOf time.Time) ([]contracts.UploadSession, error) {
	return s.repository.DeleteExpiredUploadSessions(ctx, asOf)
}
