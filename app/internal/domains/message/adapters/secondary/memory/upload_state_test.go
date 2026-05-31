package memory

import (
	"context"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryUploadStateAdapter_StoresSessionBySessionIDAndFileID(t *testing.T) {
	adapter := NewMemoryUploadStateAdapter()
	session := contracts.FileUploadSession{
		SessionID: "session-123",
		FileID:    "file-456",
		Status:    string(domain.SessionStatusActive),
	}

	require.NoError(t, adapter.CreateSession(context.Background(), session))

	bySessionID, err := adapter.GetSessionByID(context.Background(), "session-123")
	require.NoError(t, err)
	assert.Equal(t, "file-456", bySessionID.FileID)

	byFileID, err := adapter.GetSessionByFileID(context.Background(), "file-456")
	require.NoError(t, err)
	assert.Equal(t, "session-123", byFileID.SessionID)
}

func TestMemoryUploadStateAdapter_DeleteSessionRemovesFileLookup(t *testing.T) {
	adapter := NewMemoryUploadStateAdapter()
	session := contracts.FileUploadSession{
		SessionID: "session-123",
		FileID:    "file-456",
		Status:    string(domain.SessionStatusActive),
	}

	require.NoError(t, adapter.CreateSession(context.Background(), session))
	require.NoError(t, adapter.DeleteSession(context.Background(), "session-123"))

	_, err := adapter.GetSessionByID(context.Background(), "session-123")
	assert.ErrorIs(t, err, domain.ErrUploadSessionNotFound)

	_, err = adapter.GetSessionByFileID(context.Background(), "file-456")
	assert.ErrorIs(t, err, domain.ErrUploadSessionNotFound)
}

func TestMemoryUploadStateAdapter_CompleteSessionClearsKey(t *testing.T) {
	adapter := NewMemoryUploadStateAdapter()
	session := contracts.FileUploadSession{
		SessionID:     "session-123",
		FileID:        "file-456",
		Status:        string(domain.SessionStatusActive),
		EncryptionKey: []byte("super-secret-key"),
	}

	require.NoError(t, adapter.CreateSession(context.Background(), session))

	// Active session must have its key.
	active, err := adapter.GetSessionByID(context.Background(), "session-123")
	require.NoError(t, err)
	assert.NotEmpty(t, active.EncryptionKey)

	// After completion the key must be gone.
	require.NoError(t, adapter.CompleteSession(context.Background(), "session-123"))

	completed, err := adapter.GetSessionByID(context.Background(), "session-123")
	require.NoError(t, err)
	assert.Equal(t, string(domain.SessionStatusComplete), completed.Status)
	assert.Empty(t, completed.EncryptionKey, "encryption key must be cleared on completion")

	// Lookup by FileID must also return the cleared key.
	byFileID, err := adapter.GetSessionByFileID(context.Background(), "file-456")
	require.NoError(t, err)
	assert.Empty(t, byFileID.EncryptionKey, "encryption key must be cleared under file-ID index too")
}

func TestMemoryUploadStateAdapter_DeleteExpiredSessions(t *testing.T) {
	adapter := NewMemoryUploadStateAdapter()
	now := time.Now()

	expiredActive := contracts.FileUploadSession{
		SessionID: "session-expired",
		FileID:    "file-expired",
		Status:    string(domain.SessionStatusActive),
		ExpiresAt: now.Add(-time.Hour),
	}
	expiredComplete := contracts.FileUploadSession{
		SessionID: "session-complete",
		FileID:    "file-complete",
		Status:    string(domain.SessionStatusComplete),
		ExpiresAt: now.Add(-time.Hour),
	}
	activeFresh := contracts.FileUploadSession{
		SessionID: "session-fresh",
		FileID:    "file-fresh",
		Status:    string(domain.SessionStatusActive),
		ExpiresAt: now.Add(time.Hour),
	}

	require.NoError(t, adapter.CreateSession(context.Background(), expiredActive))
	require.NoError(t, adapter.CreateSession(context.Background(), expiredComplete))
	require.NoError(t, adapter.CreateSession(context.Background(), activeFresh))

	removed, err := adapter.DeleteExpiredSessions(context.Background(), now)
	require.NoError(t, err)

	// Only the expired active session is removed, and it is returned exactly once
	// even though it was indexed under both SessionID and FileID.
	require.Len(t, removed, 1)
	assert.Equal(t, "session-expired", removed[0].SessionID)

	// The expired active session is gone under both keys.
	_, err = adapter.GetSessionByID(context.Background(), "session-expired")
	assert.ErrorIs(t, err, domain.ErrUploadSessionNotFound)
	_, err = adapter.GetSessionByFileID(context.Background(), "file-expired")
	assert.ErrorIs(t, err, domain.ErrUploadSessionNotFound)

	// Completed-but-expired and active-not-expired sessions are retained.
	_, err = adapter.GetSessionByID(context.Background(), "session-complete")
	assert.NoError(t, err)
	_, err = adapter.GetSessionByID(context.Background(), "session-fresh")
	assert.NoError(t, err)
}
