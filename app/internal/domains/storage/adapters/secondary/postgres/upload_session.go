package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/jackc/pgx/v5/pgconn"
)

// sessionStatusComplete is the status value written on CompleteUploadSession. It
// must match the value stored by the web service and checked in cleanup queries.
const sessionStatusComplete = "complete"

// uploadSessionColumns is the canonical column order for all SELECT queries
// against file_upload_sessions. Keep in sync with scanUploadSession /
// scanUploadSessionRow.
const uploadSessionSelectCols = `session_id, file_id, message_id, upload_id,
		filename, content_type, total_size, total_chunks,
		status, encryption_key, completed_parts, created_at, expires_at`

// CreateUploadSession inserts a new file upload session row.
//
// EncryptionKey nil vs empty is preserved: a nil EncryptionKey is stored as
// NULL, and an empty (non-nil) EncryptionKey is stored as an empty bytea.
func (m *PostgresAdapter) CreateUploadSession(ctx context.Context, session contracts.UploadSession) error {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return openErr
	}
	defer release()

	partsJSON, err := marshalUploadSessionParts(session.CompletedParts)
	if err != nil {
		return fmt.Errorf("postgres: marshalling completed_parts: %w", err)
	}

	const q = `INSERT INTO file_upload_sessions
		(session_id, file_id, message_id, upload_id, filename, content_type,
		 total_size, total_chunks, status, encryption_key, completed_parts,
		 created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err = db.ExecContext(ctx, q,
		session.SessionID,
		session.FileID,
		session.MessageID,
		session.UploadID,
		session.Filename,
		session.ContentType,
		session.TotalSize,
		session.TotalChunks,
		session.Status,
		session.EncryptionKey,
		partsJSON,
		session.CreatedAt,
		session.ExpiresAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUploadSessionAlreadyExists
		}
		if isInvalidInputError(err) {
			return fmt.Errorf("%w: %v", domain.ErrInvalidParameter, err)
		}
		m.logger.Error().Err(err).Str("sessionID", session.SessionID).Msg("Failed to create upload session")
		return fmt.Errorf("postgres: inserting upload session: %w", err)
	}

	m.logger.Info().Str("sessionID", session.SessionID).Str("fileID", session.FileID).Msg("Upload session created")
	return nil
}

// GetUploadSession retrieves a session by either its SessionID or FileID.
// The id argument is tested against both columns so both the chunk-upload
// path (session_id) and the download path (file_id) can use this call.
//
// EncryptionKey round-trips exactly: a NULL column is returned as a nil
// EncryptionKey, and an empty bytea is returned as a non-nil empty slice.
func (m *PostgresAdapter) GetUploadSession(ctx context.Context, id string) (*contracts.UploadSession, error) {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return nil, openErr
	}
	defer release()

	q := fmt.Sprintf(`SELECT %s FROM file_upload_sessions
		WHERE session_id = $1 OR file_id = $1
		LIMIT 1`, uploadSessionSelectCols)

	row := db.QueryRowContext(ctx, q, id)
	session, err := scanUploadSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUploadSessionNotFound
	}
	if err != nil {
		m.logger.Error().Err(err).Str("id", id).Msg("Failed to get upload session")
		return nil, fmt.Errorf("postgres: querying upload session: %w", err)
	}

	return session, nil
}

// AddCompletedPart reads the current completed_parts list, merges the new
// part (replacing a duplicate PartNumber if present), and writes the updated
// JSON back to the row. It uses a database transaction with SELECT ... FOR UPDATE
// to ensure atomic updates under concurrency.
func (m *PostgresAdapter) AddCompletedPart(ctx context.Context, sessionID string, part contracts.UploadSessionPart) error {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return openErr
	}
	defer release()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("postgres: beginning transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	q := fmt.Sprintf(`SELECT %s FROM file_upload_sessions
		WHERE session_id = $1 OR file_id = $1
		LIMIT 1 FOR UPDATE`, uploadSessionSelectCols)

	row := tx.QueryRowContext(ctx, q, sessionID)
	session, err := scanUploadSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrUploadSessionNotFound
	}
	if err != nil {
		m.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to get upload session for update")
		return fmt.Errorf("postgres: querying upload session: %w", err)
	}

	merged := mergeUploadSessionPart(session.CompletedParts, part)
	partsJSON, err := marshalUploadSessionParts(merged)
	if err != nil {
		return fmt.Errorf("postgres: marshalling completed_parts: %w", err)
	}

	const updateQ = `UPDATE file_upload_sessions SET completed_parts = $1 WHERE session_id = $2`
	_, err = tx.ExecContext(ctx, updateQ, partsJSON, session.SessionID)
	if err != nil {
		m.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to add completed part")
		return fmt.Errorf("postgres: updating completed_parts: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("postgres: committing transaction: %w", err)
	}

	return nil
}

// CompleteUploadSession atomically marks the session status as complete and
// clears the stored encryption key. The key is only needed during active chunk
// uploads; once finalized the client holds the key in the share-URL fragment
// and the server no longer needs it at rest. Completing an already-complete
// session returns ErrUploadSessionNotFound by contract (explicit status guard).
func (m *PostgresAdapter) CompleteUploadSession(ctx context.Context, sessionID string) error {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return openErr
	}
	defer release()

	const q = `UPDATE file_upload_sessions SET status = $1, encryption_key = NULL
		WHERE session_id = $2 AND status <> $1`
	result, err := db.ExecContext(ctx, q, sessionStatusComplete, sessionID)
	if err != nil {
		m.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to complete upload session")
		return fmt.Errorf("postgres: completing upload session: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: checking rows affected after complete: %w", err)
	}
	if n == 0 {
		return domain.ErrUploadSessionNotFound
	}

	m.logger.Info().Str("sessionID", sessionID).Msg("Upload session completed, encryption key cleared")
	return nil
}

// DeleteUploadSession removes a single upload session row. Deletion of a
// non-existent session is treated as a no-op to match the best-effort contract
// of the in-memory adapter.
func (m *PostgresAdapter) DeleteUploadSession(ctx context.Context, sessionID string) error {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return openErr
	}
	defer release()

	const q = `DELETE FROM file_upload_sessions WHERE session_id = $1`
	_, err := db.ExecContext(ctx, q, sessionID)
	if err != nil {
		m.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to delete upload session")
		return fmt.Errorf("postgres: deleting upload session: %w", err)
	}

	return nil
}

// DeleteExpiredUploadSessions selects all incomplete sessions whose expires_at
// is before asOf, returns a copy for caller-side multipart-upload cleanup, then
// deletes them in a single statement. Completed sessions are kept so finalized
// files remain downloadable.
func (m *PostgresAdapter) DeleteExpiredUploadSessions(ctx context.Context, asOf time.Time) ([]contracts.UploadSession, error) {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return nil, openErr
	}
	defer release()

	q := fmt.Sprintf(`DELETE FROM file_upload_sessions
		WHERE status <> $1 AND expires_at < $2
		RETURNING %s`, uploadSessionSelectCols)

	rows, err := db.QueryContext(ctx, q, sessionStatusComplete, asOf)
	if err != nil {
		return nil, fmt.Errorf("postgres: deleting expired upload sessions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var removed []contracts.UploadSession
	for rows.Next() {
		session, err := scanUploadSessionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: scanning expired upload session: %w", err)
		}
		removed = append(removed, *session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterating expired upload sessions: %w", err)
	}

	m.logger.Info().Int("count", len(removed)).Msg("Deleted expired upload sessions")
	return removed, nil
}

// scanUploadSession scans a single *sql.Row into a contracts.UploadSession.
// encryption_key is NULLable (cleared on completion) so it is scanned into a
// []byte which becomes nil for NULL rows rather than returning a scan error.
func scanUploadSession(row *sql.Row) (*contracts.UploadSession, error) {
	var (
		s         contracts.UploadSession
		keyPtr    []byte
		partsJSON string
		createdAt time.Time
		expiresAt time.Time
	)
	err := row.Scan(
		&s.SessionID,
		&s.FileID,
		&s.MessageID,
		&s.UploadID,
		&s.Filename,
		&s.ContentType,
		&s.TotalSize,
		&s.TotalChunks,
		&s.Status,
		&keyPtr,
		&partsJSON,
		&createdAt,
		&expiresAt,
	)
	if err != nil {
		return nil, err
	}
	s.EncryptionKey = keyPtr // nil when NULL (completed session)
	s.CreatedAt = createdAt
	s.ExpiresAt = expiresAt
	if err := json.Unmarshal([]byte(partsJSON), &s.CompletedParts); err != nil {
		return nil, fmt.Errorf("postgres: unmarshalling completed_parts: %w", err)
	}
	return &s, nil
}

// scanUploadSessionRow scans a *sql.Rows cursor row into a contracts.UploadSession.
func scanUploadSessionRow(rows *sql.Rows) (*contracts.UploadSession, error) {
	var (
		s         contracts.UploadSession
		keyPtr    []byte
		partsJSON string
		createdAt time.Time
		expiresAt time.Time
	)
	err := rows.Scan(
		&s.SessionID,
		&s.FileID,
		&s.MessageID,
		&s.UploadID,
		&s.Filename,
		&s.ContentType,
		&s.TotalSize,
		&s.TotalChunks,
		&s.Status,
		&keyPtr,
		&partsJSON,
		&createdAt,
		&expiresAt,
	)
	if err != nil {
		return nil, err
	}
	s.EncryptionKey = keyPtr // nil when NULL (completed session)
	s.CreatedAt = createdAt
	s.ExpiresAt = expiresAt
	if err := json.Unmarshal([]byte(partsJSON), &s.CompletedParts); err != nil {
		return nil, fmt.Errorf("postgres: unmarshalling completed_parts: %w", err)
	}
	return &s, nil
}

// marshalUploadSessionParts encodes a parts slice to JSON. A nil slice is
// serialised as "[]" so the column is never NULL.
func marshalUploadSessionParts(parts []contracts.UploadSessionPart) (string, error) {
	if parts == nil {
		parts = []contracts.UploadSessionPart{}
	}
	b, err := json.Marshal(parts)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// mergeUploadSessionPart replaces the existing entry with the same PartNumber
// or appends the new part, keeping the slice free of duplicates.
func mergeUploadSessionPart(existing []contracts.UploadSessionPart, part contracts.UploadSessionPart) []contracts.UploadSessionPart {
	result := make([]contracts.UploadSessionPart, 0, len(existing)+1)
	replaced := false
	for _, p := range existing {
		if p.PartNumber == part.PartNumber {
			result = append(result, part)
			replaced = true
		} else {
			result = append(result, p)
		}
	}
	if !replaced {
		result = append(result, part)
	}
	return result
}
