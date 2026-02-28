package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		config         RateLimitConfig
		requests       int
		expectedPassed int
		expectedFailed int
	}{
		{
			name: "allows requests within limit",
			config: RateLimitConfig{
				Period: 1 * time.Hour,
				Limit:  5,
			},
			requests:       3,
			expectedPassed: 3,
			expectedFailed: 0,
		},
		{
			name: "blocks requests exceeding limit",
			config: RateLimitConfig{
				Period: 1 * time.Hour,
				Limit:  2,
			},
			requests:       5,
			expectedPassed: 2,
			expectedFailed: 3,
		},
		{
			name: "uses default key generator when not provided",
			config: RateLimitConfig{
				Period: 1 * time.Hour,
				Limit:  1,
			},
			requests:       2,
			expectedPassed: 1,
			expectedFailed: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(NewRateLimitMiddleware(tt.config))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			passed := 0
			failed := 0

			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("X-Forwarded-For", "192.168.1.1") // Consistent IP
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					passed++
				} else if w.Code == http.StatusTooManyRequests {
					failed++
				}
			}

			assert.Equal(t, tt.expectedPassed, passed, "unexpected number of passed requests")
			assert.Equal(t, tt.expectedFailed, failed, "unexpected number of failed requests")
		})
	}
}

func TestDefaultKeyGenerator(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		clientIP     string
		forwardedFor string
		expected     string
	}{
		{
			name:     "uses client IP when no forwarded header",
			clientIP: "192.168.1.1",
			expected: "192.0.2.1", // Gin test context default IP
		},
		{
			name:         "uses forwarded IP when available",
			clientIP:     "127.0.0.1",
			forwardedFor: "203.0.113.1",
			expected:     "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			result := DefaultKeyGenerator(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMessageSubmissionRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(MessageSubmissionRateLimit())
	router.POST("/api/v1/messages", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"message": "created"})
	})

	// Test that we can make 10 requests within the hour limit
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code, "request %d should succeed", i+1)
	}

	// Test that the 11th request is rate limited
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "11th request should be rate limited")
}

func TestMessageAccessRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(MessageAccessRateLimit())
	router.GET("/api/v1/messages/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test that we can make 100 requests within the hour limit
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/test-id", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
	}

	// Test that the 101st request is rate limited
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/test-id", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "101st request should be rate limited")
}

func TestMessageDecryptRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(MessageDecryptRateLimit())
	router.POST("/api/v1/messages/:id/decrypt", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "decrypted"})
	})

	// Test that we can make 20 requests within the hour limit
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/test-id/decrypt", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
	}

	// Test that the 21st request is rate limited
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/test-id/decrypt", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "21st request should be rate limited")
}

func TestHealthCheckRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(HealthCheckRateLimit())
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Test that we can make 300 requests within the hour limit
	for i := 0; i < 300; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
	}

	// Test that the 301st request is rate limited
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "301st request should be rate limited")
}

func TestCustomRateLimitReachedHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("creates proper JSON response for rate limit exceeded", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationID()) // Add correlation ID middleware
		router.GET("/test", func(c *gin.Context) {
			// Simulate rate limit headers being set by the limiter
			c.Header("X-Ratelimit-Limit", "10")
			c.Header("X-Ratelimit-Remaining", "0")
			c.Header("X-Ratelimit-Reset", "1640995200")

			// Call our custom rate limit handler
			CustomRateLimitReachedHandler(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "rate_limit_exceeded", response["error"])
		assert.Equal(t, "Rate limit exceeded. Please try again later.", response["message"])
		assert.Equal(t, "/test", response["path"])
		assert.NotEmpty(t, response["timestamp"])
		assert.NotEmpty(t, response["correlation_id"])
		assert.Equal(t, "0", response["rate_limit_remaining"])
		assert.Equal(t, "1640995200", response["rate_limit_reset"])
		assert.Equal(t, "10", response["rate_limit_limit"])
	})
}

func TestRateLimitWithDifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := RateLimitConfig{
		Period: 1 * time.Hour,
		Limit:  1, // Very restrictive for testing
	}

	router := gin.New()
	router.Use(NewRateLimitMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test that different IPs have separate rate limits
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}

	for _, ip := range ips {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Forwarded-For", ip)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "first request from IP %s should succeed", ip)

		// Second request from same IP should be rate limited
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.Header.Set("X-Forwarded-For", ip)
		w2 := httptest.NewRecorder()

		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code, "second request from IP %s should be rate limited", ip)
	}
}
