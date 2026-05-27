package secondary

// ConfigPort defines the secondary port for accessing encryption service configuration.
// Implementations supply gRPC bind addresses and related networking parameters.
type ConfigPort interface {
	// GetListenAddress returns the full host:port string the gRPC server should bind to.
	// Defaults to "0.0.0.0:50051".
	GetListenAddress() string

	// GetGRPCHost returns the host portion of the gRPC bind address.
	// Defaults to "0.0.0.0".
	GetGRPCHost() string

	// GetGRPCPort returns the port portion of the gRPC bind address.
	// Defaults to 50051.
	GetGRPCPort() int

	// ValidateListenAddress verifies that the configured listen address is well-formed.
	ValidateListenAddress() error
}
