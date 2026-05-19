package zerolog

import (
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging"
)

type ZerologAdapter struct{}

func NewZerologAdapter() secondary.LoggerPort {
	return &ZerologAdapter{}
}

func (z *ZerologAdapter) Debug() contracts.LogEvent {
	return &ZerologEvent{event: log.Debug()}
}

func (z *ZerologAdapter) Info() contracts.LogEvent {
	return &ZerologEvent{event: log.Info()}
}

func (z *ZerologAdapter) Warn() contracts.LogEvent {
	return &ZerologEvent{event: log.Warn()}
}

func (z *ZerologAdapter) Error() contracts.LogEvent {
	return &ZerologEvent{event: log.Error()}
}

type ZerologEvent struct {
	event *log.Event
}

func (e *ZerologEvent) Err(err error) contracts.LogEvent {
	e.event = e.event.Err(err)
	return e
}

func (e *ZerologEvent) Str(key, value string) contracts.LogEvent {
	e.event = e.event.Str(key, value)
	return e
}

func (e *ZerologEvent) Int(key string, value int) contracts.LogEvent {
	e.event = e.event.Int(key, value)
	return e
}

func (e *ZerologEvent) Bool(key string, value bool) contracts.LogEvent {
	e.event = e.event.Bool(key, value)
	return e
}

func (e *ZerologEvent) Dur(key string, value time.Duration) contracts.LogEvent {
	e.event = e.event.Dur(key, value)
	return e
}

func (e *ZerologEvent) Float64(key string, value float64) contracts.LogEvent {
	e.event = e.event.Float64(key, value)
	return e
}

func (e *ZerologEvent) Msg(msg string) {
	e.event.Msg(msg)
}
