package domain

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockFileEncryptionService struct{ mock.Mock }

func (m *mockFileEncryptionService) GenerateID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *mockFileEncryptionService) GenerateKey(ctx context.Context, length int32) ([]byte, error) {
	args := m.Called(ctx, length)
	var key []byte
	if value := args.Get(0); value != nil {
		key = value.([]byte)
	}
	return key, args.Error(1)
}

func (m *mockFileEncryptionService) EncryptChunk(ctx context.Context, data []byte, key []byte, meta secondary.ChunkMeta) ([]byte, error) {
	args := m.Called(ctx, data, key, meta)
	var encrypted []byte
	if value := args.Get(0); value != nil {
		encrypted = value.([]byte)
	}
	return encrypted, args.Error(1)
}

func (m *mockFileEncryptionService) DecryptFile(ctx context.Context, data []byte, key []byte, meta secondary.FileMeta) ([]byte, error) {
	args := m.Called(ctx, data, key, meta)
	var decrypted []byte
	if value := args.Get(0); value != nil {
		decrypted = value.([]byte)
	}
	return decrypted, args.Error(1)
}

func (m *mockFileEncryptionService) DecryptFileStream(ctx context.Context, src io.Reader, dst io.Writer, key []byte, meta secondary.FileMeta) error {
	args := m.Called(ctx, src, dst, key, meta)
	return args.Error(0)
}

type mockObjectStoragePort struct{ mock.Mock }

func (m *mockObjectStoragePort) InitiateMultipartUpload(ctx context.Context, fileID, contentType string) (string, error) {
	args := m.Called(ctx, fileID, contentType)
	return args.String(0), args.Error(1)
}

func (m *mockObjectStoragePort) UploadPart(ctx context.Context, fileID, uploadID string, partNumber int, data []byte) (string, error) {
	args := m.Called(ctx, fileID, uploadID, partNumber, data)
	return args.String(0), args.Error(1)
}

func (m *mockObjectStoragePort) CompleteMultipartUpload(ctx context.Context, fileID, uploadID string, parts []contracts.CompletedPart) error {
	args := m.Called(ctx, fileID, uploadID, parts)
	return args.Error(0)
}

func (m *mockObjectStoragePort) AbortMultipartUpload(ctx context.Context, fileID, uploadID string) error {
	args := m.Called(ctx, fileID, uploadID)
	return args.Error(0)
}

func (m *mockObjectStoragePort) GetObject(ctx context.Context, fileID string) (io.ReadCloser, int64, error) {
	args := m.Called(ctx, fileID)
	var reader io.ReadCloser
	if value := args.Get(0); value != nil {
		reader = value.(io.ReadCloser)
	}
	return reader, args.Get(1).(int64), args.Error(2)
}

type mockUploadStatePort struct{ mock.Mock }

func (m *mockUploadStatePort) CreateSession(ctx context.Context, session contracts.FileUploadSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockUploadStatePort) GetSessionByID(ctx context.Context, sessionID string) (*contracts.FileUploadSession, error) {
	args := m.Called(ctx, sessionID)
	var session *contracts.FileUploadSession
	if value := args.Get(0); value != nil {
		session = value.(*contracts.FileUploadSession)
	}
	return session, args.Error(1)
}

func (m *mockUploadStatePort) GetSessionByFileID(ctx context.Context, fileID string) (*contracts.FileUploadSession, error) {
	args := m.Called(ctx, fileID)
	var session *contracts.FileUploadSession
	if value := args.Get(0); value != nil {
		session = value.(*contracts.FileUploadSession)
	}
	return session, args.Error(1)
}

func (m *mockUploadStatePort) AddCompletedPart(ctx context.Context, sessionID string, part contracts.CompletedPart) ([]contracts.CompletedPart, error) {
	args := m.Called(ctx, sessionID, part)
	var parts []contracts.CompletedPart
	if value := args.Get(0); value != nil {
		parts = value.([]contracts.CompletedPart)
	}
	return parts, args.Error(1)
}

func (m *mockUploadStatePort) CompleteSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *mockUploadStatePort) DeleteSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *mockUploadStatePort) DeleteExpiredSessions(ctx context.Context, asOf time.Time) ([]contracts.FileUploadSession, error) {
	args := m.Called(ctx, asOf)
	var sessions []contracts.FileUploadSession
	if value := args.Get(0); value != nil {
		sessions = value.([]contracts.FileUploadSession)
	}
	return sessions, args.Error(1)
}

func TestFileService_InitiateUpload_Success(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 1024)

	enc.On("GenerateID", mock.Anything).Return("file-123", nil).Once()
	enc.On("GenerateID", mock.Anything).Return("session-456", nil).Once()
	enc.On("GenerateKey", mock.Anything, int32(32)).Return([]byte("enc-key"), nil).Once()
	storage.On("InitiateMultipartUpload", mock.Anything, "file-123", "application/pdf").Return("upload-789", nil).Once()
	state.On(
		"CreateSession",
		mock.Anything,
		mock.MatchedBy(func(session contracts.FileUploadSession) bool {
			return session.SessionID == "session-456" &&
				session.FileID == "file-123" &&
				session.MessageID == "msg-1" &&
				session.UploadID == "upload-789" &&
				session.Filename == "report.pdf" &&
				session.ContentType == "application/pdf" &&
				session.TotalSize == 10 &&
				session.TotalChunks == 3 &&
				session.Status == "active" &&
				bytes.Equal(session.EncryptionKey, []byte("enc-key")) &&
				session.CreatedAt.Before(session.ExpiresAt)
		}),
	).Return(nil).Once()

	response, err := service.InitiateUpload(context.Background(), contracts.InitiateUploadRequest{
		Filename:    "report.pdf",
		ContentType: "application/pdf",
		TotalSize:   10,
		ChunkSize:   4,
		MessageID:   "msg-1",
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "file-123", response.FileID)
	assert.Equal(t, "session-456", response.SessionID)
	// EncryptionKey must be returned so the client can construct the download URL.
	assert.Equal(t, []byte("enc-key"), response.EncryptionKey)

	enc.AssertExpectations(t)
	storage.AssertExpectations(t)
	state.AssertExpectations(t)
}

func TestFileService_InitiateUpload_FileTooLarge(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 1024)

	response, err := service.InitiateUpload(context.Background(), contracts.InitiateUploadRequest{
		Filename:    "video.mov",
		ContentType: "video/quicktime",
		TotalSize:   2048,
		ChunkSize:   256,
		MessageID:   "msg-1",
	})

	assert.Nil(t, response)
	assert.ErrorIs(t, err, ErrFileTooLarge)
	enc.AssertNotCalled(t, "GenerateID", mock.Anything)
	storage.AssertNotCalled(t, "InitiateMultipartUpload", mock.Anything, mock.Anything, mock.Anything)
	state.AssertNotCalled(t, "CreateSession", mock.Anything, mock.Anything)
}

func TestFileService_UploadChunk_Success(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 4096)

	session := &contracts.FileUploadSession{
		SessionID:     "session-123",
		FileID:        "file-123",
		UploadID:      "upload-123",
		TotalChunks:   2,
		Status:        "active",
		EncryptionKey: []byte("enc-key"),
		ExpiresAt:     time.Now().Add(time.Hour),
	}

	state.On("GetSessionByID", mock.Anything, "session-123").Return(session, nil).Once()
	enc.On("EncryptChunk", mock.Anything, []byte("chunk-one"), []byte("enc-key"), secondary.ChunkMeta{FileID: "file-123", ChunkIndex: 1, TotalChunks: 2}).Return([]byte("encrypted-chunk"), nil).Once()
	storage.On("UploadPart", mock.Anything, "file-123", "upload-123", 1, []byte("encrypted-chunk")).Return("etag-1", nil).Once()
	state.On("AddCompletedPart", mock.Anything, "session-123", contracts.CompletedPart{PartNumber: 1, ETag: "etag-1"}).Return([]contracts.CompletedPart{{PartNumber: 1, ETag: "etag-1"}}, nil).Once()

	response, err := service.UploadChunk(context.Background(), contracts.UploadChunkRequest{
		FileID:      "file-123",
		SessionID:   "session-123",
		ChunkIndex:  1,
		TotalChunks: 2,
		Data:        []byte("chunk-one"),
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "file-123", response.FileID)
	assert.Equal(t, 1, response.ChunkIndex)
	assert.False(t, response.Done)

	enc.AssertExpectations(t)
	storage.AssertExpectations(t)
	state.AssertExpectations(t)
}

func TestFileService_UploadChunk_FinalChunkCompletesUpload(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 4096)

	session := &contracts.FileUploadSession{
		SessionID:   "session-123",
		FileID:      "file-123",
		UploadID:    "upload-123",
		TotalChunks: 2,
		Status:      "active",
		Filename:    "report.pdf",
		ContentType: "application/pdf",
		CompletedParts: []contracts.CompletedPart{{
			PartNumber: 1,
			ETag:       "etag-1",
		}},
		EncryptionKey: []byte("enc-key"),
		ExpiresAt:     time.Now().Add(time.Hour),
	}

	allParts := []contracts.CompletedPart{
		{PartNumber: 1, ETag: "etag-1"},
		{PartNumber: 2, ETag: "etag-2"},
	}

	state.On("GetSessionByID", mock.Anything, "session-123").Return(session, nil).Once()
	enc.On("EncryptChunk", mock.Anything, []byte("chunk-two"), []byte("enc-key"), secondary.ChunkMeta{FileID: "file-123", ChunkIndex: 2, TotalChunks: 2}).Return([]byte("encrypted-two"), nil).Once()
	storage.On("UploadPart", mock.Anything, "file-123", "upload-123", 2, []byte("encrypted-two")).Return("etag-2", nil).Once()
	state.On("AddCompletedPart", mock.Anything, "session-123", contracts.CompletedPart{PartNumber: 2, ETag: "etag-2"}).Return(allParts, nil).Once()
	storage.On(
		"CompleteMultipartUpload",
		mock.Anything,
		"file-123",
		"upload-123",
		[]contracts.CompletedPart{{PartNumber: 1, ETag: "etag-1"}, {PartNumber: 2, ETag: "etag-2"}},
	).Return(nil).Once()
	state.On("CompleteSession", mock.Anything, "session-123").Return(nil).Once()

	response, err := service.UploadChunk(context.Background(), contracts.UploadChunkRequest{
		FileID:      "file-123",
		SessionID:   "session-123",
		ChunkIndex:  2,
		TotalChunks: 2,
		Data:        []byte("chunk-two"),
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "file-123", response.FileID)
	assert.Equal(t, 2, response.ChunkIndex)
	assert.True(t, response.Done)

	enc.AssertExpectations(t)
	storage.AssertExpectations(t)
	state.AssertExpectations(t)
}

func TestFileService_UploadChunk_InvalidSession(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 4096)

	state.On("GetSessionByID", mock.Anything, "missing-session").Return((*contracts.FileUploadSession)(nil), ErrUploadSessionNotFound).Once()

	response, err := service.UploadChunk(context.Background(), contracts.UploadChunkRequest{
		FileID:      "file-123",
		SessionID:   "missing-session",
		ChunkIndex:  1,
		TotalChunks: 2,
		Data:        []byte("chunk-one"),
	})

	assert.Nil(t, response)
	assert.ErrorIs(t, err, ErrUploadSessionNotFound)
	enc.AssertNotCalled(t, "EncryptChunk", mock.Anything, mock.Anything, mock.Anything)
	storage.AssertNotCalled(t, "UploadPart", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestFileService_UploadChunk_ExpiredSession(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 4096)

	session := &contracts.FileUploadSession{
		SessionID:     "session-123",
		FileID:        "file-123",
		UploadID:      "upload-123",
		TotalChunks:   2,
		Status:        "active",
		EncryptionKey: []byte("enc-key"),
		ExpiresAt:     time.Now().Add(-time.Minute),
	}

	state.On("GetSessionByID", mock.Anything, "session-123").Return(session, nil).Once()

	response, err := service.UploadChunk(context.Background(), contracts.UploadChunkRequest{
		FileID:      "file-123",
		SessionID:   "session-123",
		ChunkIndex:  1,
		TotalChunks: 2,
		Data:        []byte("chunk-one"),
	})

	assert.Nil(t, response)
	assert.ErrorIs(t, err, ErrUploadSessionExpired)
	enc.AssertNotCalled(t, "EncryptChunk", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	storage.AssertNotCalled(t, "UploadPart", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestFileService_UploadChunk_NotYetExpiredSucceeds(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 4096)

	session := &contracts.FileUploadSession{
		SessionID:     "session-123",
		FileID:        "file-123",
		UploadID:      "upload-123",
		TotalChunks:   2,
		Status:        "active",
		EncryptionKey: []byte("enc-key"),
		ExpiresAt:     time.Now().Add(time.Hour),
	}

	state.On("GetSessionByID", mock.Anything, "session-123").Return(session, nil).Once()
	enc.On("EncryptChunk", mock.Anything, []byte("chunk-one"), []byte("enc-key"), secondary.ChunkMeta{FileID: "file-123", ChunkIndex: 1, TotalChunks: 2}).Return([]byte("encrypted-chunk"), nil).Once()
	storage.On("UploadPart", mock.Anything, "file-123", "upload-123", 1, []byte("encrypted-chunk")).Return("etag-1", nil).Once()
	state.On("AddCompletedPart", mock.Anything, "session-123", contracts.CompletedPart{PartNumber: 1, ETag: "etag-1"}).Return([]contracts.CompletedPart{{PartNumber: 1, ETag: "etag-1"}}, nil).Once()

	response, err := service.UploadChunk(context.Background(), contracts.UploadChunkRequest{
		FileID:      "file-123",
		SessionID:   "session-123",
		ChunkIndex:  1,
		TotalChunks: 2,
		Data:        []byte("chunk-one"),
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.False(t, response.Done)

	enc.AssertExpectations(t)
	storage.AssertExpectations(t)
	state.AssertExpectations(t)
}

func TestFileService_CleanupExpiredSessions_AbortsMultipartUploads(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 4096)

	removed := []contracts.FileUploadSession{
		{SessionID: "session-1", FileID: "file-1", UploadID: "upload-1"},
		{SessionID: "session-2", FileID: "file-2", UploadID: "upload-2"},
	}

	state.On("DeleteExpiredSessions", mock.Anything, mock.Anything).Return(removed, nil).Once()
	storage.On("AbortMultipartUpload", mock.Anything, "file-1", "upload-1").Return(nil).Once()
	storage.On("AbortMultipartUpload", mock.Anything, "file-2", "upload-2").Return(nil).Once()

	err := service.CleanupExpiredSessions(context.Background())

	require.NoError(t, err)
	state.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestFileService_CleanupExpiredSessions_ToleratesAbortError(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 4096)

	removed := []contracts.FileUploadSession{
		{SessionID: "session-1", FileID: "file-1", UploadID: "upload-1"},
	}

	state.On("DeleteExpiredSessions", mock.Anything, mock.Anything).Return(removed, nil).Once()
	storage.On("AbortMultipartUpload", mock.Anything, "file-1", "upload-1").Return(errors.New("abort failed")).Once()

	err := service.CleanupExpiredSessions(context.Background())

	require.NoError(t, err)
	state.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestFileService_CleanupExpiredSessions_PropagatesDeleteError(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 4096)

	deleteErr := errors.New("state unavailable")
	state.On("DeleteExpiredSessions", mock.Anything, mock.Anything).Return(([]contracts.FileUploadSession)(nil), deleteErr).Once()

	err := service.CleanupExpiredSessions(context.Background())

	assert.ErrorIs(t, err, deleteErr)
	storage.AssertNotCalled(t, "AbortMultipartUpload", mock.Anything, mock.Anything, mock.Anything)
	state.AssertExpectations(t)
}

func TestFileService_DownloadFile_Success(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 4096)

	session := &contracts.FileUploadSession{
		SessionID:   "file-123",
		FileID:      "file-123",
		Filename:    "report.pdf",
		ContentType: "application/pdf",
		TotalChunks: 3,
	}

	state.On("GetSessionByFileID", mock.Anything, "file-123").Return(session, nil).Once()
	storage.On("GetObject", mock.Anything, "file-123").Return(io.NopCloser(bytes.NewReader([]byte("encrypted-file"))), int64(len("encrypted-file")), nil).Once()
	enc.On("DecryptFileStream", mock.Anything, mock.Anything, mock.Anything, []byte("download-key"), secondary.FileMeta{FileID: "file-123", TotalChunks: 3, MaxFrameSize: 4096 + 28}).Run(func(args mock.Arguments) {
		dst := args.Get(2).(io.Writer)
		_, _ = dst.Write([]byte("decrypted-content"))
	}).Return(nil).Once()

	response, err := service.DownloadFile(context.Background(), contracts.DownloadFileRequest{
		FileID: "file-123",
		Key:    []byte("download-key"),
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "report.pdf", response.Filename)
	assert.Equal(t, "application/pdf", response.ContentType)
	defer func() {
		_ = response.Data.Close()
	}()
	decryptedData, err := io.ReadAll(response.Data)
	require.NoError(t, err)
	assert.Equal(t, []byte("decrypted-content"), decryptedData)

	enc.AssertExpectations(t)
	storage.AssertExpectations(t)
	state.AssertExpectations(t)
}

func TestFileService_AbortUpload_Success(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 4096)

	session := &contracts.FileUploadSession{
		SessionID: "session-123",
		FileID:    "file-123",
		UploadID:  "upload-123",
		Status:    "active",
	}

	state.On("GetSessionByID", mock.Anything, "session-123").Return(session, nil).Once()
	storage.On("AbortMultipartUpload", mock.Anything, "file-123", "upload-123").Return(nil).Once()
	state.On("DeleteSession", mock.Anything, "session-123").Return(nil).Once()

	err := service.AbortUpload(context.Background(), "session-123")

	assert.NoError(t, err)
	storage.AssertExpectations(t)
	state.AssertExpectations(t)
}

func TestFileService_AbortUpload_PropagatesAbortFailure(t *testing.T) {
	enc := new(mockFileEncryptionService)
	storage := new(mockObjectStoragePort)
	state := new(mockUploadStatePort)
	service := NewFileService(enc, storage, state, logtest.NewNoop(), 4096)

	session := &contracts.FileUploadSession{
		SessionID: "session-123",
		FileID:    "file-123",
		UploadID:  "upload-123",
	}
	abortErr := errors.New("abort failed")

	state.On("GetSessionByID", mock.Anything, "session-123").Return(session, nil).Once()
	storage.On("AbortMultipartUpload", mock.Anything, "file-123", "upload-123").Return(abortErr).Once()

	err := service.AbortUpload(context.Background(), "session-123")

	assert.ErrorIs(t, err, abortErr)
	state.AssertNotCalled(t, "DeleteSession", mock.Anything, mock.Anything)
}

func TestAddEncryptedPayloadOverhead(t *testing.T) {
	t.Parallel()

	assert.Equal(t, int64(128), addEncryptedPayloadOverhead(100))
	assert.Equal(t, int64(math.MaxInt64), addEncryptedPayloadOverhead(math.MaxInt64))
	assert.Equal(t, int64(math.MaxInt64), addEncryptedPayloadOverhead(math.MaxInt64-secondary.EncryptedFramePayloadOverheadBytes))
}

func TestAddEncryptedStoredOverhead(t *testing.T) {
	t.Parallel()

	assert.Equal(t, int64(100)+2*secondary.EncryptedFrameStoredOverheadBytes, addEncryptedStoredOverhead(100, 2))
	assert.Equal(t, int64(math.MaxInt64), addEncryptedStoredOverhead(math.MaxInt64, 1))
	assert.Equal(t, int64(math.MaxInt64), addEncryptedStoredOverhead(math.MaxInt64-secondary.EncryptedFrameStoredOverheadBytes+1, 1))
	assert.Equal(t, int64(math.MaxInt64), addEncryptedStoredOverhead(0, int(math.MaxInt64/secondary.EncryptedFrameStoredOverheadBytes+1)))
}
