//go:build integration

package dbtest

import (
	"context"
	"database/sql"
	"testing"
	"testing/fstest"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartPostgres_AppliesMigrationsAndReturnsUsableConfig(t *testing.T) {
	fixtures := fstest.MapFS{
		"1_fixture.up.sql":   {Data: []byte("CREATE TABLE fixture (id INT PRIMARY KEY);")},
		"1_fixture.down.sql": {Data: []byte("DROP TABLE fixture;")},
	}

	db, cfg := StartPostgres(t, fixtures)

	var version string
	require.NoError(t, db.QueryRow("SELECT version()").Scan(&version))
	assert.Contains(t, version, "PostgreSQL 17")

	var exists bool
	require.NoError(t, db.QueryRow("SELECT to_regclass('public.fixture') IS NOT NULL").Scan(&exists))
	assert.True(t, exists, "fixture migration should have created the fixture table")

	// Host and Port are returned separately and must be enough on their own to
	// reach the database, independently of the *sql.DB handed back above.
	assert.NotContains(t, cfg.Host, ":", "Host must not carry the port")
	assert.NotZero(t, cfg.Port)

	viaCfg, err := sql.Open("pgx", PostgresDSN(cfg))
	require.NoError(t, err)
	t.Cleanup(func() { _ = viaCfg.Close() })
	require.NoError(t, viaCfg.PingContext(context.Background()))

	_, err = viaCfg.Exec("INSERT INTO fixture (id) VALUES (1)")
	require.NoError(t, err)
	var count int
	require.NoError(t, db.QueryRow("SELECT count(*) FROM fixture").Scan(&count))
	assert.Equal(t, 1, count)
}
