package ports

import "context"

// Logger defines the interface for logging operations.
// This interface will be implemented by both zerolog and slog adapters
// to allow for a phased migration.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any)
	With(key string, value any) Logger
	WithContext(ctx context.Context) Logger
}

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

// CorrelationIDKey is the key for storing/retrieving correlation IDs in context.
const CorrelationIDKey ContextKey = "correlationID"
