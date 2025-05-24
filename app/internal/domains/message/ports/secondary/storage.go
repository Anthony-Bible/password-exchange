package secondary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
)

// StorageServicePort defines the secondary port for message storage operations
type StorageServicePort interface {
	// StoreMessage stores an encrypted message
	StoreMessage(ctx context.Context, req domain.MessageStorageRequest) error
	
	// RetrieveMessage retrieves a stored message by ID
	RetrieveMessage(ctx context.Context, req domain.MessageRetrievalStorageRequest) (*domain.MessageStorageResponse, error)
}