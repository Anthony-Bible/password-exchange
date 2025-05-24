package secondary

// URLBuilderPort defines the secondary port for URL building operations
type URLBuilderPort interface {
	// BuildDecryptURL builds a URL for decrypting a message
	BuildDecryptURL(messageID string, encryptionKey []byte) string
}