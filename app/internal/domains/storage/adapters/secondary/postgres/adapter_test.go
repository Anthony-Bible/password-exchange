package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Unit tests for the Postgres adapter using sqlmock. These pin the SQL
// dialect translation choices the issue calls out ($n placeholders, a single
// argument for the OR lookup, the status guard, DELETE ... RETURNING) and the
// mapping of Postgres error codes to domain errors. Behaviour against a real
// Postgres 17 is pinned by the *_integration_test.go files.

type noopValidator struct{}

func (noopValidator) ValidateEmail(string) error              { return nil }
func (noopValidator) SanitizeEmailForLogging(s string) string { return s }

var errSimulatedDB = errors.New("simulated db error")

func newAdapterForTest(t *testing.T) (*PostgresAdapter, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return &PostgresAdapter{db: db, logger: logtest.NewNoop(), validator: noopValidator{}}, mock
}

func recoverPanic(f func()) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				msg = s
			} else {
				msg = "non-string panic"
			}
		}
	}()
	f()
	return ""
}

// --- construction ---

func TestNewPostgresAdapter_NilLoggerPanicsWithPostgresPrefix(t *testing.T) {
	msg := recoverPanic(func() { NewPostgresAdapter(contracts.DatabaseConfig{}, nil, noopValidator{}) })
	assert.True(t, strings.HasPrefix(msg, "storage/postgres:"), "panic message %q must start with storage/postgres:", msg)
}

func TestNewPostgresAdapter_NilValidatorPanicsWithPostgresPrefix(t *testing.T) {
	msg := recoverPanic(func() { NewPostgresAdapter(contracts.DatabaseConfig{}, logtest.NewNoop(), nil) })
	assert.True(t, strings.HasPrefix(msg, "storage/postgres:"), "panic message %q must start with storage/postgres:", msg)
}

func TestPostgresAdapter_RejectsOperationsAfterClose(t *testing.T) {
	a, mock := newAdapterForTest(t)
	mock.ExpectClose()
	require.NoError(t, a.Close())

	ctx := context.Background()
	calls := map[string]func() error{
		"InsertMessage": func() error {
			return a.InsertMessage(&contracts.Message{UniqueID: "x", Content: "y", MaxViewCount: 1})
		},
		"GetMessage":               func() error { _, e := a.GetMessage("x"); return e },
		"IncrementViewCountAndGet": func() error { _, e := a.IncrementViewCountAndGet("x"); return e },
		"DeleteExpiredMessages":    func() error { return a.DeleteExpiredMessages() },
		"LogReminderSent":          func() error { return a.LogReminderSent(1, "e@x") },
		"CreateUploadSession":      func() error { return a.CreateUploadSession(ctx, contracts.UploadSession{SessionID: "s"}) },
		"GetUploadSession":         func() error { _, e := a.GetUploadSession(ctx, "s"); return e },
		"CompleteUploadSession":    func() error { return a.CompleteUploadSession(ctx, "s") },
		"DeleteExpiredUploadSessions": func() error {
			_, e := a.DeleteExpiredUploadSessions(ctx, time.Now())
			return e
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, call(), domain.ErrRepositoryClosed)
		})
	}
}

// --- error mapping ---

func TestInsertMessage_StringTooLongAndInvalidByteSequenceMapToValidationError(t *testing.T) {
	for _, code := range []string{"22001", "22021", "22P05"} {
		t.Run(code, func(t *testing.T) {
			a, mock := newAdapterForTest(t)
			mock.ExpectExec(`INSERT INTO messages`).WillReturnError(&pgconn.PgError{Code: code})

			err := a.InsertMessage(&contracts.Message{UniqueID: "u", Content: "c", MaxViewCount: 1})

			assert.ErrorIs(t, err, domain.ErrInvalidParameter)
			assert.NotErrorIs(t, err, domain.ErrDatabaseOperation, "a rejected input is not an operational failure")
		})
	}
}

func TestCreateUploadSession_StringTooLongAndInvalidByteSequenceMapToValidationError(t *testing.T) {
	for _, code := range []string{"22001", "22021", "22P05"} {
		t.Run(code, func(t *testing.T) {
			a, mock := newAdapterForTest(t)
			mock.ExpectExec(`INSERT INTO file_upload_sessions`).WillReturnError(&pgconn.PgError{Code: code})

			err := a.CreateUploadSession(context.Background(), contracts.UploadSession{SessionID: "s", FileID: "f"})

			assert.ErrorIs(t, err, domain.ErrInvalidParameter)
		})
	}
}

func TestCreateUploadSession_UniqueViolationMapsToAlreadyExists(t *testing.T) {
	a, mock := newAdapterForTest(t)
	mock.ExpectExec(`INSERT INTO file_upload_sessions`).WillReturnError(&pgconn.PgError{Code: "23505"})

	err := a.CreateUploadSession(context.Background(), contracts.UploadSession{SessionID: "s", FileID: "f"})

	assert.ErrorIs(t, err, domain.ErrUploadSessionAlreadyExists)
}

func TestCreateUploadSession_GenericErrorUsesPostgresPrefix(t *testing.T) {
	a, mock := newAdapterForTest(t)
	mock.ExpectExec(`INSERT INTO file_upload_sessions`).WillReturnError(errSimulatedDB)

	err := a.CreateUploadSession(context.Background(), contracts.UploadSession{SessionID: "s", FileID: "f"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "postgres:")
	assert.NotContains(t, err.Error(), "mysql")
	assert.NotErrorIs(t, err, domain.ErrUploadSessionAlreadyExists)
}

// --- SQL shape ---

func TestGetUploadSession_PassesLookupIDExactlyOnce(t *testing.T) {
	a, mock := newAdapterForTest(t)
	mock.ExpectQuery(`session_id\s*=\s*\$1\s+OR\s+file_id\s*=\s*\$1`).
		WithArgs("some-id").
		WillReturnRows(sqlmock.NewRows([]string{"session_id"}))

	_, err := a.GetUploadSession(context.Background(), "some-id")

	assert.ErrorIs(t, err, domain.ErrUploadSessionNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompleteUploadSession_GuardsAgainstAlreadyCompleteSessions(t *testing.T) {
	a, mock := newAdapterForTest(t)
	mock.ExpectExec(`UPDATE file_upload_sessions SET .*encryption_key\s*=\s*NULL.*WHERE session_id\s*=\s*\$2\s+AND status\s*<>\s*\$1`).
		WithArgs("complete", "s-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, a.CompleteUploadSession(context.Background(), "s-1"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompleteUploadSession_ZeroRowsAffectedIsNotFound(t *testing.T) {
	a, mock := newAdapterForTest(t)
	mock.ExpectExec(`UPDATE file_upload_sessions .*AND status\s*<>`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := a.CompleteUploadSession(context.Background(), "s-1")

	assert.ErrorIs(t, err, domain.ErrUploadSessionNotFound)
}

func TestDeleteExpiredUploadSessions_IsASingleDeleteReturningStatement(t *testing.T) {
	a, mock := newAdapterForTest(t)
	cols := []string{
		"session_id", "file_id", "message_id", "upload_id", "filename", "content_type",
		"total_size", "total_chunks", "status", "encryption_key", "completed_parts", "created_at", "expires_at",
	}
	now := time.Now()
	// No ExpectBegin/ExpectQuery(SELECT): any extra statement fails the test.
	mock.ExpectQuery(`DELETE FROM file_upload_sessions\s+WHERE status\s*<>\s*\$1\s+AND expires_at\s*<\s*\$2\s+RETURNING`).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(
			"s1", "f1", "m", "u", "n.txt", "text/plain", int64(1), int32(1),
			"active", []byte("k"), `[]`, now, now.Add(-time.Minute),
		))

	got, err := a.DeleteExpiredUploadSessions(context.Background(), now)

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "s1", got[0].SessionID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogReminderSent_IsAnAtomicUpsertOnMessageAndEmail(t *testing.T) {
	a, mock := newAdapterForTest(t)
	mock.ExpectExec(`INSERT INTO email_reminders .*VALUES\s*\(\$1,\s*\$2.*ON CONFLICT\s*\(message_id,\s*email_address\)\s*DO UPDATE`).
		WithArgs(7, "r@example.com").
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, a.LogReminderSent(7, "r@example.com"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUnviewedMessagesForReminders_BindsParametersInDocumentedOrder(t *testing.T) {
	a, mock := newAdapterForTest(t)
	// olderThanHours=24, maxReminders=3, reminderIntervalHours=12
	mock.ExpectQuery(`make_interval\(hours\s*=>\s*\$1::int\).*reminder_count\s*<\s*\$2.*make_interval\(hours\s*=>\s*\$3::int\)`).
		WithArgs(24, 3, 12).
		WillReturnRows(sqlmock.NewRows([]string{"messageid", "uniqueid", "other_email", "created", "days_old"}))

	msgs, err := a.GetUnviewedMessagesForReminders(24, 3, 12)

	require.NoError(t, err)
	assert.Empty(t, msgs)
	assert.NoError(t, mock.ExpectationsWereMet())
}
