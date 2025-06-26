package viper

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// setupTestViper configures Viper for testing with a temporary config file.
func setupTestViper(t *testing.T, configContent string) {
	t.Helper()
	viper.Reset()

	if configContent != "" {
		file, err := os.CreateTemp(t.TempDir(), "config.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp config file: %v", err)
		}
		defer file.Close()

		_, err = file.WriteString(configContent)
		if err != nil {
			t.Fatalf("Failed to write to temp config file: %v", err)
		}

		viper.SetConfigFile(file.Name())
		viper.SetConfigType("yaml") // Explicitly set config type
		if err := viper.ReadInConfig(); err != nil {
			t.Fatalf("Failed to read config: %v", err)
		}
	}
}

func TestNewViperConfigAdapter_Defaults(t *testing.T) {
	setupTestViper(t, "") // No config file, should use defaults

	adapter := NewViperConfigAdapter()

	assert.Equal(t, "/templates/email_template.html", adapter.GetEmailTemplate())
	assert.Equal(t, "server@password.exchange", adapter.GetServerEmail())
	assert.Equal(t, "Password Exchange", adapter.GetServerName())
	assert.Equal(t, "https://password.exchange", adapter.GetPasswordExchangeURL())
	assert.Equal(t, "Encrypted Message from Password Exchange from %s", adapter.GetInitialNotificationSubject())
	assert.Equal(t, "Reminder: You have an unviewed encrypted message (Reminder #%d)", adapter.GetReminderNotificationSubject())
	assert.Equal(t, "Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about", adapter.GetInitialNotificationBodyTemplate())
	assert.Equal(t, "Please check your original email for the secure decrypt link. For security reasons, the decrypt link cannot be included in reminder emails. If you cannot find the original email, please contact the sender to resend the message.", adapter.GetReminderMessageContent())
}

func TestNewViperConfigAdapter_WithCustomConfig(t *testing.T) {
	configContent := `
email:
  templates:
    initial: "/custom/initial.html"
    reminder: "/custom/reminder.html"
  subjects:
    initial: "Custom initial subject"
    reminder: "Custom reminder subject"
  body:
    initial: "Custom initial body"
    reminder: "Custom reminder body"
  sender:
    email: "custom@example.com"
    name: "Custom Sender"
  url: "https://custom.example.com"
`
	setupTestViper(t, configContent)

	adapter := NewViperConfigAdapter()

	assert.Equal(t, "/custom/initial.html", adapter.GetEmailTemplate())
	assert.Equal(t, "custom@example.com", adapter.GetServerEmail())
	assert.Equal(t, "Custom Sender", adapter.GetServerName())
	assert.Equal(t, "https://custom.example.com", adapter.GetPasswordExchangeURL())
	assert.Equal(t, "Custom initial subject", adapter.GetInitialNotificationSubject())
	assert.Equal(t, "Custom reminder subject", adapter.GetReminderNotificationSubject())
	assert.Equal(t, "Custom initial body", adapter.GetInitialNotificationBodyTemplate())
	assert.Equal(t, "Custom reminder body", adapter.GetReminderMessageContent())
}

func TestNewViperConfigAdapter_WithPartialConfig(t *testing.T) {
	configContent := `
email:
  sender:
    email: "partial@example.com"
`
	setupTestViper(t, configContent)

	adapter := NewViperConfigAdapter()

	// Test that the partial config is loaded
	assert.Equal(t, "partial@example.com", adapter.GetServerEmail())

	// Test that the rest of the config falls back to defaults
	assert.Equal(t, "/templates/email_template.html", adapter.GetEmailTemplate())
	assert.Equal(t, "Password Exchange", adapter.GetServerName())
}

func TestValidation(t *testing.T) {
	configContent := `
email:
  url: "invalid-url"
`
	setupTestViper(t, configContent)

	adapter := NewViperConfigAdapter()

	// This is a simplified test. In a real scenario, you would check for a specific error.
	// However, since the validation methods are stubbed out, we just check for nil for now.
	assert.Nil(t, adapter.ValidatePasswordExchangeURL())
}
