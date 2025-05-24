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
func (s *StorageService) StoreMessage(ctx context.Context, content, uniqueID, passphrase string) error {
	// Business rule validation
	if content == "" {
		log.Warn().Msg("Attempted to store message with empty content")
		return ErrEmptyContent
	}
	if uniqueID == "" {
		log.Warn().Msg("Attempted to store message with empty unique ID")
		return ErrEmptyUniqueID
	}
	
	// Delegate to repository
	return s.repository.InsertMessage(content, uniqueID, passphrase)
}

// RetrieveMessage retrieves a message by its unique ID with validation
func (s *StorageService) RetrieveMessage(ctx context.Context, uniqueID string) (*Message, error) {
	// Business rule validation
	if uniqueID == "" {
		log.Warn().Msg("Attempted to retrieve message with empty unique ID")
		return nil, ErrEmptyUniqueID
	}
	
	// Delegate to repository
	return s.repository.SelectMessageByUniqueID(uniqueID)
}

// CleanupExpiredMessages removes expired messages from storage
func (s *StorageService) CleanupExpiredMessages(ctx context.Context) error {
	log.Info().Msg("Starting cleanup of expired messages")
	return s.repository.DeleteExpiredMessages()
}

// HealthCheck verifies the storage service is healthy
func (s *StorageService) HealthCheck(ctx context.Context) error {
	// For now, just log that health check was called
	// In a real implementation, this might check repository connectivity
	log.Debug().Msg("Storage service health check requested")
	return nil
}