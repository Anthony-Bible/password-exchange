package smtp

import (
	"bytes"
	"context"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
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

// parseTemplate parses a template that can be either a file path or inline template content
func (s *SMTPSender) parseTemplate(templateConfig string) (*template.Template, error) {
	// Check if templateConfig starts with '/' or contains common path separators - likely a file path
	if strings.HasPrefix(templateConfig, "/") || strings.Contains(templateConfig, "/") {
		// Check if file exists
		if _, err := os.Stat(templateConfig); err == nil {
			// It's a file path and file exists
			return template.ParseFiles(templateConfig)
		}
	}
	
	// Treat as inline template content
	return template.New("email").Parse(templateConfig)
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
		Message: req.MessageContent,
	}

	// Build email headers
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	emailHeaders := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\n%s",
		req.FromName, req.From, req.To, req.Subject, mimeHeaders)

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