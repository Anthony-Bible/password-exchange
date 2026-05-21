package logger

import (
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging"
)

// Adapter implements LoggerPort using the shared slog-based logging package
type Adapter struct{}

// NewAdapter creates a new logger adapter
func NewAdapter() secondary.LoggerPort {
	return &Adapter{}
}

func (a *Adapter) Debug() contracts.LogEvent {
	return &Event{event: logging.Debug()}
}

func (a *Adapter) Info() contracts.LogEvent {
	return &Event{event: logging.Info()}
}

func (a *Adapter) Warn() contracts.LogEvent {
	return &Event{event: logging.Warn()}
}

func (a *Adapter) Error() contracts.LogEvent {
	return &Event{event: logging.Error()}
}

// Event implements LogEvent for the shared logger
type Event struct {
	event *logging.Event
}

func (e *Event) Err(err error) contracts.LogEvent {
	e.event = e.event.Err(err)
	return e
}

func (e *Event) Str(key, value string) contracts.LogEvent {
	e.event = e.event.Str(key, value)
	return e
}

func (e *Event) Int(key string, value int) contracts.LogEvent {
	e.event = e.event.Int(key, value)
	return e
}

func (e *Event) Bool(key string, value bool) contracts.LogEvent {
	e.event = e.event.Bool(key, value)
	return e
}

func (e *Event) Dur(key string, value time.Duration) contracts.LogEvent {
	e.event = e.event.Dur(key, value)
	return e
}

func (e *Event) Float64(key string, value float64) contracts.LogEvent {
	e.event = e.event.Float64(key, value)
	return e
}

func (e *Event) Msg(msg string) {
	e.event.Msg(msg)
}
