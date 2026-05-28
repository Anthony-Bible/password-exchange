// Package http exposes a minimal HTTP primary adapter for the notification
// (email) worker. Its sole responsibility is serving a /healthz probe so
// Kubernetes can restart the pod when the underlying queue connection has
// dropped — a failure mode the consumer goroutine doesn't otherwise surface.
package http

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// QueueHealth is the minimal contract the health endpoint needs from the
// upstream queue adapter. It is declared locally so this primary adapter does
// not import the RabbitMQ secondary adapter, preserving the hexagonal
// dependency direction (adapters depend on the domain, never on each other).
type QueueHealth interface {
	IsConnectionAlive() bool
}

// HealthServer serves a /healthz endpoint that reports queue connection
// liveness. It is intentionally tiny: no metrics, no auth, no readiness — just
// enough surface for a Kubernetes liveness probe on the email worker.
type HealthServer struct {
	addr  string
	queue QueueHealth
	srv   *http.Server
}

// NewHealthServer constructs a HealthServer bound to addr (e.g. ":8080") that
// reports the given queue's connection status via GET /healthz.
func NewHealthServer(addr string, queue QueueHealth) *HealthServer {
	return &HealthServer{
		addr:  addr,
		queue: queue,
	}
}

// Handler returns the HTTP handler used by Start. It is exported so callers
// (and tests) can drive the routes directly via httptest without binding a
// listener.
func (h *HealthServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.healthz)
	return mux
}

// healthz writes 200 "ok" when the queue connection is alive, otherwise 503
// with a short diagnostic body.
func (h *HealthServer) healthz(w http.ResponseWriter, _ *http.Request) {
	if !h.queue.IsConnectionAlive() {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("queue connection closed"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// Start runs the HTTP server and blocks until ctx is cancelled, at which point
// it triggers a graceful shutdown with a short timeout. Returns nil for the
// normal shutdown path, otherwise the underlying ListenAndServe error.
func (h *HealthServer) Start(ctx context.Context) error {
	h.srv = &http.Server{
		Addr:              h.addr,
		Handler:           h.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := h.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = h.srv.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		return err
	}
}
