package domain

import "errors"

var (
	// ErrEmptyContent is returned when trying to store a message with empty content
	ErrEmptyContent = errors.New("message content cannot be empty")
	
	// ErrEmptyUniqueID is returned when unique ID is empty
	ErrEmptyUniqueID = errors.New("unique ID cannot be empty")
	
	// ErrEmptyRecipientEmail is returned when recipient email is empty
	ErrEmptyRecipientEmail = errors.New("recipient email cannot be empty")
	
	// ErrEmptyEmailAddress is returned when email address is empty
	ErrEmptyEmailAddress = errors.New("email address cannot be empty")
	
	// ErrInvalidParameter is returned when a parameter is invalid
	ErrInvalidParameter = errors.New("invalid parameter")
	
	// ErrInvalidMaxViewCount is returned when max view count is invalid
	ErrInvalidMaxViewCount = errors.New("max view count must be between 1 and 100")
	
	// ErrMessageNotFound is returned when a message is not found in storage
	ErrMessageNotFound = errors.New("message not found")
	
	// ErrDatabaseConnection is returned when database connection fails
	ErrDatabaseConnection = errors.New("database connection failed")
	
	// ErrDatabaseOperation is returned when database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")
)