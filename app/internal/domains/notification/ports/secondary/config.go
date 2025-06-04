package secondary

// ConfigPort defines the secondary port for accessing notification configuration.
// This interface abstracts configuration management from the notification domain,
// allowing flexible configuration sources (files, environment variables, remote config, etc.).
// The port provides access to email-related settings required for sending notifications.
type ConfigPort interface {
	// === Server Configuration ===
	// Basic server settings for email notifications

	// GetServerEmail returns the email address used as the sender for notifications.
	// This should be a valid email address that recipients will see as the "From" address.
	//
	// Returns:
	//   - The sender email address (e.g., "noreply@password.exchange")
	GetServerEmail() string

	// GetServerName returns the display name for the email sender.
	// This name appears alongside the sender email in most email clients.
	//
	// Returns:
	//   - The sender display name (e.g., "Password Exchange Service")
	GetServerName() string

	// GetPasswordExchangeURL returns the base URL of the password exchange service.
	// This URL is used for constructing links within notification emails and
	// should not include a trailing slash.
	//
	// Returns:
	//   - The base URL of the service (e.g., "https://password.exchange")
	GetPasswordExchangeURL() string

	// === Email Subject Configuration ===
	// Subject line templates for different notification types

	// GetInitialNotificationSubject returns the subject template for initial notification emails.
	// The template may contain format placeholders for dynamic values like sender name.
	//
	// Returns:
	//   - The subject template string (e.g., "Encrypted Message from Password Exchange from %s")
	GetInitialNotificationSubject() string

	// GetReminderNotificationSubject returns the subject template for reminder notification emails.
	// The template may contain format placeholders for dynamic values like reminder count.
	//
	// Returns:
	//   - The subject template string (e.g., "Reminder: You have an unviewed encrypted message (Reminder #%d)")
	GetReminderNotificationSubject() string

	// === Email Template Configuration ===
	// Template content and file paths for notification emails

	// GetEmailTemplate returns the path to the email template file.
	// The template should be a valid HTML or text template that can be parsed
	// and executed with notification data.
	//
	// Returns:
	//   - The file path to the email template (e.g., "templates/email_template.html")
	GetEmailTemplate() string

	// GetInitialNotificationBodyTemplate returns the body template for initial notification emails.
	// This can be either a file path to a template file or an inline template string.
	// The template should support placeholders for recipient name, sender name, and URLs.
	//
	// Returns:
	//   - The body template string or file path
	GetInitialNotificationBodyTemplate() string

	// GetReminderNotificationBodyTemplate returns the body template for reminder notification emails.
	// This can be either a file path to a template file or an inline template string.
	//
	// Returns:
	//   - The body template string or file path
	GetReminderNotificationBodyTemplate() string

	// GetReminderEmailTemplate returns the path to the reminder email template file.
	// This is specifically for HTML templates used in reminder emails.
	//
	// Returns:
	//   - The file path to the reminder email template (e.g., "/templates/reminder_email_template.html")
	GetReminderEmailTemplate() string

	// GetReminderMessageContent returns the content message used in reminder emails.
	// This message explains why the decrypt link is not included in reminders.
	//
	// Returns:
	//   - The reminder message content string
	GetReminderMessageContent() string

	// === Configuration Validation ===
	// Methods for validating configuration values

	// ValidatePasswordExchangeURL validates the password exchange URL configuration.
	// Ensures the URL is properly formatted without protocol, path, or port components.
	//
	// Returns:
	//   - An error if the URL format is invalid, nil if valid or using default
	ValidatePasswordExchangeURL() error

	// ValidateServerEmail validates the server email address configuration.
	// Uses comprehensive email validation including format, length, and pattern checks.
	//
	// Returns:
	//   - An error if the email format is invalid, nil if valid or using default
	ValidateServerEmail() error

	// ValidateTemplateFormats validates all template strings for proper formatting.
	// Checks that template placeholders match expected patterns and counts.
	//
	// Returns:
	//   - An error if any template format is invalid, nil if all are valid
	ValidateTemplateFormats() error
}