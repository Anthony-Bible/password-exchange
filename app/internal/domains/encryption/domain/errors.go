package domain

import "errors"

var (
	// ErrInsufficientRandomness indicates the system cannot generate secure random bytes
	ErrInsufficientRandomness = errors.New("insufficient randomness available")
	
	// ErrInvalidKeyLength indicates the provided key length is invalid
	ErrInvalidKeyLength = errors.New("invalid key length")
	
	// ErrInvalidCiphertext indicates the ciphertext is malformed or invalid
	ErrInvalidCiphertext = errors.New("malformed ciphertext")
	
	// ErrCipherCreationFailed indicates AES cipher creation failed
	ErrCipherCreationFailed = errors.New("failed to create cipher")
	
	// ErrGCMCreationFailed indicates GCM mode creation failed
	ErrGCMCreationFailed = errors.New("failed to create GCM")
	
	// ErrDecryptionFailed indicates decryption operation failed
	ErrDecryptionFailed = errors.New("decryption failed")
	
	// ErrEncryptionFailed indicates encryption operation failed
	ErrEncryptionFailed = errors.New("encryption failed")
	
	// ErrBase64DecodingFailed indicates base64 decoding failed
	ErrBase64DecodingFailed = errors.New("base64 decoding failed")
)