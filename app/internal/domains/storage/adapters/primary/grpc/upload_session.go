package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// uploadSessionToProto converts a contracts.UploadSession to its wire representation.
func uploadSessionToProto(s *contracts.UploadSession) *database.UploadSession {
	if s == nil {
		return nil
	}
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

// uploadSessionFromProto converts a proto UploadSession to its domain contract.
func uploadSessionFromProto(p *database.UploadSession) (contracts.UploadSession, error) {
	if p == nil {
		return contracts.UploadSession{}, fmt.Errorf("nil UploadSession in request")
	}
	createdAt, err := parseUploadTimestamp(p.GetCreatedAt())
	if err != nil {
		return contracts.UploadSession{}, fmt.Errorf("invalid created_at: %w", err)
	}
	expiresAt, err := parseUploadTimestamp(p.GetExpiresAt())
	if err != nil {
		return contracts.UploadSession{}, fmt.Errorf("invalid expires_at: %w", err)
	}

	parts := make([]contracts.UploadSessionPart, 0, len(p.GetCompletedParts()))
	for _, pp := range p.GetCompletedParts() {
		parts = append(parts, contracts.UploadSessionPart{
			PartNumber: int(pp.GetPartNumber()),
			ETag:       pp.GetEtag(),
		})
	}

	return contracts.UploadSession{
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

// parseUploadTimestamp parses an RFC3339 timestamp, returning the zero time for
// empty strings. Upload-session timestamps default to zero when absent.
func parseUploadTimestamp(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse RFC3339: %w", err)
	}
	return t, nil
}

// uploadSessionDomainErrorToStatus maps upload-session domain sentinels to
// typed gRPC status codes so clients receive deterministic, retryable signals.
func uploadSessionDomainErrorToStatus(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, domain.ErrUploadSessionNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrUploadSessionAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInvalidParameter):
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return err
}

// CreateUploadSession persists a new file upload session.
func (s *GRPCServer) CreateUploadSession(
	ctx context.Context,
	req *database.CreateUploadSessionRequest,
) (*emptypb.Empty, error) {
	if req.GetSession() == nil {
		return nil, status.Error(codes.InvalidArgument, "session is required")
	}

	sess, err := uploadSessionFromProto(req.GetSession())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid session: %v", err)
	}

	if err := s.storageService.CreateUploadSession(ctx, sess); err != nil {
		s.logger.Error().Err(err).Str("sessionID", sess.SessionID).Msg("Failed to create upload session via gRPC")
		return nil, uploadSessionDomainErrorToStatus(err)
	}

	s.logger.Info().Str("sessionID", sess.SessionID).Str("fileID", sess.FileID).Msg("Upload session created via gRPC")
	return &emptypb.Empty{}, nil
}

// GetUploadSession retrieves a session by its session_id or file_id.
// At least one of the two fields must be non-empty.
func (s *GRPCServer) GetUploadSession(
	ctx context.Context,
	req *database.GetUploadSessionRequest,
) (*database.GetUploadSessionResponse, error) {
	id := req.GetSessionId()
	if id == "" {
		id = req.GetFileId()
	}
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id or file_id is required")
	}

	sess, err := s.storageService.GetUploadSession(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id).Msg("Failed to get upload session via gRPC")
		return nil, uploadSessionDomainErrorToStatus(err)
	}

	s.logger.Info().Str("sessionID", sess.SessionID).Msg("Upload session retrieved via gRPC")
	return &database.GetUploadSessionResponse{Session: uploadSessionToProto(sess)}, nil
}

// AddCompletedPart records a successfully uploaded chunk for an upload session.
func (s *GRPCServer) AddCompletedPart(
	ctx context.Context,
	req *database.AddCompletedPartRequest,
) (*emptypb.Empty, error) {
	sessionID := req.GetSessionId()
	if sessionID == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}
	if req.GetPart() == nil {
		return nil, status.Error(codes.InvalidArgument, "part is required")
	}

	part := contracts.UploadSessionPart{
		PartNumber: int(req.GetPart().GetPartNumber()),
		ETag:       req.GetPart().GetEtag(),
	}

	if err := s.storageService.AddCompletedPart(ctx, sessionID, part); err != nil {
		s.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to add completed part via gRPC")
		return nil, uploadSessionDomainErrorToStatus(err)
	}

	s.logger.Info().Str("sessionID", sessionID).Int("partNumber", part.PartNumber).Msg("Completed part added via gRPC")
	return &emptypb.Empty{}, nil
}

// CompleteUploadSession marks the session assembled and clears its encryption key.
func (s *GRPCServer) CompleteUploadSession(
	ctx context.Context,
	req *database.CompleteUploadSessionRequest,
) (*emptypb.Empty, error) {
	sessionID := req.GetSessionId()
	if sessionID == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	if err := s.storageService.CompleteUploadSession(ctx, sessionID); err != nil {
		s.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to complete upload session via gRPC")
		return nil, uploadSessionDomainErrorToStatus(err)
	}

	s.logger.Info().Str("sessionID", sessionID).Msg("Upload session completed via gRPC, encryption key cleared")
	return &emptypb.Empty{}, nil
}

// DeleteUploadSession removes a single upload session row.
func (s *GRPCServer) DeleteUploadSession(
	ctx context.Context,
	req *database.DeleteUploadSessionRequest,
) (*emptypb.Empty, error) {
	sessionID := req.GetSessionId()
	if sessionID == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	if err := s.storageService.DeleteUploadSession(ctx, sessionID); err != nil {
		s.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to delete upload session via gRPC")
		return nil, uploadSessionDomainErrorToStatus(err)
	}

	s.logger.Info().Str("sessionID", sessionID).Msg("Upload session deleted via gRPC")
	return &emptypb.Empty{}, nil
}

// DeleteExpiredUploadSessions sweeps incomplete sessions expiring before as_of
// and returns the removed sessions for caller-side object-storage cleanup.
func (s *GRPCServer) DeleteExpiredUploadSessions(
	ctx context.Context,
	req *database.DeleteExpiredUploadSessionsRequest,
) (*database.DeleteExpiredUploadSessionsResponse, error) {
	asOf, err := parseUploadTimestamp(req.GetAsOf())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid as_of: %v", err)
	}
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}

	removed, err := s.storageService.DeleteExpiredUploadSessions(ctx, asOf)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to delete expired upload sessions via gRPC")
		return nil, uploadSessionDomainErrorToStatus(err)
	}

	proto := make([]*database.UploadSession, 0, len(removed))
	for i := range removed {
		proto = append(proto, uploadSessionToProto(&removed[i]))
	}

	s.logger.Info().Int("count", len(removed)).Msg("Expired upload sessions deleted via gRPC")
	return &database.DeleteExpiredUploadSessionsResponse{RemovedSessions: proto}, nil
}
