package domain

import (
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
)

// Type aliases to contracts - these define the domain's data contracts
type (
	LogEvent                       = contracts.LogEvent
	MessageSubmissionRequest       = contracts.MessageSubmissionRequest
	MessageSubmissionResponse      = contracts.MessageSubmissionResponse
	MessageRetrievalRequest        = contracts.MessageRetrievalRequest
	MessageRetrievalResponse       = contracts.MessageRetrievalResponse
	MessageAccessInfo              = contracts.MessageAccessInfo
	MessageStorageRequest          = contracts.MessageStorageRequest
	MessageRetrievalStorageRequest = contracts.MessageRetrievalStorageRequest
	MessageStorageResponse         = contracts.MessageStorageResponse
	MessageNotificationRequest     = contracts.MessageNotificationRequest
)

// Default settings for the message domain
const (
	DefaultMessageTTL    = 7 * 24 * time.Hour // 7 days
	MaxExpirationHours   = 720                // 30 days
	MinExpirationHours   = 1                  // 1 hour
	DefaultMaxViewCount  = 5
	AbsoluteMaxViewCount = 100
)

// Message represents the core domain entity for an encrypted message
type Message struct {
	ID             string
	Content        string
	RecipientEmail string
	ViewCount      int
	MaxViewCount   int
	ExpiresAt      *time.Time
}
