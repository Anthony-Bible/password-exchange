package viper

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/validation"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/spf13/viper"
)

// ViperConfigAdapter implements ConfigPort using Viper for configuration management.
type ViperConfigAdapter struct {
	emailConfig config.EmailConfig
	validator   *validation.ConfigValidator
}

// NewViperConfigAdapter creates a new configuration adapter and loads email settings from Viper.
func NewViperConfigAdapter() secondary.ConfigPort {
	var emailConfig config.EmailConfig
	if viper.IsSet("email") {
		if err := viper.UnmarshalKey("email", &emailConfig); err != nil {
			// Fallback to default if unmarshalling fails
			emailConfig = getDefaultEmailConfig()
		}
	} else {
		emailConfig = getDefaultEmailConfig()
	}

	// Apply defaults for any fields that were not set
	setDefaults(&emailConfig)

	return &ViperConfigAdapter{
		emailConfig: emailConfig,
		validator:   validation.NewConfigValidator(),
	}
}

// getDefaultEmailConfig returns the default email configuration.
func getDefaultEmailConfig() config.EmailConfig {
	return config.EmailConfig{
		Templates: config.EmailTemplates{
			Initial:  "/templates/email_template.html",
			Reminder: "/templates/reminder_email_template.html",
		},
		Subjects: config.EmailSubjects{
			Initial:  "Encrypted Message from Password Exchange from %s",
			Reminder: "Reminder: You have an unviewed encrypted message (Reminder #%d)",
		},
		Body: config.EmailBody{
			Initial:  "Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about",
			Reminder: "Please check your original email for the secure decrypt link. For security reasons, the decrypt link cannot be included in reminder emails. If you cannot find the original email, please contact the sender to resend the message.",
		},
		Sender: config.EmailSender{
			Email: "server@password.exchange",
			Name:  "Password Exchange",
		},
		URL: "https://password.exchange",
	}
}

// setDefaults applies default values to any empty fields in the email configuration.
func setDefaults(cfg *config.EmailConfig) {
	defaults := getDefaultEmailConfig()
	if cfg.Templates.Initial == "" {
		cfg.Templates.Initial = defaults.Templates.Initial
	}
	if cfg.Templates.Reminder == "" {
		cfg.Templates.Reminder = defaults.Templates.Reminder
	}
	if cfg.Subjects.Initial == "" {
		cfg.Subjects.Initial = defaults.Subjects.Initial
	}
	if cfg.Subjects.Reminder == "" {
		cfg.Subjects.Reminder = defaults.Subjects.Reminder
	}
	if cfg.Body.Initial == "" {
		cfg.Body.Initial = defaults.Body.Initial
	}
	if cfg.Body.Reminder == "" {
		cfg.Body.Reminder = defaults.Body.Reminder
	}
	if cfg.Sender.Email == "" {
		cfg.Sender.Email = defaults.Sender.Email
	}
	if cfg.Sender.Name == "" {
		cfg.Sender.Name = defaults.Sender.Name
	}
	if cfg.URL == "" {
		cfg.URL = defaults.URL
	}
}

func (v *ViperConfigAdapter) GetEmailTemplate() string {
	return v.emailConfig.Templates.Initial
}

func (v *ViperConfigAdapter) GetServerEmail() string {
	return v.emailConfig.Sender.Email
}

func (v *ViperConfigAdapter) GetServerName() string {
	return v.emailConfig.Sender.Name
}

func (v *ViperConfigAdapter) GetPasswordExchangeURL() string {
	return v.emailConfig.URL
}

func (v *ViperConfigAdapter) GetInitialNotificationSubject() string {
	return v.emailConfig.Subjects.Initial
}

func (v *ViperConfigAdapter) GetReminderNotificationSubject() string {
	return v.emailConfig.Subjects.Reminder
}

func (v *ViperConfigAdapter) GetInitialNotificationBodyTemplate() string {
	return v.emailConfig.Body.Initial
}

func (v *ViperConfigAdapter) GetReminderNotificationBodyTemplate() string {
	return v.emailConfig.Templates.Reminder
}

func (v *ViperConfigAdapter) GetReminderEmailTemplate() string {
	return v.emailConfig.Templates.Reminder
}

func (v *ViperConfigAdapter) GetReminderMessageContent() string {
	return v.emailConfig.Body.Reminder
}

// Validation methods

// ValidatePasswordExchangeURL validates the password exchange URL configuration.
func (v *ViperConfigAdapter) ValidatePasswordExchangeURL() error {
	return v.validator.ValidatePasswordExchangeURL(v.emailConfig.URL)
}

// ValidateServerEmail validates the server email address configuration.
func (v *ViperConfigAdapter) ValidateServerEmail() error {
	return v.validator.ValidateServerEmail(v.emailConfig.Sender.Email)
}

// ValidateTemplateFormats validates all template strings for proper formatting.
func (v *ViperConfigAdapter) ValidateTemplateFormats() error {
	return v.validator.ValidateTemplateFormats(
		v.emailConfig.Subjects.Initial,
		v.emailConfig.Subjects.Reminder,
		v.emailConfig.Body.Initial,
	)
}
