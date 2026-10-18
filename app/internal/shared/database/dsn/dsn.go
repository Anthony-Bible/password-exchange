// Package dsn builds database connection strings shared by the migrator and the
// storage adapters.
package dsn

import (
	"net"
	"net/url"
	"strconv"
)

const (
	// DefaultPostgresPort is used when Config.Port is zero.
	DefaultPostgresPort = 5432
	// DefaultSSLMode is used when Config.SSLMode is empty. TLS is off by
	// default because the Cloud SQL proxy on localhost terminates it.
	DefaultSSLMode = "disable"
)

// Config holds the connection settings for a Postgres DSN.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Postgres returns a postgres:// URL for cfg. User and password are
// percent-encoded, so any characters (including @:/?#%) are safe.
func Postgres(cfg Config) string {
	port := cfg.Port
	if port == 0 {
		port = DefaultPostgresPort
	}
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = DefaultSSLMode
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, strconv.Itoa(port)),
		Path:     "/" + cfg.Name,
		RawQuery: url.Values{"sslmode": {sslMode}}.Encode(),
	}
	return u.String()
}
