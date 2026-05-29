package secondary

import (
	logport "github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/port"
)

// LoggerPort defines the secondary port for logging operations. It is an alias
// to the shared logger interface in internal/shared/logging/port, the single
// source of truth consumed by every domain.
type LoggerPort = logport.Logger
