package logging

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	extlogging "github.com/Anthony-Bible/Logging"
)

type recordedEntry struct {
	level extlogging.Level
	msg   string
	attrs []slog.Attr
}

type recordingBackend struct {
	entries []recordedEntry
}

func (r *recordingBackend) Log(_ context.Context, level extlogging.Level, msg string, attrs ...slog.Attr) {
	r.entries = append(r.entries, recordedEntry{level: level, msg: msg, attrs: append([]slog.Attr(nil), attrs...)})
}

// installRecorder swaps the package logger for one backed by a recorder so
// tests can inspect emitted records. It returns the recorder plus a cleanup.
func installRecorder(t *testing.T, level extlogging.Level) *recordingBackend {
	t.Helper()
	original := logger
	rec := &recordingBackend{}
	logger = extlogging.New(extlogging.Config{
		Backend:               rec,
		Level:                 level,
		ErrorThreshold:        5,
		ErrorWindow:           time.Minute,
		DebugDuration:         5 * time.Minute,
		DisableSignalHandling: true,
	})
	t.Cleanup(func() { logger = original })
	return rec
}

func findAttr(attrs []slog.Attr, key string) (slog.Attr, bool) {
	for _, a := range attrs {
		if a.Key == key {
			return a, true
		}
	}
	return slog.Attr{}, false
}

func TestEventLogsStructuredFields(t *testing.T) {
	rec := installRecorder(t, extlogging.LevelDebug)

	Error().
		Str("component", "unit-test").
		Int("attempt", 2).
		Err(errors.New("boom")).
		Msg("log message")

	if len(rec.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(rec.entries))
	}
	entry := rec.entries[0]
	if entry.level != extlogging.LevelError {
		t.Errorf("expected LevelError, got %v", entry.level)
	}
	if entry.msg != "log message" {
		t.Errorf("expected msg 'log message', got %q", entry.msg)
	}
	if a, ok := findAttr(entry.attrs, "component"); !ok || a.Value.String() != "unit-test" {
		t.Errorf("expected component=unit-test, got %+v", a)
	}
	if a, ok := findAttr(entry.attrs, "attempt"); !ok || a.Value.Int64() != 2 {
		t.Errorf("expected attempt=2, got %+v", a)
	}
	if _, ok := findAttr(entry.attrs, "error"); !ok {
		t.Errorf("expected error attribute")
	}
}

func TestAllBuilderMethods(t *testing.T) {
	rec := installRecorder(t, extlogging.LevelDebug)

	Info().
		Str("s", "v").
		Int("i", 1).
		Int32("i32", 32).
		Int64("i64", 64).
		Bool("b", true).
		Dur("d", 2*time.Second).
		Float64("f", 1.5).
		Interface("any", map[string]int{"x": 1}).
		Msg("all fields")

	if len(rec.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(rec.entries))
	}
	for _, key := range []string{"s", "i", "i32", "i64", "b", "d", "f", "any"} {
		if _, ok := findAttr(rec.entries[0].attrs, key); !ok {
			t.Errorf("missing attr %q", key)
		}
	}
}

func TestMsgfAndLevels(t *testing.T) {
	rec := installRecorder(t, extlogging.LevelDebug)

	Debug().Msgf("message %d", 1)
	Info().Msg("info")
	Warn().Msg("warn")
	Error().Msg("err")

	if len(rec.entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(rec.entries))
	}
	if rec.entries[0].msg != "message 1" || rec.entries[0].level != extlogging.LevelDebug {
		t.Errorf("debug entry wrong: %+v", rec.entries[0])
	}
	if rec.entries[1].level != extlogging.LevelInfo {
		t.Errorf("info entry wrong level: %v", rec.entries[1].level)
	}
	if rec.entries[2].level != extlogging.LevelWarn {
		t.Errorf("warn entry wrong level: %v", rec.entries[2].level)
	}
	if rec.entries[3].level != extlogging.LevelError {
		t.Errorf("error entry wrong level: %v", rec.entries[3].level)
	}
}

func TestSetLevelFiltersBelowThreshold(t *testing.T) {
	rec := installRecorder(t, extlogging.LevelDebug)

	SetLevel("warn")
	t.Cleanup(func() { SetLevel("info") })

	Debug().Msg("debug")
	Info().Msg("info")
	Warn().Msg("warn")

	if len(rec.entries) != 1 {
		t.Fatalf("expected 1 entry after filtering, got %d", len(rec.entries))
	}
	if rec.entries[0].level != extlogging.LevelWarn {
		t.Errorf("expected warn entry, got %v", rec.entries[0].level)
	}
}

func TestSetLevelAliases(t *testing.T) {
	cases := map[string]extlogging.Level{
		"debug":   extlogging.LevelDebug,
		"info":    extlogging.LevelInfo,
		"":        extlogging.LevelInfo,
		"warn":    extlogging.LevelWarn,
		"warning": extlogging.LevelWarn,
		"error":   extlogging.LevelError,
		"bogus":   extlogging.LevelInfo,
	}
	t.Cleanup(func() { SetLevel("info") })
	for input, want := range cases {
		SetLevel(input)
		if got := logger.Level(); got != want {
			t.Errorf("SetLevel(%q): got %v, want %v", input, got, want)
		}
	}
}

func TestFatalCallsExit(t *testing.T) {
	rec := installRecorder(t, extlogging.LevelDebug)

	originalExit := exitFunc
	var exitCode int
	exitFunc = func(code int) { exitCode = code }
	t.Cleanup(func() { exitFunc = originalExit })

	Fatal().Str("k", "v").Msg("dying")

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
	if len(rec.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(rec.entries))
	}
	if a, ok := findAttr(rec.entries[0].attrs, "fatal"); !ok || !a.Value.Bool() {
		t.Errorf("expected fatal=true attribute")
	}
}
