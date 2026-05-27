package memory

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/secondary"
	"github.com/rs/xid"
)

// KeyGenerator implements the KeyGeneratorPort using in-memory operations
type KeyGenerator struct {
	logger secondary.LoggerPort
}

// NewKeyGenerator creates a new memory-based key generator
func NewKeyGenerator(logger secondary.LoggerPort) *KeyGenerator {
	return &KeyGenerator{logger: logger}
}

// GenerateKey generates a cryptographically secure random key
func (g *KeyGenerator) GenerateKey(ctx context.Context, length int32) (contracts.EncryptionKey, error) {
	if length != 32 {
		g.logger.Error().Int32("length", length).Msg("Invalid key length requested")
		return contracts.EncryptionKey{}, domain.ErrInvalidKeyLength
	}

	var key contracts.EncryptionKey
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		g.logger.Error().Err(err).Msg("Failed to generate random key")
		return contracts.EncryptionKey{}, fmt.Errorf("%w: %v", domain.ErrInsufficientRandomness, err)
	}

	g.logger.Debug().Msg("Successfully generated random key")
	return key, nil
}

// GenerateID generates a unique identifier using xid
func (g *KeyGenerator) GenerateID(ctx context.Context) string {
	guid := xid.New()
	id := guid.String()
	g.logger.Debug().Str("id", id).Msg("Generated unique ID")
	return id
}
