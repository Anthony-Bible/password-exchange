package logger

import (
	"errors"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/secondary"
)

// TestNewAdapter_ReturnsNonNil verifies that NewAdapter returns a non-nil LoggerPort.
func TestNewAdapter_ReturnsNonNil(t *testing.T) {
	var adapter secondary.LoggerPort = NewAdapter()
	if adapter == nil {
		t.Fatal("NewAdapter() returned nil, expected non-nil secondary.LoggerPort")
	}
}

// TestAdapter_DebugReturnsLogEvent verifies Debug() returns a non-nil LogEvent.
func TestAdapter_DebugReturnsLogEvent(t *testing.T) {
	adapter := NewAdapter()
	var event contracts.LogEvent = adapter.Debug()
	if event == nil {
		t.Fatal("adapter.Debug() returned nil, expected non-nil contracts.LogEvent")
	}
}

// TestAdapter_InfoReturnsLogEvent verifies Info() returns a non-nil LogEvent.
func TestAdapter_InfoReturnsLogEvent(t *testing.T) {
	adapter := NewAdapter()
	var event contracts.LogEvent = adapter.Info()
	if event == nil {
		t.Fatal("adapter.Info() returned nil, expected non-nil contracts.LogEvent")
	}
}

// TestAdapter_WarnReturnsLogEvent verifies Warn() returns a non-nil LogEvent.
func TestAdapter_WarnReturnsLogEvent(t *testing.T) {
	adapter := NewAdapter()
	var event contracts.LogEvent = adapter.Warn()
	if event == nil {
		t.Fatal("adapter.Warn() returned nil, expected non-nil contracts.LogEvent")
	}
}

// TestAdapter_ErrorReturnsLogEvent verifies Error() returns a non-nil LogEvent.
func TestAdapter_ErrorReturnsLogEvent(t *testing.T) {
	adapter := NewAdapter()
	var event contracts.LogEvent = adapter.Error()
	if event == nil {
		t.Fatal("adapter.Error() returned nil, expected non-nil contracts.LogEvent")
	}
}

// TestAdapter_LogEventChain_DoesNotPanic exercises the full fluent chain and
// asserts no panic occurs. If any chained call panics, the test fails.
func TestAdapter_LogEventChain_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("fluent log event chain panicked: %v", r)
		}
	}()

	adapter := NewAdapter()
	adapter.Info().
		Str("k", "v").
		Int("n", 1).
		Bool("b", true).
		Err(nil).
		Dur("d", time.Millisecond).
		Float64("f", 1.5).
		Msg("ok")
}

// TestAdapter_LogEventChain_WithError ensures Err handles a real error
// without panicking when chained through the rest of the API.
func TestAdapter_LogEventChain_WithError(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("fluent log event chain panicked: %v", r)
		}
	}()

	adapter := NewAdapter()
	adapter.Error().
		Err(errors.New("boom")).
		Str("op", "test").
		Msg("error occurred")
}
