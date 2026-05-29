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
	storeErr  error
	healthErr error

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
}

func (s *stubStorageService) StoreMessage(context.Context, *contracts.Message) error {
	return s.storeErr
}
func (s *stubStorageService) RetrieveMessage(context.Context, string) (*contracts.Message, error) {
	return &contracts.Message{}, nil
}
func (s *stubStorageService) GetMessage(context.Context, string) (*contracts.Message, error) {
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
