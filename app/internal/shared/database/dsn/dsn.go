// Package dsn builds database connection strings.
package dsn

import (
	"net"
	"net/url"
	"strconv"
)

// PostgresURL builds a postgres:// URL. Port 0 maps to 5432 and sslmode
// defaults to disable (TLS is terminated by a local proxy).
func PostgresURL(host string, port int, user, password, dbname string) string {
	if port == 0 {
		port = 5432
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, strconv.Itoa(port)),
		Path:     "/" + dbname,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}
