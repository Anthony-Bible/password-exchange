package domain

import "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"

// These aliases preserve backward compatibility for callers that import the
// domain package directly (e.g. `domain.EncryptionRequest{}`) while the
// canonical type definitions live in the contracts package.
type (
	EncryptionKey      = contracts.EncryptionKey
	EncryptionRequest  = contracts.EncryptionRequest
	EncryptionResponse = contracts.EncryptionResponse
	DecryptionRequest  = contracts.DecryptionRequest
	DecryptionResponse = contracts.DecryptionResponse
	RandomRequest      = contracts.RandomRequest
	RandomResponse     = contracts.RandomResponse
	KeyGenerator       = contracts.KeyGenerator
)
