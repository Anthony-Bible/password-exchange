package secondary

// URLBuilderPort defines the secondary port for URL building operations.
// This interface abstracts the URL construction logic for the message domain,
// allowing the domain to remain independent of specific URL formats and schemes.
// Implementations of this interface handle the creation of decryption URLs that
// include the necessary parameters for message retrieval and decryption.
type URLBuilderPort interface {
	// BuildDecryptURL constructs a complete URL for accessing and decrypting a message.
	// The URL includes the message identifier and the encryption key required for decryption.
	// The encryption key is typically base64-encoded within the URL fragment or query parameters.
	//
	// Parameters:
	//   - messageID: The unique identifier of the encrypted message
	//   - encryptionKey: The symmetric key used to decrypt the message content
	//
	// Returns:
	//   - A complete URL string that can be shared with the message recipient
	//
	// Example:
	//   https://example.com/decrypt/abc123#key=base64encodedkey
	BuildDecryptURL(messageID string, encryptionKey []byte) string
}