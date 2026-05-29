package domain

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
)

func TestEncrypt_Success(t *testing.T) {
	svc := newTestService()
	key := make([]byte, 32)
	resp, err := svc.Encrypt(context.Background(), contracts.EncryptionRequest{
		Plaintext: []string{"a", "bb"},
		Key:       key,
	})
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if len(resp.Ciphertext) != 2 {
		t.Fatalf("expected 2 ciphertexts got %d", len(resp.Ciphertext))
	}
	for i, c := range resp.Ciphertext {
		if c == "" {
			t.Errorf("ciphertext[%d] empty", i)
		}
	}
}

func TestEncrypt_InvalidKeyLength(t *testing.T) {
	svc := newTestService()
	_, err := svc.Encrypt(context.Background(), contracts.EncryptionRequest{
		Plaintext: []string{"hi"},
		Key:       make([]byte, 16),
	})
	if !errors.Is(err, ErrInvalidKeyLength) {
		t.Errorf("expected ErrInvalidKeyLength got %v", err)
	}
}

func TestEncrypt_EmptyPlaintextSlice(t *testing.T) {
	svc := newTestService()
	resp, err := svc.Encrypt(context.Background(), contracts.EncryptionRequest{
		Plaintext: []string{},
		Key:       make([]byte, 32),
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(resp.Ciphertext) != 0 {
		t.Errorf("expected empty ciphertext slice got %d entries", len(resp.Ciphertext))
	}
}

func TestDecrypt_Errors(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	enc, err := svc.Encrypt(ctx, contracts.EncryptionRequest{
		Plaintext: []string{"secret"},
		Key:       make([]byte, 32),
	})
	if err != nil {
		t.Fatalf("setup Encrypt: %v", err)
	}
	raw, err := base64.URLEncoding.DecodeString(enc.Ciphertext[0])
	if err != nil {
		t.Fatalf("setup base64 decode: %v", err)
	}
	raw[len(raw)/2] ^= 0x01
	tampered := base64.URLEncoding.EncodeToString(raw)

	tooShortForNonce := base64.URLEncoding.EncodeToString([]byte{1, 2, 3})

	tests := []struct {
		name       string
		ciphertext []string
		key        []byte
		wantErr    error
	}{
		{
			name:       "invalid key length",
			ciphertext: []string{"abc"},
			key:        make([]byte, 8),
			wantErr:    ErrInvalidKeyLength,
		},
		{
			name:       "malformed base64",
			ciphertext: []string{"!!!not_base64!!!"},
			key:        make([]byte, 32),
			wantErr:    ErrBase64DecodingFailed,
		},
		{
			name:       "ciphertext shorter than nonce",
			ciphertext: []string{tooShortForNonce},
			key:        make([]byte, 32),
			wantErr:    ErrInvalidCiphertext,
		},
		{
			name:       "tampered ciphertext fails authentication",
			ciphertext: []string{tampered},
			key:        make([]byte, 32),
			wantErr:    ErrDecryptionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Decrypt(ctx, contracts.DecryptionRequest{
				Ciphertext: tt.ciphertext,
				Key:        tt.key,
			})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestGenerateRandomKey_Success(t *testing.T) {
	var wantKey contracts.EncryptionKey
	for i := range wantKey {
		wantKey[i] = 0x42
	}
	kg := &mockKeyGen{
		GenerateKeyFunc: func(ctx context.Context, length int32) (contracts.EncryptionKey, error) {
			if length != 32 {
				t.Errorf("expected length 32 got %d", length)
			}
			return wantKey, nil
		},
	}
	svc := NewEncryptionService(kg, logtest.NewNoop())
	resp, err := svc.GenerateRandomKey(context.Background(), contracts.RandomRequest{Length: 32})
	if err != nil {
		t.Fatalf("GenerateRandomKey: %v", err)
	}
	if resp.Key != wantKey {
		t.Errorf("key mismatch")
	}
	if resp.KeyString != wantKey.String() {
		t.Errorf("keystring mismatch: got %q want %q", resp.KeyString, wantKey.String())
	}
}

func TestGenerateRandomKey_InvalidLength(t *testing.T) {
	svc := newTestService()
	_, err := svc.GenerateRandomKey(context.Background(), contracts.RandomRequest{Length: 16})
	if !errors.Is(err, ErrInvalidKeyLength) {
		t.Errorf("expected ErrInvalidKeyLength got %v", err)
	}
}

func TestGenerateRandomKey_KeygenError(t *testing.T) {
	sentinel := errors.New("keygen boom")
	kg := &mockKeyGen{
		GenerateKeyFunc: func(ctx context.Context, length int32) (contracts.EncryptionKey, error) {
			return contracts.EncryptionKey{}, sentinel
		},
	}
	svc := NewEncryptionService(kg, logtest.NewNoop())
	_, err := svc.GenerateRandomKey(context.Background(), contracts.RandomRequest{Length: 32})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error got %v", err)
	}
}

func TestGenerateID_Delegates(t *testing.T) {
	kg := &mockKeyGen{
		GenerateIDFunc: func(ctx context.Context) string { return "delegated-id" },
	}
	svc := NewEncryptionService(kg, logtest.NewNoop())
	if got := svc.GenerateID(context.Background()); got != "delegated-id" {
		t.Errorf("got %q want %q", got, "delegated-id")
	}
}
