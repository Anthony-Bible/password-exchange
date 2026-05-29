//go:build integration

// Package dbtest provides shared helpers for integration tests that need a real
// MySQL-compatible database. It is compiled only under the `integration` build
// tag so the default `go test ./...` run stays fast and Docker-free.
package dbtest

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/database/migrations"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
)

const (
	testDBName = "passwordexchange"
	testDBUser = "testuser"
	testDBPass = "testpass"
	// defaultDBImage points at the Google-hosted Docker Hub mirror rather than
	// docker.io to avoid anonymous pull-rate limits on shared CI/sandbox hosts.
	// MySQL is used (not MariaDB) because the testcontainers mysql module's
	// readiness probe keys off a MySQL-specific startup log line. Override with
	// PASSWORDEXCHANGE_TEST_MYSQL_IMAGE when a different registry or version is
	// needed.
	defaultDBImage = "mirror.gcr.io/library/mysql:8.0"
	imageEnvVar    = "PASSWORDEXCHANGE_TEST_MYSQL_IMAGE"
	pingTimeout    = 30 * time.Second
	pingInterval   = 250 * time.Millisecond

	// relaxedSQLMode drops STRICT_TRANS_TABLES / NO_ZERO_DATE from MySQL 8.0's
	// default sql_mode so the expires_at migration's zero-date backfill
	// (`WHERE expires_at = '0000-00-00 00:00:00'`) runs the same way it does on
	// the production MariaDB instance instead of erroring out.
	relaxedSQLMode = "[mysqld]\nsql_mode=NO_ENGINE_SUBSTITUTION\n"
)

// testDBImage returns the MySQL image to run, honoring the env override.
func testDBImage() string {
	if img := os.Getenv(imageEnvVar); img != "" {
		return img
	}
	return defaultDBImage
}

// StartMySQL launches a throwaway MySQL container, applies the production
// migrations from app/migrations, and returns both an open *sql.DB (for seeding
// and assertions) and the DatabaseConfig the MySQLAdapter needs to connect. All
// resources are torn down automatically via t.Cleanup.
//
// Tests using this helper require Docker to be available; they are skipped by a
// plain `go test ./...` because the file is gated behind the `integration` tag.
func StartMySQL(t *testing.T) (*sql.DB, contracts.DatabaseConfig) {
	t.Helper()

	// Disable the Ryuk reaper so testcontainers doesn't pull an extra image
	// from docker.io (which would hit the same rate limit). Container teardown
	// is handled explicitly via t.Cleanup below. Skip if the caller has already
	// configured Ryuk behavior.
	if _, ok := os.LookupEnv("TESTCONTAINERS_RYUK_DISABLED"); !ok {
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}

	cnfPath := filepath.Join(t.TempDir(), "relaxed-sqlmode.cnf")
	require.NoError(t, os.WriteFile(cnfPath, []byte(relaxedSQLMode), 0o644))

	ctx := context.Background()
	container, err := tcmysql.Run(ctx, testDBImage(),
		tcmysql.WithDatabase(testDBName),
		tcmysql.WithUsername(testDBUser),
		tcmysql.WithPassword(testDBPass),
		tcmysql.WithConfigFile(cnfPath),
	)
	require.NoError(t, err, "start mysql container")
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate mysql container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)
	mappedPort, err := container.MappedPort(ctx, "3306/tcp")
	require.NoError(t, err)
	hostPort := net.JoinHostPort(host, mappedPort.Port())

	// multiStatements lets golang-migrate apply migration files that contain
	// more than one statement (e.g. CREATE TABLE + CREATE INDEX).
	dsn := fmt.Sprintf("%s:%s@(%s)/%s?parseTime=true&multiStatements=true",
		testDBUser, testDBPass, hostPort, testDBName)

	// Run migrations on a dedicated connection: golang-migrate closes the
	// *sql.DB handed to it (via WithInstance) when it finishes, so this handle
	// must not be the one the test later uses for seeding/assertions.
	migrationDB, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	waitForPing(t, migrationDB)
	require.NoError(t, migrations.Up(migrationDB, migrationsDir(t)), "apply migrations")
	_ = migrationDB.Close()

	// Fresh connection returned to the caller for the lifetime of the test.
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	waitForPing(t, db)

	cfg := contracts.DatabaseConfig{
		Host:     hostPort,
		User:     testDBUser,
		Password: testDBPass,
		Name:     testDBName,
	}
	return db, cfg
}

// waitForPing blocks until the database answers a ping or the timeout elapses.
// The module's wait strategy usually has the server ready, but a short retry
// loop guards against the brief window before MySQL accepts TCP connections.
func waitForPing(t *testing.T, db *sql.DB) {
	t.Helper()
	deadline := time.Now().Add(pingTimeout)
	for {
		if err := db.Ping(); err == nil {
			return
		}
		if time.Now().After(deadline) {
			require.NoError(t, db.Ping(), "database did not become ready within %s", pingTimeout)
			return
		}
		time.Sleep(pingInterval)
	}
}

// migrationsDir resolves the absolute path to app/migrations relative to this
// source file so the helper works regardless of the test's working directory.
func migrationsDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "resolve caller for migrations path")
	// thisFile = app/internal/integration/dbtest/mysql.go
	// migrations = app/migrations
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "migrations")
}
