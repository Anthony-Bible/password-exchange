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

func TestIntegration_CleanDatabase(t *testing.T) {
	// 1. Setup clean DB
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "clean.db")

	// Create a separate migrations directory for the test with SQLite compatible SQL
	testMigrationsDir := filepath.Join(tmpDir, "migrations")
	err := os.MkdirAll(testMigrationsDir, 0o755)
	require.NoError(t, err)

	// Add SQLite compatible migrations
	m1Up := `CREATE TABLE messages (
	  messageid INTEGER PRIMARY KEY AUTOINCREMENT,
	  created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	  lastAccessed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	  viewed INTEGER DEFAULT 0,
	  message TEXT NOT NULL,
	  uniqueid TEXT UNIQUE
	);`
	err = os.WriteFile(filepath.Join(testMigrationsDir, "0001_initial.up.sql"), []byte(m1Up), 0o644)
	require.NoError(t, err)

	m2Up := `CREATE TABLE email_reminders (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    message_id INTEGER NOT NULL,
	    email_address TEXT NOT NULL,
	    FOREIGN KEY (message_id) REFERENCES messages(messageid)
	);`
	err = os.WriteFile(filepath.Join(testMigrationsDir, "0002_reminders.up.sql"), []byte(m2Up), 0o644)
	require.NoError(t, err)

	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer db.Close()

	// 2. Run Up
	err = Up(db, "sqlite3", testMigrationsDir)
	assert.NoError(t, err)

	// Re-open DB
	db, err = sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer db.Close()

	// 3. Verify Schema
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='messages'").Scan(&tableName)
	assert.NoError(t, err)
	assert.Equal(t, "messages", tableName)

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='email_reminders'").Scan(&tableName)
	assert.NoError(t, err)
	assert.Equal(t, "email_reminders", tableName)

	// Check version
	version, dirty, err := Version(db, "sqlite3", testMigrationsDir)
	assert.NoError(t, err)
	assert.Equal(t, uint(2), version)
	assert.False(t, dirty)
}

func TestIntegration_ExistingDatabase(t *testing.T) {
	// 1. Setup DB with existing table (simulating old system)
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "existing.db")
	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE messages (
	  messageid INTEGER PRIMARY KEY AUTOINCREMENT,
	  message TEXT NOT NULL
	);`)
	require.NoError(t, err)
	db.Close()

	// 2. Setup migrations
	testMigrationsDir := filepath.Join(tmpDir, "migrations")
	err = os.MkdirAll(testMigrationsDir, 0o755)
	require.NoError(t, err)

	// Migration 1: Initial schema (contains CREATE TABLE IF NOT EXISTS)
	m1Up := `CREATE TABLE IF NOT EXISTS messages (
	  messageid INTEGER PRIMARY KEY AUTOINCREMENT,
	  message TEXT NOT NULL,
	  uniqueid TEXT UNIQUE
	);`
	err = os.WriteFile(filepath.Join(testMigrationsDir, "0001_initial.up.sql"), []byte(m1Up), 0o644)
	require.NoError(t, err)

	// Migration 2: New functionality
	m2Up := `CREATE TABLE new_table (id INTEGER PRIMARY KEY);`
	err = os.WriteFile(filepath.Join(testMigrationsDir, "0002_new.up.sql"), []byte(m2Up), 0o644)
	require.NoError(t, err)

	// 3. Run Up
	db, err = sql.Open("sqlite3", dbPath)
	require.NoError(t, err)

	err = Up(db, "sqlite3", testMigrationsDir)
	// This might fail if the library expects to manage the schema from the start
	// But since it's "IF NOT EXISTS", it should pass.
	assert.NoError(t, err)

	// 4. Verify
	db, err = sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer db.Close()

	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='new_table'").Scan(&tableName)
	assert.NoError(t, err)
	assert.Equal(t, "new_table", tableName)

	version, _, err := Version(db, "sqlite3", testMigrationsDir)
	assert.NoError(t, err)
	assert.Equal(t, uint(2), version)
}
