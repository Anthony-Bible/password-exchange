package postgres

import (
	"errors"
	"sync"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Contract: once Close has returned, nothing may resurrect the pool, and
// operations racing with Close either succeed or get ErrRepositoryClosed
// (never a panic, never a data race).

// An explicit Connect after Close must refuse; this is the deterministic form
// of the "request passed the closed check, Close ran, Connect reopened" race.
func TestConnect_AfterCloseReturnsRepositoryClosedAndOpensNoPool(t *testing.T) {
	a := &PostgresAdapter{
		// Unroutable port: a (buggy) reopen fails with ErrDatabaseConnection
		// instead of ErrRepositoryClosed.
		config:    contracts.DatabaseConfig{Host: "127.0.0.1", Port: 1, User: "u", Password: "p", Name: "n"},
		logger:    logtest.NewNoop(),
		validator: noopValidator{},
	}
	require.NoError(t, a.Close())

	err := a.Connect()
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrRepositoryClosed),
		"Connect after Close must return ErrRepositoryClosed, got %v", err)
}

// Operations racing with Close must not panic or race, must only ever return
// nil/driver results or ErrRepositoryClosed, and must be rejected afterwards.
func TestOperationsRacingWithClose_NoPanicNoRaceAndClosedAfterwards(t *testing.T) {
	const workers = 32
	for iter := 0; iter < 20; iter++ {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)
		mock.MatchExpectationsInOrder(false)
		for i := 0; i < workers; i++ {
			mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectClose()
		a := &PostgresAdapter{db: db, logger: logtest.NewNoop(), validator: noopValidator{}}

		var wg sync.WaitGroup
		start := make(chan struct{})
		panics := make(chan any, workers)
		errsCh := make(chan error, workers)
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						panics <- r
					}
				}()
				<-start
				errsCh <- a.InsertMessage(&contracts.Message{UniqueID: "x", Content: "y", MaxViewCount: 1})
			}()
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_ = a.Close()
		}()
		close(start)
		wg.Wait()
		close(panics)
		close(errsCh)

		for p := range panics {
			t.Fatalf("operation panicked while racing with Close: %v", p)
		}
		for e := range errsCh {
			if e != nil {
				assert.True(t, errors.Is(e, domain.ErrRepositoryClosed) || errors.Is(e, domain.ErrDatabaseConnection),
					"unexpected error during Close race: %v", e)
				assert.True(t, errors.Is(e, domain.ErrRepositoryClosed),
					"op racing with Close should surface ErrRepositoryClosed, got %v", e)
			}
		}

		// After Close returned, every operation is rejected and no pool exists.
		err = a.InsertMessage(&contracts.Message{UniqueID: "x", Content: "y", MaxViewCount: 1})
		assert.True(t, errors.Is(err, domain.ErrRepositoryClosed), "got %v", err)
		assert.True(t, errors.Is(a.Connect(), domain.ErrRepositoryClosed))
		_ = mock
	}
}
