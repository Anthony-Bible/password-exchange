package secondary

import (
	"context"
	"io"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
)

// ObjectStoragePort defines the secondary port responsible for storing encrypted
// file content in an object store such as S3 or MinIO.
type ObjectStoragePort interface {
	// InitiateMultipartUpload starts a multipart upload and returns the provider upload ID.
	InitiateMultipartUpload(ctx context.Context, fileID, contentType string) (uploadID string, err error)

	// UploadPart uploads one encrypted part for the given multipart upload.
	UploadPart(ctx context.Context, fileID, uploadID string, partNumber int, data []byte) (etag string, err error)

	// CompleteMultipartUpload finalizes the multipart upload using the full ordered part list.
	CompleteMultipartUpload(ctx context.Context, fileID, uploadID string, parts []contracts.CompletedPart) error

	// AbortMultipartUpload cancels an in-progress multipart upload and discards any staged parts.
	AbortMultipartUpload(ctx context.Context, fileID, uploadID string) error

	// GetObject retrieves the completed encrypted file and its size.
	GetObject(ctx context.Context, fileID string) (io.ReadCloser, int64, error)
}
