package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// mockStorageService is a minimal stub implementing primary.StorageServicePort for testing.
type mockStorageService struct {
	storeErr       error
	healthCheckErr error
}

func (m *mockStorageService) StoreMessage(ctx context.Context, msg *domain.Message) error {
	return m.storeErr
}

func (m *mockStorageService) RetrieveMessage(ctx context.Context, uniqueID string) (*domain.Message, error) {
	return nil, nil
}

func (m *mockStorageService) GetMessage(ctx context.Context, uniqueID string) (*domain.Message, error) {
	return nil, nil
}

func (m *mockStorageService) CleanupExpiredMessages(ctx context.Context) error {
	return nil
}

func (m *mockStorageService) GetUnviewedMessagesForReminders(ctx context.Context, olderThanHours, maxReminders, reminderIntervalHours int) ([]*domain.UnviewedMessage, error) {
	return nil, nil
}

func (m *mockStorageService) LogReminderSent(ctx context.Context, messageID int, emailAddress string) error {
	return nil
}

func (m *mockStorageService) GetReminderHistory(ctx context.Context, messageID int) ([]*domain.ReminderLogEntry, error) {
	return nil, nil
}

func (m *mockStorageService) HealthCheck(ctx context.Context) error {
	return m.healthCheckErr
}

func newServerForTest() *GRPCServer {
	return &GRPCServer{storageService: nil}
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

func TestPing(t *testing.T) {
	tests := []struct {
		name           string
		healthCheckErr error
		wantCode       codes.Code
		wantMsg        string
	}{
		{
			name:           "success",
			healthCheckErr: nil,
			wantCode:       codes.OK,
			wantMsg:        "",
		},
		{
			name:           "unhealthy - generic error message",
			healthCheckErr: errors.New("database connection refused: internal detail 123"),
			wantCode:       codes.Unavailable,
			wantMsg:        "storage service unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStorageService{healthCheckErr: tt.healthCheckErr}
			s := &GRPCServer{storageService: mock}

			_, err := s.Ping(context.Background(), &emptypb.Empty{})

			if tt.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected gRPC status error, got %v", err)
			}

			if st.Code() != tt.wantCode {
				t.Errorf("expected code %v, got %v", tt.wantCode, st.Code())
			}

			if st.Message() != tt.wantMsg {
				t.Errorf("expected message %q, got %q", tt.wantMsg, st.Message())
			}
		})
	}
}
