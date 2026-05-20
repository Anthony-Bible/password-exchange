package logging

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

	var buf bytes.Buffer
	SetLevel("debug")
	Logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	if Logger == nil {
		t.Fatal("expected logger to be initialized")
	}

	Debug().Msgf("message %d", 1)
	output := buf.String()
	if !strings.Contains(output, `"level":"DEBUG"`) {
		t.Errorf("expected level DEBUG, got output: %s", output)
	}
	if !strings.Contains(output, `"msg":"message 1"`) {
		t.Errorf("expected message 1, got output: %s", output)
	}
}

func TestFatalLog(t *testing.T) {
	// We can't easily test os.Exit(1), so we just test the attribute is added
	e := Fatal()
	found := false
	for _, attr := range e.attrs {
		if a, ok := attr.(slog.Attr); ok && a.Key == "fatal" && a.Value.Bool() == true {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected fatal=true attribute in Fatal() event")
	}
}
