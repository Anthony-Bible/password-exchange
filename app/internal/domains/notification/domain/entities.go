package domain

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


// ReminderConfig holds configuration for reminder processing.
// Defines when and how often reminder emails are sent for unviewed messages.
type ReminderConfig struct {
	Enabled         bool // Whether reminder system is active
	CheckAfterHours int  // Hours to wait before first reminder (1-8760)
	MaxReminders    int  // Maximum reminders per message (1-10)
	Interval        int  // Hours between subsequent reminders (1-720)
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

// UnviewedMessage represents a message that hasn't been viewed and may need reminders
type UnviewedMessage struct {
	MessageID      int
	UniqueID       string
	RecipientEmail string
	DaysOld        int
	Created        time.Time
}

// ReminderLogEntry represents a logged reminder attempt
type ReminderLogEntry struct {
	MessageID      int
	RecipientEmail string
	ReminderCount  int
	SentAt         time.Time
}