package domain

import (
	"context"
)

// MessageSubmissionRequest represents a request to submit a new message
type MessageSubmissionRequest struct {
	Content          string
	SenderName       string
	SenderEmail      string
	RecipientName    string
	RecipientEmail   string
	Passphrase       string
	AdditionalInfo   string
	Captcha          string
	SendNotification bool
	SkipEmail        bool
	MaxViewCount     int
}

// MessageSubmissionResponse represents the response to a message submission
type MessageSubmissionResponse struct {
	MessageID  string
	Key        string
	DecryptURL string
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
	Success      bool
	Error        error
}

// MessageAccessInfo provides information about message access requirements
type MessageAccessInfo struct {
	MessageID          string
	Exists             bool
	RequiresPassphrase bool
}

// MessageStorageRequest represents a request to store an encrypted message
type MessageStorageRequest struct {
	MessageID      string
	Content        string
	Passphrase     string
	MaxViewCount   int
	RecipientEmail string // Optional, for notification purposes
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

// EncryptionService defines the interface for encryption operations
type EncryptionService interface {
	GenerateKey(ctx context.Context, length int32) ([]byte, error)
	Encrypt(ctx context.Context, plaintext []string, key []byte) ([]string, error)
	Decrypt(ctx context.Context, ciphertext []string, key []byte) ([]string, error)
	GenerateID(ctx context.Context) (string, error)
}

// StorageService defines the interface for message storage operations
type StorageService interface {
	StoreMessage(ctx context.Context, req MessageStorageRequest) error
	RetrieveMessage(ctx context.Context, req MessageRetrievalStorageRequest) (*MessageStorageResponse, error)
	GetMessage(ctx context.Context, req MessageRetrievalStorageRequest) (*MessageStorageResponse, error)
}

// NotificationService defines the interface for notification operations
type NotificationService interface {
	SendMessageNotification(ctx context.Context, req MessageNotificationRequest) error
}

// PasswordHasher defines the interface for password hashing operations
type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Verify(ctx context.Context, password, hash string) (bool, error)
}

// URLBuilder defines the interface for building message URLs
type URLBuilder interface {
	BuildDecryptURL(messageID string, encryptionKey []byte) string
}

// WebRenderer defines the interface for rendering web responses
type WebRenderer interface {
	RenderTemplate(ctx context.Context, templateName string, data interface{}) error
	RenderJSON(ctx context.Context, data interface{}) error
	Redirect(ctx context.Context, url string) error
}
