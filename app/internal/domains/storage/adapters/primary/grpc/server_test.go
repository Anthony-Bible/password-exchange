package grpc

import (
	"context"
	"testing"
	"time"

	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockStorageService is a minimal stub implementing primary.StorageServicePort for testing.
type mockStorageService struct {
	storeErr error
}

func (m *mockStorageService) StoreMessage(ctx context.Context, msg interface{}) error {
	return m.storeErr
}

// Insert_ExpiresAt tests exercise only the validation logic in GRPCServer.Insert.
// A nil storageService is safe because validation happens before StoreMessage is called.

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
