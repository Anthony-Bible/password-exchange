package grpc_clients

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ensure UploadStateGRPCAdapter implements the secondary port at compile time.
var _ secondary.UploadStatePort = (*UploadStateGRPCAdapter)(nil)

// UploadStateGRPCAdapter implements secondary.UploadStatePort by forwarding
// every call to the database service over gRPC. It reuses the DbServiceClient
// already wired for message-storage calls so no extra connection is opened.
type UploadStateGRPCAdapter struct {
	client database.DbServiceClient
}

// NewUploadStateGRPCAdapter creates an upload-state adapter backed by the
// given DbServiceClient. The caller is responsible for the client's lifecycle.
func NewUploadStateGRPCAdapter(client database.DbServiceClient) *UploadStateGRPCAdapter {
	return &UploadStateGRPCAdapter{client: client}
}

// uploadSessionToProto converts a message-domain FileUploadSession to the
// proto wire type used by the database service.
func fileUploadSessionToProto(s contracts.FileUploadSession) *database.UploadSession {
	parts := make([]*database.UploadSessionPart, 0, len(s.CompletedParts))
	for _, p := range s.CompletedParts {
		parts = append(parts, &database.UploadSessionPart{
			PartNumber: int32(p.PartNumber),
			Etag:       p.ETag,
		})
	}
	return &database.UploadSession{
		SessionId:      s.SessionID,
		FileId:         s.FileID,
		MessageId:      s.MessageID,
		UploadId:       s.UploadID,
		Filename:       s.Filename,
		ContentType:    s.ContentType,
		TotalSize:      s.TotalSize,
		TotalChunks:    int32(s.TotalChunks),
		Status:         s.Status,
		EncryptionKey:  s.EncryptionKey,
		CompletedParts: parts,
		CreatedAt:      s.CreatedAt.UTC().Format(time.RFC3339),
		ExpiresAt:      s.ExpiresAt.UTC().Format(time.RFC3339),
	}
}

// protoToFileUploadSession converts the proto wire type to the message-domain
// FileUploadSession contract.
func protoToFileUploadSession(p *database.UploadSession) (*contracts.FileUploadSession, error) {
	if p == nil {
		return nil, fmt.Errorf("grpc_clients: nil UploadSession in response")
	}

	createdAt, err := parseUploadStateTimestamp(p.GetCreatedAt())
	if err != nil {
		return nil, fmt.Errorf("grpc_clients: invalid created_at: %w", err)
	}
	expiresAt, err := parseUploadStateTimestamp(p.GetExpiresAt())
	if err != nil {
		return nil, fmt.Errorf("grpc_clients: invalid expires_at: %w", err)
	}

	parts := make([]contracts.CompletedPart, 0, len(p.GetCompletedParts()))
	for _, pp := range p.GetCompletedParts() {
		parts = append(parts, contracts.CompletedPart{
			PartNumber: int(pp.GetPartNumber()),
			ETag:       pp.GetEtag(),
		})
	}

	return &contracts.FileUploadSession{
		SessionID:      p.GetSessionId(),
		FileID:         p.GetFileId(),
		MessageID:      p.GetMessageId(),
		UploadID:       p.GetUploadId(),
		Filename:       p.GetFilename(),
		ContentType:    p.GetContentType(),
		TotalSize:      p.GetTotalSize(),
		TotalChunks:    int(p.GetTotalChunks()),
		Status:         p.GetStatus(),
		EncryptionKey:  p.GetEncryptionKey(),
		CompletedParts: parts,
		CreatedAt:      createdAt,
		ExpiresAt:      expiresAt,
	}, nil
}

// parseUploadStateTimestamp parses an RFC3339 string, returning the zero time
// for empty strings.
func parseUploadStateTimestamp(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse RFC3339: %w", err)
	}
	return t, nil
}

// grpcUploadStatusToErr maps well-known gRPC status codes from the database
// service back to message-domain sentinel errors so callers can use errors.Is.
func grpcUploadStatusToErr(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case codes.NotFound:
		return domain.ErrUploadSessionNotFound
	case codes.AlreadyExists:
		return domain.ErrUploadSessionAlreadyExists
	}
	return err
}

// CreateSession persists a new upload session via the database service.
func (a *UploadStateGRPCAdapter) CreateSession(ctx context.Context, session contracts.FileUploadSession) error {
	_, err := a.client.CreateUploadSession(ctx, &database.CreateUploadSessionRequest{
		Session: fileUploadSessionToProto(session),
	})
	if err != nil {
		return fmt.Errorf("grpc_clients: CreateSession: %w", grpcUploadStatusToErr(err))
	}
	return nil
}

// GetSessionByID retrieves a session by its SessionID via the database service.
func (a *UploadStateGRPCAdapter) GetSessionByID(ctx context.Context, sessionID string) (*contracts.FileUploadSession, error) {
	resp, err := a.client.GetUploadSession(ctx, &database.GetUploadSessionRequest{
		SessionId: sessionID,
	})
	if err != nil {
		return nil, grpcUploadStatusToErr(err)
	}
	if resp.GetSession() == nil {
		return nil, domain.ErrUploadSessionNotFound
	}
	sess, err := protoToFileUploadSession(resp.GetSession())
	if err != nil {
		return nil, fmt.Errorf("grpc_clients: GetSessionByID response decode: %w", err)
	}
	return sess, nil
}

// GetSessionByFileID retrieves a session by its FileID via the database service.
func (a *UploadStateGRPCAdapter) GetSessionByFileID(ctx context.Context, fileID string) (*contracts.FileUploadSession, error) {
	resp, err := a.client.GetUploadSession(ctx, &database.GetUploadSessionRequest{
		FileId: fileID,
	})
	if err != nil {
		return nil, grpcUploadStatusToErr(err)
	}
	if resp.GetSession() == nil {
		return nil, domain.ErrUploadSessionNotFound
	}
	sess, err := protoToFileUploadSession(resp.GetSession())
	if err != nil {
		return nil, fmt.Errorf("grpc_clients: GetSessionByFileID response decode: %w", err)
	}
	return sess, nil
}

// AddCompletedPart records a successfully uploaded part via the database service
// and returns the full post-add completed-parts list by re-fetching the session.
func (a *UploadStateGRPCAdapter) AddCompletedPart(ctx context.Context, sessionID string, part contracts.CompletedPart) ([]contracts.CompletedPart, error) {
	if part.PartNumber <= 0 || part.PartNumber > math.MaxInt32 {
		return nil, fmt.Errorf("grpc_clients: AddCompletedPart: invalid part number %d", part.PartNumber)
	}
	partNumber := int32(part.PartNumber)

	_, err := a.client.AddCompletedPart(ctx, &database.AddCompletedPartRequest{
		SessionId: sessionID,
		Part: &database.UploadSessionPart{
			PartNumber: partNumber,
			Etag:       part.ETag,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("grpc_clients: AddCompletedPart: %w", grpcUploadStatusToErr(err))
	}

	sess, err := a.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("grpc_clients: AddCompletedPart re-fetch: %w", err)
	}
	return sess.CompletedParts, nil
}

// CompleteSession marks the session assembled and clears the encryption key
// via the database service.
func (a *UploadStateGRPCAdapter) CompleteSession(ctx context.Context, sessionID string) error {
	_, err := a.client.CompleteUploadSession(ctx, &database.CompleteUploadSessionRequest{
		SessionId: sessionID,
	})
	if err != nil {
		return grpcUploadStatusToErr(err)
	}
	return nil
}

// DeleteSession removes the session via the database service.
func (a *UploadStateGRPCAdapter) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := a.client.DeleteUploadSession(ctx, &database.DeleteUploadSessionRequest{
		SessionId: sessionID,
	})
	if err != nil {
		return fmt.Errorf("grpc_clients: DeleteSession: %w", grpcUploadStatusToErr(err))
	}
	return nil
}

// DeleteExpiredSessions sweeps expired incomplete sessions via the database
// service and returns the removed sessions for caller-side cleanup.
func (a *UploadStateGRPCAdapter) DeleteExpiredSessions(ctx context.Context, asOf time.Time) ([]contracts.FileUploadSession, error) {
	resp, err := a.client.DeleteExpiredUploadSessions(ctx, &database.DeleteExpiredUploadSessionsRequest{
		AsOf: asOf.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("grpc_clients: DeleteExpiredSessions: %w", err)
	}

	removed := make([]contracts.FileUploadSession, 0, len(resp.GetRemovedSessions()))
	for _, p := range resp.GetRemovedSessions() {
		sess, err := protoToFileUploadSession(p)
		if err != nil {
			return nil, fmt.Errorf("grpc_clients: DeleteExpiredSessions decode: %w", err)
		}
		removed = append(removed, *sess)
	}
	return removed, nil
}
