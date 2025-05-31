package secondary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
)

// EmailSenderPort defines the secondary port for email sending operations
type EmailSenderPort interface {
	// SendNotification sends an email notification
	SendNotification(ctx context.Context, req contracts.NotificationRequest) (*contracts.NotificationResponse, error)
}