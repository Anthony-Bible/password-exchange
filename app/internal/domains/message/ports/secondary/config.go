package secondary

// ConfigPort defines the secondary port for accessing message configuration.
type ConfigPort interface {
	// GetDefaultMaxViewCount returns the default maximum number of views for a message.
	GetDefaultMaxViewCount() int
}
