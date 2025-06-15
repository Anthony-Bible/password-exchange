package ports

import "context"

// Logger defines the interface for logging operations.
// It's designed to be implemented by different logging libraries (e.g., slog, zerolog)
// and used throughout the application for consistent logging.
type Logger interface {
	// Info logs an informational message.
	Info(ctx context.Context, msg string, args ...any)
	// Error logs an error message.
	Error(ctx context.Context, msg string, args ...any)
	// Debug logs a debug message.
	Debug(ctx context.Context, msg string, args ...any)
	// With returns a new logger with the specified key-value pair added to its context.
	With(key string, value any) Logger
	// WithContext returns a new logger with the given context.
	WithContext(ctx context.Context) Logger
}
