package migrations

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (string, *sql.DB) {
	tmpFile := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite3", tmpFile)
	require.NoError(t, err)
	return tmpFile, db
}

func createTempMigrations(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "migrations-test")
	require.NoError(t, err)

	// Migration 1: Create users table
	m1Up := "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);"
	m1Down := "DROP TABLE users;"
	err = os.WriteFile(filepath.Join(tmpDir, "0001_initial.up.sql"), []byte(m1Up), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "0001_initial.down.sql"), []byte(m1Down), 0o644)
	require.NoError(t, err)

	// Migration 2: Add email to users
	m2Up := "ALTER TABLE users ADD COLUMN email TEXT;"
	m2Down := "ALTER TABLE users DROP COLUMN email;"
	err = os.WriteFile(filepath.Join(tmpDir, "0002_add_email.up.sql"), []byte(m2Up), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "0002_add_email.down.sql"), []byte(m2Down), 0o644)
	require.NoError(t, err)

	return tmpDir
}

func TestUp(t *testing.T) {
	t.Run("Successfully apply migrations", func(t *testing.T) {
		dbPath, db := setupTestDB(t)
		defer db.Close()
		dir := createTempMigrations(t)
		defer os.RemoveAll(dir)

		err := Up(db, "sqlite3", dir)
		assert.NoError(t, err)

		// Re-open DB since Up closes it
		db, err = sql.Open("sqlite3", dbPath)
		require.NoError(t, err)
		defer db.Close()

		// Verify table exists
		var name string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, "users", name)

		version, dirty, err := Version(db, "sqlite3", dir)
		assert.NoError(t, err)
		assert.Equal(t, uint(2), version)
		assert.False(t, dirty)
	})

	t.Run("Error on missing directory", func(t *testing.T) {
		_, db := setupTestDB(t)
		defer db.Close()

		err := Up(db, "sqlite3", "/non/existent/path")
		assert.Error(t, err)
	})

	t.Run("Error on invalid migration syntax", func(t *testing.T) {
		_, db := setupTestDB(t)
		defer db.Close()
		tmpDir, _ := os.MkdirTemp("", "bad-migration")
		defer os.RemoveAll(tmpDir)

		os.WriteFile(filepath.Join(tmpDir, "0001_bad.up.sql"), []byte("INVALID SQL;"), 0o644)

		err := Up(db, "sqlite3", tmpDir)
		assert.Error(t, err)
	})
}

func TestDown(t *testing.T) {
	t.Run("Successfully roll back last migration", func(t *testing.T) {
		dbPath, db := setupTestDB(t)
		defer db.Close()
		dir := createTempMigrations(t)
		defer os.RemoveAll(dir)

		// Apply all
		err := Up(db, "sqlite3", dir)
		require.NoError(t, err)

		// Re-open DB
		db, err = sql.Open("sqlite3", dbPath)
		require.NoError(t, err)

		// Roll back one
		err = Down(db, "sqlite3", dir)
		assert.NoError(t, err)

		// Re-open DB
		db, err = sql.Open("sqlite3", dbPath)
		require.NoError(t, err)
		defer db.Close()

		version, _, err := Version(db, "sqlite3", dir)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), version)
	})

	t.Run("Error when no migrations to roll back", func(t *testing.T) {
		_, db := setupTestDB(t)
		defer db.Close()
		dir := createTempMigrations(t)
		defer os.RemoveAll(dir)

		err := Down(db, "sqlite3", dir)
		assert.Error(t, err)
	})
}

func TestVersion(t *testing.T) {
	t.Run("Initial version is 0", func(t *testing.T) {
		_, db := setupTestDB(t)
		defer db.Close()
		dir := createTempMigrations(t)
		defer os.RemoveAll(dir)

		version, dirty, err := Version(db, "sqlite3", dir)
		assert.NoError(t, err)
		assert.Equal(t, uint(0), version)
		assert.False(t, dirty)
	})

	t.Run("Reports dirty state correctly", func(t *testing.T) {
		dbPath, db := setupTestDB(t)
		defer db.Close()
		tmpDir, _ := os.MkdirTemp("", "dirty-test")
		defer os.RemoveAll(tmpDir)

		// Create a migration that will fail
		os.WriteFile(filepath.Join(tmpDir, "0001_fail.up.sql"), []byte("CREATE TABLE test (id INT); FAIL ME;"), 0o644)

		_ = Up(db, "sqlite3", tmpDir) // This is expected to fail

		// Re-open DB
		db, _ = sql.Open("sqlite3", dbPath)
		defer db.Close()

		_, dirty, err := Version(db, "sqlite3", tmpDir)
		assert.NoError(t, err)
		assert.True(t, dirty)
	})
}
