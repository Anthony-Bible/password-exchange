//go:build integration

package database

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/integration/dbtest"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/database/dsn"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// emptyPostgresDB returns connection settings for a brand-new, empty database.
func emptyPostgresDB(t *testing.T) (host string, port int, user, pass, name string, open func() *sql.DB) {
	t.Helper()
	admin, cfg := dbtest.StartPostgres(t, fstest.MapFS{
		"1_harness.up.sql":   {Data: []byte("CREATE TABLE harness_fixture (id int);")},
		"1_harness.down.sql": {Data: []byte("DROP TABLE harness_fixture;")},
	})
	name = "cmd_database_fresh"
	_, err := admin.Exec("CREATE DATABASE " + name)
	require.NoError(t, err)
	url := dsn.PostgresURL(cfg.Host, cfg.Port, cfg.User, cfg.Password, name)
	return cfg.Host, cfg.Port, cfg.User, cfg.Password, name, func() *sql.DB {
		db, err := sql.Open("pgx", url)
		require.NoError(t, err)
		return db
	}
}

func TestPostgresMigrator_AppliesEmbeddedBaselineWithoutMigrationsDir(t *testing.T) {
	host, port, user, pass, name, open := emptyPostgresDB(t)
	resetConfig(t)
	t.Chdir(t.TempDir()) // no migrations/ directory on disk
	viper.Set("dbdriver", "postgres")
	viper.Set("dbhost", host)
	viper.Set("dbport", port)
	viper.Set("dbuser", user)
	viper.Set("dbpass", pass)
	viper.Set("dbname", name)

	require.NoError(t, initConfigAndMigrator())
	require.NoError(t, migrator.Up())

	// Honours DbPort and applies the embedded baseline.
	db := open()
	defer db.Close()
	var version int64
	var dirty bool
	require.NoError(t, db.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty))
	assert.EqualValues(t, 20260311000007, version)
	assert.False(t, dirty)
}

// Built binary, run from a cwd with no migrations/ dir, applies the baseline
// via AUTOMIGRATE and then serves gRPC (:50051).
func TestBinary_PostgresAutomigrateAppliesBaselineThenServes(t *testing.T) {
	if conn, err := net.DialTimeout("tcp", "127.0.0.1:50051", 500*time.Millisecond); err == nil {
		_ = conn.Close()
		t.Skip("port 50051 already in use")
	}
	host, port, user, pass, name, open := emptyPostgresDB(t)
	bin := buildAppBinary(t)

	cmd := exec.Command(bin, "database")
	cmd.Dir = "/" // no migrations/ dir on disk
	cmd.Env = append(os.Environ(),
		"PASSWORDEXCHANGE_DBDRIVER=postgres",
		"PASSWORDEXCHANGE_AUTOMIGRATE=true",
		"PASSWORDEXCHANGE_DBHOST="+host,
		fmt.Sprintf("PASSWORDEXCHANGE_DBPORT=%d", port),
		"PASSWORDEXCHANGE_DBUSER="+user,
		"PASSWORDEXCHANGE_DBPASS="+pass,
		"PASSWORDEXCHANGE_DBNAME="+name,
	)
	logFile := filepath.Join(t.TempDir(), "out.log")
	f, err := os.Create(logFile)
	require.NoError(t, err)
	defer f.Close()
	cmd.Stdout, cmd.Stderr = f, f
	require.NoError(t, cmd.Start())
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })

	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:50051", 300*time.Millisecond)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, 60*time.Second, 500*time.Millisecond, "gRPC server never started")

	db := open()
	defer db.Close()
	var version int64
	require.NoError(t, db.QueryRow("SELECT version FROM schema_migrations").Scan(&version))
	assert.EqualValues(t, 20260311000007, version)
}
