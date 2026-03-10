package secondary

import (
	"context"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
)

// NotificationServicePort defines the secondary port for notification operations
type NotificationServicePort interface {
	// SendMessageNotification sends a notification about a new message
	SendMessageNotification(ctx context.Context, req domain.MessageNotificationRequest) error

	// CheckHealth returns the health status of the email service
	CheckHealth(ctx context.Context) (string, error)
}
