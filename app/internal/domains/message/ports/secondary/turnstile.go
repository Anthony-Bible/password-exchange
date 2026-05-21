package secondary

import (
	"context"
)

// TurnstileValidatorPort defines the secondary port for Cloudflare Turnstile validation.
type TurnstileValidatorPort interface {
	// ValidateToken validates a Cloudflare Turnstile token.
	ValidateToken(ctx context.Context, token string, remoteIP string) (bool, error)
}
