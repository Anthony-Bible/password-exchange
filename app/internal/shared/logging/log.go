package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

var Logger = slog.Default()

type Event struct {
	logger *slog.Logger
	level  slog.Level
	attrs  []any
	fatal  bool
}

func newEvent(level slog.Level, fatal bool) *Event {
	return &Event{logger: Logger, level: level, fatal: fatal}
}

func Debug() *Event { return newEvent(slog.LevelDebug, false) }
func Info() *Event  { return newEvent(slog.LevelInfo, false) }
func Warn() *Event  { return newEvent(slog.LevelWarn, false) }
func Error() *Event { return newEvent(slog.LevelError, false) }
func Fatal() *Event { return newEvent(slog.LevelError, true) }

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
	e.logger.Log(context.Background(), e.level, msg, e.attrs...)
	if e.fatal {
		os.Exit(1)
	}
}

func (e *Event) Msgf(format string, args ...any) {
	e.Msg(fmt.Sprintf(format, args...))
}

func SetLevel(level string) {
	var l slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		l = slog.LevelDebug
	case "info", "":
		l = slog.LevelInfo
	case "warn", "warning":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l}))
}
