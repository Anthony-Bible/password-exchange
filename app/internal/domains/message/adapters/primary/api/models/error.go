package models

import (
	"time"
)

// StandardErrorResponse follows RFC 7807 Problem Details format
type StandardErrorResponse struct {
	Error     string                 `json:"error"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Path      string                 `json:"path"`
}

// ValidationErrorDetails represents validation error details
type ValidationErrorDetails struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// NewStandardError creates a new standard error response
func NewStandardError(errorCode, message, path string) *StandardErrorResponse {
	return &StandardErrorResponse{
		Error:     errorCode,
		Message:   message,
		Timestamp: time.Now(),
		Path:      path,
	}
}

// NewValidationError creates a new validation error response
func NewValidationError(path string, details map[string]interface{}) *StandardErrorResponse {
	return &StandardErrorResponse{
		Error:     "validation_failed",
		Message:   "Request validation failed",
		Details:   details,
		Timestamp: time.Now(),
		Path:      path,
	}
}

// Common error codes
const (
	ErrorCodeValidationFailed   = "validation_failed"
	ErrorCodeMessageNotFound    = "message_not_found"
	ErrorCodeInvalidPassphrase  = "invalid_passphrase"
	ErrorCodeMessageConsumed    = "message_consumed"
	ErrorCodeRateLimitExceeded  = "rate_limit_exceeded"
	ErrorCodeInternalError      = "internal_error"
	ErrorCodeServiceUnavailable = "service_unavailable"
	ErrorCodeTimeout            = "request_timeout"
)