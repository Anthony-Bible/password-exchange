package validator

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
)

// ValidationAdapter implements ValidationPort using the pkg/validation package
type ValidationAdapter struct{}

// NewValidationAdapter creates a new validation adapter
func NewValidationAdapter() secondary.ValidationPort {
	return &ValidationAdapter{}
}

func (v *ValidationAdapter) ValidateEmail(email string) error {
	return validation.ValidateEmail(email)
}

func (v *ValidationAdapter) SanitizeEmailForLogging(email string) string {
	return validation.SanitizeEmailForLogging(email)
}