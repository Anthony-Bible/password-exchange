package logging

import (
	"context"
	"log/slog"
	"os"
	"password-exchange/internal/shared/logging/ports"
)

// SlogAdapter is an adapter for the standard library's slog logger,
// making it conform to the ports.Logger interface.
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new SlogAdapter.
// It configures a slog.Logger with JSON output to stdout and a specified log level.
func NewSlogAdapter(level slog.Level) *SlogAdapter {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	return &SlogAdapter{logger: logger}
}

// Info logs an informational message using slog.
func (s *SlogAdapter) Info(ctx context.Context, msg string, args ...any) {
	s.logger.InfoContext(ctx, msg, args...)
}

// Error logs an error message using slog.
func (s *SlogAdapter) Error(ctx context.Context, msg string, args ...any) {
	s.logger.ErrorContext(ctx, msg, args...)
}

// Debug logs a debug message using slog.
func (s *SlogAdapter) Debug(ctx context.Context, msg string, args ...any) {
	s.logger.DebugContext(ctx, msg, args...)
}

// With returns a new SlogAdapter with the specified key-value pair added to its context.
func (s *SlogAdapter) With(key string, value any) ports.Logger {
	return &SlogAdapter{logger: s.logger.With(key, value)}
}

// WithContext returns a new SlogAdapter with the given context.
// Note: slog's WithContext is not part of the standard Handler interface,
// so we are returning the same logger instance. Context is typically passed directly
// to logging methods (Info, Error, Debug).
func (s *SlogAdapter) WithContext(ctx context.Context) ports.Logger {
	// Slog's InfoContext, ErrorContext, DebugContext methods handle context directly.
	// If specific context attributes need to be embedded in the logger,
	// they should be extracted and added using With.
	// For now, returning the same logger as context is passed per-call.
	return s
}
