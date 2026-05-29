package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
)

func TestGenerateKey_Valid32Bytes(t *testing.T) {
	g := NewKeyGenerator(logtest.NewNoop())
	key, err := g.GenerateKey(context.Background(), 32)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(key.Bytes()) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(key.Bytes()))
	}
	var zero contracts.EncryptionKey
	if key == zero {
		t.Fatal("generated key is all zero (vanishingly unlikely)")
	}
}

func TestGenerateKey_InvalidLength(t *testing.T) {
	g := NewKeyGenerator(logtest.NewNoop())
	_, err := g.GenerateKey(context.Background(), 16)
	if !errors.Is(err, domain.ErrInvalidKeyLength) {
		t.Fatalf("expected ErrInvalidKeyLength, got %v", err)
	}
}

func TestGenerateID_NonEmpty(t *testing.T) {
	g := NewKeyGenerator(logtest.NewNoop())
	id := g.GenerateID(context.Background())
	if id == "" {
		t.Fatal("expected non-empty id")
	}
}

func TestGenerateID_Unique(t *testing.T) {
	g := NewKeyGenerator(logtest.NewNoop())
	id1 := g.GenerateID(context.Background())
	id2 := g.GenerateID(context.Background())
	if id1 == id2 {
		t.Fatalf("expected unique ids, got %q twice", id1)
	}
}
