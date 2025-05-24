package secondary

import (
	"context"
)

// PasswordHasherPort defines the secondary port for password hashing operations
type PasswordHasherPort interface {
	// Hash creates a hash of the provided password
	Hash(ctx context.Context, password string) (string, error)
	
	// Verify checks if a password matches the provided hash
	Verify(ctx context.Context, password, hash string) (bool, error)
}