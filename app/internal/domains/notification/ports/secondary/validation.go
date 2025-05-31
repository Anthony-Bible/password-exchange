package secondary

// ValidationPort defines the secondary port for validation operations
type ValidationPort interface {
	ValidateEmail(email string) error
	SanitizeEmailForLogging(email string) string
}