//go:build integration

// Postgres harness for integration tests, mirroring StartMySQL.
//
// Running locally under podman (instead of Docker) needs testcontainers pointed
// at the podman socket and the Ryuk reaper disabled:
//
//	# macOS
//	export DOCKER_HOST=unix://$(podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}')
//	export TESTCONTAINERS_RYUK_DISABLED=true
//	go test -tags=integration ./...

package dbtest

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"net"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	// defaultPGImage uses the Google-hosted Docker Hub mirror for the same
	// rate-limit reason as defaultDBImage. Override with
	// PASSWORDEXCHANGE_TEST_PG_IMAGE.
	defaultPGImage = "mirror.gcr.io/library/postgres:17-alpine"
	pgImageEnvVar  = "PASSWORDEXCHANGE_TEST_PG_IMAGE"
)

// testPGImage returns the Postgres image to run, honoring the env override.
func testPGImage() string {
	if img := os.Getenv(pgImageEnvVar); img != "" {
		return img
	}
	return defaultPGImage
}

// StartPostgres launches a throwaway Postgres 17 container, applies the
// migrations found at the root of migrationsFS, and returns both an open
// *sql.DB (for seeding and assertions) and the DatabaseConfig a Postgres
// adapter needs to connect, with Host and Port kept separate. All resources are
// torn down automatically via t.Cleanup.
func StartPostgres(t *testing.T, migrationsFS fs.FS) (*sql.DB, contracts.DatabaseConfig) {
	t.Helper()

	// See StartMySQL: avoid pulling the Ryuk image unless the caller opted in.
	if _, ok := os.LookupEnv("TESTCONTAINERS_RYUK_DISABLED"); !ok {
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}

	ctx := context.Background()
	// BasicWaitStrategies waits for the second "ready to accept connections"
	// log line plus the listening port; Postgres logs "ready" once during the
	// init-script phase before restarting, so a single log wait is not enough.
	container, err := tcpostgres.Run(ctx, testPGImage(),
		tcpostgres.WithDatabase(testDBName),
		tcpostgres.WithUsername(testDBUser),
		tcpostgres.WithPassword(testDBPass),
		tcpostgres.BasicWaitStrategies(),
	)
	// Run can return a container alongside a startup error, so register
	// teardown before asserting; TerminateContainer is a no-op on nil.
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	})
	require.NoError(t, err, "start postgres container")

	host, err := container.Host(ctx)
	require.NoError(t, err)
	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	cfg := contracts.DatabaseConfig{
		Host:     host,
		Port:     int(mappedPort.Num()),
		User:     testDBUser,
		Password: testDBPass,
		Name:     testDBName,
	}
	dsn := PostgresDSN(cfg)

	// Run migrations on a dedicated connection: the pgx/v5 migrate driver's
	// Close() closes the *sql.DB it was given.
	migrationDB, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	waitForPing(t, migrationDB)
	applyPostgresMigrations(t, migrationDB, migrationsFS)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	waitForPing(t, db)

	return db, cfg
}

// PostgresDSN builds a pgx connection URL from cfg's separate Host and Port.
// TLS is disabled because it only targets the local test container.
func PostgresDSN(cfg contracts.DatabaseConfig) string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:     "/" + cfg.Name,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

// applyPostgresMigrations runs every up migration at the root of fsys against
// db using golang-migrate directly. db is closed when this returns.
func applyPostgresMigrations(t *testing.T, db *sql.DB, fsys fs.FS) {
	t.Helper()

	src, err := iofs.New(fsys, ".")
	require.NoError(t, err, "open migration source")
	drv, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	require.NoError(t, err, "create pgx migrate driver")
	m, err := migrate.NewWithInstance("iofs", src, "pgx5", drv)
	require.NoError(t, err, "create migrator")

	upErr := m.Up()
	srcErr, dbErr := m.Close()
	if errors.Is(upErr, migrate.ErrNoChange) {
		upErr = nil
	}
	require.NoError(t, upErr, "apply migrations")
	require.NoError(t, errors.Join(srcErr, dbErr), "close migrator")
}
