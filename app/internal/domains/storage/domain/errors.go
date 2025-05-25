package domain

import "errors"

var (
	// ErrEmptyContent is returned when trying to store a message with empty content
	ErrEmptyContent = errors.New("message content cannot be empty")
	
	// ErrEmptyUniqueID is returned when unique ID is empty
	ErrEmptyUniqueID = errors.New("unique ID cannot be empty")
	
	// ErrInvalidMaxViewCount is returned when max view count is invalid
	ErrInvalidMaxViewCount = errors.New("max view count must be greater than 0")
	
	// ErrMessageNotFound is returned when a message is not found in storage
	ErrMessageNotFound = errors.New("message not found")
	
	// ErrDatabaseConnection is returned when database connection fails
	ErrDatabaseConnection = errors.New("database connection failed")
	
	// ErrDatabaseOperation is returned when database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")
)