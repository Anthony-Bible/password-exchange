package secondary

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
)

// LoggerPort defines the secondary port for logging operations.
// Implementations produce a contracts.LogEvent at the requested severity level,
// allowing the domain and adapters to emit structured logs without depending on
// any specific logging library.
type LoggerPort interface {
	// Debug starts a debug-level log event for diagnostic output that is
	// typically silenced in production. Returns a contracts.LogEvent that
	// must be finalized with Msg to actually emit the record.
	Debug() contracts.LogEvent

	// Info starts an info-level log event for routine, high-signal operational
	// messages (service lifecycle, successful requests, etc.). Returns a
	// contracts.LogEvent that must be finalized with Msg.
	Info() contracts.LogEvent

	// Warn starts a warn-level log event for recoverable anomalies that do
	// not interrupt the current operation but may indicate degraded behavior.
	// Returns a contracts.LogEvent that must be finalized with Msg.
	Warn() contracts.LogEvent

	// Error starts an error-level log event for failures that prevented the
	// current operation from completing successfully. Returns a
	// contracts.LogEvent that must be finalized with Msg.
	Error() contracts.LogEvent
}
