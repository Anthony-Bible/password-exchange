package secondary

import (
	"context"
)

// HealthCheckPort defines the secondary port for health checking operations
type HealthCheckPort interface {
	// CheckDatabase returns the health status of the database service
	CheckDatabase(ctx context.Context) (string, error)

	// CheckEncryption returns the health status of the encryption service
	CheckEncryption(ctx context.Context) (string, error)

	// CheckEmail returns the health status of the email service
	CheckEmail(ctx context.Context) (string, error)
}
