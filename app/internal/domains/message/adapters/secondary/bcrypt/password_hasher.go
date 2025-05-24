package bcrypt

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher implements the PasswordHasherPort using bcrypt
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new bcrypt password hasher
func NewPasswordHasher(cost int) *PasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	return &PasswordHasher{
		cost: cost,
	}
}

// Hash creates a hash of the provided password
func (h *PasswordHasher) Hash(ctx context.Context, password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	hashedPassword := string(hashedBytes)
	log.Debug().Msg("Password hashed successfully")
	return hashedPassword, nil
}

// Verify checks if a password matches the provided hash
func (h *PasswordHasher) Verify(ctx context.Context, password, hash string) (bool, error) {
	// If hash is empty, consider it as no password required
	if strings.TrimSpace(hash) == "" {
		log.Debug().Msg("No password hash provided, allowing access")
		return true, nil
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			log.Debug().Msg("Password verification failed - mismatch")
			return false, nil
		}
		log.Error().Err(err).Msg("Failed to verify password")
		return false, fmt.Errorf("failed to verify password: %w", err)
	}

	log.Debug().Msg("Password verified successfully")
	return true, nil
}