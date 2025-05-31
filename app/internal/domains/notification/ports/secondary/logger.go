package secondary

import (
	"time"
)

// LoggerPort defines the secondary port for logging operations
type LoggerPort interface {
	Debug() LogEvent
	Info() LogEvent
	Warn() LogEvent
	Error() LogEvent
}

// LogEvent represents a logging event that can be enriched with context
type LogEvent interface {
	Err(error) LogEvent
	Str(string, string) LogEvent
	Int(string, int) LogEvent
	Bool(string, bool) LogEvent
	Dur(string, time.Duration) LogEvent
	Float64(string, float64) LogEvent
	Msg(string)
}