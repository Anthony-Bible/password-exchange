package viper

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
)

// ViperConfigAdapter implements ConfigPort using static configuration values
type ViperConfigAdapter struct {
	emailTemplate       string
	serverEmail        string
	serverName         string
	passwordExchangeURL string
}

// NewViperConfigAdapter creates a new configuration adapter with default values
func NewViperConfigAdapter() secondary.ConfigPort {
	return &ViperConfigAdapter{
		emailTemplate:       "/templates/email_template.html",
		serverEmail:        "server@password.exchange",
		serverName:         "Password Exchange",
		passwordExchangeURL: "https://password.exchange",
	}
}

// NewViperConfigAdapterWithValues creates a new configuration adapter with custom values
func NewViperConfigAdapterWithValues(emailTemplate, serverEmail, serverName, passwordExchangeURL string) secondary.ConfigPort {
	return &ViperConfigAdapter{
		emailTemplate:       emailTemplate,
		serverEmail:        serverEmail,
		serverName:         serverName,
		passwordExchangeURL: passwordExchangeURL,
	}
}

func (v *ViperConfigAdapter) GetEmailTemplate() string {
	return v.emailTemplate
}

func (v *ViperConfigAdapter) GetServerEmail() string {
	return v.serverEmail
}

func (v *ViperConfigAdapter) GetServerName() string {
	return v.serverName
}

func (v *ViperConfigAdapter) GetPasswordExchangeURL() string {
	return v.passwordExchangeURL
}