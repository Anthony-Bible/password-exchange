package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewServerWithFileHandler_RegistersFileRoutes(t *testing.T) {
	messageService := new(MockMessageService)
	fileService := new(MockFileService)

	handler := NewFileAPIHandler(fileService)
	server := NewServerWithFileHandler(messageService, &stubEncryptionPort{}, &stubStoragePort{}, handler)

	fileService.On("InitiateUpload", mock.Anything, domain.InitiateUploadRequest{
		Filename:    "report.pdf",
		ContentType: "application/pdf",
		TotalSize:   10,
		ChunkSize:   4,
		MessageID:   "msg-1",
	}).Return(&domain.InitiateUploadResponse{FileID: "file-123", SessionID: "session-456"}, nil).Once()

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
	server.GetRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	fileService.AssertExpectations(t)
}

func TestNewServer_WithoutFileHandler_DoesNotRegisterFileRoutes(t *testing.T) {
	messageService := new(MockMessageService)
	server := NewServer(messageService, &stubEncryptionPort{}, &stubStoragePort{})

	req, err := http.NewRequest(http.MethodPost, "/api/v1/files", bytes.NewReader([]byte(`{"filename":"report.pdf"}`)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
