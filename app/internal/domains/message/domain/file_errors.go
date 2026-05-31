package domain

import "errors"

// ErrFileTooLarge indicates the uploaded file exceeds the configured limit.
var ErrFileTooLarge = errors.New("file too large")

// ErrUploadSessionNotFound indicates the requested upload session does not exist.
var ErrUploadSessionNotFound = errors.New("upload session not found")

// ErrUploadSessionAlreadyExists indicates a session with the same ID or file ID already exists.
var ErrUploadSessionAlreadyExists = errors.New("upload session already exists")

// ErrUploadAlreadyComplete indicates the upload has already been finalized.
var ErrUploadAlreadyComplete = errors.New("upload already complete")

// ErrFileEncryptionFailed indicates file encryption or key generation failed.
var ErrFileEncryptionFailed = errors.New("file encryption failed")

// ErrFileDecryptionFailed indicates file decryption failed.
var ErrFileDecryptionFailed = errors.New("file decryption failed")

// ErrObjectStorageFailed indicates an object storage operation failed.
var ErrObjectStorageFailed = errors.New("object storage operation failed")

// ErrUploadStateFailed indicates upload-session state persistence failed.
var ErrUploadStateFailed = errors.New("upload state operation failed")

// ErrInvalidChunkIndex indicates chunk numbering or chunk counts were invalid.
var ErrInvalidChunkIndex = errors.New("invalid chunk index")

// ErrInvalidUploadRequest indicates upload initiation parameters (e.g. TotalSize or ChunkSize) are invalid.
var ErrInvalidUploadRequest = errors.New("invalid upload request")

// ErrUploadSessionExpired indicates the upload session's TTL has elapsed.
var ErrUploadSessionExpired = errors.New("upload session expired")
