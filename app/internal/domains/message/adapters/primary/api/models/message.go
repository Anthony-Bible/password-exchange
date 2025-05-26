package models

import (
	"time"
)

// MessageSubmissionRequest represents a REST API request to submit a new message
type MessageSubmissionRequest struct {
	Content          string     `json:"content" validate:"required,min=1,max=10000"`
	Sender           *Sender    `json:"sender,omitempty"`
	Recipient        *Recipient `json:"recipient,omitempty"`
	Passphrase       string     `json:"passphrase,omitempty" validate:"max=500"`
	AdditionalInfo   string     `json:"additionalInfo,omitempty"`
	SendNotification bool       `json:"sendNotification"`
	AntiSpamAnswer   string     `json:"antiSpamAnswer,omitempty"`
	MaxViewCount     int        `json:"maxViewCount,omitempty" validate:"min=0,max=100"`
}

// Sender represents sender information for message submission
type Sender struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Email string `json:"email" validate:"required,email"`
}

// Recipient represents recipient information for message submission
type Recipient struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Email string `json:"email" validate:"required,email"`
}

// MessageSubmissionResponse represents the response to a message submission
type MessageSubmissionResponse struct {
	MessageID        string    `json:"messageId"`
	DecryptURL       string    `json:"decryptUrl"`
	Key              string    `json:"key"`
	WebURL           string    `json:"webUrl"`
	ExpiresAt        time.Time `json:"expiresAt"`
	NotificationSent bool      `json:"notificationSent"`
}

// MessageAccessInfoResponse represents information about message access requirements
type MessageAccessInfoResponse struct {
	MessageID          string    `json:"messageId"`
	Exists             bool      `json:"exists"`
	RequiresPassphrase bool      `json:"requiresPassphrase"`
	HasBeenAccessed    bool      `json:"hasBeenAccessed"`
	ExpiresAt          time.Time `json:"expiresAt"`
}

// MessageDecryptRequest represents a request to decrypt a message
type MessageDecryptRequest struct {
	DecryptionKey string `json:"decryptionKey" validate:"required"`
	Passphrase    string `json:"passphrase,omitempty"`
}

// MessageDecryptResponse represents the response to a message decryption
type MessageDecryptResponse struct {
	MessageID    string    `json:"messageId"`
	Content      string    `json:"content"`
	ViewCount    int       `json:"viewCount"`
	MaxViewCount int       `json:"maxViewCount"`
	DecryptedAt  time.Time `json:"decryptedAt"`
}

// HealthCheckResponse represents the response to a health check
type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// APIInfoResponse represents information about the API
type APIInfoResponse struct {
	Version       string            `json:"version"`
	Documentation string            `json:"documentation"`
	Endpoints     map[string]string `json:"endpoints"`
	Features      map[string]bool   `json:"features"`
}
