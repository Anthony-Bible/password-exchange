package smtp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
)

// SMTPSender implements the EmailPort using SMTP
type SMTPSender struct {
	emailConn  contracts.EmailConnection
	config     secondary.ConfigPort
	logger     secondary.LoggerPort
	validation secondary.ValidationPort
}

// NewSMTPSender creates a new SMTP email sender
func NewSMTPSender(
	emailConn contracts.EmailConnection,
	config secondary.ConfigPort,
	logger secondary.LoggerPort,
	validation secondary.ValidationPort,
) *SMTPSender {
	return &SMTPSender{
		emailConn:  emailConn,
		config:     config,
		logger:     logger,
		validation: validation,
	}
}

// getSafeTemplateFunctions returns a map of safe template functions
// Only includes essential functions needed for email templates
func (s *SMTPSender) getSafeTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		// String manipulation functions
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"title":    strings.Title,
		"trim":     strings.TrimSpace,
		"replace":  strings.ReplaceAll,
		
		// HTML escaping for security
		"html": template.HTMLEscaper,
		"js":   template.JSEscaper,
		"url":  template.URLQueryEscaper,
		
		// Safe formatting
		"printf": fmt.Sprintf,
	}
}

// validateTemplateContent validates template content for security and safety
// Prevents template injection attacks by checking for dangerous functions and patterns
func (s *SMTPSender) validateTemplateContent(templateContent string) error {
	// Constants for validation limits
	const (
		maxTemplateSize   = 10 * 1024 // 10KB
		maxNestingDepth   = 50
	)

	// Check template size
	if len(templateContent) > maxTemplateSize {
		return errors.New("template too large: exceeds 10KB limit")
	}

	// Check for dangerous function patterns
	dangerousFunctions := []string{
		"exec", "system", "call", "env", "read", "write", "open", "close",
		"file", "dir", "os", "cmd", "shell", "process", "eval", "run",
	}

	// Create regex patterns for dangerous function detection
	for _, fn := range dangerousFunctions {
		// Match function calls like {{exec ...}} or {{.Data | exec}}
		pattern := fmt.Sprintf(`\{\{[^}]*\b%s\b[^}]*\}\}`, regexp.QuoteMeta(fn))
		matched, err := regexp.MatchString(pattern, templateContent)
		if err != nil {
			return fmt.Errorf("regex error checking for dangerous function %s: %w", fn, err)
		}
		if matched {
			return fmt.Errorf("dangerous function '%s' detected in template", fn)
		}
	}

	// Check for potentially dangerous function names that might be undefined
	// Use regex to find function-like patterns that aren't in our safe list
	safeFunctionNames := []string{
		"upper", "lower", "title", "trim", "replace", "html", "js", "url", "printf",
		// Go template built-ins that are safe
		"and", "or", "not", "len", "index", "print", "println", 
		"if", "else", "end", "range", "with", "template", "define", "block",
	}
	
	// Look for function calls in templates (pattern: function name followed by space/args)
	// This pattern looks for function calls like {{func arg}} but not {{.field}} or {{.field | func}}
	functionPattern := `\{\{(?:\s*-\s*)?([a-zA-Z_][a-zA-Z0-9_]*)\s+[^}|]*\}\}`
	re := regexp.MustCompile(functionPattern)
	matches := re.FindAllStringSubmatch(templateContent, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			funcName := match[1]
			// Skip template control structures and safe built-ins
			if strings.HasPrefix(funcName, ".") {
				continue
			}
			
			// Check if function is in safe list
			isSafe := false
			for _, safeName := range safeFunctionNames {
				if funcName == safeName {
					isSafe = true
					break
				}
			}
			
			if !isSafe {
				return fmt.Errorf("undefined function '%s' detected in template", funcName)
			}
		}
	}

	// Check for path traversal patterns
	pathTraversalPatterns := []string{
		`\.\./`,           // ../
		`\.\.\\`,          // ..\
		`/etc/`,           // /etc/
		`/var/`,           // /var/
		`/usr/`,           // /usr/
		`/root/`,          // /root/
		`/home/`,          // /home/
		`C:\\\\`,          // C:\
		`%SYSTEMROOT%`,    // %SYSTEMROOT%
	}

	for _, pattern := range pathTraversalPatterns {
		matched, err := regexp.MatchString(pattern, templateContent)
		if err != nil {
			return fmt.Errorf("regex error checking for path traversal: %w", err)
		}
		if matched {
			if strings.Contains(pattern, `\.\.`) {
				return errors.New("path traversal detected in template")
			}
			return errors.New("absolute path detected in template")
		}
	}

	// Check for script injection
	scriptPatterns := []string{
		`<script[^>]*>.*?</script>`,
		`javascript:`,
		`vbscript:`,
		`onload=`,
		`onerror=`,
		`onclick=`,
	}

	for _, pattern := range scriptPatterns {
		matched, err := regexp.MatchString(`(?i)`+pattern, templateContent)
		if err != nil {
			return fmt.Errorf("regex error checking for script injection: %w", err)
		}
		if matched {
			return errors.New("script tag detected in template")
		}
	}

	// Check nesting depth by counting template constructs
	nestingLevel := 0
	maxNesting := 0
	
	// Simple state machine to track nesting
	i := 0
	for i < len(templateContent) {
		if i < len(templateContent)-1 && templateContent[i] == '{' && templateContent[i+1] == '{' {
			// Found opening template tag
			j := i + 2
			for j < len(templateContent)-1 && !(templateContent[j] == '}' && templateContent[j+1] == '}') {
				j++
			}
			
			if j < len(templateContent)-1 {
				// Found closing tag, extract content
				tagContent := templateContent[i+2:j]
				
				// Check if it's a block opening tag (range, if, with, define, block)
				trimmed := strings.TrimSpace(tagContent)
				if strings.HasPrefix(trimmed, "range ") || 
				   strings.HasPrefix(trimmed, "if ") || 
				   strings.HasPrefix(trimmed, "with ") ||
				   strings.HasPrefix(trimmed, "define ") ||
				   strings.HasPrefix(trimmed, "block ") {
					nestingLevel++
					if nestingLevel > maxNesting {
						maxNesting = nestingLevel
					}
				} else if trimmed == "end" {
					nestingLevel--
				}
				
				i = j + 2
			} else {
				i++
			}
		} else {
			i++
		}
	}

	if maxNesting > maxNestingDepth {
		return fmt.Errorf("nesting too deep: %d levels (max %d)", maxNesting, maxNestingDepth)
	}

	return nil
}

// parseTemplate parses a template that can be either a file path or inline template content
// Uses restricted set of safe template functions to prevent injection attacks
func (s *SMTPSender) parseTemplate(templateConfig string) (*template.Template, error) {
	// Define safe template functions only
	safeFuncs := s.getSafeTemplateFunctions()
	
	var templateContent string
	
	// Check if templateConfig starts with '/' or contains common path separators - likely a file path
	if strings.HasPrefix(templateConfig, "/") || strings.Contains(templateConfig, "/") {
		// Check if file exists
		if _, err := os.Stat(templateConfig); err == nil {
			// It's a file path and file exists - read content
			content, err := os.ReadFile(templateConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to read template file: %w", err)
			}
			templateContent = string(content)
		} else {
			// File doesn't exist, treat as inline template
			templateContent = templateConfig
		}
	} else {
		// Treat as inline template content
		templateContent = templateConfig
	}
	
	// Validate template content before parsing
	if err := s.validateTemplateContent(templateContent); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}
	
	// Parse template with safe functions
	return template.New("email").Funcs(safeFuncs).Parse(templateContent)
}

// buildSafeEmailHeaders constructs email headers with CRLF injection protection
func (s *SMTPSender) buildSafeEmailHeaders(fromName, fromEmail, to, subject string) (string, error) {
	// Validate and sanitize all header values to prevent CRLF injection
	if err := validation.ValidateEmailHeaderValue(fromName); err != nil {
		return "", fmt.Errorf("invalid from name: %w", err)
	}
	if err := validation.ValidateEmailHeaderValue(fromEmail); err != nil {
		return "", fmt.Errorf("invalid from email: %w", err)
	}
	if err := validation.ValidateEmailHeaderValue(to); err != nil {
		return "", fmt.Errorf("invalid to email: %w", err)
	}
	if err := validation.ValidateEmailHeaderValue(subject); err != nil {
		return "", fmt.Errorf("invalid subject: %w", err)
	}

	// Sanitize all values to ensure no CRLF characters remain
	safeFromName := validation.SanitizeEmailHeaderValue(fromName)
	safeFromEmail := validation.SanitizeEmailHeaderValue(fromEmail)
	safeTo := validation.SanitizeEmailHeaderValue(to)
	safeSubject := validation.SanitizeEmailHeaderValue(subject)

	// Construct email headers with proper CRLF line endings
	mimeHeaders := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	emailHeaders := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\n%s",
		safeFromName, safeFromEmail, safeTo, safeSubject, mimeHeaders)

	return emailHeaders, nil
}

// SendEmail sends an email notification via SMTP
func (s *SMTPSender) SendNotification(ctx context.Context, req contracts.NotificationRequest) (*contracts.NotificationResponse, error) {
	s.logger.Debug().Str("to", s.validation.SanitizeEmailForLogging(req.To)).Str("subject", req.Subject).Msg("Sending email via SMTP")

	// Create SMTP authentication
	auth := smtp.PlainAuth("", s.emailConn.User, s.emailConn.Password, s.emailConn.Host)
	
	// Parse email template using injected config (supports both file paths and inline templates)
	templateConfig := s.config.GetEmailTemplate()
	tmpl, err := s.parseTemplate(templateConfig)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to parse email template")
		return nil, fmt.Errorf("%w: %v", domain.ErrTemplateNotFound, err)
	}

	// Prepare template data
	passwordExchangeURL := s.config.GetPasswordExchangeURL()
	templateData := contracts.NotificationTemplateData{
		Body: fmt.Sprintf(s.config.GetInitialNotificationBodyTemplate(),
			req.RecipientName, req.SenderName, passwordExchangeURL, passwordExchangeURL),
		Message:       req.MessageContent,
		SenderName:    req.SenderName,
		RecipientName: req.RecipientName,
		MessageURL:    req.MessageURL,
	}

	// Build email headers with CRLF injection protection
	emailHeaders, err := s.buildSafeEmailHeaders(req.FromName, req.From, req.To, req.Subject)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to build safe email headers")
		return nil, fmt.Errorf("%w: %v", domain.ErrEmailSendFailed, err)
	}

	// Render template
	body := []byte(emailHeaders)
	buf := bytes.NewBuffer(body)
	
	err = tmpl.Execute(buf, templateData)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to execute email template")
		return nil, fmt.Errorf("%w: %v", domain.ErrTemplateRenderFailed, err)
	}

	// Send email
	smtpAddr := fmt.Sprintf("%s:%d", s.emailConn.Host, s.emailConn.Port)
	err = smtp.SendMail(smtpAddr, auth, s.emailConn.From, []string{req.To}, buf.Bytes())
	if err != nil {
		s.logger.Error().Err(err).
			Str("smtpHost", s.emailConn.Host).
			Str("from", s.emailConn.From).
			Str("to", s.validation.SanitizeEmailForLogging(req.To)).
			Msg("Failed to send email via SMTP")
		return nil, fmt.Errorf("%w: %v", domain.ErrEmailSendFailed, err)
	}

	// Generate a simple message ID
	messageID := fmt.Sprintf("smtp-%d-%s", len(req.To)+len(req.Subject), req.To[:min(5, len(req.To))])
	
	response := &contracts.NotificationResponse{
		Success:   true,
		MessageID: messageID,
	}

	s.logger.Info().Str("to", s.validation.SanitizeEmailForLogging(req.To)).Str("messageId", response.MessageID).Msg("Email sent successfully via SMTP")
	return response, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}