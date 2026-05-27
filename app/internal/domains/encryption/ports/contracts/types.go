// Package contracts defines shared data types and interfaces used across the
// encryption domain's layers. Centralizing these types here prevents import
// cycles between domain, ports, and adapters.
package contracts

import (
	"context"
	"encoding/base64"
	"time"
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

// LogEvent represents a structured logging event that can be enriched with contextual data.
// This interface follows a fluent API pattern, allowing method chaining to add various
// types of contextual information before finalizing the log entry. This abstraction
// allows the encryption domain to remain independent of specific logging implementations.
type LogEvent interface {
	// Err adds an error to the log event.
	// The error will be formatted and included in the log output.
	//
	// Parameters:
	//   - err: The error to log (can be nil)
	//
	// Returns:
	//   - The LogEvent for method chaining
	Err(error) LogEvent

	// Str adds a string key-value pair to the log event.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The string value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Str(string, string) LogEvent

	// Int adds an integer key-value pair to the log event.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The integer value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Int(string, int) LogEvent

	// Int32 adds an int32 key-value pair to the log event.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The int32 value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Int32(string, int32) LogEvent

	// Bool adds a boolean key-value pair to the log event.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The boolean value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Bool(string, bool) LogEvent

	// Dur adds a duration key-value pair to the log event.
	// The duration is typically formatted in a human-readable way.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The duration value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Dur(string, time.Duration) LogEvent

	// Float64 adds a float64 key-value pair to the log event.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The float64 value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Float64(string, float64) LogEvent

	// Msg finalizes the log event with a message and writes it to the log.
	// This method should be called last in the chain.
	//
	// Parameters:
	//   - message: The log message describing the event
	Msg(string)
}
