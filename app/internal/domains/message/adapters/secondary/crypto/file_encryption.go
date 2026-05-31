// Package crypto provides an in-process implementation of the file-encryption
// secondary port. Unlike the previous gRPC-backed adapter, chunks are sealed
// and opened locally with AES-256-GCM: the symmetric key already lives in the
// caller (it is returned by GenerateKey), so round-tripping every chunk's bytes
// to the encryption service bought no key isolation while doubling base64
// overhead. Each chunk additionally binds its file ID, one-based index, and
// total chunk count as AES-GCM additional authenticated data, so reordered,
// truncated, relocated, or tampered ciphertext fails authentication on decrypt.
package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
)

// keyLength is the required AES-256 key size in bytes.
const keyLength = 32

var (
	// ErrInvalidKeyLength indicates the supplied key is not 32 bytes.
	ErrInvalidKeyLength = errors.New("crypto: encryption key must be 32 bytes")
	// ErrMalformedCiphertext indicates the stored framing could not be parsed.
	ErrMalformedCiphertext = errors.New("crypto: malformed encrypted chunk framing")
	// ErrChunkCountMismatch indicates the recovered chunk count differs from the
	// expected total, signalling truncated or padded ciphertext.
	ErrChunkCountMismatch = errors.New("crypto: encrypted chunk count does not match expected total")
)

var _ secondary.FileEncryptionServicePort = (*FileEncryptionAdapter)(nil)

// FileEncryptionAdapter encrypts and decrypts file chunks locally with
// AES-256-GCM, delegating only identifier and key generation to the shared
// encryption service.
type FileEncryptionAdapter struct {
	keySource secondary.EncryptionServicePort
}

// NewFileEncryptionAdapter creates a local file-encryption adapter. keySource is
// used solely for GenerateID and GenerateKey; per-chunk encryption never leaves
// this process. keySource may be nil when only the encrypt/decrypt methods are
// exercised (e.g. in tests).
func NewFileEncryptionAdapter(keySource secondary.EncryptionServicePort) *FileEncryptionAdapter {
	return &FileEncryptionAdapter{keySource: keySource}
}

// GenerateID delegates unique identifier generation to the shared encryption service.
func (a *FileEncryptionAdapter) GenerateID(ctx context.Context) (string, error) {
	if a.keySource == nil {
		return "", errors.New("crypto: key source is nil")
	}
	return a.keySource.GenerateID(ctx)
}

// GenerateKey delegates symmetric key generation to the shared encryption service.
func (a *FileEncryptionAdapter) GenerateKey(ctx context.Context, length int32) ([]byte, error) {
	if a.keySource == nil {
		return nil, errors.New("crypto: key source is nil")
	}
	return a.keySource.GenerateKey(ctx, length)
}

// EncryptChunk seals data with AES-256-GCM, binding meta as additional
// authenticated data, and returns a length-prefixed frame ready to be stored as
// a single object-storage part:
//
//	[uint32 big-endian length][12-byte nonce][ciphertext+16-byte tag]
func (a *FileEncryptionAdapter) EncryptChunk(_ context.Context, data []byte, key []byte, meta secondary.ChunkMeta) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: generating nonce: %w", err)
	}

	aad := chunkAAD(meta.FileID, meta.ChunkIndex, meta.TotalChunks)
	sealed := gcm.Seal(nonce, nonce, data, aad)

	frame := make([]byte, 4+len(sealed))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(sealed)))
	copy(frame[4:], sealed)
	return frame, nil
}

// DecryptFile parses the concatenated length-prefixed frames, verifies the
// recovered chunk count matches meta.TotalChunks, and opens each frame with the
// AAD reconstructed from its one-based position. Any reordering, truncation,
// relocation, or tampering fails GCM authentication.
func (a *FileEncryptionAdapter) DecryptFile(_ context.Context, data []byte, key []byte, meta secondary.FileMeta) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	frames, err := parseFrames(data)
	if err != nil {
		return nil, err
	}
	if len(frames) != meta.TotalChunks {
		return nil, fmt.Errorf("%w: got %d, want %d", ErrChunkCountMismatch, len(frames), meta.TotalChunks)
	}

	var plaintext []byte
	nonceSize := gcm.NonceSize()
	for i, sealed := range frames {
		if len(sealed) < nonceSize {
			return nil, ErrMalformedCiphertext
		}
		nonce, ciphertext := sealed[:nonceSize], sealed[nonceSize:]
		aad := chunkAAD(meta.FileID, i+1, meta.TotalChunks)
		chunk, err := gcm.Open(nil, nonce, ciphertext, aad)
		if err != nil {
			return nil, fmt.Errorf("crypto: authenticating chunk %d: %w", i+1, err)
		}
		plaintext = append(plaintext, chunk...)
	}
	return plaintext, nil
}

// newGCM builds an AES-256-GCM AEAD from key, validating its length.
func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != keyLength {
		return nil, ErrInvalidKeyLength
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: creating AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: creating GCM: %w", err)
	}
	return gcm, nil
}

// parseFrames splits the concatenated [uint32 length][payload] frames produced
// by EncryptChunk back into their individual sealed payloads.
func parseFrames(data []byte) ([][]byte, error) {
	frames := make([][]byte, 0)
	for offset := 0; offset < len(data); {
		if offset+4 > len(data) {
			return nil, fmt.Errorf("%w: dangling length prefix", ErrMalformedCiphertext)
		}
		size := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
		// Compare against the remaining length rather than offset+size: a uint32
		// above MaxInt32 casts to a negative int on 32-bit platforms, and the
		// additive form would both miss that and risk overflow. size <= 0 rejects
		// the negative-overflow case; len(data)-offset is always non-negative here.
		if size <= 0 || size > len(data)-offset {
			return nil, fmt.Errorf("%w: frame length %d exceeds remaining bytes", ErrMalformedCiphertext, size)
		}
		frames = append(frames, data[offset:offset+size])
		offset += size
	}
	return frames, nil
}

// chunkAAD builds the additional authenticated data bound to a chunk. The
// fixed-width index and total precede the variable-length file ID so the
// encoding is unambiguous, and both encrypt and decrypt sides construct it
// identically.
func chunkAAD(fileID string, chunkIndex, totalChunks int) []byte {
	aad := make([]byte, 8, 8+len(fileID))
	binary.BigEndian.PutUint32(aad[0:4], uint32(chunkIndex))
	binary.BigEndian.PutUint32(aad[4:8], uint32(totalChunks))
	aad = append(aad, fileID...)
	return aad
}
