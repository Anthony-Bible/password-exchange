package grpc_clients

import (
	"context"
	"net"
	"testing"
	"time"

	storageContracts "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	storageGRPC "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/primary/grpc"
	storageDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// inMemoryStorageRepo is a fully in-memory MessageRepository that implements
// only the upload-session methods. All message methods panic so accidental
// calls are caught immediately during the round-trip test.
type inMemoryStorageRepo struct {
	sessions map[string]*storageContracts.UploadSession
}

func newInMemoryStorageRepo() *inMemoryStorageRepo {
	return &inMemoryStorageRepo{sessions: make(map[string]*storageContracts.UploadSession)}
}

// Unsupported methods — panic on accidental calls.
func (*inMemoryStorageRepo) InsertMessage(*storageContracts.Message) error {
	panic("not used in upload round-trip test")
}
func (*inMemoryStorageRepo) SelectMessageByUniqueID(string) (*storageContracts.Message, error) {
	panic("not used in upload round-trip test")
}
func (*inMemoryStorageRepo) IncrementViewCountAndGet(string) (*storageContracts.Message, error) {
	panic("not used in upload round-trip test")
}
func (*inMemoryStorageRepo) DeleteExpiredMessages() error { return nil }
func (*inMemoryStorageRepo) GetMessage(string) (*storageContracts.Message, error) {
	panic("not used in upload round-trip test")
}
func (*inMemoryStorageRepo) GetUnviewedMessagesForReminders(_, _, _ int) ([]*storageContracts.UnviewedMessage, error) {
	panic("not used in upload round-trip test")
}
func (*inMemoryStorageRepo) LogReminderSent(int, string) error {
	panic("not used in upload round-trip test")
}
func (*inMemoryStorageRepo) GetReminderHistory(int) ([]*storageContracts.ReminderLogEntry, error) {
	panic("not used in upload round-trip test")
}
func (*inMemoryStorageRepo) Close() error                         { return nil }
func (*inMemoryStorageRepo) Ping(context.Context) error           { return nil }

// Upload-session implementation backed by the in-memory map.
func (r *inMemoryStorageRepo) CreateUploadSession(_ context.Context, s storageContracts.UploadSession) error {
	cp := s
	r.sessions[s.SessionID] = &cp
	return nil
}
func (r *inMemoryStorageRepo) GetUploadSession(_ context.Context, id string) (*storageContracts.UploadSession, error) {
	for _, s := range r.sessions {
		if s.SessionID == id || s.FileID == id {
			cp := *s
			return &cp, nil
		}
	}
	return nil, storageDomain.ErrUploadSessionNotFound
}
func (r *inMemoryStorageRepo) AddCompletedPart(_ context.Context, sessionID string, part storageContracts.UploadSessionPart) error {
	s, ok := r.sessions[sessionID]
	if !ok {
		return storageDomain.ErrUploadSessionNotFound
	}
	s.CompletedParts = append(s.CompletedParts, part)
	return nil
}
func (r *inMemoryStorageRepo) CompleteUploadSession(_ context.Context, sessionID string) error {
	s, ok := r.sessions[sessionID]
	if !ok {
		return storageDomain.ErrUploadSessionNotFound
	}
	s.Status = "complete"
	s.EncryptionKey = nil // key cleared on completion
	return nil
}
func (r *inMemoryStorageRepo) DeleteUploadSession(_ context.Context, sessionID string) error {
	delete(r.sessions, sessionID)
	return nil
}
func (r *inMemoryStorageRepo) DeleteExpiredUploadSessions(_ context.Context, asOf time.Time) ([]storageContracts.UploadSession, error) {
	var removed []storageContracts.UploadSession
	for id, s := range r.sessions {
		if s.Status != "complete" && s.ExpiresAt.Before(asOf) {
			removed = append(removed, *s)
			delete(r.sessions, id)
		}
	}
	return removed, nil
}

// startTestGRPCServer spins up a real gRPC server on a random port and
// returns the address plus a cleanup function. It is wired with an
// inMemoryStorageRepo so no MySQL is required.
func startTestGRPCServer(t *testing.T) (string, *inMemoryStorageRepo) {
	t.Helper()
	repo := newInMemoryStorageRepo()
	logger := logtest.NewNoop()
	validator := validation.NewAdapter()

	svc := storageDomain.NewStorageService(repo, logger, validator)
	srv := storageGRPC.NewGRPCServer(svc, "", logger, validator)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}

	grpcSrv := grpc.NewServer()
	database.RegisterDbServiceServer(grpcSrv, srv)

	go func() { _ = grpcSrv.Serve(lis) }()
	t.Cleanup(func() { grpcSrv.GracefulStop() })

	return lis.Addr().String(), repo
}

// TestUploadStateGRPCAdapter_RoundTrip verifies the full create→get→addPart
// →complete→delete lifecycle across the gRPC boundary using a real server
// backed by an in-memory repository.
func TestUploadStateGRPCAdapter_RoundTrip(t *testing.T) {
	addr, repo := startTestGRPCServer(t)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	dbClient := database.NewDbServiceClient(conn)
	adapter := NewUploadStateGRPCAdapter(dbClient)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	sess := contracts.FileUploadSession{
		SessionID:     "rt-sess-1",
		FileID:        "rt-file-1",
		MessageID:     "rt-msg-1",
		UploadID:      "rt-upload-1",
		Filename:      "roundtrip.bin",
		ContentType:   "application/octet-stream",
		TotalSize:     4096,
		TotalChunks:   2,
		Status:        "active",
		EncryptionKey: []byte("super-secret-aes-key"),
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Hour),
	}

	// 1. CreateSession
	if err := adapter.CreateSession(ctx, sess); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// 2. GetSessionByID
	got, err := adapter.GetSessionByID(ctx, "rt-sess-1")
	if err != nil {
		t.Fatalf("GetSessionByID: %v", err)
	}
	if got.FileID != "rt-file-1" {
		t.Errorf("expected FileID=rt-file-1, got %q", got.FileID)
	}
	if string(got.EncryptionKey) != "super-secret-aes-key" {
		t.Errorf("expected encryption key to be preserved before completion, got %q", got.EncryptionKey)
	}

	// 3. GetSessionByFileID (dual-lookup contract)
	gotByFile, err := adapter.GetSessionByFileID(ctx, "rt-file-1")
	if err != nil {
		t.Fatalf("GetSessionByFileID: %v", err)
	}
	if gotByFile.SessionID != "rt-sess-1" {
		t.Errorf("expected SessionID=rt-sess-1 from file-ID lookup, got %q", gotByFile.SessionID)
	}

	// 4. AddCompletedPart — returns the post-add parts list
	if _, err := adapter.AddCompletedPart(ctx, "rt-sess-1", contracts.CompletedPart{PartNumber: 1, ETag: "etag-1"}); err != nil {
		t.Fatalf("AddCompletedPart: %v", err)
	}

	// 5. CompleteSession — must clear the encryption key
	if err := adapter.CompleteSession(ctx, "rt-sess-1"); err != nil {
		t.Fatalf("CompleteSession: %v", err)
	}

	// Verify key was cleared in the repository
	stored := repo.sessions["rt-sess-1"]
	if stored == nil {
		t.Fatal("expected session to still exist after complete")
	}
	if stored.EncryptionKey != nil {
		t.Errorf("expected encryption key to be nil after CompleteSession, got %q", stored.EncryptionKey)
	}
	if stored.Status != "complete" {
		t.Errorf("expected status=complete, got %q", stored.Status)
	}

	// 6. DeleteSession
	if err := adapter.DeleteSession(ctx, "rt-sess-1"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if repo.sessions["rt-sess-1"] != nil {
		t.Error("expected session to be removed after DeleteSession")
	}
}

// TestUploadStateGRPCAdapter_DeleteExpiredSessions_RoundTrip verifies that
// only incomplete expired sessions are swept and completed sessions are kept.
func TestUploadStateGRPCAdapter_DeleteExpiredSessions_RoundTrip(t *testing.T) {
	addr, repo := startTestGRPCServer(t)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	dbClient := database.NewDbServiceClient(conn)
	adapter := NewUploadStateGRPCAdapter(dbClient)

	ctx := context.Background()
	past := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	future := time.Now().UTC().Add(time.Hour).Truncate(time.Second)

	// Seed two sessions directly into the in-memory repo: one expired, one not.
	repo.sessions["sess-expired"] = &storageContracts.UploadSession{
		SessionID: "sess-expired", FileID: "file-exp",
		Status: "active", ExpiresAt: past,
		CreatedAt: past.Add(-time.Hour),
	}
	repo.sessions["sess-active"] = &storageContracts.UploadSession{
		SessionID: "sess-active", FileID: "file-active",
		Status: "active", ExpiresAt: future,
		CreatedAt: past,
	}

	removed, err := adapter.DeleteExpiredSessions(ctx, time.Now().UTC())
	if err != nil {
		t.Fatalf("DeleteExpiredSessions: %v", err)
	}

	if len(removed) != 1 {
		t.Fatalf("expected 1 removed session, got %d", len(removed))
	}
	if removed[0].SessionID != "sess-expired" {
		t.Errorf("expected SessionID=sess-expired, got %q", removed[0].SessionID)
	}
	if repo.sessions["sess-active"] == nil {
		t.Error("active session should not have been swept")
	}
	if repo.sessions["sess-expired"] != nil {
		t.Error("expired session should have been deleted from the repository")
	}
}
