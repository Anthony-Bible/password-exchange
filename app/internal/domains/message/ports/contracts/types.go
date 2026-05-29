// Package contracts defines shared types and interfaces used across the message domain.
// These types serve as contracts between different layers of the hexagonal architecture,
// ensuring consistent data structures for message processing, encryption, storage, and logging.
package contracts

import (
	"time"

	logport "github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/port"
)

// LogEvent is the shared structured-logging event contract. It is an alias to
// the single definition in internal/shared/logging/port so adding a field
// happens in exactly one place across every domain.
type LogEvent = logport.LogEvent

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
