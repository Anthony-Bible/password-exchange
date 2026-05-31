package grpc_clients

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// stubDBClientForUpload is a hand-written stub implementing database.DbServiceClient.
// The non-upload-session methods panic to catch accidental calls.
type stubDBClientForUpload struct {
	createErr   error
	getResp     *database.GetUploadSessionResponse
	getErr      error
	addPartErr  error
	completeErr error
	deleteErr   error
	expiredResp *database.DeleteExpiredUploadSessionsResponse
	expiredErr  error
}

func (s *stubDBClientForUpload) Select(context.Context, *database.SelectRequest, ...grpc.CallOption) (*database.SelectResponse, error) {
	panic("unexpected call: Select")
}
func (s *stubDBClientForUpload) Insert(context.Context, *database.InsertRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("unexpected call: Insert")
}
func (s *stubDBClientForUpload) GetMessage(context.Context, *database.SelectRequest, ...grpc.CallOption) (*database.SelectResponse, error) {
	panic("unexpected call: GetMessage")
}
func (s *stubDBClientForUpload) GetUnviewedMessagesForReminders(context.Context, *database.GetUnviewedMessagesRequest, ...grpc.CallOption) (*database.GetUnviewedMessagesResponse, error) {
	panic("unexpected call: GetUnviewedMessagesForReminders")
}
func (s *stubDBClientForUpload) LogReminderSent(context.Context, *database.LogReminderRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("unexpected call: LogReminderSent")
}
func (s *stubDBClientForUpload) GetReminderHistory(context.Context, *database.GetReminderHistoryRequest, ...grpc.CallOption) (*database.GetReminderHistoryResponse, error) {
	panic("unexpected call: GetReminderHistory")
}

func (s *stubDBClientForUpload) CreateUploadSession(context.Context, *database.CreateUploadSessionRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.createErr
}
func (s *stubDBClientForUpload) GetUploadSession(context.Context, *database.GetUploadSessionRequest, ...grpc.CallOption) (*database.GetUploadSessionResponse, error) {
	return s.getResp, s.getErr
}
func (s *stubDBClientForUpload) AddCompletedPart(context.Context, *database.AddCompletedPartRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.addPartErr
}
func (s *stubDBClientForUpload) CompleteUploadSession(context.Context, *database.CompleteUploadSessionRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.completeErr
}
func (s *stubDBClientForUpload) DeleteUploadSession(context.Context, *database.DeleteUploadSessionRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.deleteErr
}
func (s *stubDBClientForUpload) DeleteExpiredUploadSessions(context.Context, *database.DeleteExpiredUploadSessionsRequest, ...grpc.CallOption) (*database.DeleteExpiredUploadSessionsResponse, error) {
	return s.expiredResp, s.expiredErr
}

// --- CreateSession ---

func TestUploadStateGRPCAdapter_CreateSession_Success(t *testing.T) {
	t.Parallel()
	a := &UploadStateGRPCAdapter{client: &stubDBClientForUpload{}}

	now := time.Now().UTC()
	err := a.CreateSession(context.Background(), contracts.FileUploadSession{
		SessionID: "s1", FileID: "f1",
		CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUploadStateGRPCAdapter_CreateSession_AlreadyExists_ReturnsErrUploadSessionAlreadyExists(t *testing.T) {
	t.Parallel()
	rpcErr := status.Error(codes.AlreadyExists, "session already exists")
	a := &UploadStateGRPCAdapter{client: &stubDBClientForUpload{createErr: rpcErr}}

	err := a.CreateSession(context.Background(), contracts.FileUploadSession{SessionID: "s1", FileID: "f1"})
	if !errors.Is(err, domain.ErrUploadSessionAlreadyExists) {
		t.Errorf("expected ErrUploadSessionAlreadyExists, got %v", err)
	}
}

// --- GetSessionByID ---

func TestUploadStateGRPCAdapter_GetSessionByID_ReturnsSession(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC().Truncate(time.Second)
	a := &UploadStateGRPCAdapter{
		client: &stubDBClientForUpload{
			getResp: &database.GetUploadSessionResponse{
				Session: &database.UploadSession{
					SessionId: "s1", FileId: "f1",
					Status:    "active",
					CreatedAt: now.Format(time.RFC3339),
					ExpiresAt: now.Add(time.Hour).Format(time.RFC3339),
				},
			},
		},
	}

	got, err := a.GetSessionByID(context.Background(), "s1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.SessionID != "s1" {
		t.Errorf("expected SessionID=s1, got %q", got.SessionID)
	}
}

func TestUploadStateGRPCAdapter_GetSessionByID_NotFound_ReturnsErrUploadSessionNotFound(t *testing.T) {
	t.Parallel()
	a := &UploadStateGRPCAdapter{
		client: &stubDBClientForUpload{getErr: status.Error(codes.NotFound, "upload session not found")},
	}

	_, err := a.GetSessionByID(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrUploadSessionNotFound) {
		t.Errorf("expected ErrUploadSessionNotFound, got %v", err)
	}
}

// --- GetSessionByFileID ---

func TestUploadStateGRPCAdapter_GetSessionByFileID_ReturnsSession(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC().Truncate(time.Second)
	a := &UploadStateGRPCAdapter{
		client: &stubDBClientForUpload{
			getResp: &database.GetUploadSessionResponse{
				Session: &database.UploadSession{
					SessionId: "s1", FileId: "f1",
					Status:    "active",
					CreatedAt: now.Format(time.RFC3339),
					ExpiresAt: now.Add(time.Hour).Format(time.RFC3339),
				},
			},
		},
	}

	got, err := a.GetSessionByFileID(context.Background(), "f1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.FileID != "f1" {
		t.Errorf("expected FileID=f1, got %q", got.FileID)
	}
}

// --- AddCompletedPart ---

func TestUploadStateGRPCAdapter_AddCompletedPart_Success(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC().Truncate(time.Second)
	a := &UploadStateGRPCAdapter{
		client: &stubDBClientForUpload{
			getResp: &database.GetUploadSessionResponse{
				Session: &database.UploadSession{
					SessionId: "s1", FileId: "f1",
					Status:    "active",
					CreatedAt: now.Format(time.RFC3339),
					ExpiresAt: now.Add(time.Hour).Format(time.RFC3339),
					CompletedParts: []*database.UploadSessionPart{
						{PartNumber: 1, Etag: "etag"},
					},
				},
			},
		},
	}

	parts, err := a.AddCompletedPart(context.Background(), "s1", contracts.CompletedPart{PartNumber: 1, ETag: "etag"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(parts) != 1 || parts[0].ETag != "etag" {
		t.Errorf("expected 1 part with ETag=etag, got %v", parts)
	}
}

// --- CompleteSession ---

func TestUploadStateGRPCAdapter_CompleteSession_Success(t *testing.T) {
	t.Parallel()
	a := &UploadStateGRPCAdapter{client: &stubDBClientForUpload{}}

	err := a.CompleteSession(context.Background(), "s1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUploadStateGRPCAdapter_CompleteSession_NotFound_ReturnsErrUploadSessionNotFound(t *testing.T) {
	t.Parallel()
	a := &UploadStateGRPCAdapter{
		client: &stubDBClientForUpload{completeErr: status.Error(codes.NotFound, "upload session not found")},
	}

	err := a.CompleteSession(context.Background(), "gone")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrUploadSessionNotFound) {
		t.Errorf("expected ErrUploadSessionNotFound, got %v", err)
	}
}

// --- DeleteSession ---

func TestUploadStateGRPCAdapter_DeleteSession_Success(t *testing.T) {
	t.Parallel()
	a := &UploadStateGRPCAdapter{client: &stubDBClientForUpload{}}

	err := a.DeleteSession(context.Background(), "s1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// --- DeleteExpiredSessions ---

func TestUploadStateGRPCAdapter_DeleteExpiredSessions_ReturnsSessions(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC().Truncate(time.Second)
	a := &UploadStateGRPCAdapter{
		client: &stubDBClientForUpload{
			expiredResp: &database.DeleteExpiredUploadSessionsResponse{
				RemovedSessions: []*database.UploadSession{
					{
						SessionId: "s-exp", FileId: "f-exp",
						Status:    "active",
						CreatedAt: now.Format(time.RFC3339),
						ExpiresAt: now.Format(time.RFC3339),
					},
				},
			},
		},
	}

	removed, err := a.DeleteExpiredSessions(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(removed) != 1 {
		t.Fatalf("expected 1 removed session, got %d", len(removed))
	}
	if removed[0].SessionID != "s-exp" {
		t.Errorf("expected SessionID=s-exp, got %q", removed[0].SessionID)
	}
}
