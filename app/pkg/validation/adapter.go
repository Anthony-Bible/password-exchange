package validation

// Adapter is the shared secondary-port implementation for email validation and
// sanitization. It wraps the package-level ValidateEmail and
// SanitizeEmailForLogging functions so that any domain whose ValidationPort
// requires those two methods (notification, storage) can reuse a single
// adapter instead of declaring a near-verbatim copy in its own adapters tree.
//
// It is returned as a concrete value rather than an interface so a single
// instance satisfies each domain's ValidationPort directly.
type Adapter struct{}

// NewAdapter returns a validation adapter backed by the package-level
// validation functions.
func NewAdapter() Adapter { return Adapter{} }

// ValidateEmail reports whether email is a valid address.
func (Adapter) ValidateEmail(email string) error { return ValidateEmail(email) }

// SanitizeEmailForLogging redacts email for safe inclusion in logs.
func (Adapter) SanitizeEmailForLogging(email string) string {
	return SanitizeEmailForLogging(email)
}
