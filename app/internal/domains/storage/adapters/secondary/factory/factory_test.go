package factory

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/mysql"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/postgres"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The factory is the single place both `app database` and `app reminder`
// consult to pick a storage engine from PASSWORDEXCHANGE_DBDRIVER.

type noopValidator struct{}

func (noopValidator) ValidateEmail(string) error              { return nil }
func (noopValidator) SanitizeEmailForLogging(s string) string { return s }

func TestNewMessageRepository_UnsetOrEmptyDriverDefaultsToMySQL(t *testing.T) {
	for _, driver := range []string{"", "mysql"} {
		repo, err := NewMessageRepository(driver, contracts.DatabaseConfig{Host: "h:3306"}, logtest.NewNoop(), noopValidator{})
		require.NoError(t, err, "driver %q", driver)
		assert.IsType(t, mysqlType(), repo, "driver %q", driver)
	}
}

func TestNewMessageRepository_PostgresDriverBuildsPostgresAdapter(t *testing.T) {
	repo, err := NewMessageRepository("postgres", contracts.DatabaseConfig{Host: "h", Port: 5432}, logtest.NewNoop(), noopValidator{})
	require.NoError(t, err)
	assert.IsType(t, &postgres.PostgresAdapter{}, repo)
}

func TestNewMessageRepository_UnknownDriverFailsNamingTheVariableAndValue(t *testing.T) {
	repo, err := NewMessageRepository("oracle", contracts.DatabaseConfig{}, logtest.NewNoop(), noopValidator{})

	require.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "PASSWORDEXCHANGE_DBDRIVER")
	assert.Contains(t, err.Error(), "oracle")
}

// DatabaseConfigFromPass must carry DbPort through so the Postgres adapter
// connects to it; this proves it end to end against a throwaway TCP listener.
func TestPostgresRepository_ConnectsToConfiguredDbPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)

	accepted := make(chan struct{}, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		accepted <- struct{}{}
		_ = conn.Close() // not a Postgres server: Connect is expected to fail
	}()

	pc := config.PassConfig{DbDriver: "postgres", DbHost: "127.0.0.1", DbPort: port, DbUser: "u", DbPass: "p", DbName: "d"}
	repo, err := NewMessageRepository(pc.DbDriver, DatabaseConfigFromPass(pc), logtest.NewNoop(), noopValidator{})
	require.NoError(t, err)
	connector, ok := repo.(interface{ Connect() error })
	require.True(t, ok, "repository must expose Connect() so cmd/reminder can fail fast")
	assert.Error(t, connector.Connect(), "a non-Postgres listener must not yield a connection")

	select {
	case <-accepted:
	case <-time.After(5 * time.Second):
		t.Fatal("postgres adapter never dialled DbPort")
	}
}

func TestDatabaseConfigFromPass_CopiesConnectionSettings(t *testing.T) {
	pc := config.PassConfig{DbHost: "db.internal", DbPort: 6543, DbUser: "app", DbPass: "s3cret", DbName: "pe", DbMaxOpenConns: 7}

	got := DatabaseConfigFromPass(pc)

	assert.Equal(t, contracts.DatabaseConfig{Host: "db.internal", Port: 6543, User: "app", Password: "s3cret", Name: "pe", MaxOpenConns: 7}, got)
}

func mysqlType() any { return mysqlAdapterPtr() }

// mysqlAdapterPtr returns a typed nil of the concrete MySQL adapter via the
// constructor's dynamic type, without depending on its unexported fields.
func mysqlAdapterPtr() any {
	return mysql.NewMySQLAdapter(contracts.DatabaseConfig{}, logtest.NewNoop(), noopValidator{})
}
