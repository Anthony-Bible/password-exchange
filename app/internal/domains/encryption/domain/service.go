package domain

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/secondary"
)

// EncryptionService provides cryptographic operations.
type EncryptionService struct {
	keyGenerator contracts.KeyGenerator
	logger       secondary.LoggerPort
}

// NewEncryptionService creates a new encryption service wired with the supplied
// key generator and logger port.
func NewEncryptionService(keyGenerator contracts.KeyGenerator, logger secondary.LoggerPort) *EncryptionService {
	return &EncryptionService{
		keyGenerator: keyGenerator,
		logger:       logger,
	}
}

// Encrypt encrypts multiple plaintext messages using AES-GCM.
func (s *EncryptionService) Encrypt(ctx context.Context, req contracts.EncryptionRequest) (*contracts.EncryptionResponse, error) {
	if len(req.Key) != 32 {
		s.logger.Error().Int("keyLength", len(req.Key)).Msg("Invalid key length for encryption")
		return nil, ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(req.Key)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create AES cipher")
		return nil, fmt.Errorf("%w: %v", ErrCipherCreationFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create GCM")
		return nil, fmt.Errorf("%w: %v", ErrGCMCreationFailed, err)
	}

	response := &contracts.EncryptionResponse{
		Ciphertext: make([]string, 0, len(req.Plaintext)),
	}

	for _, plaintext := range req.Plaintext {
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			s.logger.Error().Err(err).Msg("Failed to generate nonce")
			return nil, fmt.Errorf("%w: %v", ErrInsufficientRandomness, err)
		}

		ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
		encodedCiphertext := base64.URLEncoding.EncodeToString(ciphertext)
		response.Ciphertext = append(response.Ciphertext, encodedCiphertext)
	}

	s.logger.Debug().Int("plaintextCount", len(req.Plaintext)).Msg("Successfully encrypted messages")
	return response, nil
}

// Decrypt decrypts multiple ciphertext messages using AES-GCM.
func (s *EncryptionService) Decrypt(ctx context.Context, req contracts.DecryptionRequest) (*contracts.DecryptionResponse, error) {
	if len(req.Key) != 32 {
		s.logger.Error().Int("keyLength", len(req.Key)).Msg("Invalid key length for decryption")
		return nil, ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(req.Key)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create AES cipher")
		return nil, fmt.Errorf("%w: %v", ErrCipherCreationFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create GCM")
		return nil, fmt.Errorf("%w: %v", ErrGCMCreationFailed, err)
	}

	response := &contracts.DecryptionResponse{
		Plaintext: make([]string, 0, len(req.Ciphertext)),
	}

	for _, encodedCiphertext := range req.Ciphertext {
		ciphertext, err := base64.URLEncoding.DecodeString(encodedCiphertext)
		if err != nil {
			s.logger.Error().Err(err).Str("ciphertext", encodedCiphertext).Msg("Failed to decode base64 ciphertext")
			return nil, fmt.Errorf("%w: %v", ErrBase64DecodingFailed, err)
		}

		if len(ciphertext) < gcm.NonceSize() {
			s.logger.Error().Int("ciphertextLength", len(ciphertext)).Int("nonceSize", gcm.NonceSize()).Msg("Ciphertext too short")
			return nil, ErrInvalidCiphertext
		}

		nonce := ciphertext[:gcm.NonceSize()]
		encryptedData := ciphertext[gcm.NonceSize():]

		plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to decrypt message")
			return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
		}

		encodedPlaintext := base64.URLEncoding.EncodeToString(plaintext)
		response.Plaintext = append(response.Plaintext, encodedPlaintext)
	}

	s.logger.Debug().Int("ciphertextCount", len(req.Ciphertext)).Msg("Successfully decrypted messages")
	return response, nil
}

// GenerateRandomKey generates a new random encryption key.
func (s *EncryptionService) GenerateRandomKey(ctx context.Context, req contracts.RandomRequest) (*contracts.RandomResponse, error) {
	if req.Length != 32 {
		s.logger.Error().Int32("length", req.Length).Msg("Invalid key length requested")
		return nil, ErrInvalidKeyLength
	}

	key, err := s.keyGenerator.GenerateKey(ctx, req.Length)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate random key")
		return nil, err
	}

	response := &contracts.RandomResponse{
		Key:       key,
		KeyString: key.String(),
	}

	s.logger.Debug().Msg("Successfully generated random key")
	return response, nil
}

// GenerateID generates a new unique identifier.
func (s *EncryptionService) GenerateID(ctx context.Context) string {
	return s.keyGenerator.GenerateID(ctx)
}
