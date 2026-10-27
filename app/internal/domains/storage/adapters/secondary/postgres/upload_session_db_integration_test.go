//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"pgregory.net/rapid"
)

var ctx = context.Background()

// --- example-based acceptance criteria ---

func TestIntegration_CompleteUploadSession_FirstCallClearsKeyAndSecondCallIsNotFound(t *testing.T) {
	db, repo := startAdapter(t)
	require.NoError(t, repo.CreateUploadSession(ctx, validSession("s1", "f1")))

	require.NoError(t, repo.CompleteUploadSession(ctx, "s1"))

	var status string
	var keyIsNull bool
	require.NoError(t, db.QueryRow(`SELECT status, encryption_key IS NULL FROM file_upload_sessions WHERE session_id = 's1'`).Scan(&status, &keyIsNull))
	assert.Equal(t, "complete", status)
	assert.True(t, keyIsNull, "encryption key must be cleared on completion")

	// Epic decision: re-completing keeps returning NotFound (explicit status guard).
	assert.ErrorIs(t, repo.CompleteUploadSession(ctx, "s1"), domain.ErrUploadSessionNotFound)
	assert.ErrorIs(t, repo.CompleteUploadSession(ctx, "never-existed"), domain.ErrUploadSessionNotFound)
}

func TestIntegration_CreateUploadSession_DuplicateSessionOrFileIDIsAlreadyExists(t *testing.T) {
	_, repo := startAdapter(t)
	require.NoError(t, repo.CreateUploadSession(ctx, validSession("s1", "f1")))

	assert.ErrorIs(t, repo.CreateUploadSession(ctx, validSession("s1", "f-other")), domain.ErrUploadSessionAlreadyExists, "duplicate session_id")
	assert.ErrorIs(t, repo.CreateUploadSession(ctx, validSession("s-other", "f1")), domain.ErrUploadSessionAlreadyExists, "duplicate file_id")
}

func TestIntegration_GetUploadSession_FindsBySessionIDOrFileID(t *testing.T) {
	_, repo := startAdapter(t)
	require.NoError(t, repo.CreateUploadSession(ctx, validSession("sess-a", "file-b")))

	bySession, err := repo.GetUploadSession(ctx, "sess-a")
	require.NoError(t, err)
	byFile, err := repo.GetUploadSession(ctx, "file-b")
	require.NoError(t, err)

	assert.Equal(t, "sess-a", bySession.SessionID)
	assert.Equal(t, "sess-a", byFile.SessionID)

	_, err = repo.GetUploadSession(ctx, "nope")
	assert.ErrorIs(t, err, domain.ErrUploadSessionNotFound)
}

func TestIntegration_DeleteExpiredUploadSessions_ReturnsExactlyTheDeletedRows(t *testing.T) {
	db, repo := startAdapter(t)
	now := time.Now()
	mk := func(id, status string, expires time.Time) {
		s := validSession(id, id+"-file")
		s.Status, s.ExpiresAt = status, expires
		require.NoError(t, repo.CreateUploadSession(ctx, s))
	}
	mk("expired-1", "active", now.Add(-time.Hour))
	mk("expired-2", "aborted", now.Add(-time.Minute))
	mk("expired-but-complete", "complete", now.Add(-time.Hour))
	mk("fresh", "active", now.Add(time.Hour))

	deleted, err := repo.DeleteExpiredUploadSessions(ctx, now)
	require.NoError(t, err)

	var ids []string
	for _, s := range deleted {
		ids = append(ids, s.SessionID)
	}
	assert.ElementsMatch(t, []string{"expired-1", "expired-2"}, ids)

	var remaining []string
	rows, err := db.Query(`SELECT session_id FROM file_upload_sessions ORDER BY session_id`)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var id string
		require.NoError(t, rows.Scan(&id))
		remaining = append(remaining, id)
	}
	assert.Equal(t, []string{"expired-but-complete", "fresh"}, remaining)

	again, err := repo.DeleteExpiredUploadSessions(ctx, now)
	require.NoError(t, err)
	assert.Empty(t, again, "a second sweep has nothing left to delete")
}

func TestIntegration_DeleteExpiredUploadSessions_ConcurrentSweepsNeverReturnTheSameRowTwice(t *testing.T) {
	_, repo := startAdapter(t)
	now := time.Now()
	const n = 30
	for i := 0; i < n; i++ {
		s := validSession(fmt.Sprintf("e%02d", i), fmt.Sprintf("e%02d-f", i))
		s.ExpiresAt = now.Add(-time.Hour)
		require.NoError(t, repo.CreateUploadSession(ctx, s))
	}

	var mu sync.Mutex
	seen := map[string]int{}
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := repo.DeleteExpiredUploadSessions(ctx, now)
			assert.NoError(t, err)
			mu.Lock()
			defer mu.Unlock()
			for _, s := range got {
				seen[s.SessionID]++
			}
		}()
	}
	wg.Wait()

	assert.Len(t, seen, n, "every expired row is reported exactly once across sweepers")
	for id, c := range seen {
		assert.Equal(t, 1, c, "row %s reported %d times", id, c)
	}
}

func TestIntegration_AddCompletedPart_Concurrent(t *testing.T) {
	_, repo := startAdapter(t)
	require.NoError(t, repo.CreateUploadSession(ctx, validSession("concurrent", "concurrent-f")))

	const N = 10
	g, gctx := errgroup.WithContext(ctx)
	for i := 1; i <= N; i++ {
		part := contracts.UploadSessionPart{PartNumber: i, ETag: fmt.Sprintf("etag-%d", i)}
		g.Go(func() error { return repo.AddCompletedPart(gctx, "concurrent", part) })
	}
	require.NoError(t, g.Wait())

	got, err := repo.GetUploadSession(ctx, "concurrent")
	require.NoError(t, err)
	assert.Len(t, got.CompletedParts, N, "no lost updates")
	seen := map[int]string{}
	for _, p := range got.CompletedParts {
		seen[p.PartNumber] = p.ETag
	}
	for i := 1; i <= N; i++ {
		assert.Equal(t, fmt.Sprintf("etag-%d", i), seen[i])
	}
}

func TestIntegration_AddCompletedPart_DuplicateRetryReplacesPart(t *testing.T) {
	_, repo := startAdapter(t)
	require.NoError(t, repo.CreateUploadSession(ctx, validSession("retry", "retry-f")))

	require.NoError(t, repo.AddCompletedPart(ctx, "retry", contracts.UploadSessionPart{PartNumber: 1, ETag: "old"}))
	require.NoError(t, repo.AddCompletedPart(ctx, "retry", contracts.UploadSessionPart{PartNumber: 1, ETag: "new"}))

	got, err := repo.GetUploadSession(ctx, "retry")
	require.NoError(t, err)
	assert.Equal(t, []contracts.UploadSessionPart{{PartNumber: 1, ETag: "new"}}, got.CompletedParts)
}

func TestIntegration_AddCompletedPart_UnknownSessionIsNotFound(t *testing.T) {
	_, repo := startAdapter(t)
	err := repo.AddCompletedPart(ctx, "ghost", contracts.UploadSessionPart{PartNumber: 1, ETag: "e"})
	assert.ErrorIs(t, err, domain.ErrUploadSessionNotFound)
}

// Documented, stable contract (see godoc on CreateUploadSession/GetUploadSession):
// a nil EncryptionKey reads back as nil; an empty non-nil EncryptionKey reads
// back as a non-nil empty slice. The two are never conflated, on any read.
func TestIntegration_EncryptionKey_NilAndEmptyAreStable(t *testing.T) {
	_, repo := startAdapter(t)
	cases := map[string]struct {
		key     []byte
		wantNil bool
	}{
		"nil":   {key: nil, wantNil: true},
		"empty": {key: []byte{}, wantNil: false},
	}
	for name, c := range cases {
		s := validSession("key-"+name, "key-"+name+"-f")
		s.EncryptionKey = c.key
		require.NoError(t, repo.CreateUploadSession(ctx, s))

		for i := 0; i < 2; i++ {
			bySession, err := repo.GetUploadSession(ctx, s.SessionID)
			require.NoError(t, err)
			byFile, err := repo.GetUploadSession(ctx, s.FileID)
			require.NoError(t, err)
			for via, got := range map[string]*contracts.UploadSession{"session id": bySession, "file id": byFile} {
				assert.Empty(t, got.EncryptionKey, "%s via %s", name, via)
				if c.wantNil {
					assert.Nil(t, got.EncryptionKey, "nil key must stay nil via %s", via)
				} else {
					assert.NotNil(t, got.EncryptionKey, "empty key must stay non-nil via %s", via)
				}
			}
		}
	}
}

func TestIntegration_OverLimitAndNULAndInvalidUTF8_AreRejectedNotTruncated(t *testing.T) {
	db, repo := startAdapter(t)
	cases := map[string]string{
		"256 chars in a varchar(255)": strings.Repeat("x", 256),
		"NUL byte":                    "abc\x00def",
		"invalid UTF-8":               "abc\xffdef",
	}
	for name, bad := range cases {
		t.Run(name, func(t *testing.T) {
			s := validSession("bad-"+strings.ReplaceAll(name, " ", "-"), "bad-f-"+strings.ReplaceAll(name, " ", "-"))
			s.Filename = bad
			assert.ErrorIs(t, repo.CreateUploadSession(ctx, s), domain.ErrInvalidParameter)
		})
	}
	assert.Equal(t, 0, rowCount(t, db, "file_upload_sessions"), "no truncated row may be stored")
}

// --- property-based acceptance criteria ---

// PBT bytea: arbitrary bytes (0..64 KiB, \x00 and high bytes included) round-trip byte-identically.
func TestIntegration_Property_EncryptionKeyRoundTripsByteIdentical(t *testing.T) {
	db, repo := startAdapter(t)
	rapid.Check(t, func(rt *rapid.T) {
		resetTables(rt, db)
		key := rapid.SliceOfN(rapid.Byte(), 1, 64*1024).Draw(rt, "key")
		s := validSession("pbt-key", "pbt-key-f")
		s.EncryptionKey = key

		require.NoError(rt, repo.CreateUploadSession(ctx, s))
		got, err := repo.GetUploadSession(ctx, "pbt-key")

		require.NoError(rt, err)
		assert.Equal(rt, key, got.EncryptionKey)
	})
}

// PBT completed_parts: repeated AddCompletedPart equals "last write per
// PartNumber wins, first-seen order, no duplicates". The model is defined here
// so the test pins the contract, not the adapter's helper.
func TestIntegration_Property_CompletedPartsMatchLastWriteWinsModel(t *testing.T) {
	db, repo := startAdapter(t)
	etag := rapid.StringOfN(rapid.Rune().Filter(func(r rune) bool { return r != 0 }), 0, 24, -1)
	partGen := rapid.Custom(func(rt *rapid.T) contracts.UploadSessionPart {
		return contracts.UploadSessionPart{
			PartNumber: rapid.IntRange(1, 8).Draw(rt, "n"), // small range forces duplicate PartNumbers
			ETag:       etag.Draw(rt, "etag"),
		}
	})
	rapid.Check(t, func(rt *rapid.T) {
		resetTables(rt, db)
		parts := rapid.SliceOfN(partGen, 1, 20).Draw(rt, "parts")
		require.NoError(rt, repo.CreateUploadSession(ctx, validSession("pbt-parts", "pbt-parts-f")))

		var model []contracts.UploadSessionPart
		for _, p := range parts {
			require.NoError(rt, repo.AddCompletedPart(ctx, "pbt-parts", p))
			replaced := false
			for i := range model {
				if model[i].PartNumber == p.PartNumber {
					model[i] = p
					replaced = true
				}
			}
			if !replaced {
				model = append(model, p)
			}
		}

		got, err := repo.GetUploadSession(ctx, "pbt-parts")
		require.NoError(rt, err)
		// completed_parts is jsonb: compare decoded structures, not bytes.
		want, _ := json.Marshal(model)
		have, _ := json.Marshal(got.CompletedParts)
		assert.JSONEq(rt, string(want), string(have))
	})
}

// PBT IDs: arbitrary xid-shaped IDs round-trip, lookups are case-sensitive, and
// either session_id or file_id finds the session.
func TestIntegration_Property_IDsRoundTripAndLookupsAreCaseSensitive(t *testing.T) {
	db, repo := startAdapter(t)
	rapid.Check(t, func(rt *rapid.T) {
		resetTables(rt, db)
		sid := xidGen().Draw(rt, "sid")
		fid := xidGen().Filter(func(s string) bool { return s != sid }).Draw(rt, "fid")
		require.NoError(rt, repo.CreateUploadSession(ctx, validSession(sid, fid)))

		for _, id := range []string{sid, fid} {
			got, err := repo.GetUploadSession(ctx, id)
			require.NoError(rt, err)
			assert.Equal(rt, sid, got.SessionID)
			assert.Equal(rt, fid, got.FileID)
		}
		for _, id := range []string{sid, fid} {
			flipped := strings.ToUpper(id)
			if flipped == id {
				continue // digits only: nothing to flip
			}
			_, err := repo.GetUploadSession(ctx, flipped)
			assert.ErrorIs(rt, err, domain.ErrUploadSessionNotFound, "case-flipped %q must not match", id)
		}
	})
}

// PBT text: valid UTF-8 within the column limit round-trips exactly; anything
// over the limit, with NUL, or with invalid UTF-8 is rejected and never stored truncated.
func TestIntegration_Property_UploadSessionTextFieldsRoundTripOrAreRejected(t *testing.T) {
	db, repo := startAdapter(t)
	const limit = 255 // varchar(255): filename, content_type, message_id, upload_id
	type field struct {
		set func(s *contracts.UploadSession, v string)
		get func(s *contracts.UploadSession) string
	}
	fields := []field{
		{func(s *contracts.UploadSession, v string) { s.Filename = v }, func(s *contracts.UploadSession) string { return s.Filename }},
		{func(s *contracts.UploadSession, v string) { s.ContentType = v }, func(s *contracts.UploadSession) string { return s.ContentType }},
		{func(s *contracts.UploadSession, v string) { s.MessageID = v }, func(s *contracts.UploadSession) string { return s.MessageID }},
		{func(s *contracts.UploadSession, v string) { s.UploadID = v }, func(s *contracts.UploadSession) string { return s.UploadID }},
	}

	rapid.Check(t, func(rt *rapid.T) {
		resetTables(rt, db)
		f := fields[rapid.IntRange(0, len(fields)-1).Draw(rt, "field")]
		s := validSession("pbt-text", "pbt-text-f")

		switch rapid.IntRange(0, 3).Draw(rt, "kind") {
		case 0: // valid, within limit
			v := textGen(0, limit).Draw(rt, "v")
			f.set(&s, v)
			require.NoError(rt, repo.CreateUploadSession(ctx, s))
			got, err := repo.GetUploadSession(ctx, "pbt-text")
			require.NoError(rt, err)
			assert.Equal(rt, v, f.get(got), "stored value must equal input")
		case 1: // over limit
			f.set(&s, textGen(limit+1, limit+30).Draw(rt, "v"))
			assert.ErrorIs(rt, repo.CreateUploadSession(ctx, s), domain.ErrInvalidParameter)
			assert.Equal(rt, 0, rowCount(rt, db, "file_upload_sessions"))
		case 2: // contains NUL
			pre := textGen(0, 20).Draw(rt, "pre")
			post := textGen(0, 20).Draw(rt, "post")
			f.set(&s, pre+"\x00"+post)
			assert.ErrorIs(rt, repo.CreateUploadSession(ctx, s), domain.ErrInvalidParameter)
			assert.Equal(rt, 0, rowCount(rt, db, "file_upload_sessions"))
		case 3: // invalid UTF-8
			pre := textGen(0, 20).Draw(rt, "pre")
			bad := rapid.SampledFrom([]string{"\xff", "\xc3\x28", "\xe2\x82", "\xf0\x9f\x92"}).Draw(rt, "bad")
			require.False(rt, utf8.ValidString(pre+bad))
			f.set(&s, pre+bad)
			assert.ErrorIs(rt, repo.CreateUploadSession(ctx, s), domain.ErrInvalidParameter)
			assert.Equal(rt, 0, rowCount(rt, db, "file_upload_sessions"))
		}
	})
}

// PBT timestamps: ExpiresAt in [now, now+10y] is stored as the same instant at
// microsecond precision regardless of the time.Local the process runs under.
func TestIntegration_Property_ExpiresAtIsPreservedAtMicrosecondPrecision(t *testing.T) {
	db, repo := startAdapter(t)
	oldLocal := time.Local
	t.Cleanup(func() { time.Local = oldLocal })

	rapid.Check(t, func(rt *rapid.T) {
		resetTables(rt, db)
		offset := rapid.IntRange(-12*3600, 14*3600).Draw(rt, "tzOffset")
		time.Local = time.FixedZone("pbt", offset)
		start := time.Now()
		secs := rapid.Int64Range(0, int64(10*365*24*3600)).Draw(rt, "secs")
		nanos := rapid.Int64Range(0, 999_999_999).Draw(rt, "nanos")
		want := start.Add(time.Duration(secs)*time.Second + time.Duration(nanos)).In(time.Local)

		s := validSession("pbt-ts", "pbt-ts-f")
		s.ExpiresAt = want
		require.NoError(rt, repo.CreateUploadSession(ctx, s))
		got, err := repo.GetUploadSession(ctx, "pbt-ts")

		require.NoError(rt, err)
		assert.True(rt, got.ExpiresAt.Equal(want.Truncate(time.Microsecond)),
			"stored %v, want %v (offset %ds)", got.ExpiresAt, want.Truncate(time.Microsecond), offset)
	})
}
