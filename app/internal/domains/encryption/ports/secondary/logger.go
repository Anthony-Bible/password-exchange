package secondary

import (
	logport "github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/port"
)

// LoggerPort defines the secondary port for structured logging in the encryption domain.
// It embeds the shared logger interface (the single source of truth in
// internal/shared/logging/port) and adds Fatal, which the encryption domain
// uses for unrecoverable startup failures.
type LoggerPort interface {
	logport.Logger

	// Fatal starts a new log event at fatal severity. After Msg is called on
	// the returned LogEvent, the process terminates via os.Exit(1).
	Fatal() logport.LogEvent
}
