package web

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/middleware"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/primary"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// MessageHandler handles HTTP requests for message operations
type MessageHandler struct {
	messageService primary.MessageServicePort
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService primary.MessageServicePort) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// SubmitMessage handles POST requests to submit a new message
func (h *MessageHandler) SubmitMessage(c *gin.Context) {
	ctx := c.Request.Context()
	
	log.Info().Msg("Processing message submission request")

	// Parse max view count
	maxViewCount := 0
	if maxViewCountStr := c.PostForm("max_view_count"); maxViewCountStr != "" {
		parsed, err := strconv.Atoi(maxViewCountStr)
		if err != nil {
			log.Error().Err(err).Str("value", maxViewCountStr).Msg("Invalid max view count format")
			h.renderErrorWithField(c, "Invalid max view count: must be a number", "max_view_count")
			return
		}
		if parsed < 1 || parsed > 100 {
			log.Error().Int("value", parsed).Msg("Max view count out of range")
			h.renderErrorWithField(c, "Max view count must be between 1 and 100", "max_view_count")
			return
		}
		maxViewCount = parsed
	}

	// Parse expiration (value + unit → hours)
	expirationHours := 0
	if expirationValueStr := c.PostForm("expiration_value"); expirationValueStr != "" {
		parsedValue, err := strconv.Atoi(expirationValueStr)
		if err != nil {
			log.Error().Err(err).Str("value", expirationValueStr).Msg("Invalid expiration value format")
			h.renderErrorWithField(c, "Invalid expiration value: must be a number", "expiration_value")
			return
		}
		unit := c.PostForm("expiration_unit")
		switch unit {
		case "days":
			expirationHours = parsedValue * 24
		default: // "hours" or empty
			expirationHours = parsedValue
		}
		if expirationHours < 1 || expirationHours > domain.MaxExpirationHours {
			log.Error().Int("hours", expirationHours).Msg("Expiration out of range")
			h.renderErrorWithField(c, fmt.Sprintf("Expiration must be between 1 hour and %d days", domain.MaxExpirationHours/24), "expiration_value")
			return
		}
	}

	// Extract form data
	req := domain.MessageSubmissionRequest{
		Content:          c.PostForm("content"),
		SenderName:       c.PostForm("firstname"),
		SenderEmail:      c.PostForm("email"),
		RecipientName:    c.PostForm("other_firstname"),
		RecipientEmail:   c.PostForm("other_email"),
		Passphrase:       c.PostForm("other_lastname"),
		AdditionalInfo:   c.PostForm("other_information"),
		Captcha:          c.PostForm("h-captcha-response"),
		SendNotification: c.PostForm("enableEmail") != "" && webAntiSpamCheck(c.PostForm("questionId"), c.PostForm("color")),
		MaxViewCount:     maxViewCount,
		ExpirationHours:  expirationHours,
	}

	// Submit the message
	response, err := h.messageService.SubmitMessage(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to submit message")
		h.renderError(c, "Failed to submit message", err)
		return
	}

	// Web request - redirect to confirmation page
	c.Redirect(http.StatusSeeOther, "/confirmation?content="+response.DecryptURL)

	log.Info().Str("messageId", response.MessageID).Msg("Message submitted successfully")
}

// DisplayDecrypted handles GET requests to display the decryption page
func (h *MessageHandler) DisplayDecrypted(c *gin.Context) {
	ctx := c.Request.Context()
	messageID := c.Param("uuid")

	log.Debug().Str("messageId", messageID).Msg("Checking message access")

	// Check if message exists and requires passphrase
	accessInfo, err := h.messageService.CheckMessageAccess(ctx, messageID)
	if err != nil {
		log.Error().Err(err).Str("messageId", messageID).Msg("Failed to check message access")
		h.render404(c)
		return
	}

	if !accessInfo.Exists {
		log.Warn().Str("messageId", messageID).Msg("Message not found")
		h.render404(c)
		return
	}

	// Render decryption page 
	data := gin.H{
		"Title":       "passwordExchange Decrypted",
		"HasPassword": accessInfo.RequiresPassphrase,
	}

	c.HTML(http.StatusOK, "decryption.html", data)
}

// DecryptMessage handles POST requests to decrypt a message with passphrase
func (h *MessageHandler) DecryptMessage(c *gin.Context) {
	ctx := c.Request.Context()
	messageID := c.Param("uuid")
	keyParam := c.Param("key")
	passphrase := c.PostForm("passphrase")

	log.Debug().Str("messageId", messageID).Msg("Processing message decryption")

	// Decode the encryption key
	if strings.HasPrefix(keyParam, "/") {
		keyParam = keyParam[1:] // Remove leading slash
	}

	decryptionKey, err := base64.URLEncoding.DecodeString(keyParam)
	if err != nil {
		log.Error().Err(err).Str("messageId", messageID).Msg("Failed to decode encryption key")
		h.renderError(c, "Invalid decryption key", err)
		return
	}

	// Create retrieval request
	req := domain.MessageRetrievalRequest{
		MessageID:     messageID,
		DecryptionKey: decryptionKey,
		Passphrase:    passphrase,
	}

	// Retrieve and decrypt the message
	response, err := h.messageService.RetrieveMessage(ctx, req)
	if err != nil {
		log.Error().Err(err).Str("messageId", messageID).Msg("Failed to retrieve message")
		
		// Check if it's a passphrase error
		if err == domain.ErrInvalidPassphrase {
			data := gin.H{
				"Title":            "passwordExchange Decrypted",
				"DecryptedMessage": "Wrong Passphrase/Lastname. Please try again(can be empty)",
			}
			c.HTML(http.StatusOK, "decryption.html", data)
			return
		}

		h.render404(c)
		return
	}

	// Render the decrypted message
	data := gin.H{
		"Title":            "passwordExchange Decrypted",
		"DecryptedMessage": response.Content,
		"ViewCount":        response.ViewCount,
		"MaxViewCount":     response.MaxViewCount,
	}

	c.HTML(http.StatusOK, "decryption.html", data)
	log.Debug().Str("messageId", messageID).Msg("Message decrypted and displayed successfully")
}

// Static page handlers
func (h *MessageHandler) Home(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", gin.H{
		"Title": "Password Exchange",
	})
}

func (h *MessageHandler) About(c *gin.Context) {
	c.HTML(http.StatusOK, "about.html", gin.H{
		"Title": "About - Password Exchange",
	})
}

func (h *MessageHandler) Confirmation(c *gin.Context) {
	content := c.Query("content")
	c.HTML(http.StatusOK, "confirmation.html", gin.H{
		"Title": "passwordExchange",
		"Url":   content,
	})
}

func (h *MessageHandler) NotFound(c *gin.Context) {
	h.render404(c)
}

// Helper methods
func (h *MessageHandler) renderError(c *gin.Context, message string, err error) {
	log.Error().Err(err).Str("message", message).Msg("Rendering error page")
	
	data := gin.H{
		"Title":  "Error - Password Exchange",
		"Errors": map[string]string{"general": message},
	}
	
	c.HTML(http.StatusInternalServerError, "home.html", data)
}

func (h *MessageHandler) renderErrorWithField(c *gin.Context, message string, field string) {
	log.Error().Str("message", message).Str("field", field).Msg("Rendering validation error page")
	
	data := gin.H{
		"Title":  "Password Exchange",
		"Errors": map[string]string{field: message},
	}
	
	c.HTML(http.StatusBadRequest, "home.html", data)
}

func (h *MessageHandler) render404(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", gin.H{
		"Title": "Not Found - Password Exchange",
	})
}

// webAntiSpamCheck validates antispam answer for web form (converts string questionId to int)
func webAntiSpamCheck(questionIDStr, answer string) bool {
	if questionIDStr == "" || answer == "" {
		return false
	}
	
	questionID, err := strconv.Atoi(questionIDStr)
	if err != nil {
		return false
	}
	
	return middleware.IsValidAntiSpamAnswer(&questionID, answer)
}