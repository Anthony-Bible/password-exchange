package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	drivermysql "github.com/go-sql-driver/mysql"
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
func (m *MySQLAdapter) CreateUploadSession(ctx context.Context, session contracts.UploadSession) error {
	if err := m.requireOpen(); err != nil {
		return err
	}

	partsJSON, err := marshalUploadSessionParts(session.CompletedParts)
	if err != nil {
		return fmt.Errorf("mysql: marshalling completed_parts: %w", err)
	}

	const q = `INSERT INTO file_upload_sessions
		(session_id, file_id, message_id, upload_id, filename, content_type,
		 total_size, total_chunks, status, encryption_key, completed_parts,
		 created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = m.db.ExecContext(ctx, q,
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
		var mysqlErr *drivermysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return domain.ErrUploadSessionAlreadyExists
		}
		m.logger.Error().Err(err).Str("sessionID", session.SessionID).Msg("Failed to create upload session")
		return fmt.Errorf("mysql: inserting upload session: %w", err)
	}

	m.logger.Info().Str("sessionID", session.SessionID).Str("fileID", session.FileID).Msg("Upload session created")
	return nil
}

// GetUploadSession retrieves a session by either its SessionID or FileID.
// The id argument is tested against both columns so both the chunk-upload
// path (session_id) and the download path (file_id) can use this call.
func (m *MySQLAdapter) GetUploadSession(ctx context.Context, id string) (*contracts.UploadSession, error) {
	if err := m.requireOpen(); err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`SELECT %s FROM file_upload_sessions
		WHERE session_id = ? OR file_id = ?
		LIMIT 1`, uploadSessionSelectCols)

	row := m.db.QueryRowContext(ctx, q, id, id)
	session, err := scanUploadSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUploadSessionNotFound
	}
	if err != nil {
		m.logger.Error().Err(err).Str("id", id).Msg("Failed to get upload session")
		return nil, fmt.Errorf("mysql: querying upload session: %w", err)
	}

	return session, nil
}

// AddCompletedPart reads the current completed_parts list, merges the new
// part (replacing a duplicate PartNumber if present), and writes the updated
// JSON back to the row. It uses a database transaction with SELECT ... FOR UPDATE
// to ensure atomic updates under concurrency.
func (m *MySQLAdapter) AddCompletedPart(ctx context.Context, sessionID string, part contracts.UploadSessionPart) error {
	if err := m.requireOpen(); err != nil {
		return err
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mysql: beginning transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	q := fmt.Sprintf(`SELECT %s FROM file_upload_sessions
		WHERE session_id = ? OR file_id = ?
		LIMIT 1 FOR UPDATE`, uploadSessionSelectCols)

	row := tx.QueryRowContext(ctx, q, sessionID, sessionID)
	session, err := scanUploadSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrUploadSessionNotFound
	}
	if err != nil {
		m.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to get upload session for update")
		return fmt.Errorf("mysql: querying upload session: %w", err)
	}

	merged := mergeUploadSessionPart(session.CompletedParts, part)
	partsJSON, err := marshalUploadSessionParts(merged)
	if err != nil {
		return fmt.Errorf("mysql: marshalling completed_parts: %w", err)
	}

	const updateQ = `UPDATE file_upload_sessions SET completed_parts = ? WHERE session_id = ?`
	_, err = tx.ExecContext(ctx, updateQ, partsJSON, session.SessionID)
	if err != nil {
		m.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to add completed part")
		return fmt.Errorf("mysql: updating completed_parts: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("mysql: committing transaction: %w", err)
	}

	return nil
}

// CompleteUploadSession atomically marks the session status as complete and
// clears the stored encryption key. The key is only needed during active chunk
// uploads; once finalized the client holds the key in the share-URL fragment
// and the server no longer needs it at rest.
func (m *MySQLAdapter) CompleteUploadSession(ctx context.Context, sessionID string) error {
	if err := m.requireOpen(); err != nil {
		return err
	}

	const q = `UPDATE file_upload_sessions SET status = ?, encryption_key = NULL WHERE session_id = ?`
	result, err := m.db.ExecContext(ctx, q, sessionStatusComplete, sessionID)
	if err != nil {
		m.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to complete upload session")
		return fmt.Errorf("mysql: completing upload session: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql: checking rows affected after complete: %w", err)
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
func (m *MySQLAdapter) DeleteUploadSession(ctx context.Context, sessionID string) error {
	if err := m.requireOpen(); err != nil {
		return err
	}

	const q = `DELETE FROM file_upload_sessions WHERE session_id = ?`
	_, err := m.db.ExecContext(ctx, q, sessionID)
	if err != nil {
		m.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to delete upload session")
		return fmt.Errorf("mysql: deleting upload session: %w", err)
	}

	return nil
}

// DeleteExpiredUploadSessions selects all incomplete sessions whose expires_at
// is before asOf, returns a copy for caller-side multipart-upload cleanup, then
// deletes them in a single statement. Completed sessions are kept so finalized
// files remain downloadable.
func (m *MySQLAdapter) DeleteExpiredUploadSessions(ctx context.Context, asOf time.Time) ([]contracts.UploadSession, error) {
	if err := m.requireOpen(); err != nil {
		return nil, err
	}

	selectQ := fmt.Sprintf(`SELECT %s FROM file_upload_sessions
		WHERE status != ? AND expires_at < ?`, uploadSessionSelectCols)

	rows, err := m.db.QueryContext(ctx, selectQ, sessionStatusComplete, asOf)
	if err != nil {
		return nil, fmt.Errorf("mysql: selecting expired upload sessions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var removed []contracts.UploadSession
	for rows.Next() {
		session, err := scanUploadSessionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: scanning expired upload session: %w", err)
		}
		removed = append(removed, *session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mysql: iterating expired upload sessions: %w", err)
	}

	const deleteQ = `DELETE FROM file_upload_sessions WHERE status != ? AND expires_at < ?`
	if _, err := m.db.ExecContext(ctx, deleteQ, sessionStatusComplete, asOf); err != nil {
		return nil, fmt.Errorf("mysql: deleting expired upload sessions: %w", err)
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
		return nil, fmt.Errorf("mysql: unmarshalling completed_parts: %w", err)
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
		return nil, fmt.Errorf("mysql: unmarshalling completed_parts: %w", err)
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
