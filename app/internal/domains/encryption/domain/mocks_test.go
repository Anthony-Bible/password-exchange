package domain

import (
	"context"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/secondary"
)

// mockEvent is a no-op LogEvent used in tests. All chain methods return the
// receiver so fluent calls compile and execute without side-effects.
type mockEvent struct{}

func (m *mockEvent) Err(error) contracts.LogEvent              { return m }
func (m *mockEvent) Str(string, string) contracts.LogEvent     { return m }
func (m *mockEvent) Int(string, int) contracts.LogEvent        { return m }
func (m *mockEvent) Int32(string, int32) contracts.LogEvent    { return m }
func (m *mockEvent) Bool(string, bool) contracts.LogEvent      { return m }
func (m *mockEvent) Dur(string, time.Duration) contracts.LogEvent { return m }
func (m *mockEvent) Float64(string, float64) contracts.LogEvent { return m }
func (m *mockEvent) Msg(string)                                {}

// mockLogger implements secondary.LoggerPort for tests.
type mockLogger struct{}

func (m *mockLogger) Debug() contracts.LogEvent { return &mockEvent{} }
func (m *mockLogger) Info() contracts.LogEvent  { return &mockEvent{} }
func (m *mockLogger) Warn() contracts.LogEvent  { return &mockEvent{} }
func (m *mockLogger) Error() contracts.LogEvent { return &mockEvent{} }
func (m *mockLogger) Fatal() contracts.LogEvent { return &mockEvent{} }

// Ensure mockLogger satisfies the port interface at compile time.
var _ secondary.LoggerPort = (*mockLogger)(nil)

// mockKeyGen implements contracts.KeyGenerator with configurable behavior.
type mockKeyGen struct {
	GenerateKeyFunc func(ctx context.Context, length int32) (contracts.EncryptionKey, error)
	GenerateIDFunc  func(ctx context.Context) string
}

func (m *mockKeyGen) GenerateKey(ctx context.Context, length int32) (contracts.EncryptionKey, error) {
	if m.GenerateKeyFunc != nil {
		return m.GenerateKeyFunc(ctx, length)
	}
	var k contracts.EncryptionKey
	for i := range k {
		k[i] = 0xAA
	}
	return k, nil
}

func (m *mockKeyGen) GenerateID(ctx context.Context) string {
	if m.GenerateIDFunc != nil {
		return m.GenerateIDFunc(ctx)
	}
	return "fixed-id-123"
}

var _ contracts.KeyGenerator = (*mockKeyGen)(nil)
