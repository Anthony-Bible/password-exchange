package domain

import "errors"

var (
	// ErrInvalidMessageRequest indicates the message request is invalid
	ErrInvalidMessageRequest = errors.New("invalid message request")
	
	// ErrInvalidEmailAddress indicates the email address is malformed
	ErrInvalidEmailAddress = errors.New("invalid email address")
	
	// ErrEncryptionFailed indicates encryption operation failed
	ErrEncryptionFailed = errors.New("encryption failed")
	
	// ErrDecryptionFailed indicates decryption operation failed
	ErrDecryptionFailed = errors.New("decryption failed")
	
	// ErrStorageFailed indicates message storage failed
	ErrStorageFailed = errors.New("message storage failed")
	
	// ErrMessageNotFound indicates the requested message was not found
	ErrMessageNotFound = errors.New("message not found")
	
	// ErrPasswordHashFailed indicates password hashing failed
	ErrPasswordHashFailed = errors.New("password hashing failed")
	
	// ErrPasswordVerificationFailed indicates password verification failed
	ErrPasswordVerificationFailed = errors.New("password verification failed")
	
	// ErrInvalidPassphrase indicates the provided passphrase is incorrect
	ErrInvalidPassphrase = errors.New("invalid passphrase")
	
	// ErrGenerateIDFailed indicates ID generation failed
	ErrGenerateIDFailed = errors.New("failed to generate unique ID")
	
	// ErrDecodingFailed indicates content decoding failed
	ErrDecodingFailed = errors.New("content decoding failed")
	
	// ErrNotificationFailed indicates notification sending failed
	ErrNotificationFailed = errors.New("notification sending failed")
	
	// ErrTemplateRenderFailed indicates template rendering failed
	ErrTemplateRenderFailed = errors.New("template rendering failed")
)