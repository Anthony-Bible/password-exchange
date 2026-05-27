package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"
)

// mockLogEvent is a no-op LogEvent for tests
type mockLogEvent struct{}

func (m *mockLogEvent) Err(error) contracts.LogEvent              { return m }
func (m *mockLogEvent) Str(string, string) contracts.LogEvent     { return m }
func (m *mockLogEvent) Int(string, int) contracts.LogEvent        { return m }
func (m *mockLogEvent) Int32(string, int32) contracts.LogEvent    { return m }
func (m *mockLogEvent) Bool(string, bool) contracts.LogEvent      { return m }
func (m *mockLogEvent) Dur(string, time.Duration) contracts.LogEvent { return m }
func (m *mockLogEvent) Float64(string, float64) contracts.LogEvent   { return m }
func (m *mockLogEvent) Msg(string)                                {}

// mockLogger is a no-op LoggerPort for tests
type mockLogger struct{}

func (m *mockLogger) Debug() contracts.LogEvent { return &mockLogEvent{} }
func (m *mockLogger) Info() contracts.LogEvent  { return &mockLogEvent{} }
func (m *mockLogger) Warn() contracts.LogEvent  { return &mockLogEvent{} }
func (m *mockLogger) Error() contracts.LogEvent { return &mockLogEvent{} }
func (m *mockLogger) Fatal() contracts.LogEvent { return &mockLogEvent{} }

func TestGenerateKey_Valid32Bytes(t *testing.T) {
	g := NewKeyGenerator(&mockLogger{})
	key, err := g.GenerateKey(context.Background(), 32)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(key.Bytes()) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(key.Bytes()))
	}
	var zero contracts.EncryptionKey
	if key == zero {
		t.Fatal("generated key is all zero (vanishingly unlikely)")
	}
}

func TestGenerateKey_InvalidLength(t *testing.T) {
	g := NewKeyGenerator(&mockLogger{})
	_, err := g.GenerateKey(context.Background(), 16)
	if !errors.Is(err, domain.ErrInvalidKeyLength) {
		t.Fatalf("expected ErrInvalidKeyLength, got %v", err)
	}
}

func TestGenerateID_NonEmpty(t *testing.T) {
	g := NewKeyGenerator(&mockLogger{})
	id := g.GenerateID(context.Background())
	if id == "" {
		t.Fatal("expected non-empty id")
	}
}

func TestGenerateID_Unique(t *testing.T) {
	g := NewKeyGenerator(&mockLogger{})
	id1 := g.GenerateID(context.Background())
	id2 := g.GenerateID(context.Background())
	if id1 == id2 {
		t.Fatalf("expected unique ids, got %q twice", id1)
	}
}
