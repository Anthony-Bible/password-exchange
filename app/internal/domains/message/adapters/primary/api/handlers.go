package api

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/middleware"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/primary"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// MessageAPIHandler handles REST API requests for message operations
type MessageAPIHandler struct {
	messageService primary.MessageServicePort
}

// NewMessageAPIHandler creates a new API message handler
func NewMessageAPIHandler(messageService primary.MessageServicePort) *MessageAPIHandler {
	return &MessageAPIHandler{
		messageService: messageService,
	}
}

// SubmitMessage handles POST /api/v1/messages
// @Summary Submit a new message
// @Description Creates a new encrypted message that can be accessed via a unique URL. Optionally sends email notifications to the recipient.
// @Tags Messages
// @Accept json
// @Produce json
// @Param request body models.MessageSubmissionRequest true "Message submission request"
// @Success 201 {object} models.MessageSubmissionResponse "Message successfully created"
// @Failure 400 {object} models.StandardErrorResponse "Validation error"
// @Failure 422 {object} models.StandardErrorResponse "Anti-spam verification failed"
// @Failure 500 {object} models.StandardErrorResponse "Internal server error"
// @Router /messages [post]
func (h *MessageAPIHandler) SubmitMessage(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID, _ := c.Get(middleware.CorrelationIDKey)

	log.Info().
		Interface("correlation_id", correlationID).
		Msg("Processing API message submission request")

	// Parse and validate JSON request
	var req models.MessageSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONErrorResponse(c, http.StatusBadRequest, models.ErrorCodeValidationFailed, "Invalid request format", map[string]interface{}{
			"parse_error": err.Error(),
		})
		return
	}

	// Validate request using enhanced validation middleware
	if validationErrors := middleware.ValidateMessageSubmission(&req); validationErrors != nil {
		middleware.JSONErrorResponse(c, http.StatusBadRequest, models.ErrorCodeValidationFailed, "Request validation failed", validationErrors)
		return
	}

	// Convert API request to domain request
	domainReq := domain.MessageSubmissionRequest{
		Content:          req.Content,
		Passphrase:       req.Passphrase,
		AdditionalInfo:   req.AdditionalInfo,
		Captcha:          req.AntiSpamAnswer,
		TurnstileToken:   req.TurnstileToken,
		SendNotification: req.SendNotification,
		MaxViewCount:     req.MaxViewCount,
	}

	if req.Sender != nil {
		domainReq.SenderName = req.Sender.Name
		domainReq.SenderEmail = req.Sender.Email
	}

	if req.Recipient != nil {
		domainReq.RecipientName = req.Recipient.Name
		domainReq.RecipientEmail = req.Recipient.Email
	}

	// Add remote IP to context for Turnstile validation
	remoteIP := c.ClientIP()
	ctxWithIP := context.WithValue(ctx, "RemoteIP", remoteIP)

	// Submit the message
	response, err := h.messageService.SubmitMessage(ctxWithIP, domainReq)
	if err != nil {
		log.Error().
			Err(err).
			Interface("correlation_id", correlationID).
			Msg("Failed to submit message")

		middleware.JSONErrorResponse(c, http.StatusInternalServerError, models.ErrorCodeInternalError, "Failed to submit message", nil)
		return
	}

	// Build API response
	apiResponse := models.MessageSubmissionResponse{
		MessageID:        response.MessageID,
		DecryptURL:       response.DecryptURL,
		Key:              response.Key,
		WebURL:           response.DecryptURL,            // Same URL works for both
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour), // Messages expire after 7 days
		NotificationSent: req.SendNotification && response.Success,
	}

	log.Info().
		Str("messageId", response.MessageID).
		Interface("correlation_id", correlationID).
		Msg("Message submitted successfully via API")

	c.JSON(http.StatusCreated, apiResponse)
}

// GetMessageInfo handles GET /api/v1/messages/{id}
// @Summary Get message access information
// @Description Retrieves information about a message including whether it exists, requires a passphrase, and has been accessed.
// @Tags Messages
// @Accept json
// @Produce json
// @Param id path string true "Message ID" format(uuid)
// @Param key query string true "Base64-encoded decryption key" format(byte)
// @Success 200 {object} models.MessageAccessInfoResponse "Message information retrieved"
// @Failure 404 {object} models.StandardErrorResponse "Message not found or expired"
// @Failure 500 {object} models.StandardErrorResponse "Internal server error"
// @Router /messages/{id} [get]
func (h *MessageAPIHandler) GetMessageInfo(c *gin.Context) {
	ctx := c.Request.Context()
	messageID := c.Param("id")
	correlationID, _ := c.Get(middleware.CorrelationIDKey)

	log.Debug().
		Str("messageId", messageID).
		Interface("correlation_id", correlationID).
		Msg("Checking message access via API")

	// Check if message exists and get access info
	accessInfo, err := h.messageService.CheckMessageAccess(ctx, messageID)
	if err != nil {
		log.Error().
			Err(err).
			Str("messageId", messageID).
			Interface("correlation_id", correlationID).
			Msg("Failed to check message access")

		middleware.JSONErrorResponse(c, http.StatusInternalServerError, models.ErrorCodeInternalError, "Failed to check message access", nil)
		return
	}

	if !accessInfo.Exists {
		middleware.JSONErrorResponse(c, http.StatusNotFound, models.ErrorCodeMessageNotFound, "Message not found or has expired", nil)
		return
	}

	// Build response
	response := models.MessageAccessInfoResponse{
		MessageID:          messageID,
		Exists:             accessInfo.Exists,
		RequiresPassphrase: accessInfo.RequiresPassphrase,
		HasBeenAccessed:    false,                          // TODO: Add this to domain if needed
		ExpiresAt:          time.Now().Add(7 * 24 * time.Hour), // Messages expire after 7 days
	}

	c.JSON(http.StatusOK, response)
}

// DecryptMessage handles POST /api/v1/messages/{id}/decrypt
// @Summary Decrypt a message
// @Description Decrypts and retrieves the message content. This is a one-time operation that will delete the message after successful decryption.
// @Tags Messages
// @Accept json
// @Produce json
// @Param id path string true "Message ID" format(uuid)
// @Param request body models.MessageDecryptRequest true "Decryption request"
// @Success 200 {object} models.MessageDecryptResponse "Message successfully decrypted"
// @Failure 401 {object} models.StandardErrorResponse "Invalid passphrase"
// @Failure 404 {object} models.StandardErrorResponse "Message not found or expired"
// @Failure 410 {object} models.StandardErrorResponse "Message already consumed"
// @Failure 500 {object} models.StandardErrorResponse "Internal server error"
// @Router /messages/{id}/decrypt [post]
func (h *MessageAPIHandler) DecryptMessage(c *gin.Context) {
	ctx := c.Request.Context()
	messageID := c.Param("id")
	correlationID, _ := c.Get(middleware.CorrelationIDKey)

	log.Debug().
		Str("messageId", messageID).
		Interface("correlation_id", correlationID).
		Msg("Processing message decryption via API")

	// Parse and validate JSON request
	var req models.MessageDecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONErrorResponse(c, http.StatusBadRequest, models.ErrorCodeValidationFailed, "Invalid request format", map[string]interface{}{
			"parse_error": err.Error(),
		})
		return
	}

	// Decode the encryption key
	decryptionKey, err := base64.URLEncoding.DecodeString(req.DecryptionKey)
	if err != nil {
		middleware.JSONErrorResponse(c, http.StatusBadRequest, models.ErrorCodeValidationFailed, "Invalid decryption key format", nil)
		return
	}

	// Create domain retrieval request
	domainReq := domain.MessageRetrievalRequest{
		MessageID:     messageID,
		DecryptionKey: decryptionKey,
		Passphrase:    req.Passphrase,
	}

	// Retrieve and decrypt the message
	response, err := h.messageService.RetrieveMessage(ctx, domainReq)
	if err != nil {
		log.Error().
			Err(err).
			Str("messageId", messageID).
			Interface("correlation_id", correlationID).
			Msg("Failed to retrieve message")

		// Handle specific error types
		if err == domain.ErrInvalidPassphrase {
			middleware.JSONErrorResponse(c, http.StatusUnauthorized, models.ErrorCodeInvalidPassphrase, "Invalid passphrase provided", nil)
			return
		}

		// Check for message already consumed (this would need to be added to domain errors)
		middleware.JSONErrorResponse(c, http.StatusNotFound, models.ErrorCodeMessageNotFound, "Message not found or has expired", nil)
		return
	}

	// Build API response
	apiResponse := models.MessageDecryptResponse{
		MessageID:    messageID,
		Content:      response.Content,
		ViewCount:    response.ViewCount,
		MaxViewCount: response.MaxViewCount,
		DecryptedAt:  time.Now(),
	}

	log.Debug().
		Str("messageId", messageID).
		Interface("correlation_id", correlationID).
		Msg("Message decrypted successfully via API")

	c.JSON(http.StatusOK, apiResponse)
}

// HealthCheck handles GET /api/v1/health
// @Summary Health check
// @Description Returns the health status of the API and its dependencies
// @Tags Utility
// @Accept json
// @Produce json
// @Success 200 {object} models.HealthCheckResponse "Service health status"
// @Router /health [get]
func (h *MessageAPIHandler) HealthCheck(c *gin.Context) {
	correlationID, _ := c.Get(middleware.CorrelationIDKey)

	log.Debug().
		Interface("correlation_id", correlationID).
		Msg("Health check requested")

	// TODO: Implement actual health checks for services
	response := models.HealthCheckResponse{
		Status:    "healthy",
		Version:   "1.0.0", // TODO: Get from build info
		Timestamp: time.Now(),
		Services: map[string]string{
			"database":   "healthy", // TODO: Check database service
			"encryption": "healthy", // TODO: Check encryption service
			"email":      "healthy", // TODO: Check email service
		},
	}

	c.JSON(http.StatusOK, response)
}

// APIInfo handles GET /api/v1/info
// @Summary API information
// @Description Returns information about the API including available endpoints and features
// @Tags Utility
// @Accept json
// @Produce json
// @Success 200 {object} models.APIInfoResponse "API information"
// @Router /info [get]
func (h *MessageAPIHandler) APIInfo(c *gin.Context) {
	correlationID, _ := c.Get(middleware.CorrelationIDKey)

	log.Debug().
		Interface("correlation_id", correlationID).
		Msg("API info requested")

	response := models.APIInfoResponse{
		Version:       "1.0.0",
		Documentation: "/api/v1/docs", // TODO: Implement swagger docs
		Endpoints: map[string]string{
			"submit":  "POST /api/v1/messages",
			"access":  "GET /api/v1/messages/{id}",
			"decrypt": "POST /api/v1/messages/{id}/decrypt",
			"health":  "GET /api/v1/health",
			"info":    "GET /api/v1/info",
		},
		Features: map[string]bool{
			"emailNotifications":   true,
			"passphraseProtection": true,
			"antiSpamProtection":   true,
			"emailReminders":       true,
		},
	}

	c.JSON(http.StatusOK, response)
}
