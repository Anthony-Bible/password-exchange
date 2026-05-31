package memory

import (
	"context"
	"sync"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
)

var _ secondary.UploadStatePort = (*MemoryUploadStateAdapter)(nil)

// MemoryUploadStateAdapter stores upload sessions in memory for single-replica deployments.
// It indexes each session under both its SessionID and its FileID so that
// callers may look up a session by either key.
type MemoryUploadStateAdapter struct {
	mu       sync.RWMutex
	sessions map[string]contracts.FileUploadSession
}

// NewMemoryUploadStateAdapter creates a new in-memory upload state adapter.
func NewMemoryUploadStateAdapter() *MemoryUploadStateAdapter {
	return &MemoryUploadStateAdapter{
		sessions: make(map[string]contracts.FileUploadSession),
	}
}

// CreateSession stores a new upload session in memory.
func (a *MemoryUploadStateAdapter) CreateSession(_ context.Context, session contracts.FileUploadSession) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.storeSessionLocked(session)
	return nil
}

// GetSessionByID retrieves an upload session from memory by its SessionID.
func (a *MemoryUploadStateAdapter) GetSessionByID(_ context.Context, sessionID string) (*contracts.FileUploadSession, error) {
	a.mu.RLock()
	session, ok := a.sessions[sessionID]
	a.mu.RUnlock()
	if !ok {
		return nil, domain.ErrUploadSessionNotFound
	}
	cloned := cloneSession(session)
	return &cloned, nil
}

// GetSessionByFileID retrieves an upload session from memory by its FileID.
func (a *MemoryUploadStateAdapter) GetSessionByFileID(_ context.Context, fileID string) (*contracts.FileUploadSession, error) {
	a.mu.RLock()
	session, ok := a.sessions[fileID]
	a.mu.RUnlock()
	if !ok {
		return nil, domain.ErrUploadSessionNotFound
	}
	cloned := cloneSession(session)
	return &cloned, nil
}

// AddCompletedPart records a completed multipart upload part in memory and
// returns a copy of the full completed-parts list after the update, while still
// holding the write lock, so the caller sees a consistent snapshot.
func (a *MemoryUploadStateAdapter) AddCompletedPart(_ context.Context, sessionID string, part contracts.CompletedPart) ([]contracts.CompletedPart, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	session, ok := a.sessions[sessionID]
	if !ok {
		return nil, domain.ErrUploadSessionNotFound
	}

	// Replace an existing entry with the same PartNumber to avoid duplicates
	// that would break multipart completion when a chunk is retried.
	replaced := false
	for i, existing := range session.CompletedParts {
		if existing.PartNumber == part.PartNumber {
			session.CompletedParts[i] = part
			replaced = true
			break
		}
	}
	if !replaced {
		session.CompletedParts = append(session.CompletedParts, part)
	}
	a.storeSessionLocked(session)
	return append([]contracts.CompletedPart(nil), session.CompletedParts...), nil
}

// CompleteSession marks an in-memory upload session as complete and clears the
// stored encryption key. The key is only needed during active chunk uploads;
// once the multipart upload is finalized, the client holds the key via the
// share-URL fragment and the server no longer needs it in memory.
func (a *MemoryUploadStateAdapter) CompleteSession(_ context.Context, sessionID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	session, ok := a.sessions[sessionID]
	if !ok {
		return domain.ErrUploadSessionNotFound
	}

	session.Status = string(domain.SessionStatusComplete)
	session.EncryptionKey = nil
	a.storeSessionLocked(session)
	return nil
}

// DeleteSession removes an upload session from memory.
func (a *MemoryUploadStateAdapter) DeleteSession(_ context.Context, sessionID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	session, ok := a.sessions[sessionID]
	if !ok {
		// Best-effort deletion even when not found by the primary key.
		delete(a.sessions, sessionID)
		return nil
	}

	delete(a.sessions, session.SessionID)
	if session.FileID != "" {
		delete(a.sessions, session.FileID)
	}
	return nil
}

// DeleteExpiredSessions removes incomplete sessions whose ExpiresAt is before
// asOf, returning a clone of each removed session so the caller can release the
// associated multipart upload. Completed sessions are retained so finalized
// files remain downloadable.
func (a *MemoryUploadStateAdapter) DeleteExpiredSessions(_ context.Context, asOf time.Time) ([]contracts.FileUploadSession, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// The map double-indexes each session under both SessionID and FileID, so
	// dedupe by SessionID to avoid evaluating (and returning) a session twice.
	seen := make(map[string]bool)
	var removed []contracts.FileUploadSession
	for _, session := range a.sessions {
		if seen[session.SessionID] {
			continue
		}
		seen[session.SessionID] = true

		if session.Status == string(domain.SessionStatusComplete) {
			continue
		}
		if !session.ExpiresAt.Before(asOf) {
			continue
		}

		delete(a.sessions, session.SessionID)
		if session.FileID != "" {
			delete(a.sessions, session.FileID)
		}
		removed = append(removed, cloneSession(session))
	}

	return removed, nil
}

// storeSessionLocked writes a defensive copy of the session under both its
// SessionID and FileID keys. Callers MUST hold a.mu (write lock).
func (a *MemoryUploadStateAdapter) storeSessionLocked(session contracts.FileUploadSession) {
	cloned := cloneSession(session)
	a.sessions[session.SessionID] = cloned
	if session.FileID != "" {
		a.sessions[session.FileID] = cloned
	}
}

func cloneSession(session contracts.FileUploadSession) contracts.FileUploadSession {
	session.CompletedParts = append([]contracts.CompletedPart(nil), session.CompletedParts...)
	session.EncryptionKey = append([]byte(nil), session.EncryptionKey...)
	return session
}
