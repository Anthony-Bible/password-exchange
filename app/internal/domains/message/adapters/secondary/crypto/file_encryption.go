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
	"bufio"
	"bytes"
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

const (
	// frameLengthPrefixSize is the uint32 length-prefix size as an untyped
	// constant so it can be used as an array size in [frameLengthPrefixSize]byte.
	frameLengthPrefixSize        = 4
	defaultMaxPlaintextFrameSize = 100 * 1024 * 1024
	defaultMaxFrameSize          = defaultMaxPlaintextFrameSize + secondary.EncryptedFramePayloadOverheadBytes
)

var (
	// ErrInvalidKeyLength indicates the supplied key is not 32 bytes.
	ErrInvalidKeyLength = errors.New("crypto: encryption key must be 32 bytes")
	// ErrMalformedCiphertext aliases the shared secondary sentinel so callers in
	// both the domain and adapter layers can use errors.Is against one value.
	ErrMalformedCiphertext = secondary.ErrMalformedCiphertext
	// ErrChunkCountMismatch aliases the shared secondary sentinel.
	ErrChunkCountMismatch = secondary.ErrChunkCountMismatch
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
func (a *FileEncryptionAdapter) DecryptFile(ctx context.Context, data []byte, key []byte, meta secondary.FileMeta) ([]byte, error) {
	var plaintext bytes.Buffer
	if err := a.DecryptFileStream(ctx, bytes.NewReader(data), &plaintext, key, meta); err != nil {
		return nil, err
	}
	return plaintext.Bytes(), nil
}

// DecryptFileStream incrementally parses framed encrypted chunks from src,
// authenticates each chunk with position-bound AAD, and writes plaintext to dst.
func (a *FileEncryptionAdapter) DecryptFileStream(ctx context.Context, src io.Reader, dst io.Writer, key []byte, meta secondary.FileMeta) error {
	gcm, err := newGCM(key)
	if err != nil {
		return err
	}

	maxFrameSize := meta.MaxFrameSize
	if maxFrameSize <= 0 {
		maxFrameSize = defaultMaxFrameSize
	}

	buffered := bufio.NewReader(src)
	nonceSize := gcm.NonceSize()
	chunkCount := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		var lengthPrefix [frameLengthPrefixSize]byte
		n, err := io.ReadFull(buffered, lengthPrefix[:])
		if err != nil {
			if errors.Is(err, io.EOF) && n == 0 {
				break
			}
			// ErrUnexpectedEOF means the stored data is truncated — a format error.
			// Any other error (network failure, storage error) is passed through
			// so callers can distinguish storage failures from decryption failures.
			if errors.Is(err, io.ErrUnexpectedEOF) {
				return fmt.Errorf("%w: dangling length prefix", ErrMalformedCiphertext)
			}
			return fmt.Errorf("%w: reading frame length prefix: %w", secondary.ErrCiphertextReadFailed, err)
		}

		sizeU32 := binary.BigEndian.Uint32(lengthPrefix[:])
		size := int64(sizeU32)
		if sizeU32 == 0 || size > maxFrameSize || size > int64(int(^uint(0)>>1)) {
			return fmt.Errorf("%w: frame length %d is invalid or exceeds maximum %d", ErrMalformedCiphertext, size, maxFrameSize)
		}

		chunkCount++
		if chunkCount > meta.TotalChunks {
			return fmt.Errorf("%w: got at least %d, want %d", ErrChunkCountMismatch, chunkCount, meta.TotalChunks)
		}

		sealed := make([]byte, int(size))
		if _, err := io.ReadFull(buffered, sealed); err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				return fmt.Errorf("%w: frame body truncated at declared length %d", ErrMalformedCiphertext, size)
			}
			return fmt.Errorf("%w: reading frame body: %w", secondary.ErrCiphertextReadFailed, err)
		}
		if len(sealed) < nonceSize+gcm.Overhead() {
			return fmt.Errorf(
				"%w: frame body length %d shorter than nonce+tag minimum %d",
				ErrMalformedCiphertext,
				len(sealed),
				nonceSize+gcm.Overhead(),
			)
		}

		nonce, ciphertext := sealed[:nonceSize], sealed[nonceSize:]
		aad := chunkAAD(meta.FileID, chunkCount, meta.TotalChunks)
		plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
		if err != nil {
			return fmt.Errorf("%w: authenticating chunk %d: %v", ErrMalformedCiphertext, chunkCount, err)
		}
		n, err = dst.Write(plaintext)
		if err != nil {
			return fmt.Errorf("crypto: writing decrypted chunk %d: %w", chunkCount, err)
		}
		if n != len(plaintext) {
			return fmt.Errorf("crypto: writing decrypted chunk %d: %w", chunkCount, io.ErrShortWrite)
		}
	}

	if chunkCount != meta.TotalChunks {
		return fmt.Errorf("%w: got %d, want %d", ErrChunkCountMismatch, chunkCount, meta.TotalChunks)
	}
	return nil
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
