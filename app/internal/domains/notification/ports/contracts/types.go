// Package contracts defines shared types and interfaces used across the notification domain.
// These types serve as contracts between different layers of the hexagonal architecture,
// ensuring consistent data structures for notification processing, queue operations,
// and logging abstractions.
package contracts

import (
	"context"
	"html/template"
	"time"

	logport "github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/port"
)

// NotificationRequest represents a request to send a notification email.
// This struct contains all the information needed to construct and send
// a notification email, including recipient details, message content, and
// formatting information.
type NotificationRequest struct {
	To             string
	From           string
	FromName       string
	Subject        string
	MessageContent string
	SenderName     string
	RecipientName  string
	MessageURL     string
	Hidden         string
}

// NotificationResponse represents the result of a notification send operation.
// It provides feedback about whether the notification was successfully sent,
// along with any relevant tracking information or error details.
type NotificationResponse struct {
	Success   bool
	MessageID string
	Error     error
}

// QueueConnection represents the configuration needed to establish a connection
// to a message queue system (e.g., RabbitMQ). This struct encapsulates all
// connection parameters required for queue operations.
type QueueConnection struct {
	Host      string
	Port      int
	User      string
	Password  string
	QueueName string
}

// NotificationTemplateData represents the data structure passed to email templates
// for rendering. This struct provides the dynamic content that will be inserted
// into email template placeholders.
type NotificationTemplateData struct {
	Message       template.HTML
	SenderName    string
	RecipientName string
	MessageURL    string
}

// UnviewedMessage represents a message that has been sent but not yet viewed by
// the recipient. This struct is used by the reminder system to identify messages
// that may need follow-up notifications.
type UnviewedMessage struct {
	MessageID      int
	UniqueID       string
	RecipientEmail string
	DaysOld        int
	Created        time.Time
}

// ReminderLogEntry represents a record of a reminder notification that has been sent.
// This struct is used for tracking reminder history and preventing duplicate
// notifications within configured time windows.
type ReminderLogEntry struct {
	MessageID      int
	RecipientEmail string
	ReminderCount  int
	SentAt         time.Time
}

// QueueMessage represents a message received from the notification queue.
// This struct contains all the information needed to process a queued
// notification request, including sender and recipient details, message
// content, and security information.
type QueueMessage struct {
	Email          string
	FirstName      string
	OtherFirstName string
	OtherLastName  string
	OtherEmail     string
	UniqueID       string
	Content        string
	URL            string
	Hidden         string
	Captcha        string
}

// MessageHandler defines the interface for processing messages received from the notification queue.
// Implementations of this interface are responsible for transforming queue messages into
// actual notifications (emails) and handling any errors that occur during processing.
type MessageHandler interface {
	// HandleMessage processes a single message from the notification queue.
	// This method should transform the queue message into an email notification
	// and send it to the appropriate recipient.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - msg: The queue message containing notification details
	//
	// Returns:
	//   - nil if the message was processed successfully
	//   - An error if processing failed (e.g., invalid email, SMTP failure)
	//
	// The implementation should handle:
	//   - Email validation
	//   - Template rendering
	//   - SMTP communication
	//   - Error recovery and logging
	HandleMessage(ctx context.Context, msg QueueMessage) error
}

// LogEvent is the shared structured-logging event contract. It is an alias to
// the single definition in internal/shared/logging/port, so the notification
// domain stays independent of any concrete logging implementation while a new
// field is added in exactly one place across every domain.
type LogEvent = logport.LogEvent

// EmailConnection represents the configuration needed to establish a connection
// to an email server (SMTP). This struct encapsulates all connection parameters
// and authentication credentials required for sending emails.
type EmailConnection struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// ReminderRequest represents a request to send a reminder notification for an unviewed message.
// This struct contains all the information needed to construct a reminder email,
// including message identification, recipient details, and timing information.
type ReminderRequest struct {
	MessageID      int
	UniqueID       string
	RecipientEmail string
	DaysOld        int
	ReminderNumber int
	DecryptionURL  string
}
