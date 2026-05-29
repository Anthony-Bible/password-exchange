// Package contracts defines shared data types and interfaces used across the
// encryption domain's layers. Centralizing these types here prevents import
// cycles between domain, ports, and adapters.
package contracts

import (
	"context"
	"encoding/base64"

	logport "github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/port"
)

// EncryptionKey represents a 32-byte encryption key.
type EncryptionKey [32]byte

// String returns the base64 URL-encoded representation of the key.
func (k EncryptionKey) String() string {
	return base64.URLEncoding.EncodeToString(k[:])
}

// Bytes returns the raw bytes of the key.
func (k EncryptionKey) Bytes() []byte {
	return k[:]
}

// EncryptionRequest represents a request to encrypt plaintext.
type EncryptionRequest struct {
	Plaintext []string
	Key       []byte
}

// EncryptionResponse represents the result of encryption.
type EncryptionResponse struct {
	Ciphertext []string
}

// DecryptionRequest represents a request to decrypt ciphertext.
type DecryptionRequest struct {
	Ciphertext []string
	Key        []byte
}

// DecryptionResponse represents the result of decryption.
type DecryptionResponse struct {
	Plaintext []string
}

// RandomRequest represents a request for random key generation.
type RandomRequest struct {
	Length int32
}

// RandomResponse represents the result of random key generation.
type RandomResponse struct {
	Key       EncryptionKey
	KeyString string
}

// KeyGenerator defines the interface for generating encryption keys and unique IDs.
type KeyGenerator interface {
	// GenerateKey generates a cryptographically secure random key of the specified length.
	GenerateKey(ctx context.Context, length int32) (EncryptionKey, error)

	// GenerateID generates a unique identifier.
	GenerateID(ctx context.Context) string
}

// LogEvent is the shared structured-logging event contract. It is an alias to
// the single definition in internal/shared/logging/port, so the encryption
// domain stays independent of any concrete logging implementation while a new
// field is added in exactly one place across every domain.
type LogEvent = logport.LogEvent
