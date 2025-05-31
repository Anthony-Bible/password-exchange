package secondary

// ConfigPort defines the secondary port for configuration access
type ConfigPort interface {
	GetEmailTemplate() string
	GetServerEmail() string
	GetServerName() string
	GetPasswordExchangeURL() string
}