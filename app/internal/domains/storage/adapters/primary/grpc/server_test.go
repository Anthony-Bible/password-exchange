package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Insert_ExpiresAt tests exercise only the validation logic in GRPCServer.Insert.
// A nil storageService is safe because validation happens before StoreMessage is called.

func newServerForTest() *GRPCServer {
	return &GRPCServer{
		storageService: nil,
		logger:         logtest.NewRecorder(),
		validator:      &stubValidator{},
	}
}

// lastEntryByLevel returns the most recent recorded entry at the given level,
// or nil if the recorder captured none.
func lastEntryByLevel(rec *logtest.Recorder, level string) *logtest.Entry {
	entries := rec.Entries()
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Level == level {
			e := entries[i]
			return &e
		}
	}
	return nil
}

// stubValidator returns deterministic sanitized values so tests can assert routing.
type stubValidator struct {
	sanitizeCalls []string
}

func (v *stubValidator) ValidateEmail(string) error { return nil }
func (v *stubValidator) SanitizeEmailForLogging(email string) string {
	v.sanitizeCalls = append(v.sanitizeCalls, email)
	return "SANITIZED(" + email + ")"
}

// stubStorageService is a minimal storage service used to drive the gRPC adapter
// down the success path so we can observe logger/validator routing. The reminder
// fields let integration tests feed canned results/errors and capture the
// arguments the gRPC server forwards after protobuf decoding.
type stubStorageService struct {
	storeErr    error
	retrieveErr error
	getErr      error
	healthErr   error
	lastStored  *contracts.Message
	selectMsg   *contracts.Message

	// Reminder canned responses / errors.
	unviewedMessages []*contracts.UnviewedMessage
	unviewedErr      error
	logReminderErr   error
	reminderHistory  []*contracts.ReminderLogEntry
	reminderHistErr  error

	// Captured arguments from the most recent reminder calls.
	gotOlderThanHours, gotMaxReminders, gotIntervalHours int
	gotLogMessageID                                      int
	gotLogEmail                                          string
	gotHistoryMessageID                                  int

	// Upload-session canned responses and captured inputs.
	uploadCreateErr    error
	uploadGetErr       error
	uploadGetSession   *contracts.UploadSession
	uploadAddPartErr   error
	uploadCompleteErr  error
	uploadDeleteErr    error
	uploadExpiredErr   error
	expiredSessions    []contracts.UploadSession
	lastCreatedSession *contracts.UploadSession
}

func (s *stubStorageService) StoreMessage(_ context.Context, msg *contracts.Message) error {
	s.lastStored = &contracts.Message{}
	*s.lastStored = *msg
	return s.storeErr
}
func (s *stubStorageService) RetrieveMessage(context.Context, string) (*contracts.Message, error) {
	if s.retrieveErr != nil {
		return nil, s.retrieveErr
	}
	if s.selectMsg != nil {
		return s.selectMsg, nil
	}
	return &contracts.Message{}, nil
}
func (s *stubStorageService) GetMessage(context.Context, string) (*contracts.Message, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.selectMsg != nil {
		return s.selectMsg, nil
	}
	return &contracts.Message{}, nil
}
func (s *stubStorageService) GetUnviewedMessagesForReminders(_ context.Context, olderThanHours, maxReminders, intervalHours int) ([]*contracts.UnviewedMessage, error) {
	s.gotOlderThanHours, s.gotMaxReminders, s.gotIntervalHours = olderThanHours, maxReminders, intervalHours
	return s.unviewedMessages, s.unviewedErr
}
func (s *stubStorageService) LogReminderSent(_ context.Context, messageID int, email string) error {
	s.gotLogMessageID, s.gotLogEmail = messageID, email
	return s.logReminderErr
}
func (s *stubStorageService) GetReminderHistory(_ context.Context, messageID int) ([]*contracts.ReminderLogEntry, error) {
	s.gotHistoryMessageID = messageID
	return s.reminderHistory, s.reminderHistErr
}
func (s *stubStorageService) CleanupExpiredMessages(context.Context) error { return nil }
func (s *stubStorageService) HealthCheck(context.Context) error            { return s.healthErr }

// Upload session stubs — canned responses for server-handler tests.
func (s *stubStorageService) CreateUploadSession(_ context.Context, sess contracts.UploadSession) error {
	if s.uploadCreateErr != nil {
		return s.uploadCreateErr
	}
	s.lastCreatedSession = &sess
	return nil
}
func (s *stubStorageService) GetUploadSession(_ context.Context, id string) (*contracts.UploadSession, error) {
	if s.uploadGetErr != nil {
		return nil, s.uploadGetErr
	}
	if s.uploadGetSession != nil {
		return s.uploadGetSession, nil
	}
	return &contracts.UploadSession{SessionID: id}, nil
}
func (s *stubStorageService) AddCompletedPart(_ context.Context, _ string, _ contracts.UploadSessionPart) error {
	return s.uploadAddPartErr
}
func (s *stubStorageService) CompleteUploadSession(_ context.Context, _ string) error {
	return s.uploadCompleteErr
}
func (s *stubStorageService) DeleteUploadSession(_ context.Context, _ string) error {
	return s.uploadDeleteErr
}
func (s *stubStorageService) DeleteExpiredUploadSessions(_ context.Context, _ time.Time) ([]contracts.UploadSession, error) {
	return s.expiredSessions, s.uploadExpiredErr
}

// TestInsert_MapsDomainValidationErrorsToInvalidArgument verifies that domain
// validation sentinels returned from StoreMessage surface to gRPC clients as
// codes.InvalidArgument instead of the default codes.Unknown. Clients that
// retry on Unknown would otherwise loop forever on a deterministic validation
// failure (e.g. MaxViewCount=0, empty content, nil message).
func TestInsert_MapsDomainValidationErrorsToInvalidArgument(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		storeErr error
	}{
		{name: "nil message", storeErr: domain.ErrNilMessage},
		{name: "empty content", storeErr: domain.ErrEmptyContent},
		{name: "empty unique id", storeErr: domain.ErrEmptyUniqueID},
		{name: "invalid max view count", storeErr: domain.ErrInvalidMaxViewCount},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			svc := &stubStorageService{storeErr: tc.storeErr}
			server := NewGRPCServer(svc, "127.0.0.1:0", logtest.NewRecorder(), &stubValidator{})

			_, err := server.Insert(context.Background(), &database.InsertRequest{
				Uuid:         "abc-123",
				Content:      "ciphertext",
				MaxViewCount: 3,
			})
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected gRPC status error, got %v", err)
			}
			if got := st.Code(); got != codes.InvalidArgument {
				t.Errorf("expected codes.InvalidArgument, got %v", got)
			}
		})
	}
}

// TestRetrievalAndReminderRPCs_MapDomainErrorsToStatus verifies that every
// read/reminder RPC routes domain sentinels through domainErrorToStatus, not
// just Insert. Before this was fixed only Insert mapped errors, so a validation
// or not-found error from Select/GetMessage/GetUnviewedMessagesForReminders/
// LogReminderSent/GetReminderHistory reached clients as codes.Unknown and
// retry-on-Unknown clients looped forever on deterministic failures.
func TestRetrievalAndReminderRPCs_MapDomainErrorsToStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		want    codes.Code
		newStub func() *stubStorageService
		callRPC func(*GRPCServer) error
	}{
		{
			name:    "Select not found",
			want:    codes.NotFound,
			newStub: func() *stubStorageService { return &stubStorageService{retrieveErr: domain.ErrMessageNotFound} },
			callRPC: func(s *GRPCServer) error {
				_, err := s.Select(context.Background(), &database.SelectRequest{Uuid: "abc"})
				return err
			},
		},
		{
			name:    "GetMessage empty unique id",
			want:    codes.InvalidArgument,
			newStub: func() *stubStorageService { return &stubStorageService{getErr: domain.ErrEmptyUniqueID} },
			callRPC: func(s *GRPCServer) error {
				_, err := s.GetMessage(context.Background(), &database.SelectRequest{Uuid: ""})
				return err
			},
		},
		{
			name:    "GetUnviewedMessagesForReminders invalid parameter",
			want:    codes.InvalidArgument,
			newStub: func() *stubStorageService { return &stubStorageService{unviewedErr: domain.ErrInvalidParameter} },
			callRPC: func(s *GRPCServer) error {
				_, err := s.GetUnviewedMessagesForReminders(context.Background(), &database.GetUnviewedMessagesRequest{})
				return err
			},
		},
		{
			name:    "LogReminderSent empty email",
			want:    codes.InvalidArgument,
			newStub: func() *stubStorageService { return &stubStorageService{logReminderErr: domain.ErrEmptyEmailAddress} },
			callRPC: func(s *GRPCServer) error {
				_, err := s.LogReminderSent(context.Background(), &database.LogReminderRequest{MessageId: 1})
				return err
			},
		},
		{
			name:    "GetReminderHistory invalid parameter",
			want:    codes.InvalidArgument,
			newStub: func() *stubStorageService { return &stubStorageService{reminderHistErr: domain.ErrInvalidParameter} },
			callRPC: func(s *GRPCServer) error {
				_, err := s.GetReminderHistory(context.Background(), &database.GetReminderHistoryRequest{})
				return err
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := NewGRPCServer(tc.newStub(), "127.0.0.1:0", logtest.NewRecorder(), &stubValidator{})

			err := tc.callRPC(server)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected gRPC status error, got %v", err)
			}
			if got := st.Code(); got != tc.want {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// TestNewGRPCServer_InjectsPorts verifies the constructor accepts and stores
// the LoggerPort and ValidationPort dependencies.
func TestNewGRPCServer_InjectsPorts(t *testing.T) {
	logger := logtest.NewRecorder()
	validator := &stubValidator{}
	svc := &stubStorageService{}

	server := NewGRPCServer(svc, "127.0.0.1:0", logger, validator)

	if server.logger == nil {
		t.Fatal("expected logger to be wired, got nil")
	}
	if server.validator == nil {
		t.Fatal("expected validator to be wired, got nil")
	}
	var _ secondary.LoggerPort = server.logger
	var _ secondary.ValidationPort = server.validator
}

// TestInsert_RoutesLoggingAndValidationThroughPorts ensures Insert no longer
// reaches out to the shared logging/validation packages on the success path.
func TestInsert_RoutesLoggingAndValidationThroughPorts(t *testing.T) {
	logger := logtest.NewRecorder()
	validator := &stubValidator{}
	svc := &stubStorageService{}
	server := NewGRPCServer(svc, "127.0.0.1:0", logger, validator)

	_, err := server.Insert(context.Background(), &database.InsertRequest{
		Uuid:           "abc-123",
		Content:        "ciphertext",
		RecipientEmail: "user@example.com",
		MaxViewCount:   3,
	})
	if err != nil {
		t.Fatalf("Insert returned unexpected error: %v", err)
	}

	infoEvent := lastEntryByLevel(logger, "info")
	if infoEvent == nil {
		t.Fatal("expected an info log event, got none")
	}
	if got := infoEvent.Fields["uuid"]; got != "abc-123" {
		t.Errorf("expected uuid=abc-123 on info event, got %v", got)
	}
	if got := infoEvent.Fields["recipientEmail"]; got != "SANITIZED(user@example.com)" {
		t.Errorf("expected sanitized recipientEmail on info event, got %v", got)
	}
	if len(validator.sanitizeCalls) != 1 || validator.sanitizeCalls[0] != "user@example.com" {
		t.Errorf("expected validator.SanitizeEmailForLogging to be called once with user@example.com, got %v", validator.sanitizeCalls)
	}
}

func TestInsert_ExpiresAt_PastTimestamp(t *testing.T) {
	s := newServerForTest()

	past := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	_, err := s.Insert(context.Background(), &database.InsertRequest{
		ExpiresAt: past,
	})

	if err == nil {
		t.Fatal("expected error for past expires_at, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
	if st.Message() != "expires_at must be in the future" {
		t.Errorf("unexpected message: %q", st.Message())
	}
}

func TestInsert_ExpiresAt_ExceedsMaximum(t *testing.T) {
	s := newServerForTest()

	tooFar := time.Now().Add(maxExpirationDuration + 48*time.Hour).UTC().Format(time.RFC3339)
	_, err := s.Insert(context.Background(), &database.InsertRequest{
		ExpiresAt: tooFar,
	})

	if err == nil {
		t.Fatal("expected error for expires_at beyond max, got nil")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
	if st.Message() != "expires_at exceeds maximum allowed expiration" {
		t.Errorf("unexpected message: %q", st.Message())
	}
}

func TestInsert_ExpiresAt_InvalidFormat(t *testing.T) {
	s := newServerForTest()

	_, err := s.Insert(context.Background(), &database.InsertRequest{
		ExpiresAt: "not-a-timestamp",
	})

	if err == nil {
		t.Fatal("expected error for invalid expires_at format, got nil")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestParseExpiresAt_EmptyString(t *testing.T) {
	result, err := parseExpiresAt("")
	if err != nil {
		t.Fatalf("expected nil error for empty string, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for empty string, got %v", result)
	}
}

func TestParseExpiresAt_ValidRFC3339(t *testing.T) {
	input := "2099-01-01T00:00:00Z"
	result, err := parseExpiresAt(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	expected, _ := time.Parse(time.RFC3339, input)
	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, *result)
	}
}

// --- Upload session gRPC handler tests ---

func TestCreateUploadSession_ReturnsEmptyOnSuccess(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	svc := &stubStorageService{}
	server := NewGRPCServer(svc, "127.0.0.1:0", logtest.NewRecorder(), &stubValidator{})

	_, err := server.CreateUploadSession(context.Background(), &database.CreateUploadSessionRequest{
		Session: &database.UploadSession{
			SessionId:   "sess-1",
			FileId:      "file-1",
			MessageId:   "msg-1",
			UploadId:    "upload-1",
			Filename:    "test.txt",
			ContentType: "text/plain",
			TotalSize:   1024,
			TotalChunks: 2,
			Status:      "active",
			CreatedAt:   now.Format(time.RFC3339),
			ExpiresAt:   now.Add(time.Hour).Format(time.RFC3339),
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if svc.lastCreatedSession == nil {
		t.Fatal("expected domain CreateUploadSession to be called")
	}
	if svc.lastCreatedSession.SessionID != "sess-1" {
		t.Errorf("expected SessionID=sess-1, got %q", svc.lastCreatedSession.SessionID)
	}
}

func TestCreateUploadSession_MissingSession_ReturnsInvalidArgument(t *testing.T) {
	t.Parallel()
	server := NewGRPCServer(&stubStorageService{}, "127.0.0.1:0", logtest.NewRecorder(), &stubValidator{})

	_, err := server.CreateUploadSession(context.Background(), &database.CreateUploadSessionRequest{Session: nil})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, _ := status.FromError(err)
	if got := st.Code(); got != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", got)
	}
}

func TestGetUploadSession_ReturnsSession(t *testing.T) {
	t.Parallel()
	svc := &stubStorageService{
		uploadGetSession: &contracts.UploadSession{SessionID: "sess-1", FileID: "file-1"},
	}
	server := NewGRPCServer(svc, "127.0.0.1:0", logtest.NewRecorder(), &stubValidator{})

	resp, err := server.GetUploadSession(context.Background(), &database.GetUploadSessionRequest{
		SessionId: "sess-1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.GetSession().GetSessionId() != "sess-1" {
		t.Errorf("expected SessionId=sess-1, got %q", resp.GetSession().GetSessionId())
	}
}

func TestGetUploadSession_EmptyID_ReturnsInvalidArgument(t *testing.T) {
	t.Parallel()
	server := NewGRPCServer(&stubStorageService{}, "127.0.0.1:0", logtest.NewRecorder(), &stubValidator{})

	_, err := server.GetUploadSession(context.Background(), &database.GetUploadSessionRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, _ := status.FromError(err)
	if got := st.Code(); got != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", got)
	}
}

func TestGetUploadSession_NotFound_ReturnsNotFound(t *testing.T) {
	t.Parallel()
	svc := &stubStorageService{uploadGetErr: domain.ErrUploadSessionNotFound}
	server := NewGRPCServer(svc, "127.0.0.1:0", logtest.NewRecorder(), &stubValidator{})

	_, err := server.GetUploadSession(context.Background(), &database.GetUploadSessionRequest{SessionId: "missing"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, _ := status.FromError(err)
	if got := st.Code(); got != codes.NotFound {
		t.Errorf("expected NotFound, got %v", got)
	}
}

func TestCompleteUploadSession_NotFound_ReturnsNotFound(t *testing.T) {
	t.Parallel()
	svc := &stubStorageService{uploadCompleteErr: domain.ErrUploadSessionNotFound}
	server := NewGRPCServer(svc, "127.0.0.1:0", logtest.NewRecorder(), &stubValidator{})

	_, err := server.CompleteUploadSession(context.Background(), &database.CompleteUploadSessionRequest{SessionId: "gone"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, _ := status.FromError(err)
	if got := st.Code(); got != codes.NotFound {
		t.Errorf("expected NotFound, got %v", got)
	}
}

func TestDeleteExpiredUploadSessions_ReturnsRemovedSessions(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	svc := &stubStorageService{
		expiredSessions: []contracts.UploadSession{
			{SessionID: "expired-1", FileID: "file-exp-1"},
		},
	}
	server := NewGRPCServer(svc, "127.0.0.1:0", logtest.NewRecorder(), &stubValidator{})

	resp, err := server.DeleteExpiredUploadSessions(context.Background(), &database.DeleteExpiredUploadSessionsRequest{
		AsOf: now.Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.GetRemovedSessions()) != 1 {
		t.Errorf("expected 1 removed session, got %d", len(resp.GetRemovedSessions()))
	}
	if resp.GetRemovedSessions()[0].GetSessionId() != "expired-1" {
		t.Errorf("expected session_id=expired-1, got %q", resp.GetRemovedSessions()[0].GetSessionId())
	}
}

func TestInsert_MapsClientEncryptedFlagIntoDomainContract(t *testing.T) {
	t.Parallel()

	svc := &stubStorageService{}
	server := NewGRPCServer(svc, "127.0.0.1:0", logtest.NewRecorder(), &stubValidator{})

	req := &database.InsertRequest{
		Uuid:              "abc-123",
		Content:           "ciphertext",
		MaxViewCount:      3,
		IsClientEncrypted: true,
	}

	_, err := server.Insert(context.Background(), req)
	if err != nil {
		t.Fatalf("Insert returned unexpected error: %v", err)
	}
	if svc.lastStored == nil {
		t.Fatal("expected StoreMessage to be called")
	}
	if !svc.lastStored.IsClientEncrypted {
		t.Fatal("expected IsClientEncrypted=true in stored message contract")
	}
}

func TestSelect_MapsClientEncryptedFlagIntoProtoResponse(t *testing.T) {
	t.Parallel()

	svc := &stubStorageService{
		selectMsg: &contracts.Message{
			UniqueID:          "abc-123",
			Content:           "ciphertext",
			IsClientEncrypted: true,
		},
	}

	server := NewGRPCServer(svc, "127.0.0.1:0", logtest.NewRecorder(), &stubValidator{})
	resp, err := server.Select(context.Background(), &database.SelectRequest{Uuid: "abc-123"})
	if err != nil {
		t.Fatalf("Select returned unexpected error: %v", err)
	}
	if !resp.IsClientEncrypted {
		t.Fatal("expected IsClientEncrypted=true in gRPC response")
	}
}
