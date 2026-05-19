package log

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestEventLogsStructuredFields(t *testing.T) {
	original := Logger
	t.Cleanup(func() { Logger = original })

	var buf bytes.Buffer
	Logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	Error().
		Str("component", "unit-test").
		Int("attempt", 2).
		Err(errors.New("boom")).
		Msg("log message")

	output := buf.String()
	for _, expected := range []string{`"msg":"log message"`, `"component":"unit-test"`, `"attempt":2`, `"error":"boom"`} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %s, got: %s", expected, output)
		}
	}
}

func TestSetLevelAndMsgf(t *testing.T) {
	original := Logger
	t.Cleanup(func() { Logger = original })

	SetLevel("debug")
	if Logger == nil {
		t.Fatal("expected logger to be initialized")
	}

	Debug().Msgf("message %d", 1)
}
