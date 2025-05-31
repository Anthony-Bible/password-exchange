package contracts

import (
	"context"
	"time"
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

// NotificationResponse represents a response from sending a notification
type NotificationResponse struct {
	Success   bool
	MessageID string
	Error     error
}

// QueueConnection represents configuration for queue connections
type QueueConnection struct {
	Host     string
	Port     int
	User     string
	Password string
	QueueName string
}

// NotificationTemplateData represents data for email templates
type NotificationTemplateData struct {
	Body    string
	Message string
}

// UnviewedMessage represents an unviewed message for reminders
type UnviewedMessage struct {
	MessageID      int
	UniqueID       string
	RecipientEmail string
	DaysOld        int
	Created        time.Time
}

// ReminderLogEntry represents a logged reminder
type ReminderLogEntry struct {
	MessageID      int
	RecipientEmail string
	ReminderCount  int
	SentAt         time.Time
}

// QueueMessage represents a message received from the queue
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

// MessageHandler defines the interface for handling queue messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg QueueMessage) error
}

// LogEvent represents a logging event that can be enriched with context
type LogEvent interface {
	Err(error) LogEvent
	Str(string, string) LogEvent
	Int(string, int) LogEvent
	Bool(string, bool) LogEvent
	Dur(string, time.Duration) LogEvent
	Float64(string, float64) LogEvent
	Msg(string)
}

// EmailConnection represents configuration for email connections
type EmailConnection struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// ReminderRequest represents a request to send a reminder notification
type ReminderRequest struct {
	MessageID      int
	UniqueID       string
	RecipientEmail string
	DaysOld        int
	ReminderNumber int
	DecryptionURL  string
}