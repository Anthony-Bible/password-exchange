package domain

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
)

// Type aliases to contracts and secondary ports - these define the domain's data contracts
type (
	NotificationRequest      = contracts.NotificationRequest
	NotificationResponse     = contracts.NotificationResponse
	QueueMessage            = contracts.QueueMessage
	QueueConnection         = contracts.QueueConnection
	EmailConnection         = contracts.EmailConnection
	NotificationTemplateData = contracts.NotificationTemplateData
	UnviewedMessage         = contracts.UnviewedMessage
	ReminderLogEntry        = contracts.ReminderLogEntry
	ReminderRequest         = contracts.ReminderRequest
	MessageHandler          = contracts.MessageHandler
	LogEvent                = contracts.LogEvent
)


// ReminderConfig holds configuration for reminder processing.
// Defines when and how often reminder emails are sent for unviewed messages.
type ReminderConfig struct {
	Enabled         bool // Whether reminder system is active
	CheckAfterHours int  // Hours to wait before first reminder (1-8760)
	MaxReminders    int  // Maximum reminders per message (1-10)
	Interval        int  // Hours between subsequent reminders (1-720)
}

