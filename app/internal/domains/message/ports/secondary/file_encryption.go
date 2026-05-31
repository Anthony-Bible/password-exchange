package secondary

import "context"

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
}
