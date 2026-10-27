//go:build integration

package postgres

import (
	"database/sql"
	"errors"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/integration/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestIntegration_Message_InsertThenGetRoundTrips(t *testing.T) {
	_, repo := startAdapter(t)
	exp := time.Now().Add(48 * time.Hour)
	in := &contracts.Message{
		UniqueID: "abc123", Content: "cipher", IsClientEncrypted: true,
		Passphrase: "hash", RecipientEmail: "to@example.com", MaxViewCount: 3, ExpiresAt: &exp,
	}
	require.NoError(t, repo.InsertMessage(in))

	got, err := repo.GetMessage("abc123")

	require.NoError(t, err)
	assert.Equal(t, "cipher", got.Content)
	assert.True(t, got.IsClientEncrypted)
	assert.Equal(t, "hash", got.Passphrase)
	assert.Equal(t, "to@example.com", got.RecipientEmail)
	assert.Equal(t, 3, got.MaxViewCount)
	assert.Equal(t, 0, got.ViewCount)
	require.NotNil(t, got.ExpiresAt)
	assert.True(t, got.ExpiresAt.Equal(exp.Truncate(time.Microsecond)))
}

func TestIntegration_Message_UnknownIDIsNotFound(t *testing.T) {
	_, repo := startAdapter(t)
	_, err := repo.GetMessage("missing")
	assert.ErrorIs(t, err, domain.ErrMessageNotFound)
	_, err = repo.IncrementViewCountAndGet("missing")
	assert.ErrorIs(t, err, domain.ErrMessageNotFound)
}

func TestIntegration_DeleteExpiredMessages_RemovesOnlyExpired(t *testing.T) {
	db, repo := startAdapter(t)
	past, future := time.Now().Add(-time.Hour), time.Now().Add(time.Hour)
	for id, exp := range map[string]time.Time{"old": past, "new": future} {
		exp := exp
		m := validMessage(id)
		m.ExpiresAt = &exp
		require.NoError(t, repo.InsertMessage(m))
	}

	require.NoError(t, repo.DeleteExpiredMessages())

	_, err := repo.GetMessage("old")
	assert.ErrorIs(t, err, domain.ErrMessageNotFound)
	_, err = repo.GetMessage("new")
	assert.NoError(t, err)
	assert.Equal(t, 1, rowCount(t, db, "messages"))
}

// Concurrency: N goroutines viewing a message with max M. Exactly min(N,M)
// succeed with distinct view counts 1..min(N,M); the rest get ErrMessageNotFound;
// the row is gone iff N >= M; no other error type occurs.
func TestIntegration_IncrementViewCountAndGet_ConcurrentViewers(t *testing.T) {
	cases := []struct{ n, max int }{{3, 5}, {5, 5}, {12, 5}, {20, 1}}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			db, repo := startAdapter(t)
			m := validMessage("contended")
			m.MaxViewCount = tc.max
			require.NoError(t, repo.InsertMessage(m))

			var mu sync.Mutex
			var counts []int
			var otherErrs []error
			notFound := 0
			var wg sync.WaitGroup
			start := make(chan struct{})
			for i := 0; i < tc.n; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					got, err := repo.IncrementViewCountAndGet("contended")
					mu.Lock()
					defer mu.Unlock()
					switch {
					case err == nil:
						counts = append(counts, got.ViewCount)
					case errors.Is(err, domain.ErrMessageNotFound):
						notFound++
					default:
						otherErrs = append(otherErrs, err)
					}
				}()
			}
			close(start)
			wg.Wait()

			wantOK := tc.n
			if tc.max < wantOK {
				wantOK = tc.max
			}
			assert.Empty(t, otherErrs, "only ErrMessageNotFound is an acceptable failure")
			assert.Len(t, counts, wantOK)
			assert.Equal(t, tc.n-wantOK, notFound)
			sort.Ints(counts)
			for i, c := range counts {
				assert.Equal(t, i+1, c, "each viewer sees a distinct, gapless view count")
			}
			wantRows := 1
			if tc.n >= tc.max {
				wantRows = 0
			}
			assert.Equal(t, wantRows, rowCount(t, db, "messages"))
		})
	}
}

// PBT IDs: arbitrary xid IDs round-trip through insert/get and are case-sensitive.
func TestIntegration_Property_MessageIDsRoundTripAndAreCaseSensitive(t *testing.T) {
	db, repo := startAdapter(t)
	rapid.Check(t, func(rt *rapid.T) {
		resetTables(rt, db)
		id := xidGen().Draw(rt, "id")
		require.NoError(rt, repo.InsertMessage(validMessage(id)))

		got, err := repo.GetMessage(id)
		require.NoError(rt, err)
		assert.Equal(rt, id, got.UniqueID)

		if flipped := strings.ToUpper(id); flipped != id {
			_, err := repo.GetMessage(flipped)
			assert.ErrorIs(rt, err, domain.ErrMessageNotFound)
		}
	})
}

// PBT text: valid UTF-8 within limits round-trips exactly; over-limit, NUL and
// invalid UTF-8 are rejected with the validation error and never stored.
func TestIntegration_Property_MessageTextFieldsRoundTripOrAreRejected(t *testing.T) {
	db, repo := startAdapter(t)
	type field struct {
		limit int // 0 = unbounded text column
		set   func(m *contracts.Message, v string)
		get   func(m *contracts.Message) string
	}
	fields := []field{
		{100, func(m *contracts.Message, v string) { m.Passphrase = v }, func(m *contracts.Message) string { return m.Passphrase }},
		{255, func(m *contracts.Message, v string) { m.RecipientEmail = v }, func(m *contracts.Message) string { return m.RecipientEmail }},
		{0, func(m *contracts.Message, v string) { m.Content = v }, func(m *contracts.Message) string { return m.Content }},
	}
	rapid.Check(t, func(rt *rapid.T) {
		resetTables(rt, db)
		f := fields[rapid.IntRange(0, len(fields)-1).Draw(rt, "field")]
		m := validMessage("pbt-msg")

		maxValid := f.limit
		if maxValid == 0 {
			maxValid = 2000
		}
		switch rapid.IntRange(0, 3).Draw(rt, "kind") {
		case 0:
			v := textGen(0, maxValid).Draw(rt, "v")
			if f.limit == 0 && v == "" {
				v = "x" // message is NOT NULL but empty is storable; keep it non-empty for clarity
			}
			f.set(m, v)
			require.NoError(rt, repo.InsertMessage(m))
			got, err := repo.GetMessage("pbt-msg")
			require.NoError(rt, err)
			assert.Equal(rt, v, f.get(got))
		case 1:
			if f.limit == 0 {
				rt.Skip("unbounded column has no over-limit case")
			}
			f.set(m, textGen(f.limit+1, f.limit+30).Draw(rt, "v"))
			assert.ErrorIs(rt, repo.InsertMessage(m), domain.ErrInvalidParameter)
			assert.Equal(rt, 0, rowCount(rt, db, "messages"))
		case 2:
			f.set(m, textGen(0, 10).Draw(rt, "pre")+"\x00"+textGen(0, 10).Draw(rt, "post"))
			assert.ErrorIs(rt, repo.InsertMessage(m), domain.ErrInvalidParameter)
			assert.Equal(rt, 0, rowCount(rt, db, "messages"))
		case 3:
			f.set(m, textGen(0, 10).Draw(rt, "pre")+rapid.SampledFrom([]string{"\xff", "\xc3\x28", "\xf0\x9f\x92"}).Draw(rt, "bad"))
			assert.ErrorIs(rt, repo.InsertMessage(m), domain.ErrInvalidParameter)
			assert.Equal(rt, 0, rowCount(rt, db, "messages"))
		}
	})
}

// Pool: the adapter must never hold more server connections than its limit.
// Default (zero) is 5; an explicit MaxOpenConns is honoured. Connections are
// pinned deterministically by blocking readers behind an ACCESS EXCLUSIVE lock.
func TestIntegration_ConnectionPool_RespectsDefaultAndConfiguredLimit(t *testing.T) {
	for _, tc := range []struct {
		name     string
		maxOpen  int
		wantPeak int
	}{{"default is 5", 0, 5}, {"explicit 2", 2, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			db, cfg := dbtest.StartPostgresBaseline(t)
			cfg.MaxOpenConns = tc.maxOpen
			repo := newAdapterFor(t, cfg)
			require.NoError(t, repo.InsertMessage(validMessage("pool")))

			// Dedicated locker + counter connections, tagged so they are excluded from the count.
			lockDB := openTagged(t, cfg, "locker")
			lockDB.SetMaxOpenConns(1)
			tx, err := lockDB.Begin()
			require.NoError(t, err)
			_, err = tx.Exec(`LOCK TABLE messages IN ACCESS EXCLUSIVE MODE`)
			require.NoError(t, err)
			countDB := openTagged(t, cfg, "counter")
			countDB.SetMaxOpenConns(1)

			var wg sync.WaitGroup
			for i := 0; i < 30; i++ {
				wg.Add(1)
				go func() { defer wg.Done(); _, _ = repo.GetMessage("pool") }()
			}

			countAdapterConns := func() int {
				var n int
				require.NoError(t, countDB.QueryRow(
					`SELECT count(*) FROM pg_stat_activity
					 WHERE datname = $1 AND application_name NOT IN ('locker','counter')
					   AND pid <> pg_backend_pid() AND state = 'active' AND wait_event_type = 'Lock'`,
					cfg.Name).Scan(&n))
				return n
			}
			require.Eventually(t, func() bool { return countAdapterConns() >= tc.wantPeak }, 10*time.Second, 25*time.Millisecond)
			peak := 0
			for end := time.Now().Add(500 * time.Millisecond); time.Now().Before(end); time.Sleep(20 * time.Millisecond) {
				if n := countAdapterConns(); n > peak {
					peak = n
				}
			}
			require.NoError(t, tx.Rollback())
			wg.Wait()
			_ = db

			assert.Equal(t, tc.wantPeak, peak, "adapter connections blocked at once")
		})
	}
}

func openTagged(t *testing.T, cfg contracts.DatabaseConfig, app string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dbtest.PostgresDSN(cfg)+"&application_name="+app)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}
