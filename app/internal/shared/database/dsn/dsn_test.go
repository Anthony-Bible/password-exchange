package dsn

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// credential is a quick.Generator producing strings biased towards characters
// that break naive DSN formatting: URL delimiters, percent signs, whitespace
// and multi-byte unicode.
type credential string

const credentialAlphabet = "abcXYZ019@:/?#%&=+ \t'\"\\[]!$éüßЖ中文🔑"

func (credential) Generate(r *rand.Rand, size int) reflect.Value {
	runes := []rune(credentialAlphabet)
	// At least one rune: pgx substitutes the OS user for an empty user.
	n := 1 + r.Intn(size+1)
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteRune(runes[r.Intn(len(runes))])
	}
	return reflect.ValueOf(credential(b.String()))
}

// port is a quick.Generator for the full 0..65535 range.
type port int

func (port) Generate(r *rand.Rand, _ int) reflect.Value {
	return reflect.ValueOf(port(r.Intn(65536)))
}

func TestPostgres_RoundTripsThroughPgxParseConfig(t *testing.T) {
	property := func(user, password credential, p port) bool {
		cfg := Config{
			Host:     "db.internal",
			Port:     int(p),
			User:     string(user),
			Password: string(password),
			Name:     "passwordexchange",
		}
		parsed, err := pgx.ParseConfig(Postgres(cfg))
		if err != nil {
			t.Logf("parse %q: %v", Postgres(cfg), err)
			return false
		}
		wantPort := uint16(p)
		if p == 0 {
			wantPort = DefaultPostgresPort
		}
		return parsed.User == string(user) &&
			parsed.Password == string(password) &&
			parsed.Host == cfg.Host &&
			parsed.Port == wantPort &&
			parsed.Database == cfg.Name
	}
	require.NoError(t, quick.Check(property, &quick.Config{MaxCount: 2000}))
}

func TestPostgres(t *testing.T) {
	tests := []struct {
		name         string
		cfg          Config
		wantHost     string
		wantPort     uint16
		wantUser     string
		wantPassword string
		wantSSLMode  string
	}{
		{
			name:         "defaults port and sslmode",
			cfg:          Config{Host: "localhost", User: "u", Password: "p", Name: "db"},
			wantHost:     "localhost",
			wantPort:     5432,
			wantUser:     "u",
			wantPassword: "p",
			wantSSLMode:  "disable",
		},
		{
			name:         "honours explicit port and sslmode",
			cfg:          Config{Host: "localhost", Port: 6543, User: "u", Password: "p", Name: "db", SSLMode: "require"},
			wantHost:     "localhost",
			wantPort:     6543,
			wantUser:     "u",
			wantPassword: "p",
			wantSSLMode:  "require",
		},
		{
			name:         "ipv6 host",
			cfg:          Config{Host: "::1", Port: 5433, User: "u", Password: "p", Name: "db"},
			wantHost:     "::1",
			wantPort:     5433,
			wantUser:     "u",
			wantPassword: "p",
			wantSSLMode:  "disable",
		},
		{
			name:         "credentials with URL delimiters",
			cfg:          Config{Host: "localhost", User: "us@er:/", Password: "p@ss:w/o?r#d%20 ", Name: "db"},
			wantHost:     "localhost",
			wantPort:     5432,
			wantUser:     "us@er:/",
			wantPassword: "p@ss:w/o?r#d%20 ",
			wantSSLMode:  "disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Postgres(tt.cfg)
			assert.Contains(t, got, "sslmode="+tt.wantSSLMode)

			parsed, err := pgx.ParseConfig(got)
			require.NoError(t, err)
			assert.Equal(t, tt.wantHost, parsed.Host)
			assert.Equal(t, tt.wantPort, parsed.Port)
			assert.Equal(t, tt.wantUser, parsed.User)
			assert.Equal(t, tt.wantPassword, parsed.Password)
			assert.Equal(t, tt.cfg.Name, parsed.Database)
		})
	}
}
