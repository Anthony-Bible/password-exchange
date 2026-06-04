package domain

import (
	"errors"

	sherr "github.com/Anthony-Bible/password-exchange/app/internal/shared/errors"
)

var (
	// ErrInvalidEmailAddress indicates the email address is malformed.
	// Business: bad input, not retryable.
	ErrInvalidEmailAddress = sherr.WithCategory(
		errors.New("invalid email address"), sherr.CategoryBusiness)

	// ErrEmailSendFailed indicates the email sending operation failed.
	// Operational: typically transient SMTP issue.
	ErrEmailSendFailed = sherr.WithCategory(
		errors.New("failed to send email"), sherr.CategoryOperational)

	// ErrQueueConnectionFailed indicates queue connection failed.
	// Operational: transient broker availability.
	ErrQueueConnectionFailed = sherr.WithCategory(
		errors.New("failed to connect to message queue"), sherr.CategoryOperational)

	// ErrQueueConsumeFailed indicates queue consumption failed.
	// Operational: transient broker issue.
	ErrQueueConsumeFailed = sherr.WithCategory(
		errors.New("failed to consume from message queue"), sherr.CategoryOperational)

	// ErrTemplateRenderFailed indicates template rendering failed.
	// Operational: transient I/O reading a template file.
	ErrTemplateRenderFailed = sherr.WithCategory(
		errors.New("failed to render notification template"), sherr.CategoryOperational)

	// ErrTemplateNotFound indicates the requested template was not found.
	// Fatal: missing template is a deploy/config problem; retrying will not help.
	ErrTemplateNotFound = sherr.WithCategory(
		errors.New("notification template not found"), sherr.CategoryFatal)

	// ErrInvalidNotificationRequest indicates the notification request is invalid.
	// Business: validation failure, not retryable.
	ErrInvalidNotificationRequest = sherr.WithCategory(
		errors.New("invalid notification request"), sherr.CategoryBusiness)

	// ErrSMTPAuthFailed indicates SMTP authentication failed.
	// Operational: kept retryable because credential rotation often resolves
	// transiently; if it persists, alerting via repeated failures is preferred
	// over fail-fast here.
	ErrSMTPAuthFailed = sherr.WithCategory(
		errors.New("SMTP authentication failed"), sherr.CategoryOperational)

	// ErrMessageUnmarshalFailed indicates message unmarshaling failed.
	// Business: malformed input is not a retryable condition.
	ErrMessageUnmarshalFailed = sherr.WithCategory(
		errors.New("failed to unmarshal queue message"), sherr.CategoryBusiness)

	// ErrEmptyMessageBody indicates the message body is empty.
	// Business: validation failure.
	ErrEmptyMessageBody = sherr.WithCategory(
		errors.New("message body is empty"), sherr.CategoryBusiness)
)
