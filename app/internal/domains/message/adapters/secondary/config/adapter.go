package config

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
)

// Adapter implements ConfigPort using the shared Viper-based config package
type Adapter struct{}

// NewAdapter creates a new config adapter
func NewAdapter() secondary.ConfigPort {
	return &Adapter{}
}

func (a *Adapter) GetDefaultMaxViewCount() int {
	return config.AppConfig.DefaultMaxViewCount
}
