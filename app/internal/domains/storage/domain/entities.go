package domain

import (
	"time"
)

// Message represents a stored encrypted message with metadata
type Message struct {
	ID             int64      `json:"id"`
	Content        string     `json:"content"`        // Base64 encoded encrypted message
	UniqueID       string     `json:"unique_id"`      // UUID for message retrieval
	Passphrase     string     `json:"passphrase"`     // Additional security passphrase
	RecipientEmail string     `json:"recipient_email"` // Email address of the recipient
	ViewCount      int        `json:"view_count"`     // Number of times the message has been viewed
	MaxViewCount   int        `json:"max_view_count"` // Maximum number of views allowed
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// UnviewedMessage represents a message eligible for reminder emails
type UnviewedMessage struct {
	MessageID      int       `json:"message_id"`
	UniqueID       string    `json:"unique_id"`
	RecipientEmail string    `json:"recipient_email"`
	Created        time.Time `json:"created"`
	DaysOld        int       `json:"days_old"`
}

// ReminderLogEntry represents a logged reminder attempt
type ReminderLogEntry struct {
	MessageID         int       `json:"message_id"`
	EmailAddress      string    `json:"email_address"`
	ReminderCount     int       `json:"reminder_count"`
	LastReminderSent  time.Time `json:"last_reminder_sent"`
}

// MessageRepository defines the contract for message storage operations
type MessageRepository interface {
	InsertMessage(message *Message) error
	SelectMessageByUniqueID(uniqueID string) (*Message, error)
	IncrementViewCountAndGet(uniqueID string) (*Message, error)
	DeleteExpiredMessages() error
	GetMessage(uniqueID string) (*Message, error)
	GetUnviewedMessagesForReminders(olderThanHours, maxReminders, reminderIntervalHours int) ([]*UnviewedMessage, error)
	LogReminderSent(messageID int, emailAddress string) error
	GetReminderHistory(messageID int) ([]*ReminderLogEntry, error)
	Close() error
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	Name     string
}
