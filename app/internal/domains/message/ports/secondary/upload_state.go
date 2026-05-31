package secondary

import (
	"context"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
)

// UploadStatePort defines the secondary port used to persist multipart upload
// session metadata between chunk requests.
type UploadStatePort interface {
	// CreateSession stores a newly initiated upload session.
	CreateSession(ctx context.Context, session contracts.FileUploadSession) error

	// GetSessionByID retrieves a previously stored upload session by its SessionID.
	GetSessionByID(ctx context.Context, sessionID string) (*contracts.FileUploadSession, error)

	// GetSessionByFileID retrieves a previously stored upload session by its FileID.
	GetSessionByFileID(ctx context.Context, fileID string) (*contracts.FileUploadSession, error)

	// AddCompletedPart records a successfully uploaded part for the session and
	// returns the full post-add completed-parts list atomically, so callers can
	// check completion without a separate GetSession round-trip.
	AddCompletedPart(ctx context.Context, sessionID string, part contracts.CompletedPart) ([]contracts.CompletedPart, error)

	// CompleteSession marks the upload session as fully assembled.
	CompleteSession(ctx context.Context, sessionID string) error

	// DeleteSession removes the upload session from persistent state.
	DeleteSession(ctx context.Context, sessionID string) error

	// DeleteExpiredSessions removes upload sessions that are not yet complete and
	// whose ExpiresAt is before asOf, returning the removed sessions so the caller
	// can release their object-storage multipart uploads. Completed sessions are
	// retained so finalized files remain downloadable.
	DeleteExpiredSessions(ctx context.Context, asOf time.Time) ([]contracts.FileUploadSession, error)
}
