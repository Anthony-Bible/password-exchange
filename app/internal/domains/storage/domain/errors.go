package domain

import "errors"

var (
	// ErrNilMessage is returned when a nil message pointer is passed to StoreMessage
	ErrNilMessage = errors.New("message cannot be nil")

	// ErrEmptyContent is returned when trying to store a message with empty content
	ErrEmptyContent = errors.New("message content cannot be empty")
	
	// ErrEmptyUniqueID is returned when unique ID is empty
	ErrEmptyUniqueID = errors.New("unique ID cannot be empty")
	
	// ErrEmptyEmailAddress is returned when email address is empty
	ErrEmptyEmailAddress = errors.New("email address cannot be empty")
	
	// ErrInvalidParameter is returned when a parameter is invalid
	ErrInvalidParameter = errors.New("invalid parameter")
	
	// ErrInvalidMaxViewCount is returned when max view count is below the
	// structural minimum of 1. The upper bound is a product policy enforced by
	// the message domain (message.AbsoluteMaxViewCount), not by storage.
	ErrInvalidMaxViewCount = errors.New("max view count must be at least 1")
	
	// ErrMessageNotFound is returned when a message is not found in storage
	ErrMessageNotFound = errors.New("message not found")
	
	// ErrDatabaseConnection is returned when database connection fails
	ErrDatabaseConnection = errors.New("database connection failed")
	
	// ErrDatabaseOperation is returned when database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")

	// ErrRepositoryClosed is returned when a MessageRepository operation is
	// attempted after Close() has been called.
	ErrRepositoryClosed = errors.New("storage repository is closed")

	// ErrUploadSessionNotFound is returned when an upload session cannot be
	// found by the given session_id or file_id.
	ErrUploadSessionNotFound = errors.New("upload session not found")

	// ErrUploadSessionAlreadyExists is returned when CreateSession is called
	// with a session_id that already exists in persistent state.
	ErrUploadSessionAlreadyExists = errors.New("upload session already exists")
)