package domain

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
)

func newTestService() *EncryptionService {
	return NewEncryptionService(&mockKeyGen{}, logtest.NewNoop())
}

func randKey(t *testing.T) []byte {
	t.Helper()
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return k
}

func TestRoundtrip_RandomKey_PreservesPlaintext(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	lengths := []int{1, 2, 7, 16, 31, 32, 33, 64, 128, 4096}
	plaintexts := make([]string, 0, len(lengths))
	for i, n := range lengths {
		buf := make([]byte, n)
		if _, err := rand.Read(buf); err != nil {
			t.Fatalf("rand: %v", err)
		}
		// include some unicode in one entry
		if i == 3 {
			plaintexts = append(plaintexts, "héllo 世界 🔒")
			continue
		}
		plaintexts = append(plaintexts, string(buf))
	}

	key := randKey(t)
	encResp, err := svc.Encrypt(ctx, contracts.EncryptionRequest{Plaintext: plaintexts, Key: key})
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if len(encResp.Ciphertext) != len(plaintexts) {
		t.Fatalf("expected %d ciphertexts got %d", len(plaintexts), len(encResp.Ciphertext))
	}

	decResp, err := svc.Decrypt(ctx, contracts.DecryptionRequest{Ciphertext: encResp.Ciphertext, Key: key})
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if len(decResp.Plaintext) != len(plaintexts) {
		t.Fatalf("expected %d plaintexts got %d", len(plaintexts), len(decResp.Plaintext))
	}

	for i, encoded := range decResp.Plaintext {
		raw, err := base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			t.Fatalf("base64 decode plaintext[%d]: %v", i, err)
		}
		if string(raw) != plaintexts[i] {
			t.Errorf("plaintext[%d] mismatch: got %q want %q", i, string(raw), plaintexts[i])
		}
	}
}

// TestDecrypt_StoredWireFormat_RemainsCompatible pins the on-the-wire ciphertext
// format: base64.URLEncoding(nonce || gcm-sealed). Breaking this test means
// ciphertexts already stored in the database become unreadable — i.e. this is a
// backwards-compatibility guarantee, not an internal implementation detail.
func TestDecrypt_StoredWireFormat_RemainsCompatible(t *testing.T) {
	// Hard-coded 32-byte key.
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := []byte("hello world")

	// Build ciphertext using a fixed all-zero nonce so we exercise the exact
	// wire format produced by the service.
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("cipher.NewGCM: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize()) // 12 zero bytes
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	encoded := base64.URLEncoding.EncodeToString(sealed)

	svc := newTestService()
	resp, err := svc.Decrypt(context.Background(), contracts.DecryptionRequest{
		Ciphertext: []string{encoded},
		Key:        key,
	})
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if len(resp.Plaintext) != 1 {
		t.Fatalf("expected 1 plaintext got %d", len(resp.Plaintext))
	}
	want := base64.URLEncoding.EncodeToString(plaintext)
	if resp.Plaintext[0] != want {
		t.Errorf("wire format compatibility broken: got %q want %q", resp.Plaintext[0], want)
	}
}

func TestRoundtrip_WrongKey_Fails(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	key1 := randKey(t)
	key2 := randKey(t)

	enc, err := svc.Encrypt(ctx, contracts.EncryptionRequest{Plaintext: []string{"secret"}, Key: key1})
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = svc.Decrypt(ctx, contracts.DecryptionRequest{Ciphertext: enc.Ciphertext, Key: key2})
	if err == nil {
		t.Fatalf("expected decryption to fail with wrong key")
	}
	if !errors.Is(err, ErrDecryptionFailed) {
		t.Errorf("expected ErrDecryptionFailed, got %v", err)
	}
}
