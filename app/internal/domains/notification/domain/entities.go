package domain

import (
	"context"
)

// NotificationRequest represents a request to send a notification
type NotificationRequest struct {
	To              string
	From            string  
	FromName        string
	Subject         string
	Body            string
	MessageContent  string
	SenderName      string
	RecipientName   string
	MessageURL      string
	Hidden          string
}

// NotificationResponse represents the result of sending a notification
type NotificationResponse struct {
	Success   bool
	MessageID string
	Error     error
}

// QueueMessage represents a message consumed from the queue
type QueueMessage struct {
	Email           string
	FirstName       string
	OtherFirstName  string
	OtherLastName   string
	OtherEmail      string
	UniqueID        string
	Content         string
	URL             string
	Hidden          string
	Captcha         string
}

// QueueConnection represents connection configuration for message queues
type QueueConnection struct {
	Host     string
	Port     int
	User     string
	Password string
	QueueName string
}

// EmailConnection represents connection configuration for email sending
type EmailConnection struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// NotificationTemplateData represents data for notification templates
type NotificationTemplateData struct {
	Body    string
	Message string
}

// MessageHandler defines the interface for handling queue messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg QueueMessage) error
}

// NotificationSender defines the interface for sending notifications
type NotificationSender interface {
	SendNotification(ctx context.Context, req NotificationRequest) (*NotificationResponse, error)
}

// QueueConsumer defines the interface for consuming from message queues
type QueueConsumer interface {
	StartConsuming(ctx context.Context, queueConn QueueConnection, handler MessageHandler, concurrency int) error
	Connect(ctx context.Context, queueConn QueueConnection) error
	Close() error
}

// TemplateRenderer defines the interface for rendering notification templates
type TemplateRenderer interface {
	RenderTemplate(ctx context.Context, templateName string, data NotificationTemplateData) (string, error)
}