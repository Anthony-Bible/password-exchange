package secondary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/domain"
)

// KeyGeneratorPort defines the secondary port for key generation operations
type KeyGeneratorPort interface {
	// GenerateKey generates a cryptographically secure random key of the specified length
	GenerateKey(ctx context.Context, length int32) (domain.EncryptionKey, error)
	
	// GenerateID generates a unique identifier
	GenerateID(ctx context.Context) string
}