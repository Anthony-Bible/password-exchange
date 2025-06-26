package config

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
)

var AppConfig PassConfig

// EmailConfig holds all email-related settings.
type EmailConfig struct {
	Templates EmailTemplates `mapstructure:"templates"`
	Subjects  EmailSubjects  `mapstructure:"subjects"`
	Body      EmailBody      `mapstructure:"body"`
	Sender    EmailSender    `mapstructure:"sender"`
	URL       string         `mapstructure:"url"`
}

// EmailTemplates defines paths or inline content for email templates.
type EmailTemplates struct {
	Initial  string `mapstructure:"initial"`
	Reminder string `mapstructure:"reminder"`
}

// EmailSubjects defines the subject lines for different emails.
type EmailSubjects struct {
	Initial  string `mapstructure:"initial"`
	Reminder string `mapstructure:"reminder"`
}

// EmailBody defines the body content for different emails.
type EmailBody struct {
	Initial  string `mapstructure:"initial"`
	Reminder string `mapstructure:"reminder"`
}

// EmailSender defines the sender's details.
type EmailSender struct {
	Email string `mapstructure:"email"`
	Name  string `mapstructure:"name"`
}

// Config represents the complete application configuration
type Config struct {
	PassConfig `mapstructure:",squash"`
	Database   domain.DatabaseConfig `mapstructure:"database"`
	Reminder   ReminderConfig        `mapstructure:"reminder"`
	Email      EmailConfig           `mapstructure:"email"`
}

// ReminderConfig contains configuration for the reminder email system
type ReminderConfig struct {
	Enabled           bool `mapstructure:"enabled" default:"true"`
	CheckAfterHours   int  `mapstructure:"checkafterhours" default:"24"`   // Default: 24
	MaxReminders      int  `mapstructure:"maxreminders" default:"3"`      // Default: 3
	ReminderInterval  int  `mapstructure:"reminderinterval" default:"24"`  // Default: 24 hours
}

// NewReminderConfig creates a new ReminderConfig with proper default values
func NewReminderConfig() ReminderConfig {
	return ReminderConfig{
		Enabled:          true,
		CheckAfterHours:  24,
		MaxReminders:     3,
		ReminderInterval: 24,
	}
}

// WithDefaults applies default values to zero-value fields and ensures valid ranges
func (r *ReminderConfig) WithDefaults() {
	// Apply defaults for zero values
	if r.CheckAfterHours == 0 {
		r.CheckAfterHours = 24
	}
	if r.MaxReminders == 0 {
		r.MaxReminders = 3
	}
	if r.ReminderInterval == 0 {
		r.ReminderInterval = 24
	}
	
	// Ensure values are within valid ranges, apply defaults if out of range
	if r.CheckAfterHours < 1 || r.CheckAfterHours > 8760 {
		r.CheckAfterHours = 24
	}
	if r.MaxReminders < 1 || r.MaxReminders > 10 {
		r.MaxReminders = 3
	}
	if r.ReminderInterval < 1 || r.ReminderInterval > 720 {
		r.ReminderInterval = 24
	}
}

type PassConfig struct {
	EmailHost             string `mapstructure:"emailhost"`
	EmailUser             string `mapstructure:"emailuser"`
	EmailPass             string `mapstructure:"emailpass"`
	EmailFrom             string `mapstructure:"emailfrom"`
	RabHost               string `mapstructure:"rabhost"`
	RabUser               string `mapstructure:"rabuser"`
	RabPass               string `mapstructure:"rabpass"`
	RabQName              string `mapstructure:"rabqname"`
	DbHost                string `mapstructure:"dbhost"`
	DbUser                string `mapstructure:"dbuser"`
	DbPass                string `mapstructure:"dbpass"`
	DbName                string `mapstructure:"dbname"`
	ProdHost              string `mapstructure:"prodhost"`
	DevHost               string `mapstructure:"devhost"`
	EncryptionProdService string `mapstructure:"encryptionprodservice"`
	DatabaseProdService   string `mapstructure:"databaseprodservice"`
	EncryptionDevService  string `mapstructure:"encryptiondevservice"`
	DatabaseDevService    string `mapstructure:"databasedevservice"`
	Loglevel              string `mapstructure:"loglevel"`
	RunningEnvironment    string `mapstructure:"runningenvironment"`
	TurnstileSecret       string `mapstructure:"turnstile_secret"`
	DefaultMaxViewCount   int    `mapstructure:"defaultmaxviewcount"`
	DbPort                int    `mapstructure:"dbport"`
	EmailPort             int    `mapstructure:"emailport"`
	RabPort               int    `mapstructure:"rabport"`
}
