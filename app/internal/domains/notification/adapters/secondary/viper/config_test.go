package viper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewViperConfigAdapter(t *testing.T) {
	adapter := NewViperConfigAdapter()

	// Test existing config methods
	assert.Equal(t, "/templates/email_template.html", adapter.GetEmailTemplate())
	assert.Equal(t, "server@password.exchange", adapter.GetServerEmail())
	assert.Equal(t, "Password Exchange", adapter.GetServerName())
	assert.Equal(t, "https://password.exchange", adapter.GetPasswordExchangeURL())

	// Test new email configuration methods
	assert.Equal(t, "Encrypted Message from Password Exchange from %s", adapter.GetInitialNotificationSubject())
	assert.Equal(t, "Reminder: You have an unviewed encrypted message (Reminder #%d)", adapter.GetReminderNotificationSubject())
	assert.Equal(t, "Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about", adapter.GetInitialNotificationBodyTemplate())
	assert.Equal(t, "", adapter.GetReminderNotificationBodyTemplate())
	assert.Equal(t, "/templates/reminder_email_template.html", adapter.GetReminderEmailTemplate())
	assert.Equal(t, "Please check your original email for the secure decrypt link. For security reasons, the decrypt link cannot be included in reminder emails. If you cannot find the original email, please contact the sender to resend the message.", adapter.GetReminderMessageContent())
}

func TestNewViperConfigAdapterWithValues(t *testing.T) {
	emailTemplate := "/custom/template.html"
	serverEmail := "custom@example.com"
	serverName := "Custom Service"
	passwordExchangeURL := "https://custom.example.com"

	adapter := NewViperConfigAdapterWithValues(emailTemplate, serverEmail, serverName, passwordExchangeURL)

	// Test that custom values are used for existing methods
	assert.Equal(t, emailTemplate, adapter.GetEmailTemplate())
	assert.Equal(t, serverEmail, adapter.GetServerEmail())
	assert.Equal(t, serverName, adapter.GetServerName())
	assert.Equal(t, passwordExchangeURL, adapter.GetPasswordExchangeURL())

	// Test that default values are still used for new methods (backward compatibility)
	assert.Equal(t, "Encrypted Message from Password Exchange from %s", adapter.GetInitialNotificationSubject())
	assert.Equal(t, "Reminder: You have an unviewed encrypted message (Reminder #%d)", adapter.GetReminderNotificationSubject())
	assert.Equal(t, "Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about", adapter.GetInitialNotificationBodyTemplate())
	assert.Equal(t, "", adapter.GetReminderNotificationBodyTemplate())
	assert.Equal(t, "/templates/reminder_email_template.html", adapter.GetReminderEmailTemplate())
	assert.Equal(t, "Please check your original email for the secure decrypt link. For security reasons, the decrypt link cannot be included in reminder emails. If you cannot find the original email, please contact the sender to resend the message.", adapter.GetReminderMessageContent())
}

// TestBackwardCompatibility ensures that existing code continues to work without changes
func TestBackwardCompatibility(t *testing.T) {
	adapter := NewViperConfigAdapter()

	// Verify that the interface is properly implemented by calling all methods
	// This test will fail at compile time if interface methods are missing
	_ = adapter.GetEmailTemplate()
	_ = adapter.GetServerEmail()
	_ = adapter.GetServerName()
	_ = adapter.GetPasswordExchangeURL()
	_ = adapter.GetInitialNotificationSubject()
	_ = adapter.GetReminderNotificationSubject()
	_ = adapter.GetInitialNotificationBodyTemplate()
	_ = adapter.GetReminderNotificationBodyTemplate()
	_ = adapter.GetReminderEmailTemplate()
	_ = adapter.GetReminderMessageContent()

	// Test that we get non-empty values for critical settings
	assert.NotEmpty(t, adapter.GetInitialNotificationSubject())
	assert.NotEmpty(t, adapter.GetReminderNotificationSubject())
	assert.NotEmpty(t, adapter.GetInitialNotificationBodyTemplate())
	assert.NotEmpty(t, adapter.GetReminderMessageContent())
}

// TestConfigurationValues verifies the exact default values match the original hardcoded ones
func TestConfigurationValues(t *testing.T) {
	adapter := NewViperConfigAdapter()

	// Test that initial notification subject matches the original hardcoded value
	expectedInitialSubject := "Encrypted Message from Password Exchange from %s"
	assert.Equal(t, expectedInitialSubject, adapter.GetInitialNotificationSubject())

	// Test that reminder notification subject matches the original hardcoded value
	expectedReminderSubject := "Reminder: You have an unviewed encrypted message (Reminder #%d)"
	assert.Equal(t, expectedReminderSubject, adapter.GetReminderNotificationSubject())

	// Test that initial body template matches the original hardcoded value
	expectedInitialBody := "Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about"
	assert.Equal(t, expectedInitialBody, adapter.GetInitialNotificationBodyTemplate())

	// Test that reminder message content matches the original hardcoded value
	expectedReminderContent := "Please check your original email for the secure decrypt link. For security reasons, the decrypt link cannot be included in reminder emails. If you cannot find the original email, please contact the sender to resend the message."
	assert.Equal(t, expectedReminderContent, adapter.GetReminderMessageContent())

	// Test template paths
	assert.Equal(t, "/templates/email_template.html", adapter.GetEmailTemplate())
	assert.Equal(t, "/templates/reminder_email_template.html", adapter.GetReminderEmailTemplate())
}