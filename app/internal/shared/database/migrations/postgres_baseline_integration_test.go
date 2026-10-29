//go:build integration

// Postgres 17 integration tests for the embedded baseline migration.
// See app/internal/integration/dbtest/postgres.go for the podman setup.

package migrations_test

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/Anthony-Bible/password-exchange/app/internal/integration/dbtest"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/database/dsn"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/database/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baselineVersion = 20260311000007

// freshDatabase starts (or reuses) a PG17 container and returns an opener that
// yields a NEW *sql.DB on a brand-new empty database each time it is called
// for the same database. golang-migrate closes the *sql.DB it is given, so
// every Up/Down/Version needs its own handle.
func freshDatabase(t *testing.T) func() *sql.DB {
	t.Helper()
	admin, cfg := dbtest.StartPostgres(t, fstest.MapFS{
		"1_harness.up.sql":   {Data: []byte("CREATE TABLE harness_fixture (id int);")},
		"1_harness.down.sql": {Data: []byte("DROP TABLE harness_fixture;")},
	})
	name := fmt.Sprintf("fresh_%s", sanitizeName(t.Name()))
	_, err := admin.Exec(fmt.Sprintf(`CREATE DATABASE %q`, name))
	require.NoError(t, err)

	url := dsn.PostgresURL(cfg.Host, cfg.Port, cfg.User, cfg.Password, name)
	return func() *sql.DB {
		db, err := sql.Open("pgx", url)
		require.NoError(t, err)
		return db
	}
}

func sanitizeName(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			out = append(out, r|0x20)
		} else {
			out = append(out, '_')
		}
	}
	if len(out) > 40 {
		out = out[len(out)-40:]
	}
	return string(out)
}

func TestPostgresBaseline_UpYieldsBaselineVersionNotDirty(t *testing.T) {
	open := freshDatabase(t)

	require.NoError(t, migrations.Up(open(), migrations.DialectPostgres, ""))

	v, dirty, err := migrations.Version(open(), migrations.DialectPostgres, "")
	require.NoError(t, err)
	assert.EqualValues(t, baselineVersion, v)
	assert.False(t, dirty)
}

func TestPostgresBaseline_DownThenUpRoundTrips(t *testing.T) {
	open := freshDatabase(t)

	require.NoError(t, migrations.Up(open(), migrations.DialectPostgres, ""))
	require.NoError(t, migrations.Down(open(), migrations.DialectPostgres, ""))

	db := open()
	var n int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM information_schema.tables
		WHERE table_schema='public' AND table_name IN ('messages','email_reminders','file_upload_sessions')`).Scan(&n))
	assert.Zero(t, n, "down migration must drop all three tables")
	_ = db.Close()

	require.NoError(t, migrations.Up(open(), migrations.DialectPostgres, ""))
	v, dirty, err := migrations.Version(open(), migrations.DialectPostgres, "")
	require.NoError(t, err)
	assert.EqualValues(t, baselineVersion, v)
	assert.False(t, dirty)
}

func TestPostgresBaseline_FurtherUpIsNoOp(t *testing.T) {
	open := freshDatabase(t)

	require.NoError(t, migrations.Up(open(), migrations.DialectPostgres, ""))
	assert.NoError(t, migrations.Up(open(), migrations.DialectPostgres, ""), "ErrNoChange must be swallowed")
}

func TestPostgresBaseline_ConcurrentUpBothSucceedWithSingleVersionRow(t *testing.T) {
	open := freshDatabase(t)

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = migrations.Up(open(), migrations.DialectPostgres, "")
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		assert.NoError(t, err, "concurrent Up #%d", i)
	}

	db := open()
	defer db.Close()
	var rows int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&rows))
	assert.Equal(t, 1, rows)
}

// ---- schema parity (against the MySQL schema after migrations 0..7) ----

type colExpect struct {
	typeClass  string // information_schema.columns.data_type
	nullable   bool
	hasDefault bool // column_default present or identity column
}

var expectedColumns = map[string]map[string]colExpect{
	"messages": {
		"messageid":           {"integer", false, true},
		"created":             {"timestamp with time zone", false, true},
		"lastaccessed":        {"timestamp with time zone", false, true},
		"viewed":              {"integer", true, true},
		"firstname":           {"character varying", true, false},
		"lastname":            {"character varying", true, false},
		"other_firstname":     {"character varying", true, false},
		"other_lastname":      {"character varying", true, false},
		"message":             {"text", false, false},
		"email":               {"character varying", true, false},
		"other_email":         {"character varying", true, false},
		"uniqueid":            {"character varying", true, false},
		"is_client_encrypted": {"boolean", false, true},
		"view_count":          {"integer", false, true},
		"max_view_count":      {"integer", false, true},
		"expires_at":          {"timestamp with time zone", false, true},
	},
	"email_reminders": {
		"id":                 {"integer", false, true},
		"message_id":         {"integer", false, false},
		"email_address":      {"character varying", false, false},
		"reminder_count":     {"integer", true, true},
		"last_reminder_sent": {"timestamp with time zone", true, true},
		"created_at":         {"timestamp with time zone", true, true},
	},
	"file_upload_sessions": {
		"session_id":      {"character varying", false, false},
		"file_id":         {"character varying", false, false},
		"message_id":      {"character varying", false, true},
		"upload_id":       {"character varying", false, true},
		"filename":        {"character varying", false, true},
		"content_type":    {"character varying", false, true},
		"total_size":      {"bigint", false, true},
		"total_chunks":    {"integer", false, true},
		"status":          {"character varying", false, true},
		"encryption_key":  {"bytea", true, false},
		"completed_parts": {"jsonb", false, false},
		"created_at":      {"timestamp with time zone", false, true},
		"expires_at":      {"timestamp with time zone", false, true},
	},
}

func TestPostgresBaseline_SchemaMatchesMySQLParityExpectation(t *testing.T) {
	db, _ := dbtest.StartPostgresBaseline(t)

	for table, want := range expectedColumns {
		rows, err := db.Query(`SELECT column_name, data_type, is_nullable = 'YES',
				(column_default IS NOT NULL OR is_identity = 'YES')
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1`, table)
		require.NoError(t, err)

		got := map[string]colExpect{}
		for rows.Next() {
			var name string
			var c colExpect
			require.NoError(t, rows.Scan(&name, &c.typeClass, &c.nullable, &c.hasDefault))
			got[name] = c
		}
		require.NoError(t, rows.Err())
		_ = rows.Close()

		assert.Equal(t, want, got, "columns of %s", table)
	}
}

func TestPostgresBaseline_UniqueConstraintsExist(t *testing.T) {
	db, _ := dbtest.StartPostgresBaseline(t)

	cases := []struct {
		table string
		cols  string
	}{
		{"messages", "uniqueid"},
		{"email_reminders", "message_id,email_address"},
		{"file_upload_sessions", "file_id"},
	}
	for _, c := range cases {
		var cols string
		err := db.QueryRow(`
			SELECT string_agg(a.attname, ',' ORDER BY k.ord)
			FROM pg_constraint con
			JOIN pg_class rel ON rel.oid = con.conrelid
			CROSS JOIN LATERAL unnest(con.conkey) WITH ORDINALITY AS k(attnum, ord)
			JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = k.attnum
			WHERE con.contype = 'u' AND rel.relname = $1 AND con.conname IN ('messages_uniqueid_key','uq_email_reminders_message_email','uq_file_id')
			GROUP BY con.oid`, c.table).Scan(&cols)
		if assert.NoError(t, err, "unique constraint on %s", c.table) {
			assert.Equal(t, c.cols, cols, c.table)
		}
	}
}

func TestPostgresBaseline_ReminderFKCascadesOnDelete(t *testing.T) {
	db, _ := dbtest.StartPostgresBaseline(t)

	var deltype string
	require.NoError(t, db.QueryRow(`
		SELECT con.confdeltype::text FROM pg_constraint con
		JOIN pg_class rel ON rel.oid = con.conrelid
		WHERE con.contype = 'f' AND rel.relname = 'email_reminders'`).Scan(&deltype))
	assert.Equal(t, "c", deltype)
}

func TestPostgresBaseline_DeletingMessageDeletesItsReminders(t *testing.T) {
	db, _ := dbtest.StartPostgresBaseline(t)

	var id int
	require.NoError(t, db.QueryRow(
		`INSERT INTO messages (message, uniqueid) VALUES ('ciphertext', 'u-1') RETURNING messageid`).Scan(&id))
	_, err := db.Exec(`INSERT INTO email_reminders (message_id, email_address) VALUES ($1, 'a@example.com'), ($1, 'b@example.com')`, id)
	require.NoError(t, err)

	_, err = db.Exec(`DELETE FROM messages WHERE messageid = $1`, id)
	require.NoError(t, err)

	var n int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM email_reminders WHERE message_id = $1`, id).Scan(&n))
	assert.Zero(t, n)
}

func TestPostgresBaseline_MessageDefaultsMatchMySQLSemantics(t *testing.T) {
	db, _ := dbtest.StartPostgresBaseline(t)

	var viewCount, maxViewCount int
	var clientEnc bool
	var withinWeek bool
	require.NoError(t, db.QueryRow(`
		INSERT INTO messages (message) VALUES ('x')
		RETURNING view_count, max_view_count, is_client_encrypted,
		          abs(extract(epoch FROM (expires_at - created)) - 7*86400) < 5`).
		Scan(&viewCount, &maxViewCount, &clientEnc, &withinWeek))
	assert.Equal(t, 0, viewCount)
	assert.Equal(t, 5, maxViewCount)
	assert.False(t, clientEnc)
	assert.True(t, withinWeek, "expires_at defaults to 7 days after creation")
}

func TestPostgresBaseline_AllowsExplicitIdentityValuesForPgloader(t *testing.T) {
	db, _ := dbtest.StartPostgresBaseline(t)

	_, err := db.Exec(`INSERT INTO messages (messageid, message) VALUES (4242, 'copied row')`)
	require.NoError(t, err, "identity must be GENERATED BY DEFAULT so bulk loaders can supply ids")
}

func TestPostgresBaseline_DuplicateReminderRecipientRejected(t *testing.T) {
	db, _ := dbtest.StartPostgresBaseline(t)

	var id int
	require.NoError(t, db.QueryRow(`INSERT INTO messages (message) VALUES ('x') RETURNING messageid`).Scan(&id))
	_, err := db.Exec(`INSERT INTO email_reminders (message_id, email_address) VALUES ($1, 'a@example.com')`, id)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO email_reminders (message_id, email_address) VALUES ($1, 'a@example.com')`, id)
	assert.Error(t, err)
}

func TestPostgresBaseline_IndexNamesAreTablePrefixed(t *testing.T) {
	db, _ := dbtest.StartPostgresBaseline(t)

	want := []string{
		"idx_messages_view_count", "idx_messages_max_view_count", "idx_messages_expires_at",
		"idx_messages_reminder_lookup", "idx_email_reminders_email_address",
		"idx_email_reminders_last_reminder_sent", "idx_file_upload_sessions_expires_at",
		"idx_file_upload_sessions_status_expires_at",
	}
	for _, name := range want {
		var exists bool
		require.NoError(t, db.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE schemaname='public' AND indexname=$1)`, name).Scan(&exists))
		assert.True(t, exists, "index %s", name)
	}
}
