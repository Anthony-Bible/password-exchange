package domain

import (
	"context"

	"github.com/rs/zerolog/log"
)

// StorageService implements the primary port and provides business logic for storage operations
type StorageService struct {
	repository MessageRepository
}

// NewStorageService creates a new storage service with the given repository
func NewStorageService(repository MessageRepository) *StorageService {
	return &StorageService{
		repository: repository,
	}
}

// StoreMessage stores a new encrypted message with validation
func (s *StorageService) StoreMessage(ctx context.Context, message *Message) error {
	// Business rule validation
	if message.Content == "" {
		log.Warn().Msg("Attempted to store message with empty content")
		return ErrEmptyContent
	}
	if message.UniqueID == "" {
		log.Warn().Msg("Attempted to store message with empty unique ID")
		return ErrEmptyUniqueID
	}
	if message.MaxViewCount < 1 {
		log.Warn().Int("maxViewCount", message.MaxViewCount).Msg("Attempted to store message with invalid max view count")
		return ErrInvalidMaxViewCount
	}
	if message.RecipientEmail == "" {
		log.Warn().Msg("Attempted to store message with empty recipient email")
		return ErrEmptyRecipientEmail
	}

	// Delegate to repository
	return s.repository.InsertMessage(message)
}

// RetrieveMessage retrieves a message by its unique ID with validation and increments view count
func (s *StorageService) RetrieveMessage(ctx context.Context, uniqueID string) (*Message, error) {
	// Business rule validation
	if uniqueID == "" {
		log.Warn().Msg("Attempted to retrieve message with empty unique ID")
		return nil, ErrEmptyUniqueID
	}

	// Increment view count and get message atomically
	message, err := s.repository.IncrementViewCountAndGet(uniqueID)
	if err != nil {
		log.Warn().Err(err).Str("uniqueID", uniqueID).Msg("Failed to increment view count and retrieve message")
		return nil, err
	}

	log.Info().Str("uniqueID", uniqueID).Int("viewCount", message.ViewCount).Msg("Message retrieved and view count incremented")
	return message, nil
}

// GetMessage retrieves a message by its unique ID without incrementing view count
func (s *StorageService) GetMessage(ctx context.Context, uniqueID string) (*Message, error) {
	// Business rule validation
	if uniqueID == "" {
		log.Warn().Msg("Attempted to get message with empty unique ID")
		return nil, ErrEmptyUniqueID
	}

	// Delegate to repository
	message, err := s.repository.GetMessage(uniqueID)
	if err != nil {
		log.Warn().Err(err).Str("uniqueID", uniqueID).Msg("Failed to retrieve message")
		return nil, err
	}

	log.Info().Str("uniqueID", uniqueID).Msg("Message retrieved successfully")
	return message, nil
}

// CleanupExpiredMessages removes expired messages from storage
func (s *StorageService) CleanupExpiredMessages(ctx context.Context) error {
	log.Info().Msg("Starting cleanup of expired messages")
	return s.repository.DeleteExpiredMessages()
}

// GetUnviewedMessagesForReminders retrieves messages eligible for reminder emails
func (s *StorageService) GetUnviewedMessagesForReminders(ctx context.Context, olderThanHours, maxReminders int) ([]*UnviewedMessage, error) {
	// Business rule validation
	if olderThanHours < 1 {
		log.Warn().Int("olderThanHours", olderThanHours).Msg("Invalid olderThanHours parameter")
		return nil, ErrInvalidParameter
	}
	if maxReminders < 1 {
		log.Warn().Int("maxReminders", maxReminders).Msg("Invalid maxReminders parameter")
		return nil, ErrInvalidParameter
	}

	// Delegate to repository
	messages, err := s.repository.GetUnviewedMessagesForReminders(olderThanHours, maxReminders)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve unviewed messages for reminders")
		return nil, err
	}

	log.Info().Int("count", len(messages)).Int("olderThanHours", olderThanHours).Int("maxReminders", maxReminders).Msg("Retrieved unviewed messages for reminders")
	return messages, nil
}

// LogReminderSent records that a reminder email was sent for a message
func (s *StorageService) LogReminderSent(ctx context.Context, messageID int, emailAddress string) error {
	// Business rule validation
	if messageID < 1 {
		log.Warn().Int("messageID", messageID).Msg("Invalid messageID parameter")
		return ErrInvalidParameter
	}
	if emailAddress == "" {
		log.Warn().Msg("Attempted to log reminder with empty email address")
		return ErrEmptyEmailAddress
	}

	// Delegate to repository
	err := s.repository.LogReminderSent(messageID, emailAddress)
	if err != nil {
		log.Error().Err(err).Int("messageID", messageID).Str("emailAddress", emailAddress).Msg("Failed to log reminder sent")
		return err
	}

	log.Info().Int("messageID", messageID).Str("emailAddress", emailAddress).Msg("Reminder sent logged successfully")
	return nil
}

// GetReminderHistory retrieves the reminder history for a specific message
func (s *StorageService) GetReminderHistory(ctx context.Context, messageID int) ([]*ReminderLogEntry, error) {
	// Business rule validation
	if messageID < 1 {
		log.Warn().Int("messageID", messageID).Msg("Invalid messageID parameter")
		return nil, ErrInvalidParameter
	}

	// Delegate to repository
	history, err := s.repository.GetReminderHistory(messageID)
	if err != nil {
		log.Error().Err(err).Int("messageID", messageID).Msg("Failed to retrieve reminder history")
		return nil, err
	}

	log.Info().Int("messageID", messageID).Int("count", len(history)).Msg("Retrieved reminder history")
	return history, nil
}

// HealthCheck verifies the storage service is healthy
func (s *StorageService) HealthCheck(ctx context.Context) error {
	// For now, just log that health check was called
	// In a real implementation, this might check repository connectivity
	log.Debug().Msg("Storage service health check requested")
	return nil
}
