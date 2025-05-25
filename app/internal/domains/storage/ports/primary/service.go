package primary

import (
	"context"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
)

// StorageServicePort defines the primary interface for storage operations
// This will be implemented by the storage service and used by external adapters
type StorageServicePort interface {
	// StoreMessage stores a new encrypted message with unique ID, passphrase, and max view count
	StoreMessage(ctx context.Context, content, uniqueID, passphrase string, maxViewCount int) error

	// RetrieveMessage retrieves a message by its unique ID
	RetrieveMessage(ctx context.Context, uniqueID string) (*domain.Message, error)

	// GetMessage retrieves a message by its unique ID without incrementing view count
	GetMessage(ctx context.Context, uniqueID string) (*domain.Message, error)

	// CleanupExpiredMessages removes expired messages from storage
	CleanupExpiredMessages(ctx context.Context) error

	// HealthCheck verifies the storage service is healthy
	HealthCheck(ctx context.Context) error
}
