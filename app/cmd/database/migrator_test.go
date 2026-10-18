package database

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/database/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveMigrationTarget_MySQLIsDefault(t *testing.T) {
	for _, driver := range []string{"", "mysql"} {
		t.Run("dbdriver="+driver, func(t *testing.T) {
			target, err := resolveMigrationTarget(config.PassConfig{
				DbDriver: driver,
				DbUser:   "user",
				DbPass:   "pass",
				DbHost:   "mariadb",
				DbPort:   3306,
				DbName:   "passwordexchange",
			})
			require.NoError(t, err)

			assert.Equal(t, "mysql", target.sqlDriver)
			assert.Equal(t, "user:pass@tcp(mariadb:3306)/passwordexchange?parseTime=true&multiStatements=true", target.dsn)
			assert.Equal(t, migrations.MySQL, target.dialect)
			assert.Equal(t, "migrations", target.createDir)
			assert.Equal(t, os.DirFS("migrations"), target.source)
		})
	}
}

func TestResolveMigrationTarget_Postgres(t *testing.T) {
	target, err := resolveMigrationTarget(config.PassConfig{
		DbDriver: "postgres",
		DbUser:   "user@x",
		DbPass:   "p:a/s#s",
		DbHost:   "127.0.0.1",
		DbPort:   6543,
		DbName:   "passwordexchange",
	})
	require.NoError(t, err)

	assert.Equal(t, "pgx", target.sqlDriver)
	assert.Equal(t, migrations.Postgres, target.dialect)
	assert.Equal(t, filepath.Join("migrations", "postgres"), target.createDir)

	parsed, err := pgx.ParseConfig(target.dsn)
	require.NoError(t, err)
	assert.Equal(t, "user@x", parsed.User)
	assert.Equal(t, "p:a/s#s", parsed.Password)
	assert.Equal(t, "127.0.0.1", parsed.Host)
	assert.Equal(t, uint16(6543), parsed.Port)
	assert.Equal(t, "passwordexchange", parsed.Database)

	// The source is the embedded baseline, not a directory on disk.
	_, err = fs.Stat(target.source, "20260311000007_baseline.up.sql")
	assert.NoError(t, err)
}

func TestResolveMigrationTarget_RejectsUnknownDriver(t *testing.T) {
	_, err := resolveMigrationTarget(config.PassConfig{DbDriver: "oracle"})
	assert.ErrorContains(t, err, "oracle")
}

func TestDefaultMigratorCreate_WritesToCreateDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "migrations", "postgres")
	m, err := NewMigrator(nil, migrations.Postgres, os.DirFS(dir), dir)
	require.NoError(t, err)

	require.NoError(t, m.Create("add_widgets"))

	ups, err := filepath.Glob(filepath.Join(dir, "*_add_widgets.up.sql"))
	require.NoError(t, err)
	downs, err := filepath.Glob(filepath.Join(dir, "*_add_widgets.down.sql"))
	require.NoError(t, err)
	assert.Len(t, ups, 1)
	assert.Len(t, downs, 1)
}
