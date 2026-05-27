package primary

import (
	"context"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"
)

// EncryptionServicePort defines the primary port for encryption operations
type EncryptionServicePort interface {
	// Encrypt encrypts multiple plaintext messages
	Encrypt(ctx context.Context, req contracts.EncryptionRequest) (*contracts.EncryptionResponse, error)

	// Decrypt decrypts multiple ciphertext messages
	Decrypt(ctx context.Context, req contracts.DecryptionRequest) (*contracts.DecryptionResponse, error)

	// GenerateRandomKey generates a new random encryption key
	GenerateRandomKey(ctx context.Context, req contracts.RandomRequest) (*contracts.RandomResponse, error)

	// GenerateID generates a new unique identifier
	GenerateID(ctx context.Context) string
}
