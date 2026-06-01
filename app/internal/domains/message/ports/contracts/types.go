// Package contracts defines shared types and interfaces used across the message domain.
// These types serve as contracts between different layers of the hexagonal architecture,
// ensuring consistent data structures for message processing, encryption, storage, and logging.
package contracts

import (
	"io"
	"time"

	logport "github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/port"
)

// LogEvent is the shared structured-logging event contract. It is an alias to
// the single definition in internal/shared/logging/port so adding a field
// happens in exactly one place across every domain.
type LogEvent = logport.LogEvent

// MessageSubmissionRequest represents a request to share a new encrypted message
type MessageSubmissionRequest struct {
	Content           string
	IsClientEncrypted bool
	SenderName        string
	SenderEmail       string
	RecipientName     string
	RecipientEmail    string
	Passphrase        string
	AdditionalInfo    string
	Captcha           string
	TurnstileToken    string
	SendNotification  bool
	MaxViewCount      int
	ExpirationHours   int
}

// MessageSubmissionResponse represents the response to a message submission
type MessageSubmissionResponse struct {
	MessageID         string
	Key               string
	DecryptURL        string
	IsClientEncrypted bool
	ExpiresAt         *time.Time
	Success           bool
	Error             error
}

// MessageRetrievalRequest represents a request to retrieve and decrypt a message
type MessageRetrievalRequest struct {
	MessageID     string
	DecryptionKey []byte
	Passphrase    string
}

// MessageRetrievalResponse represents the response to a message retrieval
type MessageRetrievalResponse struct {
	MessageID         string
	Content           string
	IsClientEncrypted bool
	ViewCount         int
	MaxViewCount      int
	ExpiresAt         *time.Time
	Success           bool
	Error             error
}

// MessageAccessInfo provides information about message access requirements
type MessageAccessInfo struct {
	MessageID          string
	Exists             bool
	RequiresPassphrase bool
	IsClientEncrypted  bool
	ExpiresAt          *time.Time
}

// MessageStorageRequest represents a request to store an encrypted message
type MessageStorageRequest struct {
	MessageID         string
	Content           string
	IsClientEncrypted bool
	Passphrase        string
	MaxViewCount      int
	RecipientEmail    string
	ExpiresAt         *time.Time
}

// MessageRetrievalStorageRequest represents a request to retrieve a stored message
type MessageRetrievalStorageRequest struct {
	MessageID string
}

// MessageStorageResponse represents a stored message from storage
type MessageStorageResponse struct {
	MessageID         string
	EncryptedContent  string
	IsClientEncrypted bool
	HashedPassphrase  string
	HasPassphrase     bool
	ViewCount         int
	MaxViewCount      int
	ExpiresAt         *time.Time
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

// FileUploadSession represents an in-progress or completed chunked file upload session.
type FileUploadSession struct {
	// SessionID identifies the upload session tracked across chunk requests.
	SessionID string
	// FileID identifies the stored encrypted object.
	FileID string
	// MessageID links the uploaded file back to its parent message.
	MessageID string
	// UploadID is the object storage provider's multipart upload identifier.
	UploadID string
	// Filename is the original client-provided filename.
	Filename string
	// ContentType is the client-provided media type for the file.
	ContentType string
	// TotalSize is the total number of plaintext bytes expected for the upload.
	TotalSize int64
	// TotalChunks is the expected number of chunks for the upload.
	TotalChunks int
	// CompletedParts holds the parts successfully uploaded so far.
	CompletedParts []CompletedPart
	// Status tracks whether the upload is active, complete, or aborted.
	Status string
	// EncryptionKey stores the symmetric key used for chunk encryption.
	EncryptionKey []byte
	// CreatedAt records when the upload session was created.
	CreatedAt time.Time
	// ExpiresAt records when the upload session should no longer be accepted.
	ExpiresAt time.Time
}

// CompletedPart represents a successfully uploaded part in a multipart upload.
type CompletedPart struct {
	// PartNumber identifies the uploaded multipart part.
	PartNumber int
	// ETag stores the object storage checksum token returned for the part.
	ETag string
}

// InitiateUploadRequest is the request to start a new chunked file upload.
type InitiateUploadRequest struct {
	// Filename is the client-provided filename for the upload.
	Filename string
	// ContentType is the client-provided MIME type for the upload.
	ContentType string
	// TotalSize is the full plaintext size of the file.
	TotalSize int64
	// ChunkSize is the expected plaintext size for each uploaded chunk.
	ChunkSize int64
	// MessageID identifies the parent message that owns the file.
	MessageID string
}

// InitiateUploadResponse is returned after successfully initiating an upload.
type InitiateUploadResponse struct {
	// FileID identifies the file object created for the upload.
	FileID string
	// SessionID identifies the resumable upload session.
	SessionID string
	// EncryptionKey is the symmetric key used to encrypt the file chunks.
	// The client must store this key and supply it when constructing the download
	// URL so that the recipient can authenticate and decrypt the file.
	EncryptionKey []byte
}

// UploadChunkRequest is the request to upload a single chunk.
type UploadChunkRequest struct {
	// FileID identifies the file object receiving the chunk.
	FileID string
	// SessionID identifies the upload session authorizing the chunk.
	SessionID string
	// ChunkIndex is the one-based position of the chunk in the upload.
	ChunkIndex int
	// TotalChunks is the client-reported total number of chunks.
	TotalChunks int
	// Data contains the plaintext chunk bytes before encryption.
	Data []byte
}

// UploadChunkResponse is returned after successfully uploading a chunk.
type UploadChunkResponse struct {
	// FileID identifies the file object that received the chunk.
	FileID string
	// ChunkIndex is the one-based position of the chunk that was accepted.
	ChunkIndex int
	// Done indicates whether the chunk completed the multipart upload.
	Done bool
}

// DownloadFileRequest is the request to download and decrypt a file.
type DownloadFileRequest struct {
	// FileID identifies the stored encrypted file object.
	FileID string
	// Key is the symmetric decryption key supplied by the client.
	Key []byte
}

// DownloadFileResponse contains the decrypted file content.
type DownloadFileResponse struct {
	// Filename is the original filename associated with the file.
	Filename string
	// ContentType is the media type associated with the file.
	ContentType string
	// Data streams decrypted file bytes to the caller.
	Data io.ReadCloser
}
