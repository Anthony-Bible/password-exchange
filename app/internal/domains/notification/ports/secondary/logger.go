package secondary

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
)

// LoggerPort defines the secondary port for logging operations
type LoggerPort interface {
	Debug() contracts.LogEvent
	Info() contracts.LogEvent
	Warn() contracts.LogEvent
	Error() contracts.LogEvent
}