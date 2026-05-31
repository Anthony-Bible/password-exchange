package api

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/middleware"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging"
	"github.com/gin-gonic/gin"
)

type fileServicePort interface {
	InitiateUpload(ctx context.Context, req domain.InitiateUploadRequest) (*domain.InitiateUploadResponse, error)
	UploadChunk(ctx context.Context, req domain.UploadChunkRequest) (*domain.UploadChunkResponse, error)
	DownloadFile(ctx context.Context, req domain.DownloadFileRequest) (*domain.DownloadFileResponse, error)
	AbortUpload(ctx context.Context, sessionID string) error
}

// FileAPIHandler handles file upload/download endpoints.
type FileAPIHandler struct {
	fileService fileServicePort
}

// NewFileAPIHandler creates a new file API handler.
func NewFileAPIHandler(fileService fileServicePort) *FileAPIHandler {
	return &FileAPIHandler{fileService: fileService}
}

type initiateUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	TotalSize   int64  `json:"totalSize"`
	ChunkSize   int64  `json:"chunkSize"`
	MessageID   string `json:"messageID"`
}

// InitiateUpload handles POST /api/v1/files/initiate.
func (h *FileAPIHandler) InitiateUpload(c *gin.Context) {
	var req initiateUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, "Invalid request format", map[string]interface{}{"parse_error": err.Error()})
		return
	}

	if req.Filename == "" || req.MessageID == "" || req.TotalSize <= 0 || req.ChunkSize <= 0 {
		writeValidationError(c, "Request validation failed", map[string]interface{}{
			"filename":  req.Filename,
			"messageID": req.MessageID,
			"totalSize": req.TotalSize,
			"chunkSize": req.ChunkSize,
		})
		return
	}

	response, err := h.fileService.InitiateUpload(c.Request.Context(), domain.InitiateUploadRequest{
		Filename:    req.Filename,
		ContentType: req.ContentType,
		TotalSize:   req.TotalSize,
		ChunkSize:   req.ChunkSize,
		MessageID:   req.MessageID,
	})
	if err != nil {
		writeFileServiceError(c, err, "Failed to initiate upload")
		return
	}

	encodedKey := base64.URLEncoding.EncodeToString(response.EncryptionKey)
	c.JSON(http.StatusCreated, gin.H{"fileID": response.FileID, "sessionID": response.SessionID, "encodedKey": encodedKey})
}

// UploadChunk handles POST /api/v1/files/:fileID/chunks.
func (h *FileAPIHandler) UploadChunk(c *gin.Context) {
	sessionID := c.PostForm("sessionID")
	chunkIndex, ciErr := strconv.Atoi(c.PostForm("chunkIndex"))
	totalChunks, tcErr := strconv.Atoi(c.PostForm("totalChunks"))
	fileHeader, dataErr := c.FormFile("data")
	if sessionID == "" || ciErr != nil || chunkIndex <= 0 || chunkIndex > math.MaxUint32 ||
		tcErr != nil || totalChunks <= 0 || totalChunks > math.MaxUint32 || dataErr != nil {
		writeValidationError(c, "Request validation failed", map[string]interface{}{
			"sessionID":   sessionID,
			"chunkIndex":  c.PostForm("chunkIndex"),
			"totalChunks": c.PostForm("totalChunks"),
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		writeValidationError(c, "Invalid file upload", nil)
		return
	}
	defer func() {
		_ = file.Close()
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		middleware.JSONErrorResponse(c, http.StatusInternalServerError, models.ErrorCodeInternalError, "Failed to read chunk", nil)
		return
	}

	response, err := h.fileService.UploadChunk(c.Request.Context(), domain.UploadChunkRequest{
		FileID:      c.Param("fileID"),
		SessionID:   sessionID,
		ChunkIndex:  chunkIndex,
		TotalChunks: totalChunks,
		Data:        data,
	})
	if err != nil {
		writeFileServiceError(c, err, "Failed to upload chunk")
		return
	}

	c.JSON(http.StatusOK, gin.H{"fileID": response.FileID, "chunkIndex": response.ChunkIndex, "done": response.Done})
}

// fileKeyHeader is the HTTP header used to supply the base64url-encoded AES-256
// file decryption key. Using a request header keeps the key out of server
// access logs, proxy logs, browser history, and Referer headers.
const fileKeyHeader = "X-File-Key"

// DownloadFile handles GET /api/v1/files/:fileID.
//
// The decryption key MUST be supplied via the X-File-Key request header.
// The ?key= query parameter is no longer accepted — it was removed because
// query strings appear in server access logs and proxy logs. Clients must use
// a fetch() call with the header set; direct browser navigation to this
// endpoint will fail (browsers cannot attach custom headers to navigations).
func (h *FileAPIHandler) DownloadFile(c *gin.Context) {
	encodedKey := c.GetHeader(fileKeyHeader)
	if encodedKey == "" {
		writeValidationError(c, "Missing key: supply via X-File-Key request header", nil)
		return
	}

	key, err := base64.URLEncoding.DecodeString(encodedKey)
	if err != nil {
		writeValidationError(c, "Invalid key", nil)
		return
	}

	response, err := h.fileService.DownloadFile(c.Request.Context(), domain.DownloadFileRequest{FileID: c.Param("fileID"), Key: key})
	if err != nil {
		writeFileServiceError(c, err, "Failed to download file")
		return
	}

	filename := sanitizeDownloadFilename(response.Filename)
	c.Header("Content-Type", response.ContentType)
	c.Header("Content-Disposition", `attachment; filename=`+strconv.Quote(filename))
	c.Data(http.StatusOK, response.ContentType, response.Data)
}

func sanitizeDownloadFilename(filename string) string {
	filename = strings.ReplaceAll(filename, "\\", "/")
	filename = filepath.Base(filename)
	filename = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r < 32 || r == 127 {
			return -1
		}
		return r
	}, filename)
	filename = strings.TrimSpace(filename)
	if filename == "" || filename == "." || filename == ".." {
		return "download"
	}
	return filename
}

func writeValidationError(c *gin.Context, message string, details map[string]interface{}) {
	middleware.JSONErrorResponse(c, http.StatusBadRequest, models.ErrorCodeValidationFailed, message, details)
}

func writeFileServiceError(c *gin.Context, err error, fallbackMessage string) {
	status := http.StatusInternalServerError
	code := models.ErrorCodeInternalError
	message := fallbackMessage

	switch {
	case errors.Is(err, domain.ErrFileTooLarge), errors.Is(err, domain.ErrInvalidChunkIndex), errors.Is(err, domain.ErrInvalidUploadRequest), errors.Is(err, domain.ErrUploadAlreadyComplete):
		status = http.StatusBadRequest
		code = models.ErrorCodeValidationFailed
		message = err.Error()
	case errors.Is(err, domain.ErrUploadSessionNotFound):
		status = http.StatusNotFound
		code = models.ErrorCodeUploadSessionNotFound
		message = err.Error()
	case errors.Is(err, domain.ErrUploadSessionExpired):
		status = http.StatusGone
		code = models.ErrorCodeSessionExpired
		message = err.Error()
	}

	if status == http.StatusInternalServerError {
		correlationID, _ := c.Get(middleware.CorrelationIDKey)
		logging.Error().
			Err(err).
			Interface("correlation_id", correlationID).
			Str("path", c.Request.URL.Path).
			Msg("file service internal error")
	}

	middleware.JSONErrorResponse(c, status, code, message, nil)
}
