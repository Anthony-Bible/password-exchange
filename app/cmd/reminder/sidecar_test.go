package reminder

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShutdownSidecar(t *testing.T) {
	// Test case 1: Istio sidecar is available
	t.Run("Istio sidecar available", func(t *testing.T) {
		istioServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/quitquitquit", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer istioServer.Close()

		shutdownSidecar(istioServer.URL, "http://localhost:12346")
	})

	// Test case 2: Linkerd sidecar is available
	t.Run("Linkerd sidecar available", func(t *testing.T) {
		linkerdServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/shutdown", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer linkerdServer.Close()

		shutdownSidecar("http://localhost:12345", linkerdServer.URL)
	})

	// Test case 3: No sidecar is available
	t.Run("No sidecar available", func(t *testing.T) {
		shutdownSidecar("http://localhost:12345", "http://localhost:12346")
	})
}
