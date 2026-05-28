// Package contracts defines shared types and interfaces used across the storage domain.
// These types serve as contracts between different layers of the hexagonal architecture,
// ensuring consistent data structures for logging abstractions and other cross-layer
// concerns without introducing import cycles between ports and adapters.
package contracts

import (
	"time"
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

// LogEvent represents a structured logging event that can be enriched with contextual data.
// This interface follows a fluent API pattern, allowing method chaining to add various
// types of contextual information before finalizing the log entry. This abstraction
// allows the storage domain to remain independent of specific logging implementations.
type LogEvent interface {
	// Err adds an error to the log event.
	// The error will be formatted and included in the log output.
	//
	// Parameters:
	//   - err: The error to log (can be nil)
	//
	// Returns:
	//   - The LogEvent for method chaining
	Err(error) LogEvent

	// Str adds a string key-value pair to the log event.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The string value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Str(string, string) LogEvent

	// Int adds an integer key-value pair to the log event.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The integer value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Int(string, int) LogEvent

	// Int32 adds an int32 key-value pair to the log event.
	// Useful for fields that arrive as int32 (e.g. protobuf scalars) without
	// requiring callers to widen them at the call site.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The int32 value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Int32(string, int32) LogEvent

	// Int64 adds an int64 key-value pair to the log event.
	// Useful for fields that naturally exceed 32-bit range such as row
	// counts returned by sql.Result.RowsAffected.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The int64 value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Int64(string, int64) LogEvent

	// Bool adds a boolean key-value pair to the log event.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The boolean value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Bool(string, bool) LogEvent

	// Dur adds a duration key-value pair to the log event.
	// The duration is typically formatted in a human-readable way.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The duration value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Dur(string, time.Duration) LogEvent

	// Float64 adds a float64 key-value pair to the log event.
	//
	// Parameters:
	//   - key: The field name
	//   - value: The float64 value
	//
	// Returns:
	//   - The LogEvent for method chaining
	Float64(string, float64) LogEvent

	// Msg finalizes the log event with a message and writes it to the log.
	// This method should be called last in the chain.
	//
	// Parameters:
	//   - message: The log message describing the event
	Msg(string)
}
