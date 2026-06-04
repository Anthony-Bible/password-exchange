//go:build integration

package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/integration/dbtest"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func newUploadSessionAdapter(db *sql.DB) *MySQLAdapter {
	return &MySQLAdapter{db: db, logger: logtest.NewNoop(), validator: noopValidator{}}
}

// seedUploadSession inserts a minimal file_upload_sessions row for testing.
func seedUploadSession(t *testing.T, adapter *MySQLAdapter, sessionID string) {
	t.Helper()
	session := contracts.UploadSession{
		SessionID:   sessionID,
		FileID:      sessionID + "-file",
		MessageID:   sessionID + "-msg",
		UploadID:    sessionID + "-upload",
		Filename:    "test.txt",
		ContentType: "text/plain",
		TotalSize:   1024,
		TotalChunks: 10,
		Status:      "active",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, adapter.CreateUploadSession(context.Background(), session))
}

// TestIntegration_AddCompletedPart_Concurrent fires N concurrent AddCompletedPart
// calls for a single session and asserts all N distinct parts are present afterward.
// This validates the SELECT ... FOR UPDATE transaction fix prevents lost updates.
func TestIntegration_AddCompletedPart_Concurrent(t *testing.T) {
	db, _ := dbtest.StartMySQL(t)
	adapter := newUploadSessionAdapter(db)

	const sessionID = "concurrent-session"
	const N = 10

	seedUploadSession(t, adapter, sessionID)

	g, ctx := errgroup.WithContext(context.Background())
	for i := 1; i <= N; i++ {
		partNum := i
		g.Go(func() error {
			part := contracts.UploadSessionPart{
				PartNumber: partNum,
				ETag:       fmt.Sprintf("etag-%d", partNum),
			}
			return adapter.AddCompletedPart(ctx, sessionID, part)
		})
	}
	require.NoError(t, g.Wait())

	session, err := adapter.GetUploadSession(context.Background(), sessionID)
	require.NoError(t, err)

	assert.Len(t, session.CompletedParts, N, "all %d parts must be present after concurrent writes", N)

	partNums := make(map[int]string, len(session.CompletedParts))
	for _, p := range session.CompletedParts {
		partNums[p.PartNumber] = p.ETag
	}
	for i := 1; i <= N; i++ {
		assert.Equal(t, fmt.Sprintf("etag-%d", i), partNums[i], "part %d must be present with correct ETag", i)
	}
}

// TestIntegration_AddCompletedPart_DeduplicatesRetry verifies that calling
// AddCompletedPart twice with the same PartNumber replaces the first entry
// rather than appending a duplicate.
func TestIntegration_AddCompletedPart_DeduplicatesRetry(t *testing.T) {
	db, _ := dbtest.StartMySQL(t)
	adapter := newUploadSessionAdapter(db)

	const sessionID = "dedup-session"

	seedUploadSession(t, adapter, sessionID)

	ctx := context.Background()

	first := contracts.UploadSessionPart{PartNumber: 1, ETag: "etag-first"}
	require.NoError(t, adapter.AddCompletedPart(ctx, sessionID, first))

	second := contracts.UploadSessionPart{PartNumber: 1, ETag: "etag-second"}
	require.NoError(t, adapter.AddCompletedPart(ctx, sessionID, second))

	session, err := adapter.GetUploadSession(ctx, sessionID)
	require.NoError(t, err)

	require.Len(t, session.CompletedParts, 1, "duplicate PartNumber must not create a second entry")
	assert.Equal(t, 1, session.CompletedParts[0].PartNumber)
	assert.Equal(t, "etag-second", session.CompletedParts[0].ETag, "second call must replace the first ETag")
}
