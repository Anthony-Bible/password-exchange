package primary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
)

// MessageServicePort defines the primary port for message operations
type MessageServicePort interface {
	// SubmitMessage handles the submission of a new encrypted message
	SubmitMessage(ctx context.Context, req domain.MessageSubmissionRequest) (*domain.MessageSubmissionResponse, error)
	
	// RetrieveMessage handles the retrieval and decryption of a stored message
	RetrieveMessage(ctx context.Context, req domain.MessageRetrievalRequest) (*domain.MessageRetrievalResponse, error)
	
	// CheckMessageAccess checks if a message exists and whether it requires a passphrase
	CheckMessageAccess(ctx context.Context, messageID string) (*domain.MessageAccessInfo, error)
}