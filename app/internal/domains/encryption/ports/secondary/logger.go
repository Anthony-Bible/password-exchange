package secondary

import "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"

// LoggerPort defines the secondary port for structured logging in the encryption domain.
// Implementations adapt a specific logging library to the domain's needs without
// leaking that library's types into the domain layer.
type LoggerPort interface {
	// Debug starts a new log event at debug level.
	Debug() contracts.LogEvent

	// Info starts a new log event at info level.
	Info() contracts.LogEvent

	// Warn starts a new log event at warn level.
	Warn() contracts.LogEvent

	// Error starts a new log event at error level.
	Error() contracts.LogEvent

	// Fatal starts a new log event at fatal severity. After Msg is called on
	// the returned LogEvent, the process terminates via os.Exit(1).
	Fatal() contracts.LogEvent
}
