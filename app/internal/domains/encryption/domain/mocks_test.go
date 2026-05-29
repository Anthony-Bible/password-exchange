package domain

import (
	"context"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"
)

// mockKeyGen implements contracts.KeyGenerator with configurable behavior.
type mockKeyGen struct {
	GenerateKeyFunc func(ctx context.Context, length int32) (contracts.EncryptionKey, error)
	GenerateIDFunc  func(ctx context.Context) string
}

func (m *mockKeyGen) GenerateKey(ctx context.Context, length int32) (contracts.EncryptionKey, error) {
	if m.GenerateKeyFunc != nil {
		return m.GenerateKeyFunc(ctx, length)
	}
	var k contracts.EncryptionKey
	for i := range k {
		k[i] = 0xAA
	}
	return k, nil
}

func (m *mockKeyGen) GenerateID(ctx context.Context) string {
	if m.GenerateIDFunc != nil {
		return m.GenerateIDFunc(ctx)
	}
	return "fixed-id-123"
}

var _ contracts.KeyGenerator = (*mockKeyGen)(nil)
