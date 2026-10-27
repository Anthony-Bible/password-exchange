package migrations_test

import (
	"io/fs"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/migrations"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baselineVersion = 20260311000007

func TestPostgresFS_ContainsBaselinePair(t *testing.T) {
	for _, name := range []string{
		"postgres/20260311000007_baseline.up.sql",
		"postgres/20260311000007_baseline.down.sql",
	} {
		b, err := fs.ReadFile(migrations.PostgresFS, name)
		require.NoError(t, err, name)
		assert.NotEmpty(t, strings.TrimSpace(string(b)), name)
	}
}

func TestPostgresFS_IsConsumableByIofsAtPostgresSubdir(t *testing.T) {
	src, err := iofs.New(migrations.PostgresFS, "postgres")
	require.NoError(t, err)
	t.Cleanup(func() { _ = src.Close() })

	first, err := src.First()
	require.NoError(t, err)
	assert.EqualValues(t, baselineVersion, first)

	_, _, err = src.ReadUp(first)
	require.NoError(t, err)
	_, _, err = src.ReadDown(first)
	require.NoError(t, err)
}

func TestPostgresBaseline_VersionMatchesLatestMySQLMigration(t *testing.T) {
	re := regexp.MustCompile(`^(\d+)_.*\.up\.sql$`)
	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	var versions []int
	for _, e := range entries {
		if m := re.FindStringSubmatch(e.Name()); m != nil && !e.IsDir() {
			v, convErr := strconv.Atoi(m[1])
			require.NoError(t, convErr)
			versions = append(versions, v)
		}
	}
	require.NotEmpty(t, versions, "expected the MySQL migrations to remain in app/migrations")
	sort.Ints(versions)
	assert.Equal(t, baselineVersion, versions[len(versions)-1],
		"the Postgres baseline must sit at the latest MySQL migration version")
}

func TestPostgresDownMigration_DropsAllTablesAndWarnsDestructive(t *testing.T) {
	b, err := fs.ReadFile(migrations.PostgresFS, "postgres/20260311000007_baseline.down.sql")
	require.NoError(t, err)
	body := string(b)

	assert.Contains(t, strings.ToUpper(body), "WARNING", "down migration needs a destructive-warning header")
	for _, table := range []string{"messages", "email_reminders", "file_upload_sessions"} {
		assert.Contains(t, body, table)
	}
	assert.Contains(t, strings.ToUpper(body), "DROP TABLE IF EXISTS")
}

func TestPostgresMigrationsDoNotLeakIntoMySQLSource(t *testing.T) {
	// golang-migrate's file source skips subdirectories; make sure the
	// Postgres files live only under postgres/ and not beside the MySQL ones.
	entries, err := os.ReadDir(".")
	require.NoError(t, err)
	for _, e := range entries {
		assert.NotContains(t, e.Name(), "baseline", "baseline must live in migrations/postgres/")
	}
}
