package secondary

import (
	"context"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
)

// StorageServicePort defines the secondary port for message storage operations
type StorageServicePort interface {
	// StoreMessage stores an encrypted message
	StoreMessage(ctx context.Context, req contracts.MessageStorageRequest) error

	// RetrieveMessage retrieves a stored message by ID
	RetrieveMessage(ctx context.Context, req contracts.MessageRetrievalStorageRequest) (*contracts.MessageStorageResponse, error)

	// GetMessage retrieves message metadata without incrementing view count
	GetMessage(ctx context.Context, req contracts.MessageRetrievalStorageRequest) (*contracts.MessageStorageResponse, error)

	// HealthCheck verifies the underlying storage service is reachable and
	// serving. Implementations SHOULD honour ctx for cancellation/timeout so
	// the web service's /readyz probe cannot wedge on a hung backend.
	HealthCheck(ctx context.Context) error
}
