package grpc_clients

import (
	"context"
	"fmt"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	db "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// StorageClient implements the StorageServicePort using gRPC
type StorageClient struct {
	client db.DbServiceClient
	conn   *grpc.ClientConn
}

// NewStorageClient creates a new storage gRPC client
func NewStorageClient(endpoint string) (*StorageClient, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to connect to storage service")
		return nil, fmt.Errorf("failed to connect to storage service: %w", err)
	}

	client := db.NewDbServiceClient(conn)

	return &StorageClient{
		client: client,
		conn:   conn,
	}, nil
}

// StoreMessage stores an encrypted message
func (c *StorageClient) StoreMessage(ctx context.Context, req domain.MessageStorageRequest) error {
	grpcReq := &db.InsertRequest{
		Uuid:           req.MessageID,
		Content:        req.Content,
		Passphrase:     req.Passphrase,
		MaxViewCount:   int32(req.MaxViewCount),
		RecipientEmail: req.RecipientEmail,
	}

	_, err := c.client.Insert(ctx, grpcReq)
	if err != nil {
		log.Error().Err(err).Str("messageId", req.MessageID).Msg("Failed to store message")
		return fmt.Errorf("failed to store message: %w", err)
	}

	log.Debug().Str("messageId", req.MessageID).Int("maxViewCount", req.MaxViewCount).Msg("Stored message successfully")
	return nil
}

// parseExpiresAt parses an RFC3339 timestamp string into a *time.Time, returning nil on empty or error.
func parseExpiresAt(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		log.Warn().
			Err(err).
			Str("value", s).
			Msg("Failed to parse expires_at timestamp from storage; treating as no expiry")
		return nil
	}
	return &t
}

// RetrieveMessage retrieves a stored message by ID
func (c *StorageClient) RetrieveMessage(
	ctx context.Context,
	req domain.MessageRetrievalStorageRequest,
) (*domain.MessageStorageResponse, error) {
	grpcReq := &db.SelectRequest{
		Uuid: req.MessageID,
	}

	resp, err := c.client.Select(ctx, grpcReq)
	if err != nil {
		log.Error().Err(err).Str("messageId", req.MessageID).Msg("Failed to retrieve message")
		return nil, fmt.Errorf("failed to retrieve message: %w", err)
	}

	hasPassphrase := resp.GetPassphrase() != ""

	response := &domain.MessageStorageResponse{
		MessageID:        req.MessageID,
		EncryptedContent: resp.GetContent(),
		HashedPassphrase: resp.GetPassphrase(),
		HasPassphrase:    hasPassphrase,
		ViewCount:        int(resp.GetViewCount()),
		MaxViewCount:     int(resp.GetMaxViewCount()),
		ExpiresAt:        parseExpiresAt(resp.GetExpiresAt()),
	}

	log.Debug().
		Str("messageId", req.MessageID).
		Bool("hasPassphrase", hasPassphrase).
		Msg("Retrieved message successfully")
	return response, nil
}

// GetMessage retrieves a message by its unique ID without incrementing view count
func (c *StorageClient) GetMessage(
	ctx context.Context,
	req domain.MessageRetrievalStorageRequest,
) (*domain.MessageStorageResponse, error) {
	grpcReq := &db.SelectRequest{
		Uuid: req.MessageID,
	}

	resp, err := c.client.GetMessage(ctx, grpcReq)
	if err != nil {
		log.Error().
			Err(err).
			Str("messageId", req.MessageID).
			Msg("Failed to retrieve message without incrementing view count")
		return nil, fmt.Errorf("failed to retrieve message without incrementing view count: %w", err)
	}

	hasPassphrase := resp.GetPassphrase() != ""

	response := &domain.MessageStorageResponse{
		MessageID:        req.MessageID,
		EncryptedContent: resp.GetContent(),
		HashedPassphrase: resp.GetPassphrase(),
		HasPassphrase:    hasPassphrase,
		ViewCount:        int(resp.GetViewCount()),
		MaxViewCount:     int(resp.GetMaxViewCount()),
		ExpiresAt:        parseExpiresAt(resp.GetExpiresAt()),
	}

	log.Debug().
		Str("messageId", req.MessageID).
		Bool("hasPassphrase", hasPassphrase).
		Msg("Retrieved message without incrementing view count successfully")
	return response, nil
}

// Close closes the gRPC connection
func (c *StorageClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
