// Package contracts defines shared types and interfaces used across the storage domain.
// These types serve as contracts between different layers of the hexagonal architecture,
// ensuring consistent data structures for logging abstractions and other cross-layer
// concerns without introducing import cycles between ports and adapters.
package contracts

import (
	"time"

	logport "github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/port"
)

// Message represents a stored encrypted message with metadata.
// It is the canonical contract type used by the storage domain, its ports,
// and any adapters that need to exchange message records.
type Message struct {
	ID             int64      `json:"id"`
	Content        string     `json:"content"`         // Base64 encoded encrypted message
	UniqueID       string     `json:"unique_id"`       // UUID for message retrieval
	Passphrase     string     `json:"passphrase"`      // Additional security passphrase
	RecipientEmail string     `json:"recipient_email"` // Email address of the recipient
	ViewCount      int        `json:"view_count"`      // Number of times the message has been viewed
	MaxViewCount   int        `json:"max_view_count"`  // Maximum number of views allowed
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// UnviewedMessage represents a message that is still unviewed and eligible to
// receive a reminder email. It captures the subset of message metadata that
// reminder workers need without exposing the full Message contract.
type UnviewedMessage struct {
	MessageID      int       `json:"message_id"`
	UniqueID       string    `json:"unique_id"`
	RecipientEmail string    `json:"recipient_email"`
	Created        time.Time `json:"created"`
	DaysOld        int       `json:"days_old"`
}

// ReminderLogEntry represents a single recipient's reminder history for a message,
// including how many reminders have been sent and when the most recent one fired.
type ReminderLogEntry struct {
	MessageID        int       `json:"message_id"`
	EmailAddress     string    `json:"email_address"`
	ReminderCount    int       `json:"reminder_count"`
	LastReminderSent time.Time `json:"last_reminder_sent"`
}

// DatabaseConfig contains the connection settings required to reach the
// underlying relational database used by the storage domain.
type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	Name     string
}

// LogEvent is the shared structured-logging event contract. It is an alias to
// the single definition in internal/shared/logging/port, so the storage domain
// stays independent of any concrete logging implementation while a new field
// is added in exactly one place across every domain.
type LogEvent = logport.LogEvent
