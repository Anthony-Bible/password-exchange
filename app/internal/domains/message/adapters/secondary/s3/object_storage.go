package s3

import (
	"bytes"
	"context"
	"io"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var _ secondary.ObjectStoragePort = (*S3Adapter)(nil)

// multipartCore is satisfied by *minio.Core and allows unit tests to inject a fake.
type multipartCore interface {
	NewMultipartUpload(ctx context.Context, bucket, object string, opts minio.PutObjectOptions) (string, error)
	PutObjectPart(ctx context.Context, bucket, object, uploadID string, partID int, data io.Reader, size int64, opts minio.PutObjectPartOptions) (minio.ObjectPart, error)
	CompleteMultipartUpload(ctx context.Context, bucket, object, uploadID string, parts []minio.CompletePart, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	AbortMultipartUpload(ctx context.Context, bucket, object, uploadID string) error
}

// readerStat is satisfied by *minio.Object; it combines io.ReadCloser with the
// Stat method needed to retrieve object metadata before streaming.
type readerStat interface {
	io.ReadCloser
	Stat() (minio.ObjectInfo, error)
}

// objectGetter abstracts *minio.Client.GetObject so tests can inject a fake.
type objectGetter interface {
	GetObject(ctx context.Context, bucket, object string, opts minio.GetObjectOptions) (readerStat, error)
}

// minioObjectGetter wraps *minio.Client so its GetObject return type satisfies
// objectGetter (which returns readerStat instead of *minio.Object).
type minioObjectGetter struct{ client *minio.Client }

func (g *minioObjectGetter) GetObject(ctx context.Context, bucket, object string, opts minio.GetObjectOptions) (readerStat, error) {
	return g.client.GetObject(ctx, bucket, object, opts)
}

// S3Adapter implements ObjectStoragePort with an S3-compatible backend.
type S3Adapter struct {
	getter objectGetter
	core   multipartCore
	bucket string
}

// NewS3Adapter creates a new S3-compatible object storage adapter.
func NewS3Adapter(endpoint, accessKeyID, secretAccessKey, bucket string, useSSL bool) (*S3Adapter, error) {
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	}

	s3Client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, err
	}

	coreClient, err := minio.NewCore(endpoint, opts)
	if err != nil {
		return nil, err
	}

	return &S3Adapter{
		getter: &minioObjectGetter{client: s3Client},
		core:   coreClient,
		bucket: bucket,
	}, nil
}

// InitiateMultipartUpload starts a multipart upload for the provided file.
func (a *S3Adapter) InitiateMultipartUpload(ctx context.Context, fileID, contentType string) (string, error) {
	return a.core.NewMultipartUpload(ctx, a.bucket, fileID, minio.PutObjectOptions{ContentType: contentType})
}

// UploadPart uploads a single multipart chunk and returns its ETag.
func (a *S3Adapter) UploadPart(ctx context.Context, fileID, uploadID string, partNumber int, data []byte) (string, error) {
	part, err := a.core.PutObjectPart(ctx, a.bucket, fileID, uploadID, partNumber, bytes.NewReader(data), int64(len(data)), minio.PutObjectPartOptions{})
	if err != nil {
		return "", err
	}
	return part.ETag, nil
}

// CompleteMultipartUpload finalizes the multipart upload with all completed parts.
func (a *S3Adapter) CompleteMultipartUpload(ctx context.Context, fileID, uploadID string, parts []contracts.CompletedPart) error {
	completedParts := make([]minio.CompletePart, 0, len(parts))
	for _, part := range parts {
		completedParts = append(completedParts, minio.CompletePart{PartNumber: part.PartNumber, ETag: part.ETag})
	}

	_, err := a.core.CompleteMultipartUpload(ctx, a.bucket, fileID, uploadID, completedParts, minio.PutObjectOptions{})
	return err
}

// AbortMultipartUpload cancels an in-progress multipart upload.
func (a *S3Adapter) AbortMultipartUpload(ctx context.Context, fileID, uploadID string) error {
	return a.core.AbortMultipartUpload(ctx, a.bucket, fileID, uploadID)
}

// GetObject retrieves the stored object and returns its size.
func (a *S3Adapter) GetObject(ctx context.Context, fileID string) (io.ReadCloser, int64, error) {
	obj, err := a.getter.GetObject(ctx, a.bucket, fileID, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, err
	}

	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, 0, err
	}

	return obj, info.Size, nil
}
