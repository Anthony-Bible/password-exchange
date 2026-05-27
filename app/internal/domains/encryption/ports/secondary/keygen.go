package secondary

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"
)

// KeyGeneratorPort is an alias for contracts.KeyGenerator. The contracts
// package owns the canonical interface; this alias preserves the
// hexagonal-conventional "*Port" name without introducing a parallel
// interface that could drift over time.
type KeyGeneratorPort = contracts.KeyGenerator
