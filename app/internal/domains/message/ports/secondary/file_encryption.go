package secondary

import (
	"context"
	"errors"
	"io"
)

// Wire-format constants for the length-prefixed encrypted frame layout:
//
//	[uint32 big-endian length (4 bytes)][12-byte nonce][ciphertext + 16-byte GCM tag]
const (
	// FrameLengthPrefixBytes is the size of the uint32 big-endian length prefix.
	FrameLengthPrefixBytes = int64(4)
	// EncryptedFramePayloadOverheadBytes is the per-frame overhead from the
	// AES-256-GCM nonce (12 bytes) and authentication tag (16 bytes).
	EncryptedFramePayloadOverheadBytes = int64(28)
	// EncryptedFrameStoredOverheadBytes is the total per-frame storage overhead
	// including the length prefix.
	EncryptedFrameStoredOverheadBytes = FrameLengthPrefixBytes + EncryptedFramePayloadOverheadBytes
)

var (
	// ErrMalformedCiphertext indicates the stored framing could not be parsed.
	ErrMalformedCiphertext = errors.New("crypto: malformed encrypted chunk framing")
	// ErrChunkCountMismatch indicates the recovered chunk count differs from the
	// expected total, signalling truncated or padded ciphertext.
	ErrChunkCountMismatch = errors.New("crypto: encrypted chunk count does not match expected total")
	// ErrCiphertextReadFailed indicates an underlying I/O error occurred while
	// reading the ciphertext stream (e.g., object storage/network failure).
	ErrCiphertextReadFailed = errors.New("crypto: failed to read ciphertext stream")
)

// ChunkMeta carries the per-chunk context bound into the authenticated
// encryption of a single uploaded chunk. Binding these values as AES-GCM
// additional authenticated data (AAD) ensures a stored chunk cannot be
// silently reordered or relocated to a different file without failing
// authentication on decrypt.
type ChunkMeta struct {
	// FileID identifies the file the chunk belongs to.
	FileID string
	// ChunkIndex is the one-based position of the chunk within the file.
	ChunkIndex int
	// TotalChunks is the total number of chunks expected for the file.
	TotalChunks int
}

// FileMeta carries the whole-file context required to authenticate every
// chunk during decryption. TotalChunks lets the decrypter detect truncated
// or padded ciphertext, while FileID binds the content to its file.
type FileMeta struct {
	// FileID identifies the file being decrypted.
	FileID string
	// TotalChunks is the number of chunks the file is expected to contain.
	TotalChunks int
	// MaxFrameSize is the maximum allowed encrypted frame payload length in bytes
	// (nonce + ciphertext + tag), used to bound per-frame allocations while
	// streaming decryption.
	MaxFrameSize int64
}

// FileEncryptionServicePort defines the secondary port used by the file service
// to generate identifiers, derive encryption keys, encrypt uploaded chunks, and
// decrypt fully assembled files.
type FileEncryptionServicePort interface {
	// GenerateID creates a unique identifier for a file or upload session.
	GenerateID(ctx context.Context) (string, error)

	// GenerateKey creates a symmetric encryption key with the requested length.
	GenerateKey(ctx context.Context, length int32) ([]byte, error)

	// EncryptChunk encrypts a single uploaded chunk with the provided key,
	// binding meta as authenticated data so the chunk's position and owning
	// file are cryptographically verified on decrypt.
	EncryptChunk(ctx context.Context, data []byte, key []byte, meta ChunkMeta) ([]byte, error)

	// DecryptFile decrypts the stored encrypted file contents with the provided
	// key. It verifies that the recovered chunk count and each chunk's bound
	// metadata match meta, rejecting reordered, truncated, or tampered content.
	DecryptFile(ctx context.Context, data []byte, key []byte, meta FileMeta) ([]byte, error)
	// DecryptFileStream decrypts the framed ciphertext from src and streams the
	// recovered plaintext to dst, enforcing the same integrity checks as
	// DecryptFile while keeping memory bounded to one frame.
	DecryptFileStream(ctx context.Context, src io.Reader, dst io.Writer, key []byte, meta FileMeta) error
}
