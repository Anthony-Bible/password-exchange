//go:build integration

package postgres

import (
	"sync"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/integration/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The adapter connects lazily on first use, and `app database` never connects
// eagerly, so the first concurrent requests race. This must neither be a data
// race (run with -race) nor open more than one pool: the db-g1-small budget
// (shared by dev and prod) assumes at most MaxOpenConns connections per pod.
func TestIntegration_ConcurrentFirstUse_OpensSingleBoundedPool(t *testing.T) {
	const maxOpen = 2
	const goroutines = 24

	db, cfg := dbtest.StartPostgresBaseline(t)
	cfg.MaxOpenConns = maxOpen
	repo := newAdapterFor(t, cfg)

	// Dedicated connection so measuring does not disturb the baseline count.
	probe, err := db.Conn(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = probe.Close() })
	countConns := func() int {
		var n int
		require.NoError(t, probe.QueryRowContext(ctx,
			`SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND pid <> pg_backend_pid()`).Scan(&n))
		return n
	}
	before := countConns()

	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			if i%4 == 0 {
				errs <- repo.Ping(ctx)
				return
			}
			_, err := repo.GetMessage("does-not-exist")
			if err == domain.ErrMessageNotFound {
				err = nil
			}
			errs <- err
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		assert.NoError(t, err)
	}

	assert.LessOrEqual(t, countConns()-before, maxOpen,
		"concurrent first use must share one pool capped at MaxOpenConns, not open one pool per caller")
}
