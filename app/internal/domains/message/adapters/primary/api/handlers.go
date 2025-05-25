package api

import (
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

	// Validate conditional requirements
	if err := h.validateSubmissionRequest(&req); err != nil {
		middleware.JSONErrorResponse(c, http.StatusBadRequest, models.ErrorCodeValidationFailed, "Request validation failed", map[string]interface{}{
			"validation_errors": err,
		})
		return
	}

	// Convert API request to domain request
	domainReq := domain.MessageSubmissionRequest{
		Content:          req.Content,
		Passphrase:       req.Passphrase,
		AdditionalInfo:   req.AdditionalInfo,
		SendNotification: req.SendNotification,
		SkipEmail:        !req.SendNotification,
	}

	if req.Sender != nil {
		domainReq.SenderName = req.Sender.Name
		domainReq.SenderEmail = req.Sender.Email
	}

	if req.Recipient != nil {
		domainReq.RecipientName = req.Recipient.Name
		domainReq.RecipientEmail = req.Recipient.Email
	}

	// Submit the message
	response, err := h.messageService.SubmitMessage(ctx, domainReq)
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
		WebURL:           response.DecryptURL, // Same URL works for both
		ExpiresAt:        time.Now().Add(24 * time.Hour), // TODO: Get from config
		NotificationSent: req.SendNotification && response.Success,
	}

	log.Info().
		Str("messageId", response.MessageID).
		Interface("correlation_id", correlationID).
		Msg("Message submitted successfully via API")

	c.JSON(http.StatusCreated, apiResponse)
}

// GetMessageInfo handles GET /api/v1/messages/{id}
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
		HasBeenAccessed:    false, // TODO: Add this to domain if needed
		ExpiresAt:          time.Now().Add(24 * time.Hour), // TODO: Get from storage
	}

	c.JSON(http.StatusOK, response)
}

// DecryptMessage handles POST /api/v1/messages/{id}/decrypt
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
		MessageID:   messageID,
		Content:     response.Content,
		DecryptedAt: time.Now(),
	}

	log.Debug().
		Str("messageId", messageID).
		Interface("correlation_id", correlationID).
		Msg("Message decrypted successfully via API")

	c.JSON(http.StatusOK, apiResponse)
}

// HealthCheck handles GET /api/v1/health
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
		},
	}

	c.JSON(http.StatusOK, response)
}

// validateSubmissionRequest validates the message submission request
func (h *MessageAPIHandler) validateSubmissionRequest(req *models.MessageSubmissionRequest) map[string]string {
	errors := make(map[string]string)

	// If notifications are enabled, sender and recipient info is required
	if req.SendNotification {
		if req.Sender == nil {
			errors["sender"] = "Sender information is required when notifications are enabled"
		} else {
			if req.Sender.Name == "" {
				errors["sender.name"] = "Sender name is required when notifications are enabled"
			}
			if req.Sender.Email == "" {
				errors["sender.email"] = "Sender email is required when notifications are enabled"
			}
		}

		if req.Recipient == nil {
			errors["recipient"] = "Recipient information is required when notifications are enabled"
		} else {
			if req.Recipient.Name == "" {
				errors["recipient.name"] = "Recipient name is required when notifications are enabled"
			}
			if req.Recipient.Email == "" {
				errors["recipient.email"] = "Recipient email is required when notifications are enabled"
			}
		}

		// Anti-spam validation
		if req.AntiSpamAnswer != "blue" {
			errors["antiSpamAnswer"] = "Anti-spam answer must be 'blue'"
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}