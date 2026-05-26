// Package contracts defines shared types and interfaces used across the message domain.
// These types serve as contracts between different layers of the hexagonal architecture,
// ensuring consistent data structures for message processing, encryption, storage, and logging.
package contracts

import (
	"time"
)

// LogEvent represents a structured logging event that can be enriched with contextual data.
type LogEvent interface {
	Err(error) LogEvent
	Str(string, string) LogEvent
	Int(string, int) LogEvent
	Bool(string, bool) LogEvent
	Dur(string, time.Duration) LogEvent
	Float64(string, float64) LogEvent
	Msg(string)
}

// MessageSubmissionRequest represents a request to share a new encrypted message
type MessageSubmissionRequest struct {
	Content          string
	SenderName       string
	SenderEmail      string
	RecipientName    string
	RecipientEmail   string
	Passphrase       string
	AdditionalInfo   string
	Captcha          string
	TurnstileToken   string
	SendNotification bool
	MaxViewCount     int
	ExpirationHours  int
}

// MessageSubmissionResponse represents the response to a message submission
type MessageSubmissionResponse struct {
	MessageID  string
	Key        string
	DecryptURL string
	ExpiresAt  *time.Time
	Success    bool
	Error      error
}

// MessageRetrievalRequest represents a request to retrieve and decrypt a message
type MessageRetrievalRequest struct {
	MessageID     string
	DecryptionKey []byte
	Passphrase    string
}

// MessageRetrievalResponse represents the response to a message retrieval
type MessageRetrievalResponse struct {
	MessageID    string
	Content      string
	ViewCount    int
	MaxViewCount int
	ExpiresAt    *time.Time
	Success      bool
	Error        error
}

// MessageAccessInfo provides information about message access requirements
type MessageAccessInfo struct {
	MessageID          string
	Exists             bool
	RequiresPassphrase bool
	ExpiresAt          *time.Time
}

// MessageStorageRequest represents a request to store an encrypted message
type MessageStorageRequest struct {
	MessageID      string
	Content        string
	Passphrase     string
	MaxViewCount   int
	RecipientEmail string
	ExpiresAt      *time.Time
}

// MessageRetrievalStorageRequest represents a request to retrieve a stored message
type MessageRetrievalStorageRequest struct {
	MessageID string
}

// MessageStorageResponse represents a stored message from storage
type MessageStorageResponse struct {
	MessageID        string
	EncryptedContent string
	HashedPassphrase string
	HasPassphrase    bool
	ViewCount        int
	MaxViewCount     int
	ExpiresAt        *time.Time
}

// MessageNotificationRequest represents a request to send a message notification
type MessageNotificationRequest struct {
	SenderName     string
	SenderEmail    string
	RecipientName  string
	RecipientEmail string
	MessageURL     string
	AdditionalInfo string
}
