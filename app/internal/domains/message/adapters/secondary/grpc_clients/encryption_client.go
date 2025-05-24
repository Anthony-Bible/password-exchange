package grpc_clients

import (
	"context"
	"fmt"

	pb "github.com/Anthony-Bible/password-exchange/app/pkg/pb/encryption"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// EncryptionClient implements the EncryptionServicePort using gRPC
type EncryptionClient struct {
	client pb.MessageServiceClient
	conn   *grpc.ClientConn
}

// NewEncryptionClient creates a new encryption gRPC client
func NewEncryptionClient(endpoint string) (*EncryptionClient, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to connect to encryption service")
		return nil, fmt.Errorf("failed to connect to encryption service: %w", err)
	}

	client := pb.NewMessageServiceClient(conn)

	return &EncryptionClient{
		client: client,
		conn:   conn,
	}, nil
}

// GenerateKey generates a new encryption key
func (c *EncryptionClient) GenerateKey(ctx context.Context, length int32) ([]byte, error) {
	req := &pb.Randomrequest{
		RandomLength: length,
	}

	resp, err := c.client.GenerateRandomString(ctx, req)
	if err != nil {
		log.Error().Err(err).Int32("length", length).Msg("Failed to generate encryption key")
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	log.Debug().Int32("length", length).Msg("Generated encryption key successfully")
	return resp.GetEncryptionbytes(), nil
}

// Encrypt encrypts plaintext using the provided key
func (c *EncryptionClient) Encrypt(ctx context.Context, plaintext []string, key []byte) ([]string, error) {
	req := &pb.EncryptedMessageRequest{
		PlainText: plaintext,
		Key:       key,
	}

	resp, err := c.client.EncryptMessage(ctx, req)
	if err != nil {
		log.Error().Err(err).Int("plaintextCount", len(plaintext)).Msg("Failed to encrypt message")
		return nil, fmt.Errorf("failed to encrypt message: %w", err)
	}

	log.Debug().Int("plaintextCount", len(plaintext)).Int("ciphertextCount", len(resp.GetCiphertext())).Msg("Encrypted message successfully")
	return resp.GetCiphertext(), nil
}

// Decrypt decrypts ciphertext using the provided key
func (c *EncryptionClient) Decrypt(ctx context.Context, ciphertext []string, key []byte) ([]string, error) {
	req := &pb.DecryptedMessageRequest{
		Ciphertext: ciphertext,
		Key:        key,
	}

	resp, err := c.client.DecryptMessage(ctx, req)
	if err != nil {
		log.Error().Err(err).Int("ciphertextCount", len(ciphertext)).Msg("Failed to decrypt message")
		return nil, fmt.Errorf("failed to decrypt message: %w", err)
	}

	log.Debug().Int("ciphertextCount", len(ciphertext)).Int("plaintextCount", len(resp.GetPlaintext())).Msg("Decrypted message successfully")
	return resp.GetPlaintext(), nil
}

// GenerateID generates a unique identifier
func (c *EncryptionClient) GenerateID(ctx context.Context) (string, error) {
	// For now, we'll use the encryption service to generate a key and use it as an ID
	// In a real implementation, this might be a separate endpoint
	req := &pb.Randomrequest{
		RandomLength: 32,
	}

	resp, err := c.client.GenerateRandomString(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate unique ID")
		return "", fmt.Errorf("failed to generate unique ID: %w", err)
	}

	// Use the string representation of the key as the ID
	id := resp.GetEncryptionString()
	log.Debug().Str("id", id).Msg("Generated unique ID successfully")
	return id, nil
}

// Close closes the gRPC connection
func (c *EncryptionClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}