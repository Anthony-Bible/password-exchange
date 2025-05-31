package secondary

// ConfigPort defines the secondary port for accessing notification configuration.
// This interface abstracts configuration management from the notification domain,
// allowing flexible configuration sources (files, environment variables, remote config, etc.).
// The port provides access to email-related settings required for sending notifications.
type ConfigPort interface {
	// GetEmailTemplate returns the path to the email template file.
	// The template should be a valid HTML or text template that can be parsed
	// and executed with notification data.
	//
	// Returns:
	//   - The file path to the email template (e.g., "templates/email_template.html")
	GetEmailTemplate() string

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
}