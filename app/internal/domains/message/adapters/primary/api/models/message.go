package models

import (
	"time"
)

// MessageSubmissionRequest represents a REST API request to submit a new message
type MessageSubmissionRequest struct {
	// Required. The message content to encrypt and store. 1–10000 characters.
	Content string `json:"content" validate:"required,min=1"`
	// Optional. Set to true if the content is already client-side encrypted (E2E mode).
	// When true, the server stores the ciphertext as-is without re-encrypting.
	IsClientEncrypted bool `json:"isClientEncrypted,omitempty"`
	// Optional. Base64url-encoded client-generated AES key for E2E encrypted messages.
	// Required when isClientEncrypted is true and sendNotification is true so the
	// email link includes the #key fragment. Not stored server-side. Max 512 characters.
	E2EKey string `json:"e2eKey,omitempty" validate:"max=512"`
	// Optional. Sender details. Required when sendNotification is true.
	Sender *Sender `json:"sender,omitempty"`
	// Optional. Recipient details. Required when sendNotification is true.
	Recipient *Recipient `json:"recipient,omitempty"`
	// Optional. Passphrase that the recipient must supply before decrypting. Max 500 characters.
	Passphrase string `json:"passphrase,omitempty" validate:"max=500"`
	// Optional. Free-form text included in the notification email body.
	AdditionalInfo string `json:"additionalInfo,omitempty"`
	// Optional. Set to true to send an email notification to the recipient.
	// Requires sender and recipient to be set.
	SendNotification bool `json:"sendNotification"`
	// Optional. Answer to the anti-spam question identified by questionId.
	// Required when sendNotification is true.
	AntiSpamAnswer string `json:"antiSpamAnswer,omitempty"`
	// Optional. ID of the anti-spam question the user answered.
	// Required when sendNotification is true.
	QuestionID *int `json:"questionId,omitempty"`
	// Optional. Maximum number of times the message may be viewed before it is deleted. 0 means unlimited.
	MaxViewCount int `json:"maxViewCount,omitempty" validate:"min=0,max=100"`
	// Optional. Cloudflare Turnstile token for bot protection. Max 2048 characters.
	TurnstileToken string `json:"turnstileToken,omitempty" validate:"max=2048"`
	// Optional. Custom expiration in hours. When 0 or omitted the server default (168 h / 7 days) applies.
	// Valid range: 1–2160 (1 hour to 90 days).
	ExpirationHours int `json:"expirationHours,omitempty" validate:"min=0,max=2160"`
}

// Sender represents sender information for message submission
type Sender struct {
	// Required. Sender's display name. 1–100 characters.
	Name string `json:"name" validate:"required,min=1,max=100"`
	// Required. Sender's email address.
	Email string `json:"email" validate:"required,email"`
}

// Recipient represents recipient information for message submission
type Recipient struct {
	// Optional. Recipient's display name. 1–100 characters.
	Name string `json:"name" validate:"min=1,max=100"`
	// Required. Recipient's email address.
	Email string `json:"email" validate:"required,email"`
}

// MessageSubmissionResponse represents the response to a message submission
type MessageSubmissionResponse struct {
	// Unique ID of the created message.
	MessageID string `json:"messageId"`
	// Full URL the recipient should visit to decrypt the message.
	DecryptURL string `json:"decryptUrl"`
	// Base64url-encoded decryption key. Include this in the share URL fragment so it never reaches the server.
	Key string `json:"key"`
	// True when the stored content is client-side encrypted (E2E mode).
	IsClientEncrypted bool `json:"isClientEncrypted"`
	// Web UI URL for the message (without the key fragment).
	WebURL string `json:"webUrl"`
	// Time at which the message will be automatically deleted. Null for legacy messages.
	ExpiresAt *time.Time `json:"expiresAt"`
	// True if an email notification was successfully sent to the recipient.
	NotificationSent bool `json:"notificationSent"`
}

// MessageAccessInfoResponse represents information about message access requirements
type MessageAccessInfoResponse struct {
	// Unique ID of the message.
	MessageID string `json:"messageId"`
	// True if the message exists and has not yet been consumed.
	Exists bool `json:"exists"`
	// True if the message requires a passphrase before it can be decrypted.
	RequiresPassphrase bool `json:"requiresPassphrase"`
	// True when the stored content is client-side encrypted (E2E mode); decryption happens in the browser.
	IsClientEncrypted bool `json:"isClientEncrypted"`
	// True if the message has already been viewed at least once.
	HasBeenAccessed bool `json:"hasBeenAccessed"`
	// Time at which the message will be automatically deleted. Null for legacy messages.
	ExpiresAt *time.Time `json:"expiresAt"`
}

// MessageDecryptRequest represents a request to decrypt a message.
type MessageDecryptRequest struct {
	// Optional. Base64url-encoded decryption key. Omit for client-encrypted (E2E) messages
	// where decryption happens client-side. Max 4096 characters.
	DecryptionKey string `json:"decryptionKey,omitempty" validate:"max=4096"`
	// Optional. Passphrase required if the message was created with one.
	Passphrase string `json:"passphrase,omitempty"`
}

// MessageDecryptResponse represents the response to a message decryption
type MessageDecryptResponse struct {
	// Unique ID of the message.
	MessageID string `json:"messageId"`
	// Decrypted message content. For E2E messages this is the raw ciphertext; the client decrypts it.
	Content string `json:"content"`
	// True when the content is client-side encrypted (E2E mode).
	IsClientEncrypted bool `json:"isClientEncrypted"`
	// Number of times this message has been viewed.
	ViewCount int `json:"viewCount"`
	// Maximum allowed views before deletion. 0 means unlimited.
	MaxViewCount int `json:"maxViewCount"`
	// Timestamp of this decryption.
	DecryptedAt time.Time `json:"decryptedAt"`
	// Time at which the message will be automatically deleted. Null for legacy messages.
	ExpiresAt *time.Time `json:"expiresAt"`
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
