package logging

import (
	"context"
	"password-exchange/internal/shared/logging/ports"
)

// DualLogger acts as a bridge, forwarding logging calls to two underlying loggers (e.g., zerolog and slog).
// This is useful for migrating from one logging system to another, allowing for comparison and verification.
type DualLogger struct {
	logger1 ports.Logger // Primary logger (e.g., the new slog implementation)
	logger2 ports.Logger // Secondary logger (e.g., the old zerolog implementation)
	useNew  bool         // Feature flag to control which logger is primary
}

// NewDualLogger creates a new DualLogger instance.
// It takes two loggers that implement the ports.Logger interface and a feature flag.
// If useNew is true, logger1 is primary; otherwise, logger2 is primary.
func NewDualLogger(logger1, logger2 ports.Logger, useNew bool) *DualLogger {
	return &DualLogger{
		logger1: logger1,
		logger2: logger2,
		useNew:  useNew,
	}
}

// Info logs an informational message using both loggers.
func (d *DualLogger) Info(ctx context.Context, msg string, args ...any) {
	d.logger1.Info(ctx, msg, args...)
	d.logger2.Info(ctx, msg, args...)
}

// Error logs an error message using both loggers.
func (d *DualLogger) Error(ctx context.Context, msg string, args ...any) {
	d.logger1.Error(ctx, msg, args...)
	d.logger2.Error(ctx, msg, args...)
}

// Debug logs a debug message using both loggers.
func (d *DualLogger) Debug(ctx context.Context, msg string, args ...any) {
	d.logger1.Debug(ctx, msg, args...)
	d.logger2.Debug(ctx, msg, args...)
}

// With returns a new DualLogger with the specified key-value pair added to both underlying loggers.
func (d *DualLogger) With(key string, value any) ports.Logger {
	return NewDualLogger(
		d.logger1.With(key, value),
		d.logger2.With(key, value),
		d.useNew,
	)
}

// WithContext returns a new DualLogger with the given context applied to both underlying loggers.
func (d *DualLogger) WithContext(ctx context.Context) ports.Logger {
	return NewDualLogger(
		d.logger1.WithContext(ctx),
		d.logger2.WithContext(ctx),
		d.useNew,
	)
}
