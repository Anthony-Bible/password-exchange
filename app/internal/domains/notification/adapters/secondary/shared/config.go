package shared

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/validation"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
)

// SharedConfigAdapter provides a shared implementation of the ConfigPort interface
// using the existing PassConfig structure. This eliminates code duplication between
// the email and reminder commands while maintaining consistent configuration behavior.
type SharedConfigAdapter struct {
	config    config.PassConfig
	validator *validation.ConfigValidator
}

// NewSharedConfigAdapter creates a new shared config adapter with the provided configuration.
//
// Parameters:
//   - cfg: The PassConfig containing email and service configuration
//
// Returns:
//   - A ConfigPort implementation using the shared adapter
func NewSharedConfigAdapter(cfg config.PassConfig) secondary.ConfigPort {
	return &SharedConfigAdapter{
		config:    cfg,
		validator: validation.NewConfigValidator(),
	}
}

// GetEmailTemplate returns the default email template path.
func (c *SharedConfigAdapter) GetEmailTemplate() string {
	return "/templates/email_template.html"
}

// GetServerEmail returns the server email address from configuration.
// Falls back to default if not configured.
func (c *SharedConfigAdapter) GetServerEmail() string {
	if c.config.EmailFrom != "" {
		return c.config.EmailFrom
	}
	return "server@password.exchange"
}

// GetServerName returns the display name for the email server.
func (c *SharedConfigAdapter) GetServerName() string {
	return "Password Exchange"
}

// GetPasswordExchangeURL returns the password exchange service URL.
// Uses production host from configuration or falls back to default.
func (c *SharedConfigAdapter) GetPasswordExchangeURL() string {
	if c.config.ProdHost != "" {
		return "https://" + c.config.ProdHost
	}
	return "https://password.exchange"
}

// GetInitialNotificationSubject returns the subject template for initial notifications.
func (c *SharedConfigAdapter) GetInitialNotificationSubject() string {
	return "Encrypted Message from Password Exchange from %s"
}

// GetReminderNotificationSubject returns the subject template for reminder notifications.
func (c *SharedConfigAdapter) GetReminderNotificationSubject() string {
	return "Reminder: You have an unviewed encrypted message (Reminder #%d)"
}

// GetInitialNotificationBodyTemplate returns the body template for initial notifications.
func (c *SharedConfigAdapter) GetInitialNotificationBodyTemplate() string {
	return "Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about"
}

// GetReminderNotificationBodyTemplate returns the body template for reminder notifications.
func (c *SharedConfigAdapter) GetReminderNotificationBodyTemplate() string {
	return ""
}

// GetReminderEmailTemplate returns the path to the reminder email template.
func (c *SharedConfigAdapter) GetReminderEmailTemplate() string {
	return "/templates/reminder_email_template.html"
}

// GetReminderMessageContent returns the content message for reminder emails.
func (c *SharedConfigAdapter) GetReminderMessageContent() string {
	return "Please check your original email for the secure decrypt link. For security reasons, the decrypt link cannot be included in reminder emails. If you cannot find the original email, please contact the sender to resend the message."
}

// Validation methods

// ValidatePasswordExchangeURL validates the password exchange URL configuration.
func (c *SharedConfigAdapter) ValidatePasswordExchangeURL() error {
	return c.validator.ValidatePasswordExchangeURL(c.config.ProdHost)
}

// ValidateServerEmail validates the server email address configuration.
func (c *SharedConfigAdapter) ValidateServerEmail() error {
	return c.validator.ValidateServerEmail(c.config.EmailFrom)
}

// ValidateTemplateFormats validates all template strings for proper formatting.
func (c *SharedConfigAdapter) ValidateTemplateFormats() error {
	return c.validator.ValidateTemplateFormats(
		c.GetInitialNotificationSubject(),
		c.GetReminderNotificationSubject(),
		c.GetInitialNotificationBodyTemplate(),
	)
}