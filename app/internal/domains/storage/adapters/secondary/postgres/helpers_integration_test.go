//go:build integration

package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/integration/dbtest"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// These tests need a Postgres 17 container. Under podman, see the header of
// internal/integration/dbtest/postgres.go (DOCKER_HOST + TESTCONTAINERS_RYUK_DISABLED).

// startAdapter boots Postgres with the embedded baseline migration and
// returns a raw *sql.DB (for seeding and assertions) plus the adapter under
// test, built only through its public constructor.
func startAdapter(t *testing.T) (*sql.DB, secondary.MessageRepository) {
	t.Helper()
	db, cfg := dbtest.StartPostgresBaseline(t)
	return db, newAdapterFor(t, cfg)
}

func newAdapterFor(t *testing.T, cfg contracts.DatabaseConfig) secondary.MessageRepository {
	t.Helper()
	repo := NewPostgresAdapter(cfg, logtest.NewNoop(), noopValidator{})
	t.Cleanup(func() { _ = repo.Close() })
	return repo
}

// resetTables empties every application table so property iterations start clean.
// testingTB is a minimal interface because *rapid.T does not implement testing.TB on newer Go versions.
type testingTB interface {
	require.TestingT
	Helper()
}

func resetTables(tb testingTB, db *sql.DB) {
	tb.Helper()
	_, err := db.Exec(`TRUNCATE messages, email_reminders, file_upload_sessions RESTART IDENTITY CASCADE`)
	require.NoError(tb, err)
}

func rowCount(tb testingTB, db *sql.DB, table string) int {
	tb.Helper()
	var n int
	require.NoError(tb, db.QueryRow(`SELECT count(*) FROM `+table).Scan(&n))
	return n
}

// xidGen generates strings shaped like rs/xid IDs (20 chars, base32hex lowercase).
func xidGen() *rapid.Generator[string] {
	return rapid.StringMatching(`[0-9a-v]{20}`)
}

// textGen generates valid UTF-8 text without NUL, between minRunes and maxRunes
// runes, including 2-, 3- and 4-byte runes.
func textGen(minRunes, maxRunes int) *rapid.Generator[string] {
	return rapid.StringOfN(rapid.Rune().Filter(func(r rune) bool { return r != 0 }), minRunes, maxRunes, -1)
}

func validSession(sessionID, fileID string) contracts.UploadSession {
	now := time.Now()
	return contracts.UploadSession{
		SessionID: sessionID, FileID: fileID, MessageID: "msg", UploadID: "upload",
		Filename: "f.txt", ContentType: "text/plain", TotalSize: 10, TotalChunks: 2,
		Status: "active", EncryptionKey: []byte("key"),
		CompletedParts: []contracts.UploadSessionPart{},
		CreatedAt:      now, ExpiresAt: now.Add(time.Hour),
	}
}

func validMessage(uniqueID string) *contracts.Message {
	return &contracts.Message{UniqueID: uniqueID, Content: "ciphertext", MaxViewCount: 5}
}
