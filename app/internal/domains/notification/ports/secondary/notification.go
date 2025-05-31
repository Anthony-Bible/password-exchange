package secondary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
)

// NotificationPort defines the secondary port for publishing notification messages to a queue
type NotificationPort interface {
	// PublishNotification publishes a notification message to the queue for processing
	PublishNotification(ctx context.Context, req contracts.NotificationRequest) error
}