package viper

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

const (
	envHost = "PASSWORDEXCHANGE_ENCRYPTION_HOST"
	envPort = "PASSWORDEXCHANGE_ENCRYPTION_PORT"
)

// cleanState resets global viper and re-installs the env-binding configuration
// that cmd/root.go normally applies at app startup (prefix, key replacer,
// AutomaticEnv). It also clears any encryption env vars inherited from the
// test runner so per-test t.Setenv calls drive behavior cleanly.
func cleanState(t *testing.T) {
	t.Helper()
	viper.Reset()
	viper.SetEnvPrefix("passwordexchange")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	t.Setenv(envHost, "")
	t.Setenv(envPort, "")
	t.Cleanup(viper.Reset)
}

// seedConfig loads encryption.host/port from an in-memory YAML buffer, mimicking
// values supplied via a real config file (Viper's "config" precedence tier).
func seedConfig(t *testing.T, host string, port int) {
	t.Helper()
	viper.SetConfigType("yaml")
	yaml := "encryption:\n  host: " + host + "\n  port: " + strconv.Itoa(port) + "\n"
	if err := viper.ReadConfig(strings.NewReader(yaml)); err != nil {
		t.Fatalf("seed config: %v", err)
	}
}

func TestViperConfigAdapter_Defaults(t *testing.T) {
	cleanState(t)
	a := NewViperConfigAdapter()
	if got := a.GetGRPCHost(); got != "0.0.0.0" {
		t.Errorf("default host: expected 0.0.0.0, got %q", got)
	}
	if got := a.GetGRPCPort(); got != 50051 {
		t.Errorf("default port: expected 50051, got %d", got)
	}
	if got := a.GetListenAddress(); got != "0.0.0.0:50051" {
		t.Errorf("default address: expected 0.0.0.0:50051, got %q", got)
	}
	if err := a.ValidateListenAddress(); err != nil {
		t.Errorf("default address should validate, got %v", err)
	}
}

func TestViperConfigAdapter_ConfigOverrides(t *testing.T) {
	cleanState(t)
	seedConfig(t, "127.0.0.1", 9090)
	a := NewViperConfigAdapter()
	if got := a.GetGRPCHost(); got != "127.0.0.1" {
		t.Errorf("host: expected 127.0.0.1, got %q", got)
	}
	if got := a.GetGRPCPort(); got != 9090 {
		t.Errorf("port: expected 9090, got %d", got)
	}
	if got := a.GetListenAddress(); got != "127.0.0.1:9090" {
		t.Errorf("address: expected 127.0.0.1:9090, got %q", got)
	}
}

func TestViperConfigAdapter_EnvOverridesConfig(t *testing.T) {
	cleanState(t)
	seedConfig(t, "127.0.0.1", 9090)
	t.Setenv(envHost, "10.0.0.5")
	t.Setenv(envPort, "31337")

	a := NewViperConfigAdapter()
	if got := a.GetGRPCHost(); got != "10.0.0.5" {
		t.Errorf("env host should win over config; got %q", got)
	}
	if got := a.GetGRPCPort(); got != 31337 {
		t.Errorf("env port should win over config; got %d", got)
	}
}

func TestViperConfigAdapter_EnvOnly(t *testing.T) {
	cleanState(t)
	t.Setenv(envHost, "0.0.0.0")
	t.Setenv(envPort, "60000")

	a := NewViperConfigAdapter()
	if got := a.GetListenAddress(); got != "0.0.0.0:60000" {
		t.Errorf("env-only address: expected 0.0.0.0:60000, got %q", got)
	}
}

func TestViperConfigAdapter_GetListenAddress_IPv6(t *testing.T) {
	cleanState(t)
	t.Setenv(envHost, "::1")
	t.Setenv(envPort, "50052")

	a := NewViperConfigAdapter()
	if got := a.GetListenAddress(); got != "[::1]:50052" {
		t.Errorf("ipv6 listen address: expected [::1]:50052, got %q", got)
	}
}

func TestViperConfigAdapter_EnvPortBadFormatFallsBack(t *testing.T) {
	cleanState(t)
	t.Setenv(envPort, "not-a-number")
	a := NewViperConfigAdapter()
	if got := a.GetGRPCPort(); got != defaultPort {
		t.Errorf("invalid env port should fall back to default %d, got %d", defaultPort, got)
	}
}

func TestViperConfigAdapter_ValidateListenAddress(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		port    int
		wantErr bool
	}{
		{"valid", "127.0.0.1", 8080, false},
		{"valid ipv6", "::1", 8080, false},
		{"host whitespace", "   ", 8080, true},
		{"host unparsable", "bad host", 8080, true},
		{"port too high", "127.0.0.1", 70000, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanState(t)
			viper.Set("encryption.host", tt.host)
			viper.Set("encryption.port", tt.port)
			a := NewViperConfigAdapter()
			err := a.ValidateListenAddress()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidListenAddress) {
				t.Fatalf("expected ErrInvalidListenAddress, got %v", err)
			}
		})
	}
}
