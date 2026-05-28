// Package domain contains the storage domain's business logic and error
// sentinels. The entity types previously declared in this file (Message,
// UnviewedMessage, ReminderLogEntry, DatabaseConfig) have been relocated to
// internal/domains/storage/ports/contracts so the domain can depend on the
// secondary ports without producing an import cycle.
package domain
