package domain

import "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"

// FileUploadSessionStatus describes the lifecycle state of a file upload session.
type FileUploadSessionStatus string

// SessionStatusActive indicates the upload session can accept additional chunks.
const SessionStatusActive FileUploadSessionStatus = "active"

// SessionStatusComplete indicates the multipart upload has been finalized.
const SessionStatusComplete FileUploadSessionStatus = "complete"

// InitiateUploadRequest is the domain alias for the upload-initiation request contract.
type InitiateUploadRequest = contracts.InitiateUploadRequest

// InitiateUploadResponse is the domain alias for the upload-initiation response contract.
type InitiateUploadResponse = contracts.InitiateUploadResponse

// UploadChunkRequest is the domain alias for the chunk-upload request contract.
type UploadChunkRequest = contracts.UploadChunkRequest

// UploadChunkResponse is the domain alias for the chunk-upload response contract.
type UploadChunkResponse = contracts.UploadChunkResponse

// DownloadFileRequest is the domain alias for the file-download request contract.
type DownloadFileRequest = contracts.DownloadFileRequest

// DownloadFileResponse is the domain alias for the file-download response contract.
type DownloadFileResponse = contracts.DownloadFileResponse

// FileUploadSession is the domain alias for the persisted upload-session contract.
type FileUploadSession = contracts.FileUploadSession

// CompletedPart is the domain alias for a successfully uploaded multipart part.
type CompletedPart = contracts.CompletedPart
