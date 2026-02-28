package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimitConfig defines configuration for rate limiting
type RateLimitConfig struct {
	// Period is the time window for rate limiting
	Period time.Duration
	// Limit is the maximum number of requests allowed in the period
	Limit int64
	// KeyGenerator generates the key for rate limiting (default: IP-based)
	KeyGenerator func(*gin.Context) string
}

// DefaultKeyGenerator generates rate limit key based on client IP
func DefaultKeyGenerator(c *gin.Context) string {
	return c.ClientIP()
}

// NewRateLimitMiddleware creates a new rate limiting middleware with the given configuration
func NewRateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	// Use default key generator if not provided
	if config.KeyGenerator == nil {
		config.KeyGenerator = DefaultKeyGenerator
	}

	// Create memory store for rate limiting
	store := memory.NewStore()

	// Create rate limiter
	rate := limiter.Rate{
		Period: config.Period,
		Limit:  config.Limit,
	}

	// Create limiter instance
	instance := limiter.New(store, rate, limiter.WithTrustForwardHeader(true))

	// Return gin middleware with custom error handler
	return ginlimiter.NewMiddleware(instance,
		ginlimiter.WithKeyGetter(config.KeyGenerator),
		ginlimiter.WithLimitReachedHandler(CustomRateLimitReachedHandler),
	)
}

// MessageSubmissionRateLimit creates rate limiter for message submission
// 10 requests per hour per IP
func MessageSubmissionRateLimit() gin.HandlerFunc {
	config := RateLimitConfig{
		Period: 1 * time.Hour,
		Limit:  10,
	}
	return NewRateLimitMiddleware(config)
}

// MessageAccessRateLimit creates rate limiter for message access
// 100 requests per hour per IP
func MessageAccessRateLimit() gin.HandlerFunc {
	config := RateLimitConfig{
		Period: 1 * time.Hour,
		Limit:  100,
	}
	return NewRateLimitMiddleware(config)
}

// MessageDecryptRateLimit creates rate limiter for message decryption
// 20 requests per hour per IP
func MessageDecryptRateLimit() gin.HandlerFunc {
	config := RateLimitConfig{
		Period: 1 * time.Hour,
		Limit:  20,
	}
	return NewRateLimitMiddleware(config)
}

// CustomRateLimitErrorHandler creates a rate limiter middleware with custom JSON error responses
func CustomRateLimitErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// This middleware handles rate limit responses via the custom limiter options
	}
}

// CustomRateLimitReachedHandler creates a custom rate limit reached handler that returns JSON
func CustomRateLimitReachedHandler(c *gin.Context) {
	correlationID, _ := c.Get(CorrelationIDKey)

	errorResponse := gin.H{
		"error":     "rate_limit_exceeded",
		"message":   "Rate limit exceeded. Please try again later.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"path":      c.Request.URL.Path,
	}

	if correlationID != nil {
		errorResponse["correlation_id"] = correlationID
	}

	// Add rate limit headers if available from response headers
	if remaining := c.Writer.Header().Get("X-Ratelimit-Remaining"); remaining != "" {
		errorResponse["rate_limit_remaining"] = remaining
	}
	if resetTime := c.Writer.Header().Get("X-Ratelimit-Reset"); resetTime != "" {
		errorResponse["rate_limit_reset"] = resetTime
	}
	if limit := c.Writer.Header().Get("X-Ratelimit-Limit"); limit != "" {
		errorResponse["rate_limit_limit"] = limit
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusTooManyRequests, errorResponse)
}

// HealthCheckRateLimit creates a more lenient rate limiter for health checks
// 300 requests per hour per IP
func HealthCheckRateLimit() gin.HandlerFunc {
	config := RateLimitConfig{
		Period: 1 * time.Hour,
		Limit:  300,
	}
	return NewRateLimitMiddleware(config)
}
