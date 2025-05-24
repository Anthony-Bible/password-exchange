package secondary

import (
	"context"
)

// EncryptionServicePort defines the secondary port for encryption operations
type EncryptionServicePort interface {
	// GenerateKey generates a new encryption key of the specified length
	GenerateKey(ctx context.Context, length int32) ([]byte, error)
	
	// Encrypt encrypts plaintext using the provided key
	Encrypt(ctx context.Context, plaintext []string, key []byte) ([]string, error)
	
	// Decrypt decrypts ciphertext using the provided key
	Decrypt(ctx context.Context, ciphertext []string, key []byte) ([]string, error)
	
	// GenerateID generates a unique identifier
	GenerateID(ctx context.Context) (string, error)
}