package primary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/domain"
)

// EncryptionServicePort defines the primary port for encryption operations
type EncryptionServicePort interface {
	// Encrypt encrypts multiple plaintext messages
	Encrypt(ctx context.Context, req domain.EncryptionRequest) (*domain.EncryptionResponse, error)
	
	// Decrypt decrypts multiple ciphertext messages
	Decrypt(ctx context.Context, req domain.DecryptionRequest) (*domain.DecryptionResponse, error)
	
	// GenerateRandomKey generates a new random encryption key
	GenerateRandomKey(ctx context.Context, req domain.RandomRequest) (*domain.RandomResponse, error)
	
	// GenerateID generates a new unique identifier
	GenerateID() string
}