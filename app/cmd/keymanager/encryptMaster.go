package keymanager

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// Encrypt the data with AES-GCM and given master password
func encryptMasterPassword(data []byte, password string) ([]byte, error) {
	// Use a key derivation function on the password to get the encryption key
	// For the sake of simplicity, here we're just padding or truncating the password to get a 32-byte key.
	// In a real-world scenario, this approach is NOT secure.
	key := []byte(password)
	for len(key) < 32 {
		key = append(key, '0')
	}
	key = key[:32]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aead.Seal(nonce, nonce, data, nil), nil
}

const (
	keyLen     = 32
	saltLen    = 8
	iterations = 4096
)

// generateKey generates a secure encryption key from the given password.
func generateKey(password string) ([]byte, []byte, error) {
	// Generate a new random salt
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, nil, err
	}

	// Use PBKDF2 to derive a key from the password
	key := pbkdf2.Key([]byte(password), salt, iterations, keyLen, sha256.New)

	return key, salt, nil
}

// encryptData encrypts the given data using AES-GCM and the given key.
func encryptData(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aead.Seal(nonce, nonce, data, nil), nil
}

// encryptWithPassword encrypts the given data with a key derived from the given password.
func encryptWithPassword(data []byte, password string) ([]byte, []byte, []byte, error) {
	key, salt, err := generateKey(password)
	if err != nil {
		return nil, nil, nil, err
	}

	encryptedData, err := encryptData(data, key)
	if err != nil {
		return nil, nil, nil, err
	}

	return encryptedData, key, salt, nil
}
