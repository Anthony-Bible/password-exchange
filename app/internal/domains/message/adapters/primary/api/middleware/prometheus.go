package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMetrics contains all Prometheus metrics for the API
type PrometheusMetrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestsInFlight prometheus.Gauge
}

// NewPrometheusMetrics creates and registers Prometheus metrics
func NewPrometheusMetrics(registry *prometheus.Registry) *PrometheusMetrics {
	metrics := &PrometheusMetrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests by method, endpoint, and status code",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets, // Default buckets: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
			},
			[]string{"method", "endpoint"},
		),
		RequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),
	}

	// Register metrics with the provided registry
	// Use MustRegister to ensure all metrics are registered properly
	registry.MustRegister(
		metrics.RequestsTotal,
		metrics.RequestDuration,
		metrics.RequestsInFlight,
	)

	return metrics
}

// PrometheusMiddleware creates a Gin middleware that collects Prometheus metrics
func PrometheusMiddleware(metrics *PrometheusMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment in-flight requests
		metrics.RequestsInFlight.Inc()
		defer metrics.RequestsInFlight.Dec()

		// Process request
		c.Next()

		// Collect metrics after request completion
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		endpoint := c.FullPath()

		// Use request path if FullPath is empty (for unmatched routes)
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// Record metrics
		labels := []string{method, endpoint, strconv.Itoa(statusCode)}
		metrics.RequestsTotal.WithLabelValues(labels...).Inc()
		metrics.RequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
	}
}

// PrometheusHandler creates a Gin handler that exposes Prometheus metrics
func PrometheusHandler(registry *prometheus.Registry) gin.HandlerFunc {
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	return gin.WrapH(handler)
}
