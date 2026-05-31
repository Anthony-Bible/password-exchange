package domain

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
)

const fileUploadSessionTTL = 24 * time.Hour

// FileService provides chunked encrypted file upload operations.
type FileService struct {
	encryptionService secondary.FileEncryptionServicePort
	objectStorage     secondary.ObjectStoragePort
	uploadState       secondary.UploadStatePort
	logger            secondary.LoggerPort
	maxFileSize       int64
}

// NewFileService creates a new file service.
func NewFileService(
	encryptionService secondary.FileEncryptionServicePort,
	objectStorage secondary.ObjectStoragePort,
	uploadState secondary.UploadStatePort,
	logger secondary.LoggerPort,
	maxFileSize int64,
) *FileService {
	return &FileService{
		encryptionService: encryptionService,
		objectStorage:     objectStorage,
		uploadState:       uploadState,
		logger:            logger,
		maxFileSize:       maxFileSize,
	}
}

// InitiateUpload starts a new multipart upload session.
func (s *FileService) InitiateUpload(ctx context.Context, req InitiateUploadRequest) (*InitiateUploadResponse, error) {
	if req.TotalSize <= 0 || req.ChunkSize <= 0 {
		return nil, ErrInvalidUploadRequest
	}
	if req.TotalSize > s.maxFileSize {
		return nil, ErrFileTooLarge
	}

	fileID, err := s.encryptionService.GenerateID(ctx)
	if err != nil {
		return nil, wrapWithSentinel(ErrFileEncryptionFailed, err)
	}

	sessionID, err := s.encryptionService.GenerateID(ctx)
	if err != nil {
		return nil, wrapWithSentinel(ErrFileEncryptionFailed, err)
	}

	encryptionKey, err := s.encryptionService.GenerateKey(ctx, 32)
	if err != nil {
		return nil, wrapWithSentinel(ErrFileEncryptionFailed, err)
	}

	uploadID, err := s.objectStorage.InitiateMultipartUpload(ctx, fileID, req.ContentType)
	if err != nil {
		return nil, wrapWithSentinel(ErrObjectStorageFailed, err)
	}

	totalChunks := int((req.TotalSize + req.ChunkSize - 1) / req.ChunkSize)
	now := time.Now()
	session := contracts.FileUploadSession{
		SessionID:     sessionID,
		FileID:        fileID,
		MessageID:     req.MessageID,
		UploadID:      uploadID,
		Filename:      req.Filename,
		ContentType:   req.ContentType,
		TotalSize:     req.TotalSize,
		TotalChunks:   totalChunks,
		Status:        string(SessionStatusActive),
		EncryptionKey: encryptionKey,
		CreatedAt:     now,
		ExpiresAt:     now.Add(fileUploadSessionTTL),
	}

	if err := s.uploadState.CreateSession(ctx, session); err != nil {
		// Best-effort abort to avoid leaving a dangling multipart upload in object storage.
		// Use a fresh context with a short timeout so that a cancelled or timed-out
		// request context does not prevent the cleanup from reaching object storage.
		abortCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.objectStorage.AbortMultipartUpload(abortCtx, fileID, uploadID)
		return nil, wrapWithSentinel(ErrUploadStateFailed, err)
	}

	return &InitiateUploadResponse{FileID: fileID, SessionID: sessionID, EncryptionKey: encryptionKey}, nil
}

// UploadChunk uploads a single encrypted chunk.
func (s *FileService) UploadChunk(ctx context.Context, req UploadChunkRequest) (*UploadChunkResponse, error) {
	if err := validateChunkRequest(req); err != nil {
		return nil, err
	}

	session, err := s.uploadState.GetSessionByID(ctx, req.SessionID)
	if err != nil {
		if errors.Is(err, ErrUploadSessionNotFound) {
			return nil, wrapWithSentinel(ErrUploadSessionNotFound, err)
		}
		return nil, wrapWithSentinel(ErrUploadStateFailed, err)
	}
	if session == nil || session.FileID != req.FileID {
		return nil, ErrUploadSessionNotFound
	}
	if session.Status == string(SessionStatusComplete) {
		return nil, ErrUploadAlreadyComplete
	}
	if session.Status != string(SessionStatusActive) {
		return nil, ErrUploadSessionNotFound
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrUploadSessionExpired
	}
	if session.TotalChunks > 0 && req.TotalChunks != session.TotalChunks {
		return nil, ErrInvalidChunkIndex
	}

	chunkMeta := secondary.ChunkMeta{
		FileID:      session.FileID,
		ChunkIndex:  req.ChunkIndex,
		TotalChunks: session.TotalChunks,
	}
	encryptedChunk, err := s.encryptionService.EncryptChunk(ctx, req.Data, session.EncryptionKey, chunkMeta)
	if err != nil {
		return nil, wrapWithSentinel(ErrFileEncryptionFailed, err)
	}

	etag, err := s.objectStorage.UploadPart(ctx, session.FileID, session.UploadID, req.ChunkIndex, encryptedChunk)
	if err != nil {
		return nil, wrapWithSentinel(ErrObjectStorageFailed, err)
	}

	part := contracts.CompletedPart{PartNumber: req.ChunkIndex, ETag: etag}
	parts, err := s.uploadState.AddCompletedPart(ctx, req.SessionID, part)
	if err != nil {
		return nil, wrapWithSentinel(ErrUploadStateFailed, err)
	}

	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})
	done := allPartsPresent(parts, session.TotalChunks)
	if done {
		// Multipart completion APIs require the complete ordered list of uploaded
		// parts, not only the final chunk, so use the atomically returned list.
		if err := s.objectStorage.CompleteMultipartUpload(ctx, session.FileID, session.UploadID, parts); err != nil {
			return nil, wrapWithSentinel(ErrObjectStorageFailed, err)
		}
		if err := s.uploadState.CompleteSession(ctx, req.SessionID); err != nil {
			return nil, wrapWithSentinel(ErrUploadStateFailed, err)
		}
	}

	return &UploadChunkResponse{FileID: req.FileID, ChunkIndex: req.ChunkIndex, Done: done}, nil
}

// DownloadFile downloads and decrypts a file.
func (s *FileService) DownloadFile(ctx context.Context, req DownloadFileRequest) (*DownloadFileResponse, error) {
	session, err := s.uploadState.GetSessionByFileID(ctx, req.FileID)
	if err != nil {
		if errors.Is(err, ErrUploadSessionNotFound) {
			return nil, wrapWithSentinel(ErrUploadSessionNotFound, err)
		}
		return nil, wrapWithSentinel(ErrUploadStateFailed, err)
	}
	if session == nil {
		return nil, ErrUploadSessionNotFound
	}

	reader, size, err := s.objectStorage.GetObject(ctx, req.FileID)
	if err != nil {
		return nil, wrapWithSentinel(ErrObjectStorageFailed, err)
	}

	if session.TotalSize > 0 && session.TotalChunks > 0 {
		maxEncryptedSize := addEncryptedStoredOverhead(session.TotalSize, session.TotalChunks)
		if size > maxEncryptedSize {
			_ = reader.Close()
			return nil, ErrFileTooLarge
		}
	} else if s.maxFileSize > 0 {
		chunkCount := session.TotalChunks
		if chunkCount <= 0 {
			chunkCount = 1
		}
		maxEncryptedSize := addEncryptedStoredOverhead(s.maxFileSize, chunkCount)
		if size > maxEncryptedSize {
			_ = reader.Close()
			return nil, ErrFileTooLarge
		}
	}

	// maxFrameSize caps per-frame allocations in DecryptFileStream.
	//
	// Note: TotalChunks is derived from the client-provided chunk size at upload
	// initiation (TotalChunks = ceil(TotalSize/ChunkSize)), but ChunkSize is not
	// persisted in the session. Using TotalSize/TotalChunks here would therefore
	// under-estimate the maximum chunk size for many valid uploads (especially when
	// the last chunk is small) and cause downloads to fail. The only safe bound
	// available from persisted metadata is the full plaintext size.
	maxFrameSize := int64(0)
	if session.TotalSize > 0 {
		maxFrameSize = addEncryptedPayloadOverhead(session.TotalSize)
	} else if s.maxFileSize > 0 {
		maxFrameSize = addEncryptedPayloadOverhead(s.maxFileSize)
	}

	fileMeta := secondary.FileMeta{FileID: req.FileID, TotalChunks: session.TotalChunks, MaxFrameSize: maxFrameSize}
	decryptedReader, decryptedWriter := io.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = reader.Close() }()
		if err := s.encryptionService.DecryptFileStream(ctx, reader, decryptedWriter, req.Key, fileMeta); err != nil {
			// Only I/O errors while reading ciphertext should be treated as storage failures.
			if errors.Is(err, secondary.ErrCiphertextReadFailed) {
				_ = decryptedWriter.CloseWithError(wrapWithSentinel(ErrObjectStorageFailed, err))
			} else {
				_ = decryptedWriter.CloseWithError(wrapWithSentinel(ErrFileDecryptionFailed, err))
			}
			return
		}
		_ = decryptedWriter.Close()
	}()
	// Safety net: if the caller discards response.Data without closing it the
	// goroutine above would block forever in dst.Write. Closing the write end
	// when the request context finishes guarantees the goroutine exits even
	// when the caller abandons the reader.
	go func() {
		select {
		case <-ctx.Done():
			_ = decryptedWriter.CloseWithError(ctx.Err())
		case <-done:
		}
	}()

	return &DownloadFileResponse{
		Filename:    session.Filename,
		ContentType: session.ContentType,
		Data:        decryptedReader,
	}, nil
}

// AbortUpload aborts an in-progress multipart upload.
func (s *FileService) AbortUpload(ctx context.Context, sessionID string) error {
	session, err := s.uploadState.GetSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ErrUploadSessionNotFound) {
			return wrapWithSentinel(ErrUploadSessionNotFound, err)
		}
		return wrapWithSentinel(ErrUploadStateFailed, err)
	}
	if session == nil {
		return ErrUploadSessionNotFound
	}

	if err := s.objectStorage.AbortMultipartUpload(ctx, session.FileID, session.UploadID); err != nil {
		return wrapWithSentinel(ErrObjectStorageFailed, err)
	}

	if err := s.uploadState.DeleteSession(ctx, sessionID); err != nil {
		return wrapWithSentinel(ErrUploadStateFailed, err)
	}

	return nil
}

// CleanupExpiredSessions removes expired incomplete sessions and best-effort
// aborts their dangling multipart uploads.
func (s *FileService) CleanupExpiredSessions(ctx context.Context) error {
	removed, err := s.uploadState.DeleteExpiredSessions(ctx, time.Now())
	if err != nil {
		return wrapWithSentinel(ErrUploadStateFailed, err)
	}

	for _, session := range removed {
		if session.UploadID == "" {
			continue
		}
		// Best-effort: a failed abort leaves a dangling multipart upload, but the
		// session is already gone from state so retrying is not possible here.
		// Mirror the best-effort abort used when session creation fails.
		_ = s.objectStorage.AbortMultipartUpload(ctx, session.FileID, session.UploadID)
	}

	return nil
}

// RunSessionCleanup sweeps expired sessions on a ticker until ctx is cancelled.
func (s *FileService) RunSessionCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.CleanupExpiredSessions(ctx); err != nil {
				s.logger.Error().Err(err).Msg("Failed to cleanup expired upload sessions")
			}
		}
	}
}

// allPartsPresent returns true when parts contains exactly one entry for each
// part number from 1 through totalChunks, ensuring completion is triggered
// only when every chunk has been received regardless of arrival order.
func allPartsPresent(parts []contracts.CompletedPart, totalChunks int) bool {
	if len(parts) < totalChunks {
		return false
	}
	seen := make(map[int]bool, len(parts))
	for _, p := range parts {
		seen[p.PartNumber] = true
	}
	for i := 1; i <= totalChunks; i++ {
		if !seen[i] {
			return false
		}
	}
	return true
}

func validateChunkRequest(req UploadChunkRequest) error {
	if req.ChunkIndex <= 0 || req.TotalChunks <= 0 || req.ChunkIndex > req.TotalChunks {
		return ErrInvalidChunkIndex
	}
	return nil
}

func wrapWithSentinel(sentinel error, err error) error {
	if err == nil {
		return sentinel
	}
	return fmt.Errorf("%w: %w", sentinel, err)
}

// addEncryptedPayloadOverhead adds per-frame AEAD payload overhead to a
// plaintext size while saturating on int64 overflow.
func addEncryptedPayloadOverhead(plaintextSize int64) int64 {
	if plaintextSize > math.MaxInt64-secondary.EncryptedFramePayloadOverheadBytes {
		return math.MaxInt64
	}
	return plaintextSize + secondary.EncryptedFramePayloadOverheadBytes
}

// addEncryptedStoredOverhead adds per-frame storage overhead for total chunks
// while saturating on int64 overflow.
func addEncryptedStoredOverhead(baseSize int64, totalChunks int) int64 {
	chunkCount := int64(totalChunks)
	if chunkCount > 0 && chunkCount > math.MaxInt64/secondary.EncryptedFrameStoredOverheadBytes {
		return math.MaxInt64
	}
	overhead := chunkCount * secondary.EncryptedFrameStoredOverheadBytes
	if baseSize > math.MaxInt64-overhead {
		return math.MaxInt64
	}
	return baseSize + overhead
}
