package logging

import (
	"context"
	"log/slog"
	"os" // For default handler, will be configurable
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/ports"
)

// SlogAdapter wraps slog.Logger to implement the ports.Logger interface.
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new SlogAdapter.
// Configuration for handler type (JSON, text), output (stdout, file), and level
// will be passed via LogConfig.
func NewSlogAdapter(cfg LogConfig, serviceName string) (ports.Logger, error) {
	// Determine log level
	levelStr := cfg.Level // Default level
	if serviceName != "" {
		if serviceLevel, ok := cfg.Services[serviceName]; ok {
			levelStr = serviceLevel // Use service-specific level if defined
		}
	}

	logLevel := new(slog.LevelVar)
	switch strings.ToLower(levelStr) {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelInfo) // Default to Info
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
		// AddSource: true, // Optionally add source file and line
	}

	var handler slog.Handler
	// TODO: Implement output selection based on cfg.Output (stdout, file, multi)
	// TODO: Implement format selection based on cfg.Format (json, text)
	// For now, defaults to JSON handler on Stdout
	if strings.ToLower(cfg.Format) == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	return &SlogAdapter{logger: logger}, nil
}

func (s *SlogAdapter) Info(ctx context.Context, msg string, args ...any) {
	s.logger.InfoContext(ctx, msg, args...)
}

func (s *SlogAdapter) Error(ctx context.Context, msg string, args ...any) {
	s.logger.ErrorContext(ctx, msg, args...)
}

func (s *SlogAdapter) Debug(ctx context.Context, msg string, args ...any) {
	s.logger.DebugContext(ctx, msg, args...)
}

func (s *SlogAdapter) With(key string, value any) ports.Logger {
	// slog.With creates a new logger with the field, so we return a new adapter
	return &SlogAdapter{logger: s.logger.With(key, value)}
}

// WithContext for SlogAdapter can enhance the logger with values from context if needed.
// Slog's InfoContext etc. automatically handle context, but this can be used for structured logging from context.
func (s *SlogAdapter) WithContext(ctx context.Context) ports.Logger {
	if corrID, ok := ctx.Value(ports.CorrelationIDKey).(string); ok && corrID != "" {
		return &SlogAdapter{logger: s.logger.With("correlationID", corrID)}
	}
	return s
}
