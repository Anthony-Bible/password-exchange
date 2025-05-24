package domain

import (
	"context"
	"encoding/base64"
)

// EncryptionKey represents a 32-byte encryption key
type EncryptionKey [32]byte

// String returns the base64 URL-encoded representation of the key
func (k EncryptionKey) String() string {
	return base64.URLEncoding.EncodeToString(k[:])
}

// Bytes returns the raw bytes of the key
func (k EncryptionKey) Bytes() []byte {
	return k[:]
}

// EncryptionRequest represents a request to encrypt plaintext
type EncryptionRequest struct {
	Plaintext []string
	Key       []byte
}

// EncryptionResponse represents the result of encryption
type EncryptionResponse struct {
	Ciphertext []string
}

// DecryptionRequest represents a request to decrypt ciphertext
type DecryptionRequest struct {
	Ciphertext []string
	Key        []byte
}

// DecryptionResponse represents the result of decryption
type DecryptionResponse struct {
	Plaintext []string
}

// RandomRequest represents a request for random key generation
type RandomRequest struct {
	Length int32
}

// RandomResponse represents the result of random key generation
type RandomResponse struct {
	Key       EncryptionKey
	KeyString string
}

// KeyGenerator defines the interface for generating encryption keys
type KeyGenerator interface {
	GenerateKey(ctx context.Context, length int32) (EncryptionKey, error)
	GenerateID(ctx context.Context) string
}