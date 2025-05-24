package secondary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
)

// EmailSenderPort defines the secondary port for email sending operations
type EmailSenderPort interface {
	// SendEmail sends an email notification
	SendEmail(ctx context.Context, req domain.NotificationRequest) (*domain.NotificationResponse, error)
}