package grpc

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/secondary"
	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Insert_ExpiresAt tests exercise only the validation logic in GRPCServer.Insert.
// A nil storageService is safe because validation happens before StoreMessage is called.

func newServerForTest() *GRPCServer {
	return &GRPCServer{
		storageService: nil,
		logger:         &recordingLogger{},
		validator:      &stubValidator{},
	}
}

// recordingLogger captures log calls for assertion in tests.
type recordingLogger struct {
	mu     sync.Mutex
	events []*recordingEvent
}

func (l *recordingLogger) Debug() contracts.LogEvent { return l.start("debug") }
func (l *recordingLogger) Info() contracts.LogEvent  { return l.start("info") }
func (l *recordingLogger) Warn() contracts.LogEvent  { return l.start("warn") }
func (l *recordingLogger) Error() contracts.LogEvent { return l.start("error") }

func (l *recordingLogger) start(level string) contracts.LogEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := &recordingEvent{level: level, fields: map[string]interface{}{}}
	l.events = append(l.events, e)
	return e
}

func (l *recordingLogger) lastByLevel(level string) *recordingEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := len(l.events) - 1; i >= 0; i-- {
		if l.events[i].level == level {
			return l.events[i]
		}
	}
	return nil
}

type recordingEvent struct {
	level  string
	fields map[string]interface{}
	err    error
	msg    string
}

func (e *recordingEvent) Err(err error) contracts.LogEvent { e.err = err; return e }
func (e *recordingEvent) Str(k, v string) contracts.LogEvent {
	e.fields[k] = v
	return e
}
func (e *recordingEvent) Int(k string, v int) contracts.LogEvent     { e.fields[k] = v; return e }
func (e *recordingEvent) Int32(k string, v int32) contracts.LogEvent { e.fields[k] = v; return e }
func (e *recordingEvent) Int64(k string, v int64) contracts.LogEvent { e.fields[k] = v; return e }
func (e *recordingEvent) Bool(k string, v bool) contracts.LogEvent   { e.fields[k] = v; return e }
func (e *recordingEvent) Dur(k string, v time.Duration) contracts.LogEvent {
	e.fields[k] = v
	return e
}
func (e *recordingEvent) Float64(k string, v float64) contracts.LogEvent {
	e.fields[k] = v
	return e
}
func (e *recordingEvent) Msg(msg string) { e.msg = msg }

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
// down the success path so we can observe logger/validator routing.
type stubStorageService struct {
	storeErr error
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
func (s *stubStorageService) GetUnviewedMessagesForReminders(context.Context, int, int, int) ([]*contracts.UnviewedMessage, error) {
	return nil, nil
}
func (s *stubStorageService) LogReminderSent(context.Context, int, string) error { return nil }
func (s *stubStorageService) GetReminderHistory(context.Context, int) ([]*contracts.ReminderLogEntry, error) {
	return nil, nil
}
func (s *stubStorageService) CleanupExpiredMessages(context.Context) error { return nil }
func (s *stubStorageService) HealthCheck(context.Context) error            { return nil }

// TestInsert_MapsDomainValidationErrorsToInvalidArgument verifies that domain
// validation sentinels returned from StoreMessage surface to gRPC clients as
// codes.InvalidArgument instead of the default codes.Unknown. Clients that
// retry on Unknown would otherwise loop forever on a deterministic validation
// failure (e.g. MaxViewCount=0, empty content, nil message).
func TestInsert_MapsDomainValidationErrorsToInvalidArgument(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
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
			server := NewGRPCServer(svc, "127.0.0.1:0", &recordingLogger{}, &stubValidator{})

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
	logger := &recordingLogger{}
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

func TestNewGRPCServer_PanicsOnNilStorageService(t *testing.T) {
	defer func() {
		if r := recover(); r != "storage/grpc: NewGRPCServer requires a non-nil StorageServicePort" {
			t.Fatalf("expected nil storage service panic, got %v", r)
		}
	}()

	NewGRPCServer(nil, "127.0.0.1:0", &recordingLogger{}, &stubValidator{})
}

func TestNewGRPCServer_PanicsOnNilLogger(t *testing.T) {
	defer func() {
		if r := recover(); r != "storage/grpc: NewGRPCServer requires a non-nil LoggerPort" {
			t.Fatalf("expected nil logger panic, got %v", r)
		}
	}()

	NewGRPCServer(&stubStorageService{}, "127.0.0.1:0", nil, &stubValidator{})
}

func TestNewGRPCServer_PanicsOnNilValidator(t *testing.T) {
	defer func() {
		if r := recover(); r != "storage/grpc: NewGRPCServer requires a non-nil ValidationPort" {
			t.Fatalf("expected nil validator panic, got %v", r)
		}
	}()

	NewGRPCServer(&stubStorageService{}, "127.0.0.1:0", &recordingLogger{}, nil)
}

// TestInsert_RoutesLoggingAndValidationThroughPorts ensures Insert no longer
// reaches out to the shared logging/validation packages on the success path.
func TestInsert_RoutesLoggingAndValidationThroughPorts(t *testing.T) {
	logger := &recordingLogger{}
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

	infoEvent := logger.lastByLevel("info")
	if infoEvent == nil {
		t.Fatal("expected an info log event, got none")
	}
	if got := infoEvent.fields["uuid"]; got != "abc-123" {
		t.Errorf("expected uuid=abc-123 on info event, got %v", got)
	}
	if got := infoEvent.fields["recipientEmail"]; got != "SANITIZED(user@example.com)" {
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
