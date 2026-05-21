package validation

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
)

// Adapter implements ValidationPort using the shared validation package
type Adapter struct{}

// NewAdapter creates a new validation adapter
func NewAdapter() secondary.ValidationPort {
	return &Adapter{}
}

func (a *Adapter) SanitizeEmailForLogging(email string) string {
	return validation.SanitizeEmailForLogging(email)
}
