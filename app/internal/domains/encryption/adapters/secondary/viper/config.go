// Package viper provides a Viper-backed implementation of the encryption
// domain's ConfigPort, supplying gRPC bind address configuration.
package viper

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/secondary"
	"github.com/spf13/viper"
)

// ErrInvalidListenAddress indicates the configured listen address is malformed.
var ErrInvalidListenAddress = errors.New("invalid listen address")

const (
	keyHost = "encryption.host"
	keyPort = "encryption.port"

	defaultHost = "0.0.0.0"
	defaultPort = 50051
)

// ViperConfigAdapter implements ConfigPort using Viper for both config-file
// and environment-variable resolution. Env vars are picked up via Viper's
// prefix + key-replacer machinery (configured in cmd/root.go at startup):
// PASSWORDEXCHANGE_ENCRYPTION_HOST and PASSWORDEXCHANGE_ENCRYPTION_PORT map
// to encryption.host and encryption.port respectively.
//
// Precedence (highest to lowest): explicit override (viper.Set) > env var >
// config file > built-in default.
type ViperConfigAdapter struct{}

// NewViperConfigAdapter creates a new Viper-backed configuration adapter
// and registers env-var bindings for the encryption listen address keys.
func NewViperConfigAdapter() secondary.ConfigPort {
	viper.BindEnv(keyHost)
	viper.BindEnv(keyPort)
	return &ViperConfigAdapter{}
}

// GetGRPCHost returns the host portion of the gRPC bind address.
func (a *ViperConfigAdapter) GetGRPCHost() string {
	if v := viper.GetString(keyHost); v != "" {
		return v
	}
	return defaultHost
}

// GetGRPCPort returns the port portion of the gRPC bind address. If the
// configured value is missing, unparseable, or non-positive, the built-in
// default port is used.
func (a *ViperConfigAdapter) GetGRPCPort() int {
	if p := viper.GetInt(keyPort); p > 0 {
		return p
	}
	return defaultPort
}

// GetListenAddress returns the full host:port bind string.
func (a *ViperConfigAdapter) GetListenAddress() string {
	return net.JoinHostPort(strings.TrimSpace(a.GetGRPCHost()), strconv.Itoa(a.GetGRPCPort()))
}

// ValidateListenAddress verifies host is non-empty and port is in 1..65535.
func (a *ViperConfigAdapter) ValidateListenAddress() error {
	host := strings.TrimSpace(a.GetGRPCHost())
	port := a.GetGRPCPort()
	if host == "" {
		return fmt.Errorf("%w: host is empty", ErrInvalidListenAddress)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("%w: port %d out of range 1..65535", ErrInvalidListenAddress, port)
	}
	address := net.JoinHostPort(host, strconv.Itoa(port))
	if _, _, err := net.SplitHostPort(address); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidListenAddress, err)
	}
	if _, err := net.ResolveTCPAddr("tcp", address); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidListenAddress, err)
	}
	return nil
}
