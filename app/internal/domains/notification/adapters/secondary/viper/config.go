package viper

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
)

// ViperConfigAdapter implements ConfigPort using static configuration values
type ViperConfigAdapter struct {
	emailTemplate                     string
	serverEmail                       string
	serverName                        string
	passwordExchangeURL               string
	initialNotificationSubject        string
	reminderNotificationSubject       string
	initialNotificationBodyTemplate   string
	reminderNotificationBodyTemplate  string
	reminderEmailTemplate             string
	reminderMessageContent            string
}

// NewViperConfigAdapter creates a new configuration adapter with default values
func NewViperConfigAdapter() secondary.ConfigPort {
	return &ViperConfigAdapter{
		emailTemplate:                     "/templates/email_template.html",
		serverEmail:                       "server@password.exchange",
		serverName:                        "Password Exchange",
		passwordExchangeURL:               "https://password.exchange",
		initialNotificationSubject:        "Encrypted Message from Password Exchange from %s",
		reminderNotificationSubject:       "Reminder: You have an unviewed encrypted message (Reminder #%d)",
		initialNotificationBodyTemplate:   "Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about",
		reminderNotificationBodyTemplate:  "",
		reminderEmailTemplate:             "/templates/reminder_email_template.html",
		reminderMessageContent:            "Please check your original email for the secure decrypt link. For security reasons, the decrypt link cannot be included in reminder emails. If you cannot find the original email, please contact the sender to resend the message.",
	}
}

// NewViperConfigAdapterWithValues creates a new configuration adapter with custom values
func NewViperConfigAdapterWithValues(emailTemplate, serverEmail, serverName, passwordExchangeURL string) secondary.ConfigPort {
	return &ViperConfigAdapter{
		emailTemplate:                     emailTemplate,
		serverEmail:                       serverEmail,
		serverName:                        serverName,
		passwordExchangeURL:               passwordExchangeURL,
		initialNotificationSubject:        "Encrypted Message from Password Exchange from %s",
		reminderNotificationSubject:       "Reminder: You have an unviewed encrypted message (Reminder #%d)",
		initialNotificationBodyTemplate:   "Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about",
		reminderNotificationBodyTemplate:  "",
		reminderEmailTemplate:             "/templates/reminder_email_template.html",
		reminderMessageContent:            "Please check your original email for the secure decrypt link. For security reasons, the decrypt link cannot be included in reminder emails. If you cannot find the original email, please contact the sender to resend the message.",
	}
}

func (v *ViperConfigAdapter) GetEmailTemplate() string {
	return v.emailTemplate
}

func (v *ViperConfigAdapter) GetServerEmail() string {
	return v.serverEmail
}

func (v *ViperConfigAdapter) GetServerName() string {
	return v.serverName
}

func (v *ViperConfigAdapter) GetPasswordExchangeURL() string {
	return v.passwordExchangeURL
}

func (v *ViperConfigAdapter) GetInitialNotificationSubject() string {
	return v.initialNotificationSubject
}

func (v *ViperConfigAdapter) GetReminderNotificationSubject() string {
	return v.reminderNotificationSubject
}

func (v *ViperConfigAdapter) GetInitialNotificationBodyTemplate() string {
	return v.initialNotificationBodyTemplate
}

func (v *ViperConfigAdapter) GetReminderNotificationBodyTemplate() string {
	return v.reminderNotificationBodyTemplate
}

func (v *ViperConfigAdapter) GetReminderEmailTemplate() string {
	return v.reminderEmailTemplate
}

func (v *ViperConfigAdapter) GetReminderMessageContent() string {
	return v.reminderMessageContent
}