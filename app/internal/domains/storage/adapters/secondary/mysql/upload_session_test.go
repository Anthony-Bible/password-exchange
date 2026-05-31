package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
	"github.com/DATA-DOG/go-sqlmock"
	drivermysql "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errSimulatedDB is a package-level sentinel for database error injection in tests.
var errSimulatedDB = errors.New("simulated db error")

// newUploadAdapterForTest creates a *MySQLAdapter backed by a go-sqlmock *sql.DB.
// The caller receives both the adapter and the mock controller for expectation setup.
func newUploadAdapterForTest(t *testing.T) (*MySQLAdapter, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	a := &MySQLAdapter{
		db:        db,
		logger:    logtest.NewNoop(),
		validator: noopValidator{},
	}
	return a, mock
}

// uploadSessionColumns is the canonical column list that all SELECT helpers return.
var uploadSessionColumns = []string{
	"session_id", "file_id", "message_id", "upload_id",
	"filename", "content_type", "total_size", "total_chunks",
	"status", "encryption_key", "completed_parts", "created_at", "expires_at",
}

func makeSessionRow(sessionID, fileID string, now, exp time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(uploadSessionColumns).AddRow(
		sessionID, fileID, "msg-1", "upload-1",
		"test.txt", "text/plain", int64(1024), int32(2),
		"active", []byte("key"), `[]`, now, exp,
	)
}

// --- CreateUploadSession ---

func TestCreateUploadSession_InsertsRow(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	now := time.Now().UTC().Truncate(time.Second)
	sess := contracts.UploadSession{
		SessionID: "sess-1", FileID: "file-1",
		MessageID: "msg-1", UploadID: "upload-1",
		Filename: "test.txt", ContentType: "text/plain",
		TotalSize: 1024, TotalChunks: 2,
		Status: "active", EncryptionKey: []byte("secret"),
		CompletedParts: []contracts.UploadSessionPart{},
		CreatedAt:      now, ExpiresAt: now.Add(time.Hour),
	}

	mock.ExpectExec(`INSERT INTO file_upload_sessions`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := a.CreateUploadSession(context.Background(), sess)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUploadSession_DBError_ReturnsError(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	mock.ExpectExec(`INSERT INTO file_upload_sessions`).
		WillReturnError(errSimulatedDB)

	err := a.CreateUploadSession(context.Background(), contracts.UploadSession{SessionID: "s1", FileID: "f1"})
	require.Error(t, err)
}

func TestCreateUploadSession_DuplicateKey_ReturnsErrUploadSessionAlreadyExists(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	// Simulating a duplicate entry error (MySQL error code 1062)
	mysqlErr := &drivermysql.MySQLError{
		Number:  1062,
		Message: "Duplicate entry 's1' for key 'PRIMARY'",
	}

	mock.ExpectExec(`INSERT INTO file_upload_sessions`).
		WillReturnError(mysqlErr)

	err := a.CreateUploadSession(context.Background(), contracts.UploadSession{SessionID: "s1", FileID: "f1"})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUploadSessionAlreadyExists)
}

// --- GetUploadSession ---

func TestGetUploadSession_ReturnsSession(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	now := time.Now().UTC().Truncate(time.Second)
	exp := now.Add(time.Hour)

	mock.ExpectQuery(`SELECT .+ FROM file_upload_sessions WHERE`).
		WithArgs("sess-1", "sess-1").
		WillReturnRows(makeSessionRow("sess-1", "file-1", now, exp))

	got, err := a.GetUploadSession(context.Background(), "sess-1")
	require.NoError(t, err)
	assert.Equal(t, "sess-1", got.SessionID)
	assert.Equal(t, "file-1", got.FileID)
	assert.Equal(t, []byte("key"), got.EncryptionKey)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUploadSession_NotFound_ReturnsErrUploadSessionNotFound(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	mock.ExpectQuery(`SELECT .+ FROM file_upload_sessions WHERE`).
		WithArgs("missing", "missing").
		WillReturnRows(sqlmock.NewRows(uploadSessionColumns)) // empty result set

	_, err := a.GetUploadSession(context.Background(), "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUploadSessionNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- AddCompletedPart ---

func TestAddCompletedPart_MergesPartAndUpdatesRow(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	now := time.Now().UTC().Truncate(time.Second)
	exp := now.Add(time.Hour)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .+ FROM file_upload_sessions WHERE .+ FOR UPDATE`).
		WithArgs("sess-1", "sess-1").
		WillReturnRows(makeSessionRow("sess-1", "file-1", now, exp))

	mock.ExpectExec(`UPDATE file_upload_sessions SET completed_parts`).
		WithArgs(`[{"PartNumber":1,"ETag":"etag-abc"}]`, "sess-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := a.AddCompletedPart(context.Background(), "sess-1", contracts.UploadSessionPart{
		PartNumber: 1, ETag: "etag-abc",
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddCompletedPart_RollbackOnSelectError(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .+ FROM file_upload_sessions WHERE .+ FOR UPDATE`).
		WithArgs("sess-1", "sess-1").
		WillReturnError(errSimulatedDB)
	mock.ExpectRollback()

	err := a.AddCompletedPart(context.Background(), "sess-1", contracts.UploadSessionPart{
		PartNumber: 1, ETag: "etag-abc",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "simulated db error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddCompletedPart_RollbackOnUpdateError(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	now := time.Now().UTC().Truncate(time.Second)
	exp := now.Add(time.Hour)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .+ FROM file_upload_sessions WHERE .+ FOR UPDATE`).
		WithArgs("sess-1", "sess-1").
		WillReturnRows(makeSessionRow("sess-1", "file-1", now, exp))

	mock.ExpectExec(`UPDATE file_upload_sessions SET completed_parts`).
		WithArgs(`[{"PartNumber":1,"ETag":"etag-abc"}]`, "sess-1").
		WillReturnError(errSimulatedDB)
	mock.ExpectRollback()

	err := a.AddCompletedPart(context.Background(), "sess-1", contracts.UploadSessionPart{
		PartNumber: 1, ETag: "etag-abc",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "simulated db error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- CompleteUploadSession ---

func TestCompleteUploadSession_ClearsEncryptionKey(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	mock.ExpectExec(`UPDATE file_upload_sessions SET status = .+, encryption_key = NULL WHERE session_id = .+`).
		WithArgs("complete", "sess-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := a.CompleteUploadSession(context.Background(), "sess-1")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompleteUploadSession_NotFound_ReturnsErrUploadSessionNotFound(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	mock.ExpectExec(`UPDATE file_upload_sessions SET status = .+, encryption_key = NULL WHERE session_id = .+`).
		WithArgs("complete", "not-there").
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected → not found

	err := a.CompleteUploadSession(context.Background(), "not-there")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUploadSessionNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- DeleteUploadSession ---

func TestDeleteUploadSession_DeletesRow(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	mock.ExpectExec(`DELETE FROM file_upload_sessions WHERE session_id = .+`).
		WithArgs("sess-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := a.DeleteUploadSession(context.Background(), "sess-1")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUploadSession_NonExistent_IsNoOp(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	mock.ExpectExec(`DELETE FROM file_upload_sessions WHERE session_id = .+`).
		WithArgs("gone").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := a.DeleteUploadSession(context.Background(), "gone")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- DeleteExpiredUploadSessions ---

func TestDeleteExpiredUploadSessions_ReturnsAndDeletesExpired(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	asOf := time.Now().UTC()
	now := asOf.Add(-2 * time.Hour).Truncate(time.Second)
	exp := asOf.Add(-1 * time.Hour).Truncate(time.Second)

	mock.ExpectQuery(`SELECT .+ FROM file_upload_sessions WHERE status != .+ AND expires_at < .+`).
		WillReturnRows(makeSessionRow("sess-expired", "file-exp", now, exp))

	mock.ExpectExec(`DELETE FROM file_upload_sessions WHERE status != .+ AND expires_at < .+`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	removed, err := a.DeleteExpiredUploadSessions(context.Background(), asOf)
	require.NoError(t, err)
	require.Len(t, removed, 1)
	assert.Equal(t, "sess-expired", removed[0].SessionID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteExpiredUploadSessions_NoneExpired_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	a, mock := newUploadAdapterForTest(t)

	mock.ExpectQuery(`SELECT .+ FROM file_upload_sessions WHERE status != .+ AND expires_at < .+`).
		WillReturnRows(sqlmock.NewRows(uploadSessionColumns))

	mock.ExpectExec(`DELETE FROM file_upload_sessions WHERE status != .+ AND expires_at < .+`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	removed, err := a.DeleteExpiredUploadSessions(context.Background(), time.Now())
	require.NoError(t, err)
	assert.Empty(t, removed)
	assert.NoError(t, mock.ExpectationsWereMet())
}
