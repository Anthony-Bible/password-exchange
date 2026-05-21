package secondary

// ValidationPort defines the secondary port for validation operations.
type ValidationPort interface {
	// SanitizeEmailForLogging prepares an email address for safe logging.
	SanitizeEmailForLogging(email string) string
}
