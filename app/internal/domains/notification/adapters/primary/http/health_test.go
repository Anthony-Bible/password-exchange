package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// stubQueue is a fixed-state QueueHealth used by the table tests.
type stubQueue struct{ alive bool }

func (s stubQueue) IsConnectionAlive() bool { return s.alive }

func TestHealthzHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		alive      bool
		wantStatus int
		wantBody   string
	}{
		{
			name:       "queue connection alive returns 200 ok",
			alive:      true,
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
		{
			name:       "queue connection dead returns 503",
			alive:      false,
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   "queue connection closed",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := NewHealthServer(":0", stubQueue{alive: tc.alive})
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()

			srv.Handler().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", rec.Code, tc.wantStatus)
			}
			if got := rec.Body.String(); got != tc.wantBody {
				t.Errorf("body: got %q, want %q", got, tc.wantBody)
			}
		})
	}
}

func TestHealthzRejectsNonGET(t *testing.T) {
	t.Parallel()

	srv := NewHealthServer(":0", stubQueue{alive: true})

	// The Go 1.22 mux pattern "GET /healthz" will return 405 for other methods
	// when the path matches but the method does not.
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHealthServerStartReturnsListenError(t *testing.T) {
	t.Parallel()

	srv := NewHealthServer("invalid-addr", stubQueue{alive: true})
	err := srv.Start(context.Background())
	if err == nil {
		t.Fatal("expected listen error, got nil")
	}
}

func TestHealthServerStartStopsOnContextCancel(t *testing.T) {
	srv := NewHealthServer("127.0.0.1:0", stubQueue{alive: true})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for srv.srv == nil && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if srv.srv == nil {
		t.Fatal("server did not initialize")
	}

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start returned error after context cancel: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after context cancel")
	}
}
