package api

import (
	"context"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/middleware"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/primary"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging"
	"github.com/gin-gonic/gin"
)

// readyzTimeout bounds how long Readyz waits on its downstream HealthCheck
// calls before failing. Must stay tight so k8s readiness probes don't queue
// up behind a wedged dependency.
const readyzTimeout = 2 * time.Second

// healthCheckTimeout bounds how long the public /api/v1/health endpoint waits
// on its downstream dependency checks. Unlike readyzTimeout this is not a k8s
// probe, so it can be a touch more lenient.
const healthCheckTimeout = 5 * time.Second

// MessageAPIHandler handles REST API requests for message operations
type MessageAPIHandler struct {
	messageService    primary.MessageServicePort
	encryptionService secondary.EncryptionServicePort
	storageService    secondary.StorageServicePort
}

// NewMessageAPIHandler creates a new API message handler. The two secondary
// ports are required for the dependency-aware /readyz probe; SubmitMessage
// and friends route through messageService as before.
func NewMessageAPIHandler(
	messageService primary.MessageServicePort,
	encryptionService secondary.EncryptionServicePort,
	storageService secondary.StorageServicePort,
) *MessageAPIHandler {
	return &MessageAPIHandler{
		messageService:    messageService,
		encryptionService: encryptionService,
		storageService:    storageService,
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

	logging.Info().
		Interface("correlation_id", correlationID).
		Msg("Processing API message submission request")

	// Parse and validate JSON request
	var req models.MessageSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONErrorResponse(
			c,
			http.StatusBadRequest,
			models.ErrorCodeValidationFailed,
			"Invalid request format",
			map[string]interface{}{
				"parse_error": err.Error(),
			},
		)
		return
	}

	// Validate request using enhanced validation middleware
	if validationErrors := middleware.ValidateMessageSubmission(&req); validationErrors != nil {
		middleware.JSONErrorResponse(
			c,
			http.StatusBadRequest,
			models.ErrorCodeValidationFailed,
			"Request validation failed",
			validationErrors,
		)
		return
	}

	// Convert API request to domain request
	domainReq := domain.MessageSubmissionRequest{
		Content:           req.Content,
		IsClientEncrypted: req.IsClientEncrypted,
		Passphrase:        req.Passphrase,
		AdditionalInfo:    req.AdditionalInfo,
		Captcha:           req.AntiSpamAnswer,
		TurnstileToken:    req.TurnstileToken,
		SendNotification:  req.SendNotification,
		MaxViewCount:      req.MaxViewCount,
		ExpirationHours:   req.ExpirationHours,
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
		logging.Error().
			Err(err).
			Interface("correlation_id", correlationID).
			Msg("Failed to submit message")

		middleware.JSONErrorResponse(
			c,
			http.StatusInternalServerError,
			models.ErrorCodeInternalError,
			"Failed to submit message",
			nil,
		)
		return
	}

	// Build API response — pass ExpiresAt pointer directly from domain (nil for legacy messages)
	apiResponse := buildSubmissionResponse(response, req.SendNotification)

	logging.Info().
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

	logging.Debug().
		Str("messageId", messageID).
		Interface("correlation_id", correlationID).
		Msg("Checking message access via API")

	// Check if message exists and get access info
	accessInfo, err := h.messageService.CheckMessageAccess(ctx, messageID)
	if err != nil {
		logging.Error().
			Err(err).
			Str("messageId", messageID).
			Interface("correlation_id", correlationID).
			Msg("Failed to check message access")

		middleware.JSONErrorResponse(
			c,
			http.StatusInternalServerError,
			models.ErrorCodeInternalError,
			"Failed to check message access",
			nil,
		)
		return
	}

	if !accessInfo.Exists {
		middleware.JSONErrorResponse(
			c,
			http.StatusNotFound,
			models.ErrorCodeMessageNotFound,
			"Message not found or has expired",
			nil,
		)
		return
	}

	// Build response — pass ExpiresAt pointer directly from domain (nil for legacy messages)
	response := models.MessageAccessInfoResponse{
		MessageID:          messageID,
		Exists:             accessInfo.Exists,
		RequiresPassphrase: accessInfo.RequiresPassphrase,
		IsClientEncrypted:  accessInfo.IsClientEncrypted,
		HasBeenAccessed:    false, // TODO: Add this to domain if needed
		ExpiresAt:          accessInfo.ExpiresAt,
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

	logging.Debug().
		Str("messageId", messageID).
		Interface("correlation_id", correlationID).
		Msg("Processing message decryption via API")

	// Parse and validate JSON request
	var req models.MessageDecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONErrorResponse(
			c,
			http.StatusBadRequest,
			models.ErrorCodeValidationFailed,
			"Invalid request format",
			map[string]interface{}{
				"parse_error": err.Error(),
			},
		)
		return
	}

	// Decode the encryption key when provided (client-encrypted payloads don't send one).
	decryptionKey, err := decodeDecryptionKey(req.DecryptionKey)
	if err != nil {
		middleware.JSONErrorResponse(
			c,
			http.StatusBadRequest,
			models.ErrorCodeValidationFailed,
			"Invalid decryption key format",
			nil,
		)
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
		logging.Error().
			Err(err).
			Str("messageId", messageID).
			Interface("correlation_id", correlationID).
			Msg("Failed to retrieve message")

		// Handle specific error types
		if err == domain.ErrInvalidPassphrase {
			middleware.JSONErrorResponse(
				c,
				http.StatusUnauthorized,
				models.ErrorCodeInvalidPassphrase,
				"Invalid passphrase provided",
				nil,
			)
			return
		}

		// Check for message already consumed (this would need to be added to domain errors)
		middleware.JSONErrorResponse(
			c,
			http.StatusNotFound,
			models.ErrorCodeMessageNotFound,
			"Message not found or has expired",
			nil,
		)
		return
	}

	// Build API response
	apiResponse := models.MessageDecryptResponse{
		MessageID:         messageID,
		Content:           response.Content,
		IsClientEncrypted: response.IsClientEncrypted,
		ViewCount:         response.ViewCount,
		MaxViewCount:      response.MaxViewCount,
		DecryptedAt:       time.Now(),
		ExpiresAt:         response.ExpiresAt,
	}

	logging.Debug().
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

	logging.Debug().
		Interface("correlation_id", correlationID).
		Msg("Health check requested")

	ctx, cancel := context.WithTimeout(c.Request.Context(), healthCheckTimeout)
	defer cancel()

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		services = make(map[string]string)
	)

	checks := []struct {
		name string
		fn   func(context.Context) error
	}{
		// "database" is reported by the storage gRPC client port.
		{"database", h.storageService.HealthCheck},
		{"encryption", h.encryptionService.HealthCheck},
	}

	for _, check := range checks {
		wg.Add(1)
		go func(name string, fn func(context.Context) error) {
			defer wg.Done()
			status := "healthy"
			if err := fn(ctx); err != nil {
				status = "unhealthy"
				logging.Warn().Err(err).Str("dep", name).Msg("Health check dependency failed")
			}
			mu.Lock()
			services[name] = status
			mu.Unlock()
		}(check.name, check.fn)
	}
	wg.Wait()

	failed := 0
	for _, status := range services {
		if status == "unhealthy" {
			failed++
		}
	}

	overall := "healthy"
	switch {
	case failed == len(services):
		overall = "unhealthy"
	case failed > 0:
		overall = "degraded"
	}

	response := models.HealthCheckResponse{
		Status:    overall,
		Version:   "1.0.0", // TODO: Get from build info
		Timestamp: time.Now(),
		Services:  services,
	}

	c.JSON(http.StatusOK, response)
}

// Livez handles GET /livez. It returns 200 unconditionally without touching
// any downstream dependency so kubelet only restarts the process when it is
// genuinely wedged (not when a database hiccup degrades readiness).
func (h *MessageAPIHandler) Livez(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readyz handles GET /readyz. It probes both downstream gRPC services in
// parallel under a short deadline; if either fails the pod is taken off the
// service load balancer with a 503 and a small JSON body naming the failed
// dep(s). On success it returns 200 with {"status":"ok"}.
func (h *MessageAPIHandler) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), readyzTimeout)
	defer cancel()

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		failed []string
	)

	checks := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"encryption", h.encryptionService.HealthCheck},
		{"storage", h.storageService.HealthCheck},
	}

	for _, check := range checks {
		wg.Add(1)
		go func(name string, fn func(context.Context) error) {
			defer wg.Done()
			if err := fn(ctx); err != nil {
				mu.Lock()
				failed = append(failed, name)
				mu.Unlock()
				logging.Warn().Err(err).Str("dep", name).Msg("Readyz dependency check failed")
			}
		}(check.name, check.fn)
	}
	wg.Wait()

	if len(failed) > 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"failed": failed,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
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

	logging.Debug().
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

func buildSubmissionResponse(
	response *domain.MessageSubmissionResponse,
	sendNotification bool,
) models.MessageSubmissionResponse {
	return models.MessageSubmissionResponse{
		MessageID:         response.MessageID,
		DecryptURL:        response.DecryptURL,
		Key:               response.Key,
		IsClientEncrypted: response.IsClientEncrypted,
		WebURL:            response.DecryptURL, // Same URL works for both
		ExpiresAt:         response.ExpiresAt,
		NotificationSent:  sendNotification && response.Success,
	}
}

func decodeDecryptionKey(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, nil
	}
	return base64.URLEncoding.DecodeString(encoded)
}
