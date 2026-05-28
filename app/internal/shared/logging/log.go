package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	extlogging "github.com/Anthony-Bible/Logging"
)

// exitFunc is overridable for tests.
var exitFunc = os.Exit

var (
	slogLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// extlogging.New installs SIGUSR1/SIGUSR2 handlers itself when
	// DisableSignalHandling is false, so no extra wiring is needed here.
	logger = extlogging.New(extlogging.Config{
		Backend:        extlogging.NewSlogBackend(slogLogger),
		Level:          extlogging.LevelInfo,
		ErrorThreshold: 5,
		ErrorWindow:    time.Minute,
		DebugDuration:  5 * time.Minute,
	})
)

// Logger preserves the previously exported *slog.Logger handle for backwards
// compatibility with any code that imported it directly.
var Logger = slogLogger

type Event struct {
	level extlogging.Level
	attrs []slog.Attr
	fatal bool
}

func newEvent(level extlogging.Level, fatal bool) *Event {
	return &Event{level: level, fatal: fatal}
}

func Debug() *Event { return newEvent(extlogging.LevelDebug, false) }
func Info() *Event  { return newEvent(extlogging.LevelInfo, false) }
func Warn() *Event  { return newEvent(extlogging.LevelWarn, false) }
func Error() *Event { return newEvent(extlogging.LevelError, false) }
func Fatal() *Event {
	e := newEvent(extlogging.LevelError, true)
	e.attrs = append(e.attrs, slog.Bool("fatal", true))
	return e
}

func (e *Event) Err(err error) *Event {
	if err != nil {
		e.attrs = append(e.attrs, slog.Any("error", err))
	}
	return e
}

func (e *Event) Str(key, value string) *Event {
	e.attrs = append(e.attrs, slog.String(key, value))
	return e
}

func (e *Event) Int(key string, value int) *Event {
	e.attrs = append(e.attrs, slog.Int(key, value))
	return e
}

func (e *Event) Int64(key string, value int64) *Event {
	e.attrs = append(e.attrs, slog.Int64(key, value))
	return e
}

func (e *Event) Int32(key string, value int32) *Event {
	e.attrs = append(e.attrs, slog.Int64(key, int64(value)))
	return e
}

func (e *Event) Bool(key string, value bool) *Event {
	e.attrs = append(e.attrs, slog.Bool(key, value))
	return e
}

func (e *Event) Dur(key string, value time.Duration) *Event {
	e.attrs = append(e.attrs, slog.Duration(key, value))
	return e
}

func (e *Event) Float64(key string, value float64) *Event {
	e.attrs = append(e.attrs, slog.Float64(key, value))
	return e
}

func (e *Event) Interface(key string, value any) *Event {
	e.attrs = append(e.attrs, slog.Any(key, value))
	return e
}

func (e *Event) Msg(msg string) {
	ctx := context.Background()
	switch e.level {
	case extlogging.LevelDebug:
		logger.Debug(ctx, msg, e.attrs...)
	case extlogging.LevelInfo:
		logger.Info(ctx, msg, e.attrs...)
	case extlogging.LevelWarn:
		logger.Warn(ctx, msg, e.attrs...)
	case extlogging.LevelError:
		logger.Error(ctx, msg, e.attrs...)
	default:
		logger.Log(ctx, e.level, msg, e.attrs...)
	}
	if e.fatal {
		exitFunc(1)
	}
}

func (e *Event) Msgf(format string, args ...any) {
	e.Msg(fmt.Sprintf(format, args...))
}

func SetLevel(level string) {
	var l extlogging.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		l = extlogging.LevelDebug
	case "info", "":
		l = extlogging.LevelInfo
	case "warn", "warning":
		l = extlogging.LevelWarn
	case "error":
		l = extlogging.LevelError
	default:
		l = extlogging.LevelInfo
	}
	logger.SetLevel(l)
}
