package secondary

// ValidationPort defines the secondary port for validation operations in the notification domain.
// This interface abstracts validation logic from the domain, allowing for flexible validation
// rules and sanitization strategies. It ensures that notification-related data meets security
// and format requirements before processing.
type ValidationPort interface {
	// ValidateEmail checks if an email address is valid according to the implementation's rules.
	// This method should verify both the format (RFC compliance) and any business-specific
	// requirements (e.g., allowed domains, blacklists).
	//
	// Parameters:
	//   - email: The email address to validate
	//
	// Returns:
	//   - nil if the email is valid
	//   - An error describing why the email is invalid
	//
	// Common validation checks:
	//   - RFC 5322 format compliance
	//   - Domain existence (optional)
	//   - Blacklist/whitelist rules
	//   - Length constraints
	ValidateEmail(email string) error

	// SanitizeEmailForLogging prepares an email address for safe logging.
	// This method should redact or mask sensitive portions of the email to prevent
	// PII exposure in logs while maintaining enough information for debugging.
	//
	// Parameters:
	//   - email: The email address to sanitize
	//
	// Returns:
	//   - A sanitized version of the email safe for logging
	//
	// Example transformations:
	//   - "user@example.com" -> "u***@example.com"
	//   - "john.doe@company.org" -> "j***.***@company.org"
	//   - Invalid emails might return "[INVALID_EMAIL]"
	SanitizeEmailForLogging(email string) string
}