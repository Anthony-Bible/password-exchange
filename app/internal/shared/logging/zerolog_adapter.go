package logging

import (
	"context"
	"os" // For default handler, will be configurable
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/ports"
	"github.com/rs/zerolog"
)

// ZerologAdapter wraps zerolog.Logger to implement the ports.Logger interface.
type ZerologAdapter struct {
	logger zerolog.Logger
}

// NewZerologAdapter creates a new ZerologAdapter.
// Configuration for output and level will be passed via LogConfig.
func NewZerologAdapter(cfg LogConfig, serviceName string) (ports.Logger, error) {
	var logger zerolog.Logger

	// TODO: Implement output selection based on cfg.Output (stdout, file)
	// TODO: Implement format selection based on cfg.Format (json, console/text)
	// For now, defaults to JSON-like output on Stdout
	if strings.ToLower(cfg.Format) == "text" { // Or "console"
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	// Determine log level
	levelStr := cfg.Level // Default level
	if serviceName != "" {
		if serviceLevel, ok := cfg.Services[serviceName]; ok {
			levelStr = serviceLevel // Use service-specific level if defined
		}
	}

	parsedLevel, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		parsedLevel = zerolog.InfoLevel // Default to Info if parsing fails or level is invalid
	}
	logger = logger.Level(parsedLevel)

	return &ZerologAdapter{logger: logger}, nil
}

func (z *ZerologAdapter) Info(ctx context.Context, msg string, args ...any) {
	event := z.logger.Info()
	z.logWithArgs(event, msg, args...)
}

func (z *ZerologAdapter) Error(ctx context.Context, msg string, args ...any) {
	event := z.logger.Error()
	z.logWithArgs(event, msg, args...)
}

func (z *ZerologAdapter) Debug(ctx context.Context, msg string, args ...any) {
	event := z.logger.Debug()
	z.logWithArgs(event, msg, args...)
}

// logWithArgs helper to handle slog-style arguments for zerolog events.
func (z *ZerologAdapter) logWithArgs(event *zerolog.Event, msg string, args ...any) {
	// Zerolog's primary API is .Str(key, val).Msg()
	// Slog uses Info(msg, key1, val1, key2, val2...)
	// This adapter needs to bridge that.
	if len(args) > 0 {
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				key, ok := args[i].(string)
				if ok {
					event = event.Interface(key, args[i+1])
				}
			}
		}
	}
	event.Msg(msg)
}

func (z *ZerologAdapter) With(key string, value any) ports.Logger {
	// zerolog.Logger.With() returns a zerolog.Context, then .Logger() makes it a new logger.
	return &ZerologAdapter{logger: z.logger.With().Interface(key, value).Logger()}
}

func (z *ZerologAdapter) WithContext(ctx context.Context) ports.Logger {
	if corrID, ok := ctx.Value(ports.CorrelationIDKey).(string); ok && corrID != "" {
		return &ZerologAdapter{logger: z.logger.With().Str("correlationID", corrID).Logger()}
	}
	return z
}
