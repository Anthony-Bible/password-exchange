package migrations

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMigrator_RejectsUnknownDialect(t *testing.T) {
	_, db := setupTestDB(t)
	defer db.Close()

	_, err := NewMigrator(db, Dialect("oracle"), t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "oracle")
}

func TestNewMigrator_AcceptsAllSupportedDialects(t *testing.T) {
	_, db := setupTestDB(t)
	defer db.Close()

	for _, d := range []Dialect{DialectMySQL, DialectPostgres, DialectSQLite3} {
		m, err := NewMigrator(db, d, t.TempDir())
		require.NoError(t, err, string(d))
		assert.NotNil(t, m)
	}
}

func TestDialectConstants_UseConfigSpellings(t *testing.T) {
	// These strings are what PASSWORDEXCHANGE_DBDRIVER carries.
	assert.Equal(t, "mysql", string(DialectMySQL))
	assert.Equal(t, "postgres", string(DialectPostgres))
	assert.Equal(t, "sqlite3", string(DialectSQLite3))
}

// The dialect is chosen explicitly by the caller; it must not be inferred from
// the concrete driver type of the *sql.DB. A sqlite connection driven with the
// sqlite3 dialect keeps working (regression guard for the %T sniffing removal).
func TestUp_SQLiteDialectStillAppliesFileMigrations(t *testing.T) {
	dbPath, db := setupTestDB(t)
	defer db.Close()
	dir := createTempMigrations(t)

	require.NoError(t, Up(db, DialectSQLite3, dir))

	db2, _ := setupTestDBAt(t, dbPath)
	defer db2.Close()
	v, dirty, err := Version(db2, DialectSQLite3, dir)
	require.NoError(t, err)
	assert.EqualValues(t, 2, v)
	assert.False(t, dirty)
}

func setupTestDBAt(t *testing.T, path string) (*sql.DB, string) {
	t.Helper()
	db, err := sql.Open("sqlite3", path)
	require.NoError(t, err)
	return db, path
}
