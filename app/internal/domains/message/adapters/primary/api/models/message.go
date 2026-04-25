package models

import (
	"time"
)

// MessageSubmissionRequest represents a REST API request to submit a new message
type MessageSubmissionRequest struct {
	// Secret content to be encrypted
	Content string `json:"content" validate:"required,min=1,max=10000" example:"This is a secret message"`
	// Sender information (required if notifications enabled)
	Sender *Sender `json:"sender,omitempty"`
	// Recipient information (required if notifications enabled)
	Recipient *Recipient `json:"recipient,omitempty"`
	// Optional passphrase for additional security
	Passphrase string `json:"passphrase,omitempty" validate:"max=500" example:"correct-horse-battery-staple"`
	// Optional additional information for the recipient
	AdditionalInfo string `json:"additionalInfo,omitempty" example:"Valid for 24 hours"`
	// Whether to send email notifications
	SendNotification bool `json:"sendNotification" example:"true"`
	// Answer to the anti-spam question (required if notifications enabled)
	AntiSpamAnswer string `json:"antiSpamAnswer,omitempty" example:"blue"`
	// ID of the anti-spam question
	QuestionID *int `json:"questionId,omitempty" example:"0"`
	// Maximum number of times the message can be viewed (0 for unlimited)
	MaxViewCount int `json:"maxViewCount,omitempty" validate:"min=0,max=100" example:"1"`
	// Cloudflare Turnstile token for verification
	TurnstileToken string `json:"turnstileToken,omitempty" validate:"max=2048"`
	// ExpirationHours specifies a custom expiration in hours. When 0 or omitted, the server default (7 days / 168 hours) applies.
	// Valid range: 1–2160 (1 hour to 90 days).
	ExpirationHours int `json:"expirationHours,omitempty" validate:"min=0,max=2160" example:"24"`
}

// Sender represents sender information for message submission
type Sender struct {
	// Sender's name
	Name string `json:"name" validate:"required,min=1,max=100" example:"John Doe"`
	// Sender's email address
	Email string `json:"email" validate:"required,email" example:"john@example.com"`
}

// Recipient represents recipient information for message submission
type Recipient struct {
	// Recipient's name
	Name string `json:"name" validate:"min=1,max=100" example:"Jane Smith"`
	// Recipient's email address
	Email string `json:"email" validate:"required,email" example:"jane@example.com"`
}

// MessageSubmissionResponse represents the response to a message submission
type MessageSubmissionResponse struct {
	// Unique identifier for the created message
	MessageID string `json:"messageId" example:"123e4567-e89b-12d3-a456-426614174000"`
	// API URL to decrypt the message
	DecryptURL string `json:"decryptUrl" example:"https://api.password.exchange/api/v1/messages/123e4567-e89b-12d3-a456-426614174000/decrypt?key=YWJjZGVmZ2hpams="`
	// Base64 encoded encryption key
	Key string `json:"key" example:"YWJjZGVmZ2hpams="`
	// Web URL for human-friendly access
	WebURL string `json:"webUrl" example:"https://password.exchange/decrypt/123e4567-e89b-12d3-a456-426614174000/YWJjZGVmZ2hpams="`
	// When the message will expire
	ExpiresAt *time.Time `json:"expiresAt"`
	// Whether an email notification was successfully sent
	NotificationSent bool `json:"notificationSent" example:"true"`
}

// MessageAccessInfoResponse represents information about message access requirements
type MessageAccessInfoResponse struct {
	// Unique identifier for the message
	MessageID string `json:"messageId" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Whether the message exists and is still available
	Exists bool `json:"exists" example:"true"`
	// Whether a passphrase is required to decrypt the message
	RequiresPassphrase bool `json:"requiresPassphrase" example:"true"`
	// Whether the message has already been accessed
	HasBeenAccessed bool `json:"hasBeenAccessed" example:"false"`
	// When the message will expire
	ExpiresAt *time.Time `json:"expiresAt"`
}

// MessageDecryptRequest represents a request to decrypt a message
type MessageDecryptRequest struct {
	// Base64 encoded encryption key
	DecryptionKey string `json:"decryptionKey" validate:"required" example:"YWJjZGVmZ2hpams="`
	// Optional passphrase (required if message is passphrase protected)
	Passphrase string `json:"passphrase,omitempty" example:"correct-horse-battery-staple"`
}

// MessageDecryptResponse represents the response to a message decryption
type MessageDecryptResponse struct {
	// Unique identifier for the message
	MessageID string `json:"messageId" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Decrypted content of the message
	Content string `json:"content" example:"This was a secret message"`
	// Current number of times the message has been viewed
	ViewCount int `json:"viewCount" example:"1"`
	// Maximum allowed views for this message
	MaxViewCount int `json:"maxViewCount" example:"5"`
	// When the message was decrypted
	DecryptedAt time.Time `json:"decryptedAt"`
	// When the message will expire
	ExpiresAt *time.Time `json:"expiresAt"`
}

// HealthCheckResponse represents the response to a health check
type HealthCheckResponse struct {
	// Overall status of the service
	Status string `json:"status"    example:"healthy"`
	// Service version
	Version string `json:"version"   example:"1.0.0"`
	// Timestamp of the health check
	Timestamp time.Time `json:"timestamp"`
	// Detailed status of individual components
	Services map[string]string `json:"services"  example:"database:healthy,encryption:healthy"`
}

// APIInfoResponse represents information about the API
type APIInfoResponse struct {
	// API version
	Version string `json:"version" example:"1.0.0"`
	// URL to the API documentation
	Documentation string `json:"documentation" example:"/api/v1/docs"`
	// Map of available API endpoints
	Endpoints map[string]string `json:"endpoints" example:"submit:POST /api/v1/messages"`
	// Map of enabled features
	Features map[string]bool `json:"features" example:"emailNotifications:true"`
}
