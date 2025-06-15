package logging

import (
	"context"
	"io"
	"os"
	"password-exchange/internal/shared/logging/ports"

	"github.com/rs/zerolog"
)

// ZerologAdapter is an adapter for the zerolog logger,
// making it conform to the ports.Logger interface.
type ZerologAdapter struct {
	logger zerolog.Logger
}

// NewZerologAdapter creates a new ZerologAdapter.
// It configures a zerolog.Logger with a specified log level and output writer.
// By default, it writes to os.Stdout.
func NewZerologAdapter(level zerolog.Level, writer io.Writer) *ZerologAdapter {
	if writer == nil {
		writer = os.Stdout
	}
	logger := zerolog.New(writer).With().Timestamp().Logger().Level(level)
	return &ZerologAdapter{logger: logger}
}

// Info logs an informational message using zerolog.
func (z *ZerologAdapter) Info(ctx context.Context, msg string, args ...any) {
	z.logger.Info().Fields(convertToFields(args...)).Msg(msg)
}

// Error logs an error message using zerolog.
func (z *ZerologAdapter) Error(ctx context.Context, msg string, args ...any) {
	z.logger.Error().Fields(convertToFields(args...)).Msg(msg)
}

// Debug logs a debug message using zerolog.
func (z *ZerologAdapter) Debug(ctx context.Context, msg string, args ...any) {
	z.logger.Debug().Fields(convertToFields(args...)).Msg(msg)
}

// With returns a new ZerologAdapter with the specified key-value pair added to its context.
func (z *ZerologAdapter) With(key string, value any) ports.Logger {
	return &ZerologAdapter{logger: z.logger.With().Interface(key, value).Logger()}
}

// WithContext returns a new ZerologAdapter with context-derived fields.
// Zerolog's context integration is often about adding fields from the context
// rather than replacing the logger's context wholesale.
// This example extracts common fields; customize as needed.
func (z *ZerologAdapter) WithContext(ctx context.Context) ports.Logger {
	// Example: Extract a correlation ID from context and add it to the logger
	// correlationID, ok := ctx.Value("correlationID").(string)
	// if ok {
	//	 return &ZerologAdapter{logger: z.logger.With().Str("correlationID", correlationID).Logger()}
	// }
	// For now, returning the same logger as context handling can be complex
	// and application-specific with zerolog.
	return z
}

// convertToFields converts a list of key-value pairs (args) into a map
// suitable for zerolog's Fields method.
// It expects args to be in the format: key1, value1, key2, value2, ...
func convertToFields(args ...any) map[string]any {
	fields := make(map[string]any)
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if ok {
				fields[key] = args[i+1]
			}
		}
	}
	return fields
}
