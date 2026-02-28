package middleware

import (
	"net/http"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ErrorHandler middleware handles panics and converts them to JSON error responses
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		correlationID, _ := c.Get(CorrelationIDKey)

		log.Error().
			Interface("panic", recovered).
			Interface("correlation_id", correlationID).
			Str("path", c.Request.URL.Path).
			Str("method", c.Request.Method).
			Msg("Panic recovered in API handler")

		// Create standard error response
		errorResponse := models.NewStandardError(
			models.ErrorCodeInternalError,
			"An internal server error occurred",
			c.Request.URL.Path,
		)

		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
	})
}

// JSONErrorResponse sends a standardized JSON error response
func JSONErrorResponse(c *gin.Context, statusCode int, errorCode, message string, details map[string]interface{}) {
	correlationID, _ := c.Get(CorrelationIDKey)

	log.Error().
		Interface("correlation_id", correlationID).
		Str("error_code", errorCode).
		Str("message", message).
		Interface("details", details).
		Int("status_code", statusCode).
		Str("path", c.Request.URL.Path).
		Msg("API error response")

	var errorResponse *models.StandardErrorResponse
	if details != nil {
		errorResponse = models.NewValidationError(c.Request.URL.Path, details)
	} else {
		errorResponse = models.NewStandardError(errorCode, message, c.Request.URL.Path)
	}

	c.AbortWithStatusJSON(statusCode, errorResponse)
}
