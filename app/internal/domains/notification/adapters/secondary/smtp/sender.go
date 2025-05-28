package smtp

import (
	"bytes"
	"context"
	"fmt"
	"net/smtp"
	"text/template"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
	"github.com/rs/zerolog/log"
)

// SMTPSender implements the EmailSenderPort using SMTP
type SMTPSender struct {
	emailConn domain.EmailConnection
}

// NewSMTPSender creates a new SMTP email sender
func NewSMTPSender(emailConn domain.EmailConnection) *SMTPSender {
	return &SMTPSender{
		emailConn: emailConn,
	}
}

// SendEmail sends an email notification via SMTP
func (s *SMTPSender) SendNotification(ctx context.Context, req domain.NotificationRequest) (*domain.NotificationResponse, error) {
	log.Debug().Str("to", validation.SanitizeEmailForLogging(req.To)).Str("subject", req.Subject).Msg("Sending email via SMTP")

	// Create SMTP authentication
	auth := smtp.PlainAuth("", s.emailConn.User, s.emailConn.Password, s.emailConn.Host)
	
	// Parse email template
	tmpl, err := template.ParseFiles("/templates/email_template.html")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse email template")
		return nil, fmt.Errorf("%w: %v", domain.ErrTemplateNotFound, err)
	}

	// Prepare template data
	templateData := domain.NotificationTemplateData{
		Body: fmt.Sprintf("Hi %s, \n %s used our service at <a href=\"https://password.exchange\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to https://password.exchange/about", 
			req.RecipientName, req.SenderName),
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
		log.Error().Err(err).Msg("Failed to execute email template")
		return nil, fmt.Errorf("%w: %v", domain.ErrTemplateRenderFailed, err)
	}

	// Send email
	smtpAddr := fmt.Sprintf("%s:%d", s.emailConn.Host, s.emailConn.Port)
	err = smtp.SendMail(smtpAddr, auth, s.emailConn.From, []string{req.To}, buf.Bytes())
	if err != nil {
		log.Error().Err(err).
			Str("smtpHost", s.emailConn.Host).
			Str("from", s.emailConn.From).
			Str("to", validation.SanitizeEmailForLogging(req.To)).
			Msg("Failed to send email via SMTP")
		return nil, fmt.Errorf("%w: %v", domain.ErrEmailSendFailed, err)
	}

	// Generate a simple message ID
	messageID := fmt.Sprintf("smtp-%d-%s", len(req.To)+len(req.Subject), req.To[:min(5, len(req.To))])
	
	response := &domain.NotificationResponse{
		Success:   true,
		MessageID: messageID,
	}

	log.Info().Str("to", validation.SanitizeEmailForLogging(req.To)).Str("messageId", response.MessageID).Msg("Email sent successfully via SMTP")
	return response, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}