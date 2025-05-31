package secondary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
)

// QueuePort defines the secondary port for message queue operations
type QueuePort interface {
	// Connect establishes a connection to the message queue
	Connect(ctx context.Context, queueConn contracts.QueueConnection) error
	
	// StartConsuming starts consuming messages from the queue
	StartConsuming(ctx context.Context, queueConn contracts.QueueConnection, handler contracts.MessageHandler, concurrency int) error
	
	// Close closes the queue connection
	Close() error
}