package domain

import "errors"

var (
	// ErrInvalidEmailAddress indicates the email address is malformed
	ErrInvalidEmailAddress = errors.New("invalid email address")
	
	// ErrEmailSendFailed indicates the email sending operation failed
	ErrEmailSendFailed = errors.New("failed to send email")
	
	// ErrQueueConnectionFailed indicates queue connection failed
	ErrQueueConnectionFailed = errors.New("failed to connect to message queue")
	
	// ErrQueueConsumeFailed indicates queue consumption failed
	ErrQueueConsumeFailed = errors.New("failed to consume from message queue")
	
	// ErrTemplateRenderFailed indicates template rendering failed
	ErrTemplateRenderFailed = errors.New("failed to render notification template")
	
	// ErrTemplateNotFound indicates the requested template was not found
	ErrTemplateNotFound = errors.New("notification template not found")
	
	// ErrInvalidNotificationRequest indicates the notification request is invalid
	ErrInvalidNotificationRequest = errors.New("invalid notification request")
	
	// ErrSMTPAuthFailed indicates SMTP authentication failed
	ErrSMTPAuthFailed = errors.New("SMTP authentication failed")
	
	// ErrMessageUnmarshalFailed indicates message unmarshaling failed
	ErrMessageUnmarshalFailed = errors.New("failed to unmarshal queue message")
	
	// ErrEmptyMessageBody indicates the message body is empty
	ErrEmptyMessageBody = errors.New("message body is empty")
)