package secondary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
)

// QueueConsumerPort defines the secondary port for message queue operations
type QueueConsumerPort interface {
	// Connect establishes a connection to the message queue
	Connect(ctx context.Context, queueConn domain.QueueConnection) error
	
	// StartConsuming starts consuming messages from the queue
	StartConsuming(ctx context.Context, queueConn domain.QueueConnection, handler domain.MessageHandler, concurrency int) error
	
	// Close closes the queue connection
	Close() error
}