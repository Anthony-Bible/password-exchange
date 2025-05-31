package secondary

import "context"

// URLBuilderPort defines the secondary port for URL construction operations in the notification domain.
// This interface provides URL building capabilities for various notification scenarios,
// abstracting the URL structure and format from the domain logic. Implementations handle
// the construction of URLs for message decryption, service homepage, and information pages.
type URLBuilderPort interface {
	// BuildDecryptionURL constructs a URL for accessing and decrypting a specific message.
	// Unlike the message domain's URL builder, this method only requires the message ID
	// as the encryption key is typically retrieved separately in the notification flow.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline propagation
	//   - uniqueID: The unique identifier of the encrypted message
	//
	// Returns:
	//   - The complete decryption URL
	//   - An error if URL construction fails
	//
	// Example:
	//   https://password.exchange/decrypt/abc123
	BuildDecryptionURL(ctx context.Context, uniqueID string) (string, error)

	// BuildPasswordExchangeURL returns the base URL of the password exchange service.
	// This is typically used in email templates to provide a link to the main service.
	//
	// Returns:
	//   - The base URL of the password exchange service
	//
	// Example:
	//   https://password.exchange
	BuildPasswordExchangeURL() string

	// BuildAboutURL constructs the URL to the service's about or information page.
	// This URL is often included in notification emails to provide recipients
	// with more information about the service.
	//
	// Returns:
	//   - The URL to the about page
	//
	// Example:
	//   https://password.exchange/about
	BuildAboutURL() string
}