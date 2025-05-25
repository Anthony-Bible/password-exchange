package domain

import (
	"time"
)

// Message represents a stored encrypted message with metadata
type Message struct {
	ID           int64      `json:"id"`
	Content      string     `json:"content"`      // Base64 encoded encrypted message
	UniqueID     string     `json:"unique_id"`    // UUID for message retrieval
	Passphrase   string     `json:"passphrase"`   // Additional security passphrase
	ViewCount    int        `json:"view_count"`   // Number of times the message has been viewed
	MaxViewCount int        `json:"max_view_count"` // Maximum number of views allowed
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// MessageRepository defines the contract for message storage operations
type MessageRepository interface {
	InsertMessage(content, uniqueID, passphrase string, maxViewCount int) error
	SelectMessageByUniqueID(uniqueID string) (*Message, error)
	IncrementViewCountAndGet(uniqueID string) (*Message, error)
	DeleteExpiredMessages() error
	GetMessage(uniqueID string) (*Message, error)
	Close() error
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	Name     string
}
