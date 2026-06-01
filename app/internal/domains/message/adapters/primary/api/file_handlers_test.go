package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockFileService struct{ mock.Mock }

type partialErrorReadCloser struct {
	data []byte
	err  error
	read bool
}

func (r *partialErrorReadCloser) Read(p []byte) (int, error) {
	if !r.read {
		r.read = true
		n := copy(p, r.data)
		if n > 0 {
			return n, nil
		}
	}
	if r.err == nil {
		return 0, io.EOF
	}
	return 0, r.err
}

func (r *partialErrorReadCloser) Close() error { return nil }

func (m *MockFileService) InitiateUpload(ctx context.Context, req domain.InitiateUploadRequest) (*domain.InitiateUploadResponse, error) {
	args := m.Called(ctx, req)
	var response *domain.InitiateUploadResponse
	if value := args.Get(0); value != nil {
		response = value.(*domain.InitiateUploadResponse)
	}
	return response, args.Error(1)
}

func (m *MockFileService) UploadChunk(ctx context.Context, req domain.UploadChunkRequest) (*domain.UploadChunkResponse, error) {
	args := m.Called(ctx, req)
	var response *domain.UploadChunkResponse
	if value := args.Get(0); value != nil {
		response = value.(*domain.UploadChunkResponse)
	}
	return response, args.Error(1)
}

func (m *MockFileService) DownloadFile(ctx context.Context, req domain.DownloadFileRequest) (*domain.DownloadFileResponse, error) {
	args := m.Called(ctx, req)
	var response *domain.DownloadFileResponse
	if value := args.Get(0); value != nil {
		response = value.(*domain.DownloadFileResponse)
	}
	return response, args.Error(1)
}

func (m *MockFileService) AbortUpload(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func setupFileTestRouter(fileService *MockFileService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	handler := NewFileAPIHandler(fileService)
	router := gin.New()
	// Mirror the production custom recovery so tests exercise the same panic
	// propagation behaviour as the real server.
	router.Use(gin.CustomRecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, err any) {
		if err == http.ErrAbortHandler { //nolint:errorlint // sentinel value identity check, not error chain
			panic(err)
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	v1 := router.Group("/api/v1")
	files := v1.Group("/files")
	files.POST("/initiate", handler.InitiateUpload)
	files.POST("/:fileID/chunks", handler.UploadChunk)
	files.GET("/:fileID", handler.DownloadFile)

	return router
}

func TestInitiateUploadHandler_Success(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	expectedRequest := domain.InitiateUploadRequest{
		Filename:    "report.pdf",
		ContentType: "application/pdf",
		TotalSize:   10,
		ChunkSize:   4,
		MessageID:   "msg-1",
	}
	key32 := make([]byte, 32)
	for i := range key32 {
		key32[i] = byte(i + 1)
	}
	service.On("InitiateUpload", mock.Anything, expectedRequest).Return(&domain.InitiateUploadResponse{
		FileID:        "file-123",
		SessionID:     "session-456",
		EncryptionKey: key32,
	}, nil).Once()

	body, err := json.Marshal(map[string]any{
		"filename":    "report.pdf",
		"contentType": "application/pdf",
		"totalSize":   10,
		"chunkSize":   4,
		"messageID":   "msg-1",
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/files/initiate", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response struct {
		FileID     string `json:"fileID"`
		SessionID  string `json:"sessionID"`
		EncodedKey string `json:"encodedKey"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "file-123", response.FileID)
	assert.Equal(t, "session-456", response.SessionID)
	assert.Equal(t, base64.URLEncoding.EncodeToString(key32), response.EncodedKey)

	service.AssertExpectations(t)
}

func TestInitiateUploadHandler_MissingRequiredFields(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/files/initiate", bytes.NewBufferString(`{"contentType":"application/pdf"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.StandardErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeValidationFailed, response.Error)
	service.AssertNotCalled(t, "InitiateUpload", mock.Anything, mock.Anything)
}

func TestInitiateUploadHandler_ServiceError(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	service.On("InitiateUpload", mock.Anything, mock.Anything).Return((*domain.InitiateUploadResponse)(nil), errors.New("boom")).Once()

	body := `{"filename":"report.pdf","contentType":"application/pdf","totalSize":10,"chunkSize":4,"messageID":"msg-1"}`
	req, err := http.NewRequest(http.MethodPost, "/api/v1/files/initiate", bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	service.AssertExpectations(t)
}

func TestInitiateUploadHandler_FileTooLarge(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	service.On("InitiateUpload", mock.Anything, mock.Anything).Return((*domain.InitiateUploadResponse)(nil), domain.ErrFileTooLarge).Once()

	body := `{"filename":"video.mov","contentType":"video/quicktime","totalSize":999999999,"chunkSize":4194304,"messageID":"msg-1"}`
	req, err := http.NewRequest(http.MethodPost, "/api/v1/files/initiate", bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	service.AssertExpectations(t)
}

func TestUploadChunkHandler_Success(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	expected := domain.UploadChunkRequest{
		FileID:      "file-123",
		SessionID:   "session-456",
		ChunkIndex:  1,
		TotalChunks: 3,
		Data:        []byte("chunk-one"),
	}
	service.On("UploadChunk", mock.Anything, expected).Return(&domain.UploadChunkResponse{
		FileID:     "file-123",
		ChunkIndex: 1,
		Done:       false,
	}, nil).Once()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("sessionID", "session-456"))
	require.NoError(t, writer.WriteField("chunkIndex", "1"))
	require.NoError(t, writer.WriteField("totalChunks", "3"))
	fileWriter, err := writer.CreateFormFile("data", "chunk.bin")
	require.NoError(t, err)
	_, err = fileWriter.Write([]byte("chunk-one"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req, err := http.NewRequest(http.MethodPost, "/api/v1/files/file-123/chunks", &body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		FileID     string `json:"fileID"`
		ChunkIndex int    `json:"chunkIndex"`
		Done       bool   `json:"done"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "file-123", response.FileID)
	assert.Equal(t, 1, response.ChunkIndex)
	assert.False(t, response.Done)

	service.AssertExpectations(t)
}

func TestUploadChunkHandler_MissingRequiredFields(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("chunkIndex", "1"))
	require.NoError(t, writer.WriteField("totalChunks", "3"))
	require.NoError(t, writer.Close())

	req, err := http.NewRequest(http.MethodPost, "/api/v1/files/file-123/chunks", &body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	service.AssertNotCalled(t, "UploadChunk", mock.Anything, mock.Anything)
}

func TestUploadChunkHandler_ServiceError(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	service.On("UploadChunk", mock.Anything, mock.Anything).Return((*domain.UploadChunkResponse)(nil), errors.New("boom")).Once()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("sessionID", "session-456"))
	require.NoError(t, writer.WriteField("chunkIndex", "1"))
	require.NoError(t, writer.WriteField("totalChunks", "3"))
	fileWriter, err := writer.CreateFormFile("data", "chunk.bin")
	require.NoError(t, err)
	_, err = fileWriter.Write([]byte("chunk-one"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req, err := http.NewRequest(http.MethodPost, "/api/v1/files/file-123/chunks", &body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	service.AssertExpectations(t)
}

func TestUploadChunkHandler_InvalidSession(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	service.On("UploadChunk", mock.Anything, mock.Anything).Return((*domain.UploadChunkResponse)(nil), domain.ErrUploadSessionNotFound).Once()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("sessionID", "missing-session"))
	require.NoError(t, writer.WriteField("chunkIndex", "1"))
	require.NoError(t, writer.WriteField("totalChunks", "3"))
	fileWriter, err := writer.CreateFormFile("data", "chunk.bin")
	require.NoError(t, err)
	_, err = fileWriter.Write([]byte("chunk-one"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req, err := http.NewRequest(http.MethodPost, "/api/v1/files/file-123/chunks", &body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.StandardErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeUploadSessionNotFound, response.Error)

	service.AssertExpectations(t)
}

func TestUploadChunkHandler_ExpiredSession(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	service.On("UploadChunk", mock.Anything, mock.Anything).Return((*domain.UploadChunkResponse)(nil), domain.ErrUploadSessionExpired).Once()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("sessionID", "session-456"))
	require.NoError(t, writer.WriteField("chunkIndex", "1"))
	require.NoError(t, writer.WriteField("totalChunks", "3"))
	fileWriter, err := writer.CreateFormFile("data", "chunk.bin")
	require.NoError(t, err)
	_, err = fileWriter.Write([]byte("chunk-one"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req, err := http.NewRequest(http.MethodPost, "/api/v1/files/file-123/chunks", &body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusGone, w.Code)

	var response models.StandardErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeSessionExpired, response.Error)
	service.AssertExpectations(t)
}

func TestDownloadFileHandler_Success_ViaHeader(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	encodedKey := base64.URLEncoding.EncodeToString([]byte("download-key"))
	service.On("DownloadFile", mock.Anything, domain.DownloadFileRequest{
		FileID: "file-123",
		Key:    []byte("download-key"),
	}).Return(&domain.DownloadFileResponse{
		Filename:    "report.pdf",
		ContentType: "application/pdf",
		Data:        io.NopCloser(bytes.NewReader([]byte("decrypted-content"))),
	}, nil).Once()

	req, err := http.NewRequest(http.MethodGet, "/api/v1/files/file-123", nil)
	require.NoError(t, err)
	req.Header.Set("X-File-Key", encodedKey)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), `filename="report.pdf"`)
	assert.Equal(t, []byte("decrypted-content"), w.Body.Bytes())
	service.AssertExpectations(t)
}

func TestDownloadFileHandler_QueryParamRejected(t *testing.T) {
	// After removing the ?key= fallback, a request with only ?key= and no
	// X-File-Key header must return 400 — the query path is no longer supported.
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	encodedKey := base64.URLEncoding.EncodeToString([]byte("download-key"))
	req, err := http.NewRequest(http.MethodGet, "/api/v1/files/file-123?key="+encodedKey, nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	service.AssertNotCalled(t, "DownloadFile", mock.Anything, mock.Anything)
}

func TestDownloadFileHandler_MissingRequiredFields(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/files/file-123", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	service.AssertNotCalled(t, "DownloadFile", mock.Anything, mock.Anything)
}

func TestDownloadFileHandler_ServiceError(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	encodedKey := base64.URLEncoding.EncodeToString([]byte("download-key"))
	service.On("DownloadFile", mock.Anything, mock.Anything).Return((*domain.DownloadFileResponse)(nil), errors.New("boom")).Once()

	req, err := http.NewRequest(http.MethodGet, "/api/v1/files/file-123", nil)
	require.NoError(t, err)
	req.Header.Set("X-File-Key", encodedKey)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	service.AssertExpectations(t)
}

func TestDownloadFileHandler_StreamReadErrorAbortsConnection(t *testing.T) {
	service := new(MockFileService)
	router := setupFileTestRouter(service)

	encodedKey := base64.URLEncoding.EncodeToString([]byte("download-key"))
	service.On("DownloadFile", mock.Anything, mock.Anything).Return(&domain.DownloadFileResponse{
		Filename:    "report.pdf",
		ContentType: "application/pdf",
		Data: &partialErrorReadCloser{
			data: []byte("partial"),
			err:  errors.New("decrypt failed mid-stream"),
		},
	}, nil).Once()

	req, err := http.NewRequest(http.MethodGet, "/api/v1/files/file-123", nil)
	require.NoError(t, err)
	req.Header.Set("X-File-Key", encodedKey)

	w := httptest.NewRecorder()
	require.PanicsWithValue(t, http.ErrAbortHandler, func() {
		router.ServeHTTP(w, req)
	})
	service.AssertExpectations(t)
}
