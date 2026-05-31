package crypto

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubKeySource provides GenerateID/GenerateKey delegation and satisfies the
// shared EncryptionServicePort interface. Encrypt/Decrypt are unused by the
// local adapter and must never be called.
type stubKeySource struct {
	generateIDFn  func(ctx context.Context) (string, error)
	generateKeyFn func(ctx context.Context, length int32) ([]byte, error)
}

func (s *stubKeySource) GenerateKey(ctx context.Context, length int32) ([]byte, error) {
	return s.generateKeyFn(ctx, length)
}

func (s *stubKeySource) GenerateID(ctx context.Context) (string, error) {
	return s.generateIDFn(ctx)
}

func (s *stubKeySource) Encrypt(context.Context, []string, []byte) ([]string, error) {
	panic("Encrypt must not be called by the local file-encryption adapter")
}

func (s *stubKeySource) Decrypt(context.Context, []string, []byte) ([]string, error) {
	panic("Decrypt must not be called by the local file-encryption adapter")
}

func (s *stubKeySource) HealthCheck(context.Context) error { return nil }

func newTestKey() []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return key
}

// encryptAll runs each chunk through EncryptChunk and concatenates the frames
// the way object storage multipart completion would.
func encryptAll(t *testing.T, a *FileEncryptionAdapter, key []byte, fileID string, chunks [][]byte) []byte {
	t.Helper()
	total := len(chunks)
	var blob []byte
	for i, chunk := range chunks {
		meta := secondary.ChunkMeta{FileID: fileID, ChunkIndex: i + 1, TotalChunks: total}
		frame, err := a.EncryptChunk(context.Background(), chunk, key, meta)
		require.NoError(t, err)
		blob = append(blob, frame...)
	}
	return blob
}

func TestEncryptChunk_DecryptFile_RoundTrip(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()
	fileID := "file-abc"
	chunks := [][]byte{[]byte("hello "), []byte("brave "), []byte("world")}

	blob := encryptAll(t, a, key, fileID, chunks)

	plaintext, err := a.DecryptFile(context.Background(), blob, key, secondary.FileMeta{FileID: fileID, TotalChunks: len(chunks)})
	require.NoError(t, err)
	assert.Equal(t, []byte("hello brave world"), plaintext)
}

func TestDecryptFileStream_RoundTrip(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()
	fileID := "file-abc"
	chunks := [][]byte{[]byte("hello "), []byte("brave "), []byte("world")}
	blob := encryptAll(t, a, key, fileID, chunks)

	var out bytes.Buffer
	err := a.DecryptFileStream(context.Background(), bytes.NewReader(blob), &out, key, secondary.FileMeta{FileID: fileID, TotalChunks: len(chunks), MaxFrameSize: 1024})
	require.NoError(t, err)
	assert.Equal(t, []byte("hello brave world"), out.Bytes())
}

func TestEncryptChunk_NoDoubleEncoding(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()
	data := []byte("0123456789")
	meta := secondary.ChunkMeta{FileID: "f", ChunkIndex: 1, TotalChunks: 1}

	frame, err := a.EncryptChunk(context.Background(), data, key, meta)
	require.NoError(t, err)

	// Frame = uint32 length prefix + nonce(12) + ciphertext(len(data)) + tag(16).
	// Overhead is exactly 4+12+16 = 32 bytes regardless of content; no base64 bloat.
	declared := binary.BigEndian.Uint32(frame[:4])
	assert.Equal(t, len(frame)-4, int(declared))
	assert.Equal(t, len(data)+32, len(frame))
}

func TestDecryptFile_RejectsReorderedChunks(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()
	fileID := "file-abc"

	c1, err := a.EncryptChunk(context.Background(), []byte("AAAA"), key, secondary.ChunkMeta{FileID: fileID, ChunkIndex: 1, TotalChunks: 2})
	require.NoError(t, err)
	c2, err := a.EncryptChunk(context.Background(), []byte("BBBB"), key, secondary.ChunkMeta{FileID: fileID, ChunkIndex: 2, TotalChunks: 2})
	require.NoError(t, err)

	// Swap the two frames: each authenticates alone, but the bound index no
	// longer matches its position, so decrypt must fail.
	swapped := append(append([]byte{}, c2...), c1...)
	_, err = a.DecryptFile(context.Background(), swapped, key, secondary.FileMeta{FileID: fileID, TotalChunks: 2})
	assert.Error(t, err)
}

func TestDecryptFile_RejectsTruncation(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()
	fileID := "file-abc"
	chunks := [][]byte{[]byte("AAAA"), []byte("BBBB"), []byte("CCCC")}

	blob := encryptAll(t, a, key, fileID, chunks)
	// Drop the final frame by trimming to the first two frames' bytes.
	firstTwo := encryptAll(t, a, key, fileID, chunks[:2])
	truncated := blob[:len(firstTwo)]

	_, err := a.DecryptFile(context.Background(), truncated, key, secondary.FileMeta{FileID: fileID, TotalChunks: 3})
	assert.Error(t, err)
}

func TestDecryptFile_RejectsWrongFile(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()
	chunks := [][]byte{[]byte("secret")}

	blob := encryptAll(t, a, key, "file-one", chunks)

	// Same bytes, different FileID in the decrypt metadata -> AAD mismatch.
	_, err := a.DecryptFile(context.Background(), blob, key, secondary.FileMeta{FileID: "file-two", TotalChunks: 1})
	assert.Error(t, err)
}

func TestDecryptFile_RejectsTamperedCiphertext(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()
	fileID := "file-abc"

	blob := encryptAll(t, a, key, fileID, [][]byte{[]byte("hello world")})
	// Flip a byte inside the ciphertext (after the 4-byte length + 12-byte nonce).
	tampered := append([]byte{}, blob...)
	tampered[4+12] ^= 0xFF
	_, err := a.DecryptFile(context.Background(), tampered, key, secondary.FileMeta{FileID: fileID, TotalChunks: 1})
	assert.Error(t, err)
}

func TestEncryptChunk_RejectsBadKeyLength(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	_, err := a.EncryptChunk(context.Background(), []byte("data"), []byte("short"), secondary.ChunkMeta{FileID: "f", ChunkIndex: 1, TotalChunks: 1})
	assert.Error(t, err)
}

func TestDecryptFile_RejectsMalformedFraming(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()
	// A length prefix that claims more bytes than are present.
	bad := make([]byte, 4)
	binary.BigEndian.PutUint32(bad, 1000)
	_, err := a.DecryptFile(context.Background(), bad, key, secondary.FileMeta{FileID: "f", TotalChunks: 1})
	assert.Error(t, err)
}

func TestDecryptFile_RejectsOversizedFrameLength(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()

	// A length prefix declaring far more bytes than follow it must be rejected
	// without panicking, including on platforms where int(uint32) could go
	// negative.
	blob := make([]byte, 4+8)
	binary.BigEndian.PutUint32(blob[:4], 0xFFFFFFFF)
	_, err := a.DecryptFile(context.Background(), blob, key, secondary.FileMeta{FileID: "f", TotalChunks: 1})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrMalformedCiphertext)
}

func TestDecryptFileStream_RejectsFrameOverConfiguredMax(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()
	fileID := "file-abc"
	blob := encryptAll(t, a, key, fileID, [][]byte{bytes.Repeat([]byte("a"), 64)})

	var out bytes.Buffer
	err := a.DecryptFileStream(context.Background(), bytes.NewReader(blob), &out, key, secondary.FileMeta{FileID: fileID, TotalChunks: 1, MaxFrameSize: 16})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrMalformedCiphertext)
}

func TestDecryptFile_HonorsMaxFrameSize(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()
	fileID := "file-abc"
	blob := encryptAll(t, a, key, fileID, [][]byte{bytes.Repeat([]byte("a"), 64)})

	_, err := a.DecryptFile(context.Background(), blob, key, secondary.FileMeta{FileID: fileID, TotalChunks: 1, MaxFrameSize: 16})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrMalformedCiphertext)
}

func TestDecryptFile_RejectsZeroLengthFrame(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()

	blob := make([]byte, 4)
	binary.BigEndian.PutUint32(blob[:4], 0)
	_, err := a.DecryptFile(context.Background(), blob, key, secondary.FileMeta{FileID: "f", TotalChunks: 1})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrMalformedCiphertext)
}

func TestGenerateID_And_GenerateKey_Delegate(t *testing.T) {
	t.Parallel()
	expectedErr := errors.New("boom")
	source := &stubKeySource{
		generateIDFn: func(ctx context.Context) (string, error) { return "file-123", nil },
		generateKeyFn: func(ctx context.Context, length int32) ([]byte, error) {
			assert.Equal(t, int32(32), length)
			return nil, expectedErr
		},
	}
	a := NewFileEncryptionAdapter(source)

	id, err := a.GenerateID(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "file-123", id)

	key, err := a.GenerateKey(context.Background(), 32)
	assert.Nil(t, key)
	assert.ErrorIs(t, err, expectedErr)
}

func TestAdapterSatisfiesPort(t *testing.T) {
	t.Parallel()
	var _ secondary.FileEncryptionServicePort = NewFileEncryptionAdapter(nil)
}

// Guard: an empty blob with zero expected chunks decrypts to empty output.
func TestDecryptFile_EmptyWithZeroChunks(t *testing.T) {
	t.Parallel()
	a := NewFileEncryptionAdapter(nil)
	key := newTestKey()
	out, err := a.DecryptFile(context.Background(), []byte{}, key, secondary.FileMeta{FileID: "f", TotalChunks: 0})
	require.NoError(t, err)
	assert.True(t, bytes.Equal(out, []byte{}) || len(out) == 0)
}
