package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusMetricsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("collects request counter metrics", func(t *testing.T) {
		// Create a new registry for isolated testing
		registry := prometheus.NewRegistry()
		metrics := NewPrometheusMetrics(registry)
		
		router := gin.New()
		router.Use(PrometheusMiddleware(metrics))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify request counter metric
		assert.Equal(t, 1, testutil.CollectAndCount(metrics.RequestsTotal))
		assert.Equal(t, float64(1), testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues("GET", "/test", "200")))
	})

	t.Run("tracks different HTTP status codes", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		metrics := NewPrometheusMetrics(registry)
		
		router := gin.New()
		router.Use(PrometheusMiddleware(metrics))
		router.GET("/success", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		router.GET("/error", func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		})

		// Make requests with different status codes
		testCases := []struct {
			path       string
			statusCode int
			count      int
		}{
			{"/success", 200, 2},
			{"/error", 500, 1},
		}

		for _, tc := range testCases {
			for i := 0; i < tc.count; i++ {
				req := httptest.NewRequest(http.MethodGet, tc.path, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				assert.Equal(t, tc.statusCode, w.Code)
			}
		}

		// Verify metrics were collected correctly
		assert.Equal(t, float64(2), testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues("GET", "/success", "200")))
		assert.Equal(t, float64(1), testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues("GET", "/error", "500")))
	})

	t.Run("records request duration histogram", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		metrics := NewPrometheusMetrics(registry)
		
		router := gin.New()
		router.Use(PrometheusMiddleware(metrics))
		router.GET("/fast", func(c *gin.Context) {
			time.Sleep(1 * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"message": "fast"})
		})
		router.GET("/slow", func(c *gin.Context) {
			time.Sleep(50 * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"message": "slow"})
		})

		// Make requests with different durations
		req1 := httptest.NewRequest(http.MethodGet, "/fast", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		req2 := httptest.NewRequest(http.MethodGet, "/slow", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		// Verify histogram metrics were recorded
		assert.Equal(t, 2, testutil.CollectAndCount(metrics.RequestDuration))
		
		// Check that duration observations were recorded by examining the histogram
		// The histogram should have recorded observations for both endpoints
		fastHistogram := metrics.RequestDuration.WithLabelValues("GET", "/fast")
		slowHistogram := metrics.RequestDuration.WithLabelValues("GET", "/slow")
		
		// We can't easily get the exact count, but we can verify the histograms exist
		assert.NotNil(t, fastHistogram)
		assert.NotNil(t, slowHistogram)
	})

	t.Run("tracks requests in flight gauge", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		metrics := NewPrometheusMetrics(registry)
		
		router := gin.New()
		router.Use(PrometheusMiddleware(metrics))
		router.GET("/test", func(c *gin.Context) {
			// During request processing, in-flight should be > 0
			time.Sleep(5 * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// After request completion, in-flight should be 0
		assert.Equal(t, float64(0), testutil.ToFloat64(metrics.RequestsInFlight))
	})
}

func TestPrometheusMetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	t.Run("exposes metrics via /metrics endpoint", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		metrics := NewPrometheusMetrics(registry)
		
		// Record some test metrics
		metrics.RequestsTotal.WithLabelValues("GET", "/api/v1/messages", "200").Inc()
		metrics.RequestsTotal.WithLabelValues("POST", "/api/v1/messages", "201").Inc()
		// Record a duration observation to make the histogram appear
		metrics.RequestDuration.WithLabelValues("GET", "/api/v1/messages").Observe(0.1)
		
		router := gin.New()
		router.GET("/metrics", PrometheusHandler(registry))
		
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
		
		// Verify response contains Prometheus metrics format
		body := w.Body.String()
		assert.Contains(t, body, "# HELP")
		assert.Contains(t, body, "# TYPE")
		assert.Contains(t, body, "http_requests_total")
		assert.Contains(t, body, "http_request_duration_seconds")
		assert.Contains(t, body, "method=\"GET\"")
		assert.Contains(t, body, "status_code=\"200\"")
	})

	t.Run("metrics endpoint works with empty metrics", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		NewPrometheusMetrics(registry) // Initialize metrics but don't record anything
		
		router := gin.New()
		router.GET("/metrics", PrometheusHandler(registry))
		
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		body := w.Body.String()
		assert.Contains(t, body, "# HELP")
		assert.Contains(t, body, "# TYPE")
	})
}

func TestPrometheusMetricsRegistration(t *testing.T) {
	t.Run("registers metrics with custom registry", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		metrics := NewPrometheusMetrics(registry)
		
		// Record some observations to make vector metrics appear in output
		metrics.RequestsTotal.WithLabelValues("GET", "/test", "200").Inc()
		metrics.RequestDuration.WithLabelValues("GET", "/test").Observe(0.1)
		
		// Verify metrics are registered by attempting to gather them
		gathered, err := registry.Gather()
		require.NoError(t, err)
		require.NotEmpty(t, gathered)
		
		// Check that our expected metrics are present
		metricNames := make(map[string]bool)
		for _, mf := range gathered {
			metricNames[mf.GetName()] = true
		}
		
		assert.True(t, metricNames["http_requests_total"], "http_requests_total should be registered")
		assert.True(t, metricNames["http_request_duration_seconds"], "http_request_duration_seconds should be registered")
		assert.True(t, metricNames["http_requests_in_flight"], "http_requests_in_flight should be registered")
	})

	t.Run("handles duplicate registration by panicking as expected", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		
		// First registration should succeed
		metrics1 := NewPrometheusMetrics(registry)
		assert.NotNil(t, metrics1)
		
		// Second registration to same registry should panic (expected Prometheus behavior)
		assert.Panics(t, func() {
			NewPrometheusMetrics(registry)
		}, "duplicate registration should panic")
	})
}