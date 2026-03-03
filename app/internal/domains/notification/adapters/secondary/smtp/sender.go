package smtp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"regexp"
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Pre-compiled regexes and safe function data (initialized once at package load).
var (
	// dangerousFuncRegexes maps dangerous function names to their pre-compiled pattern.
	dangerousFuncRegexes map[string]*regexp.Regexp

	// funcCallRegex matches function-call-like patterns in templates: {{func arg}}.
	funcCallRegex *regexp.Regexp

	// pathTraversalRegexes are pre-compiled patterns for path traversal detection.
	pathTraversalRegexes []pathPattern

	// scriptRegexes are pre-compiled patterns for script injection detection.
	scriptRegexes []*regexp.Regexp

	// safeFuncMap is the immutable set of template functions available to email templates.
	safeFuncMap template.FuncMap

	// safeFuncNames is the set of allowed function names (FuncMap keys + Go template builtins).
	safeFuncNames map[string]bool
)

type pathPattern struct {
	re          *regexp.Regexp
	isTraversal bool // true for ../ patterns vs absolute path patterns
}

func init() {
	// Build safe template functions once.
	titler := cases.Title(language.Und, cases.NoLower)
	safeFuncMap = template.FuncMap{
		"upper":   strings.ToUpper,
		"lower":   strings.ToLower,
		"title":   titler.String,
		"trim":    strings.TrimSpace,
		"replace": strings.ReplaceAll,
		"url":     template.URLQueryEscaper,
		"printf":  fmt.Sprintf,
	}

	// Derive safe function names from the FuncMap keys plus Go template builtins.
	// Single source of truth: adding a function to safeFuncMap automatically allows it in validation.
	safeFuncNames = make(map[string]bool)
	for name := range safeFuncMap {
		safeFuncNames[name] = true
	}
	for _, name := range []string{
		"and", "or", "not", "len", "index", "print", "println",
		"if", "else", "end", "range", "with", "template", "define", "block",
	} {
		safeFuncNames[name] = true
	}

	// Pre-compile dangerous function detection patterns.
	dangerousFuncRegexes = make(map[string]*regexp.Regexp)
	for _, fn := range []string{
		"exec", "system", "call", "env", "read", "write", "open", "close",
		"file", "dir", "os", "cmd", "shell", "process", "eval", "run",
	} {
		pattern := fmt.Sprintf(`\{\{[^}]*\b%s\b[^}]*\}\}`, regexp.QuoteMeta(fn))
		dangerousFuncRegexes[fn] = regexp.MustCompile(pattern)
	}

	// Pre-compile function call pattern.
	funcCallRegex = regexp.MustCompile(`\{\{(?:\s*-\s*)?([a-zA-Z_][a-zA-Z0-9_]*)\s+[^}|]*\}\}`)

	// Pre-compile path traversal patterns.
	for _, p := range []struct {
		pattern     string
		isTraversal bool
	}{
		{`\.\./`, true},
		{`\.\.\\`, true},
		{`/etc/`, false},
		{`/var/`, false},
		{`/usr/`, false},
		{`/root/`, false},
		{`/home/`, false},
		{`C:\\\\`, false},
		{`%SYSTEMROOT%`, false},
	} {
		pathTraversalRegexes = append(pathTraversalRegexes, pathPattern{
			re:          regexp.MustCompile(p.pattern),
			isTraversal: p.isTraversal,
		})
	}

	// Pre-compile script injection patterns (case insensitive).
	for _, p := range []string{
		`(?i)<script[^>]*>.*?</script>`,
		`(?i)javascript:`,
		`(?i)vbscript:`,
		`(?i)onload=`,
		`(?i)onerror=`,
		`(?i)onclick=`,
	} {
		scriptRegexes = append(scriptRegexes, regexp.MustCompile(p))
	}
}

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

// getSafeTemplateFunctions returns the package-level safe template FuncMap.
// html/template handles contextual escaping automatically, so no manual
// html/js escape helpers are needed (they would defeat auto-escaping).
func (s *SMTPSender) getSafeTemplateFunctions() template.FuncMap {
	return safeFuncMap
}

// validateTemplateContent validates template content for security and safety.
// Prevents template injection attacks by checking for dangerous functions and patterns.
// Uses pre-compiled regexes from package init for efficiency.
func (s *SMTPSender) validateTemplateContent(templateContent string) error {
	const (
		maxTemplateSize = 10 * 1024 // 10KB
		maxNestingDepth = 50
	)

	if len(templateContent) > maxTemplateSize {
		return errors.New("template too large: exceeds 10KB limit")
	}

	// Check for dangerous function patterns using pre-compiled regexes.
	for fn, re := range dangerousFuncRegexes {
		if re.MatchString(templateContent) {
			return fmt.Errorf("dangerous function '%s' detected in template", fn)
		}
	}

	// Check for undefined function names using pre-compiled regex and derived safe list.
	matches := funcCallRegex.FindAllStringSubmatch(templateContent, -1)
	for _, match := range matches {
		if len(match) > 1 {
			funcName := match[1]
			if strings.HasPrefix(funcName, ".") {
				continue
			}
			if !safeFuncNames[funcName] {
				return fmt.Errorf("undefined function '%s' detected in template", funcName)
			}
		}
	}

	// Check for path traversal patterns using pre-compiled regexes.
	for _, pp := range pathTraversalRegexes {
		if pp.re.MatchString(templateContent) {
			if pp.isTraversal {
				return errors.New("path traversal detected in template")
			}
			return errors.New("absolute path detected in template")
		}
	}

	// Check for script injection using pre-compiled regexes.
	for _, re := range scriptRegexes {
		if re.MatchString(templateContent) {
			return errors.New("script tag detected in template")
		}
	}

	// Check nesting depth by counting template constructs.
	nestingLevel := 0
	maxNesting := 0

	i := 0
	for i < len(templateContent) {
		if i < len(templateContent)-1 && templateContent[i] == '{' && templateContent[i+1] == '{' {
			j := i + 2
			for j < len(templateContent)-1 && !(templateContent[j] == '}' && templateContent[j+1] == '}') {
				j++
			}

			if j < len(templateContent)-1 {
				tagContent := templateContent[i+2 : j]
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

	// Check if templateConfig looks like a file path - attempt to read directly.
	// Avoids TOCTOU by skipping os.Stat and handling os.ReadFile errors instead.
	if strings.HasPrefix(templateConfig, "/") || strings.Contains(templateConfig, "/") {
		content, err := os.ReadFile(templateConfig)
		if err == nil {
			templateContent = string(content)
		} else if errors.Is(err, os.ErrNotExist) {
			// File doesn't exist, treat as inline template
			templateContent = templateConfig
		} else {
			return nil, fmt.Errorf("failed to read template file: %w", err)
		}
	} else {
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
func (s *SMTPSender) SendNotification(
	ctx context.Context,
	req contracts.NotificationRequest,
) (*contracts.NotificationResponse, error) {
	s.logger.Debug().
		Str("to", s.validation.SanitizeEmailForLogging(req.To)).
		Str("subject", req.Subject).
		Msg("Sending email via SMTP")

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
	templateData := contracts.NotificationTemplateData{
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

	s.logger.Info().
		Str("to", s.validation.SanitizeEmailForLogging(req.To)).
		Str("messageId", response.MessageID).
		Msg("Email sent successfully via SMTP")
	return response, nil
}
