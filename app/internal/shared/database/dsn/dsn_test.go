package dsn

import (
	"math/rand"
	"strings"
	"testing"
	"testing/quick"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresURL_Example(t *testing.T) {
	u := PostgresURL("db.internal", 6543, "app", "s3cret", "passwordexchange")

	cfg, err := pgx.ParseConfig(u)
	require.NoError(t, err)
	assert.Equal(t, "db.internal", cfg.Host)
	assert.EqualValues(t, 6543, cfg.Port)
	assert.Equal(t, "app", cfg.User)
	assert.Equal(t, "s3cret", cfg.Password)
	assert.Equal(t, "passwordexchange", cfg.Database)
}

func TestPostgresURL_DefaultsPortTo5432WhenZero(t *testing.T) {
	cfg, err := pgx.ParseConfig(PostgresURL("localhost", 0, "u", "p", "d"))
	require.NoError(t, err)
	assert.EqualValues(t, 5432, cfg.Port)
}

func TestPostgresURL_DefaultsSSLModeToDisable(t *testing.T) {
	// The Cloud SQL Auth Proxy terminates TLS on localhost, so the client must
	// not attempt TLS by default (pgx leaves TLSConfig nil for sslmode=disable).
	cfg, err := pgx.ParseConfig(PostgresURL("127.0.0.1", 5432, "u", "p", "d"))
	require.NoError(t, err)
	assert.Nil(t, cfg.TLSConfig, "sslmode must default to disable")
}

func TestPostgresURL_UsesPostgresScheme(t *testing.T) {
	assert.True(t, strings.HasPrefix(PostgresURL("h", 1, "u", "p", "d"), "postgres://"))
}

func TestPostgresURL_NastyCredentialsExample(t *testing.T) {
	user := "us er@:/?#%"
	pass := "p@ss:w/rd?#% é日本 &=+"
	cfg, err := pgx.ParseConfig(PostgresURL("localhost", 5432, user, pass, "d"))
	require.NoError(t, err)
	assert.Equal(t, user, cfg.User)
	assert.Equal(t, pass, cfg.Password)
}

// Property: for arbitrary user/password strings and any port in 0..65535,
// parsing the built URL yields the same user, password, host and port, with
// port 0 mapped to 5432.
func TestPostgresURL_RoundTripsThroughPgxParser(t *testing.T) {
	special := []string{"@", ":", "/", "?", "#", "%", " ", "é", "日本", "&", "=", "+", "\\", "'", "\""}
	sanitize := func(s string) string {
		s = strings.ToValidUTF8(s, "x")
		return strings.ReplaceAll(s, "\x00", "n")
	}

	prop := func(user, pass string, port uint16, spice uint8) bool {
		user = sanitize(user)
		pass = sanitize(pass)
		// Force some of the nasty characters into every sample.
		user += special[int(spice)%len(special)]
		pass += special[(int(spice)+3)%len(special)] + special[(int(spice)+7)%len(special)]
		if strings.TrimSpace(user) == "" {
			return true
		}

		cfg, err := pgx.ParseConfig(PostgresURL("db.example.com", int(port), user, pass, "passwordexchange"))
		if err != nil {
			t.Logf("ParseConfig failed for user=%q pass=%q port=%d: %v", user, pass, port, err)
			return false
		}
		wantPort := uint16(5432)
		if port != 0 {
			wantPort = port
		}
		return cfg.User == user && cfg.Password == pass &&
			cfg.Host == "db.example.com" && cfg.Port == wantPort &&
			cfg.Database == "passwordexchange"
	}

	require.NoError(t, quick.Check(prop, &quick.Config{
		MaxCount: 2000,
		Rand:     rand.New(rand.NewSource(565)),
	}))
}

func TestPostgresURL_IPv6HostIsBracketed(t *testing.T) {
	cfg, err := pgx.ParseConfig(PostgresURL("::1", 5433, "u", "p", "d"))
	require.NoError(t, err)
	assert.Equal(t, "::1", cfg.Host)
	assert.EqualValues(t, 5433, cfg.Port)
}
