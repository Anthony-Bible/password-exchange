package secondary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
)

// NotificationServicePort defines the secondary port for notification operations
type NotificationServicePort interface {
	// SendMessageNotification sends a notification about a new message
	SendMessageNotification(ctx context.Context, req domain.MessageNotificationRequest) error
}