package memory

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/domain"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

// KeyGenerator implements the KeyGeneratorPort using in-memory operations
type KeyGenerator struct{}

// NewKeyGenerator creates a new memory-based key generator
func NewKeyGenerator() *KeyGenerator {
	return &KeyGenerator{}
}

// GenerateKey generates a cryptographically secure random key
func (g *KeyGenerator) GenerateKey(ctx context.Context, length int32) (domain.EncryptionKey, error) {
	if length != 32 {
		log.Error().Int32("length", length).Msg("Invalid key length requested")
		return domain.EncryptionKey{}, domain.ErrInvalidKeyLength
	}

	var key domain.EncryptionKey
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate random key")
		return domain.EncryptionKey{}, fmt.Errorf("%w: %v", domain.ErrInsufficientRandomness, err)
	}

	log.Debug().Msg("Successfully generated random key")
	return key, nil
}

// GenerateID generates a unique identifier using xid
func (g *KeyGenerator) GenerateID(ctx context.Context) string {
	guid := xid.New()
	id := guid.String()
	log.Debug().Str("id", id).Msg("Generated unique ID")
	return id
}