package web

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/middleware"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/primary"
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// MessageHandler handles HTTP requests for message operations
type MessageHandler struct {
	messageService primary.MessageServicePort
}

const (
	acceptHeaderName        = "Accept"
	markdownMediaType       = "text/markdown"
	markdownContentType     = markdownMediaType + "; charset=utf-8"
	wrongPassphraseSentinel = "Wrong Passphrase/Lastname. Please try again(can be empty)"
)

// markdownBuilder produces markdown directly from handler data, bypassing the
// HTML template. Used for pages whose templates rely on JavaScript to inject
// the user-facing content, where HTML→markdown conversion would yield an
// empty shell.
type markdownBuilder func(data gin.H) string

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
			if parsedValue > domain.MaxExpirationHours/24 {
				log.Error().Int("days", parsedValue).Msg("Expiration out of range for days unit")
				h.renderErrorWithField(
					c,
					fmt.Sprintf("Expiration must be between 1 hour and %d days", domain.MaxExpirationHours/24),
					"expiration_value",
				)
				return
			}
			expirationHours = parsedValue * 24
		case "hours", "":
			expirationHours = parsedValue
		default:
			log.Error().Str("unit", unit).Msg("Invalid expiration unit")
			h.renderErrorWithField(c, "Invalid expiration unit: must be 'hours' or 'days'", "expiration_value")
			return
		}
		if expirationHours < 1 || expirationHours > domain.MaxExpirationHours {
			log.Error().Int("hours", expirationHours).Msg("Expiration out of range")
			h.renderErrorWithField(
				c,
				fmt.Sprintf("Expiration must be between 1 hour and %d days", domain.MaxExpirationHours/24),
				"expiration_value",
			)
			return
		}
	}

	// Extract form data
	req := domain.MessageSubmissionRequest{
		Content:        c.PostForm("content"),
		SenderName:     c.PostForm("firstname"),
		SenderEmail:    c.PostForm("email"),
		RecipientName:  c.PostForm("other_firstname"),
		RecipientEmail: c.PostForm("other_email"),
		Passphrase:     c.PostForm("other_lastname"),
		AdditionalInfo: c.PostForm("other_information"),
		Captcha:        c.PostForm("h-captcha-response"),
		SendNotification: c.PostForm("enableEmail") != "" &&
			webAntiSpamCheck(c.PostForm("questionId"), c.PostForm("color")),
		MaxViewCount:    maxViewCount,
		ExpirationHours: expirationHours,
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

	h.renderHTMLOrMarkdown(c, http.StatusOK, "decryption.html", data, displayDecryptedMarkdown)
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
				"DecryptedMessage": wrongPassphraseSentinel,
			}
			h.renderHTMLOrMarkdown(c, http.StatusOK, "decryption.html", data, decryptMessageMarkdown)
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

	h.renderHTMLOrMarkdown(c, http.StatusOK, "decryption.html", data, decryptMessageMarkdown)
	log.Debug().Str("messageId", messageID).Msg("Message decrypted and displayed successfully")
}

// Static page handlers
func (h *MessageHandler) Home(c *gin.Context) {
	c.Writer.Header().Add("Link", `</.well-known/api-catalog>; rel="api-catalog"`)
	c.Writer.Header().Add("Link", `</api/v1/docs/>; rel="service-doc"`)
	data := gin.H{
		"Title": "Password Exchange",
	}
	h.renderHTMLOrMarkdown(c, http.StatusOK, "home.html", data, nil)
}

func (h *MessageHandler) About(c *gin.Context) {
	data := gin.H{
		"Title": "About - Password Exchange",
	}
	h.renderHTMLOrMarkdown(c, http.StatusOK, "about.html", data, nil)
}

func (h *MessageHandler) Confirmation(c *gin.Context) {
	content := c.Query("content")
	data := gin.H{
		"Title": "passwordExchange",
		"Url":   content,
	}
	h.renderHTMLOrMarkdown(c, http.StatusOK, "confirmation.html", data, nil)
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

	h.renderHTMLOrMarkdown(c, http.StatusInternalServerError, "home.html", data, nil)
}

func (h *MessageHandler) renderErrorWithField(c *gin.Context, message string, field string) {
	log.Error().Str("message", message).Str("field", field).Msg("Rendering validation error page")

	data := gin.H{
		"Title":  "Password Exchange",
		"Errors": map[string]string{field: message},
	}

	h.renderHTMLOrMarkdown(c, http.StatusBadRequest, "home.html", data, nil)
}

func (h *MessageHandler) render404(c *gin.Context) {
	data := gin.H{
		"Title": "Not Found - Password Exchange",
	}
	h.renderHTMLOrMarkdown(c, http.StatusNotFound, "404.html", data, nil)
}

func (h *MessageHandler) renderHTMLOrMarkdown(
	c *gin.Context,
	statusCode int,
	templateName string,
	data gin.H,
	mdBuilder markdownBuilder,
) {
	c.Header("Vary", acceptHeaderName)

	if !wantsMarkdown(c) {
		c.HTML(statusCode, templateName, data)
		return
	}

	// Pages whose templates rely on JavaScript to inject content provide a
	// direct builder so we don't ship an empty HTML shell as markdown.
	if mdBuilder != nil {
		c.Data(statusCode, markdownContentType, []byte(mdBuilder(data)))
		return
	}

	htmlOutput, capturedStatus := captureHTMLResponse(c, statusCode, templateName, data)
	md, err := htmltomarkdown.ConvertString(htmlOutput)
	if err != nil {
		log.Error().Err(err).Str("template", templateName).Msg("html→markdown conversion failed")
		c.Data(http.StatusInternalServerError, markdownContentType, []byte("# Conversion error\n"))
		return
	}
	c.Writer.Header().Del("Content-Type") // c.HTML set this on the original writer's header map
	c.Data(capturedStatus, markdownContentType, []byte(md))
}

func captureHTMLResponse(c *gin.Context, statusCode int, templateName string, data gin.H) (string, int) {
	originalWriter := c.Writer
	captureWriter := &markdownCaptureWriter{
		ResponseWriter: originalWriter,
		body:           &bytes.Buffer{},
		status:         statusCode,
		size:           -1,
	}

	c.Writer = captureWriter
	defer func() {
		c.Writer = originalWriter
	}()
	c.HTML(statusCode, templateName, data)

	return captureWriter.body.String(), captureWriter.Status()
}

// markdownCaptureWriter buffers template output for HTML→markdown conversion.
// Streaming/upgrade methods are stubbed so promoted gin.ResponseWriter methods
// cannot leak bytes to the underlying socket while we're capturing.
type markdownCaptureWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
	size   int
}

func (w *markdownCaptureWriter) WriteHeader(code int) {
	if code >= 100 && code <= 999 {
		w.status = code
	}
}

func (w *markdownCaptureWriter) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0
	}
}

func (w *markdownCaptureWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	n, err := w.body.Write(data)
	w.size += n
	return n, err
}

func (w *markdownCaptureWriter) WriteString(s string) (int, error) {
	w.WriteHeaderNow()
	n, err := w.body.WriteString(s)
	w.size += n
	return n, err
}

func (w *markdownCaptureWriter) Status() int { return w.status }
func (w *markdownCaptureWriter) Size() int   { return w.size }
func (w *markdownCaptureWriter) Written() bool {
	return w.size != -1
}

// Flush is a no-op: callers must not flush mid-capture, since the buffer is
// the only sink while the wrapper is active.
func (w *markdownCaptureWriter) Flush() {}

// CloseNotify returns a channel that never fires; capture is in-process and
// cannot observe client disconnects on the wrapped socket.
func (w *markdownCaptureWriter) CloseNotify() <-chan bool {
	return make(chan bool, 1)
}

// Pusher returns nil so the captured response cannot trigger HTTP/2 push on
// the underlying socket while we're buffering.
func (w *markdownCaptureWriter) Pusher() http.Pusher { return nil }

// Hijack refuses to surrender the underlying socket during capture.
func (w *markdownCaptureWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, http.ErrNotSupported
}

// displayDecryptedMarkdown emits markdown for the GET /decrypt page. The HTML
// template is a JavaScript shell — the secret is only populated client-side —
// so we explain the situation and point agents to the POST flow.
func displayDecryptedMarkdown(data gin.H) string {
	requires, _ := data["HasPassword"].(bool)
	var b strings.Builder
	b.WriteString("# Decrypt Message\n\n")
	b.WriteString("This page is rendered by JavaScript in a browser. ")
	b.WriteString("To retrieve the message programmatically, POST to this URL ")
	b.WriteString("with `Accept: text/markdown` and form field `passphrase` ")
	b.WriteString("(may be empty if no passphrase was set).\n\n")
	fmt.Fprintf(&b, "- requires_passphrase: %t\n", requires)
	return b.String()
}

// decryptMessageMarkdown emits markdown for the POST /decrypt response,
// covering both the success path and the wrong-passphrase sentinel.
func decryptMessageMarkdown(data gin.H) string {
	msg, _ := data["DecryptedMessage"].(string)
	if msg == wrongPassphraseSentinel {
		return "# Decryption failed\n\nWrong passphrase. Please try again (may be empty).\n"
	}
	if msg == "" {
		return "# Decrypted message\n\n(empty)\n"
	}
	fence := pickFence(msg)
	var b strings.Builder
	b.WriteString("# Decrypted message\n\n")
	b.WriteString(fence)
	b.WriteByte('\n')
	b.WriteString(msg)
	if !strings.HasSuffix(msg, "\n") {
		b.WriteByte('\n')
	}
	b.WriteString(fence)
	b.WriteByte('\n')
	if vc, ok := data["ViewCount"].(int); ok {
		mvc, _ := data["MaxViewCount"].(int)
		fmt.Fprintf(&b, "\n- view_count: %d\n- max_view_count: %d\n", vc, mvc)
	}
	return b.String()
}

// pickFence returns a backtick run long enough to safely fence content
// containing any number of consecutive backticks.
func pickFence(content string) string {
	longest, run := 0, 0
	for _, r := range content {
		if r == '`' {
			run++
			if run > longest {
				longest = run
			}
		} else {
			run = 0
		}
	}
	length := longest + 1
	if length < 3 {
		length = 3
	}
	return strings.Repeat("`", length)
}

// wantsMarkdown negotiates Accept preferences and serves markdown only when
// markdown has positive quality and is at least as preferred as HTML.
func wantsMarkdown(c *gin.Context) bool {
	acceptHeader := strings.TrimSpace(c.GetHeader(acceptHeaderName))
	if acceptHeader == "" {
		return false
	}

	markdownQuality, markdownFound := mediaTypeMaxQuality(acceptHeader, markdownMediaType)
	if !markdownFound || markdownQuality <= 0 {
		return false
	}

	htmlQuality, htmlFound := mediaTypeMaxQuality(acceptHeader, "text/html")
	if !htmlFound {
		htmlQuality = 1
	}

	return markdownQuality >= htmlQuality
}

// mediaTypeMaxQuality returns the highest valid q value for a media type in an Accept header.
func mediaTypeMaxQuality(acceptHeader, mediaType string) (float64, bool) {
	highest := 0.0
	found := false

	for _, entry := range strings.Split(acceptHeader, ",") {
		parsedMediaType, params, err := mime.ParseMediaType(strings.TrimSpace(entry))
		if err != nil || !strings.EqualFold(parsedMediaType, mediaType) {
			continue
		}

		quality, ok := parseQuality(params)
		if !ok {
			continue
		}

		if !found || quality > highest {
			highest = quality
			found = true
		}
	}

	return highest, found
}

// parseQuality parses the q parameter and returns the RFC default of 1 when q is omitted.
func parseQuality(params map[string]string) (float64, bool) {
	qValue, ok := params["q"]
	if !ok {
		return 1, true
	}

	q, err := strconv.ParseFloat(strings.TrimSpace(qValue), 64)
	if err != nil || q < 0 || q > 1 {
		return 0, false
	}

	return q, true
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
