package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetConfig isolates package-level state (viper, cfg, migrator) per test.
func resetConfig(t *testing.T) {
	t.Helper()
	oldMigrator, oldCfg := migrator, cfg
	migrator, cfg = nil, Config{}
	viper.Reset()
	viper.SetEnvPrefix("passwordexchange")
	viper.AutomaticEnv()
	t.Cleanup(func() {
		migrator, cfg = oldMigrator, oldCfg
		viper.Reset()
	})
}

func TestDBDriver_DefaultsToMySQL(t *testing.T) {
	resetConfig(t)
	viper.Set("dbhost", "127.0.0.1")
	viper.Set("dbport", 3306)

	require.NoError(t, initConfigAndMigrator())

	assert.Equal(t, "mysql", cfg.PassConfig.DbDriver)
	assert.NotNil(t, migrator)
}

func TestDBDriver_ReadFromEnvironment(t *testing.T) {
	resetConfig(t)
	t.Setenv("PASSWORDEXCHANGE_DBDRIVER", "postgres")
	t.Setenv("PASSWORDEXCHANGE_DBHOST", "127.0.0.1")
	t.Setenv("PASSWORDEXCHANGE_DBPORT", "5432")

	require.NoError(t, initConfigAndMigrator())

	assert.Equal(t, "postgres", cfg.PassConfig.DbDriver)
	assert.NotNil(t, migrator, "postgres must yield a usable migrator (pgx driver registered)")
}

func TestDBDriver_PostgresInitSucceedsWithoutMigrationsDirOnDisk(t *testing.T) {
	resetConfig(t)
	t.Chdir(t.TempDir())
	viper.Set("dbdriver", "postgres")
	viper.Set("dbhost", "127.0.0.1")
	viper.Set("dbport", 5432)

	require.NoError(t, initConfigAndMigrator())
	assert.NotNil(t, migrator)
}

func TestDBDriver_UnknownDriverIsRejected(t *testing.T) {
	resetConfig(t)
	viper.Set("dbdriver", "oracle")

	err := initConfigAndMigrator()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "oracle")
}

func TestMigrateCreate_PostgresWritesUnderPostgresSubdir(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m, err := NewMigrator(nil, "postgres", "migrations")
	require.NoError(t, err)
	require.NoError(t, m.Create("add_widget"))

	ups, err := filepath.Glob(filepath.Join(dir, "migrations", "postgres", "*_add_widget.up.sql"))
	require.NoError(t, err)
	downs, err := filepath.Glob(filepath.Join(dir, "migrations", "postgres", "*_add_widget.down.sql"))
	require.NoError(t, err)
	assert.Len(t, ups, 1)
	assert.Len(t, downs, 1)

	top, err := filepath.Glob(filepath.Join(dir, "migrations", "*_add_widget.*.sql"))
	require.NoError(t, err)
	assert.Empty(t, top, "postgres migrations must not pollute the MySQL directory")
}

func TestMigrateCreate_MySQLKeepsWritingToMigrationsDir(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m, err := NewMigrator(nil, "mysql", "migrations")
	require.NoError(t, err)
	require.NoError(t, m.Create("add_widget"))

	ups, err := filepath.Glob(filepath.Join(dir, "migrations", "*_add_widget.up.sql"))
	require.NoError(t, err)
	assert.Len(t, ups, 1)
	_, statErr := os.Stat(filepath.Join(dir, "migrations", "postgres"))
	assert.True(t, os.IsNotExist(statErr))
}

// A failing automigrate must make the `database` process exit non-zero so
// Kubernetes restarts the pod, instead of returning cleanly without serving.
func TestBinary_FailingAutomigrateExitsNonZero(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the binary")
	}
	bin := buildAppBinary(t)

	res := runBinary(t, bin, t.TempDir(), 60, map[string]string{
		"PASSWORDEXCHANGE_DBDRIVER":    "postgres",
		"PASSWORDEXCHANGE_AUTOMIGRATE": "true",
		"PASSWORDEXCHANGE_DBHOST":      "127.0.0.1",
		"PASSWORDEXCHANGE_DBPORT":      "1", // nothing listens here
		"PASSWORDEXCHANGE_DBUSER":      "u",
		"PASSWORDEXCHANGE_DBPASS":      "p",
		"PASSWORDEXCHANGE_DBNAME":      "d",
	})

	require.False(t, res.timedOut, "process kept running instead of exiting: %s", res.output)
	assert.NotZero(t, res.exitCode, "output: %s", res.output)
}
