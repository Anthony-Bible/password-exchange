//go:build integration

package migrations_test

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/integration/dbtest"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/database/migrations"
	appmigrations "github.com/Anthony-Bible/password-exchange/app/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// baselineVersion is the Postgres baseline's version, equal to the latest
// MySQL migration.
const baselineVersion = uint(20260311000007)

var dbCounter atomic.Int64

// pgEnv is one PG17 container shared by the tests in this file; each test
// gets its own empty database via freshDatabase.
type pgEnv struct {
	admin *sql.DB
	cfg   contracts.DatabaseConfig
}

func startPG(t *testing.T) pgEnv {
	t.Helper()
	admin, cfg := dbtest.StartPostgresBaseline(t)
	return pgEnv{admin: admin, cfg: cfg}
}

// freshDatabase creates an empty database and returns a function that opens a
// new *sql.DB to it. golang-migrate closes the *sql.DB it is given, so every
// migrator call needs its own handle.
func (e pgEnv) freshDatabase(t *testing.T) func() *sql.DB {
	t.Helper()
	name := fmt.Sprintf("mig_%d", dbCounter.Add(1))
	_, err := e.admin.Exec("CREATE DATABASE " + name)
	require.NoError(t, err)

	cfg := e.cfg
	cfg.Name = name
	return func() *sql.DB {
		db, err := sql.Open("pgx", dbtest.PostgresDSN(cfg))
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		return db
	}
}

func baselineSource(t *testing.T) fs.FS {
	t.Helper()
	sub, err := fs.Sub(appmigrations.PostgresFS, "postgres")
	require.NoError(t, err)
	return sub
}

func TestPostgresBaseline(t *testing.T) {
	env := startPG(t)
	src := baselineSource(t)

	t.Run("Up, Down, Up round-trips", func(t *testing.T) {
		open := env.freshDatabase(t)

		require.NoError(t, migrations.Up(open(), migrations.Postgres, src))
		version, dirty, err := migrations.Version(open(), migrations.Postgres, src)
		require.NoError(t, err)
		assert.Equal(t, baselineVersion, version)
		assert.False(t, dirty)

		require.NoError(t, migrations.Down(open(), migrations.Postgres, src))
		version, _, err = migrations.Version(open(), migrations.Postgres, src)
		require.NoError(t, err)
		assert.Equal(t, uint(0), version)
		var tables int
		require.NoError(t, open().QueryRow(
			`SELECT count(*) FROM information_schema.tables
			 WHERE table_schema = 'public' AND table_name <> 'schema_migrations'`).Scan(&tables))
		assert.Zero(t, tables, "down migration should drop every application table")

		require.NoError(t, migrations.Up(open(), migrations.Postgres, src))
		version, dirty, err = migrations.Version(open(), migrations.Postgres, src)
		require.NoError(t, err)
		assert.Equal(t, baselineVersion, version)
		assert.False(t, dirty)

		// A further Up is a no-op: ErrNoChange is swallowed.
		assert.NoError(t, migrations.Up(open(), migrations.Postgres, src))
	})

	t.Run("concurrent Up serialises on the advisory lock", func(t *testing.T) {
		open := env.freshDatabase(t)
		dbs := []*sql.DB{open(), open()}

		var wg sync.WaitGroup
		errs := make([]error, len(dbs))
		for i, db := range dbs {
			wg.Add(1)
			go func(i int, db *sql.DB) {
				defer wg.Done()
				errs[i] = migrations.Up(db, migrations.Postgres, src)
			}(i, db)
		}
		wg.Wait()

		for i, err := range errs {
			assert.NoError(t, err, "goroutine %d", i)
		}
		var rows int
		require.NoError(t, open().QueryRow("SELECT count(*) FROM schema_migrations").Scan(&rows))
		assert.Equal(t, 1, rows)
	})

	t.Run("deleting a message cascades to its reminders", func(t *testing.T) {
		open := env.freshDatabase(t)
		require.NoError(t, migrations.Up(open(), migrations.Postgres, src))
		db := open()

		var id int
		require.NoError(t, db.QueryRow(
			`INSERT INTO messages (message, uniqueid) VALUES ('ciphertext', 'u-1') RETURNING messageid`).Scan(&id))
		_, err := db.Exec(
			`INSERT INTO email_reminders (message_id, email_address) VALUES ($1, 'a@example.com'), ($1, 'b@example.com')`, id)
		require.NoError(t, err)

		_, err = db.Exec("DELETE FROM messages WHERE messageid = $1", id)
		require.NoError(t, err)

		var remaining int
		require.NoError(t, db.QueryRow(
			"SELECT count(*) FROM email_reminders WHERE message_id = $1", id).Scan(&remaining))
		assert.Zero(t, remaining)
	})

	t.Run("schema matches MySQL after all migrations", func(t *testing.T) {
		open := env.freshDatabase(t)
		require.NoError(t, migrations.Up(open(), migrations.Postgres, src))
		db := open()

		for table, want := range expectedSchema {
			assert.Equal(t, want, actualColumns(t, db, table), "columns of %s", table)
		}

		for table, cols := range map[string]string{
			"messages":             "uniqueid",
			"email_reminders":      "message_id,email_address",
			"file_upload_sessions": "file_id",
		} {
			assert.Contains(t, uniqueConstraints(t, db, table), cols, "unique constraint on %s", table)
		}

		var delType string
		require.NoError(t, db.QueryRow(
			`SELECT confdeltype FROM pg_constraint
			 WHERE contype = 'f' AND conrelid = 'email_reminders'::regclass
			   AND confrelid = 'messages'::regclass`).Scan(&delType))
		assert.Equal(t, "c", delType, "email_reminders.message_id must be ON DELETE CASCADE")
	})
}

// column is the slice of information_schema.columns the parity contract
// compares. MaxLength is set for character varying columns only.
type column struct {
	Name       string
	Nullable   bool
	Type       string
	MaxLength  int
	HasDefault bool // a DEFAULT expression or an identity (MySQL AUTO_INCREMENT)
}

// expectedSchema is the MySQL schema after migrations 20260311000000 through
// 20260311000007, translated to Postgres types: int → integer,
// TIMESTAMP → timestamptz, TINYINT(1) → boolean, BLOB → bytea, JSON → jsonb.
// MySQL's `lastAccessed` is folded to lower case.
var expectedSchema = map[string][]column{
	"messages": {
		{Name: "messageid", Type: "integer", HasDefault: true},
		{Name: "created", Type: "timestamp with time zone", HasDefault: true},
		{Name: "lastaccessed", Type: "timestamp with time zone", HasDefault: true},
		{Name: "viewed", Nullable: true, Type: "integer", HasDefault: true},
		{Name: "firstname", Nullable: true, Type: "character varying", MaxLength: 100},
		{Name: "lastname", Nullable: true, Type: "character varying", MaxLength: 100},
		{Name: "other_firstname", Nullable: true, Type: "character varying", MaxLength: 100},
		{Name: "other_lastname", Nullable: true, Type: "character varying", MaxLength: 100},
		{Name: "message", Type: "text"},
		{Name: "email", Nullable: true, Type: "character varying", MaxLength: 255},
		{Name: "other_email", Nullable: true, Type: "character varying", MaxLength: 255},
		{Name: "uniqueid", Nullable: true, Type: "character varying", MaxLength: 255},
		{Name: "is_client_encrypted", Type: "boolean", HasDefault: true},
		{Name: "view_count", Type: "integer", HasDefault: true},
		{Name: "max_view_count", Type: "integer", HasDefault: true},
		{Name: "expires_at", Type: "timestamp with time zone", HasDefault: true},
	},
	"email_reminders": {
		{Name: "id", Type: "integer", HasDefault: true},
		{Name: "message_id", Type: "integer"},
		{Name: "email_address", Type: "character varying", MaxLength: 255},
		{Name: "reminder_count", Nullable: true, Type: "integer", HasDefault: true},
		{Name: "last_reminder_sent", Nullable: true, Type: "timestamp with time zone", HasDefault: true},
		{Name: "created_at", Nullable: true, Type: "timestamp with time zone", HasDefault: true},
	},
	"file_upload_sessions": {
		{Name: "session_id", Type: "character varying", MaxLength: 255},
		{Name: "file_id", Type: "character varying", MaxLength: 255},
		{Name: "message_id", Type: "character varying", MaxLength: 255, HasDefault: true},
		{Name: "upload_id", Type: "character varying", MaxLength: 255, HasDefault: true},
		{Name: "filename", Type: "character varying", MaxLength: 255, HasDefault: true},
		{Name: "content_type", Type: "character varying", MaxLength: 255, HasDefault: true},
		{Name: "total_size", Type: "bigint", HasDefault: true},
		{Name: "total_chunks", Type: "integer", HasDefault: true},
		{Name: "status", Type: "character varying", MaxLength: 50, HasDefault: true},
		{Name: "encryption_key", Nullable: true, Type: "bytea"},
		{Name: "completed_parts", Type: "jsonb"},
		{Name: "created_at", Type: "timestamp with time zone", HasDefault: true},
		{Name: "expires_at", Type: "timestamp with time zone", HasDefault: true},
	},
}

func actualColumns(t *testing.T, db *sql.DB, table string) []column {
	t.Helper()
	rows, err := db.Query(`
		SELECT column_name,
		       is_nullable = 'YES',
		       data_type,
		       coalesce(character_maximum_length, 0),
		       column_default IS NOT NULL OR is_identity = 'YES'
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
		ORDER BY ordinal_position`, table)
	require.NoError(t, err)
	defer rows.Close()

	var cols []column
	for rows.Next() {
		var c column
		require.NoError(t, rows.Scan(&c.Name, &c.Nullable, &c.Type, &c.MaxLength, &c.HasDefault))
		cols = append(cols, c)
	}
	require.NoError(t, rows.Err())
	return cols
}

// uniqueConstraints returns each UNIQUE constraint on table as its
// comma-joined column list, in key order.
func uniqueConstraints(t *testing.T, db *sql.DB, table string) []string {
	t.Helper()
	rows, err := db.Query(`
		SELECT string_agg(a.attname, ',' ORDER BY k.ord)
		FROM pg_constraint c
		CROSS JOIN LATERAL unnest(c.conkey) WITH ORDINALITY AS k(attnum, ord)
		JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = k.attnum
		WHERE c.contype = 'u' AND c.conrelid = $1::regclass
		GROUP BY c.oid`, table)
	require.NoError(t, err)
	defer rows.Close()

	var out []string
	for rows.Next() {
		var cols string
		require.NoError(t, rows.Scan(&cols))
		out = append(out, strings.TrimSpace(cols))
	}
	require.NoError(t, rows.Err())
	return out
}
