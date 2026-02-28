package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// CorrelationIDHeader is the header name for correlation ID
	CorrelationIDHeader = "X-Correlation-ID"
	// CorrelationIDKey is the context key for correlation ID
	CorrelationIDKey = "correlation_id"
)

// CorrelationID middleware adds a correlation ID to each request
func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if correlation ID is already provided in headers
		correlationID := c.GetHeader(CorrelationIDHeader)

		// Generate a new one if not provided
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Add to context and response headers
		c.Set(CorrelationIDKey, correlationID)
		c.Header(CorrelationIDHeader, correlationID)

		c.Next()
	}
}
