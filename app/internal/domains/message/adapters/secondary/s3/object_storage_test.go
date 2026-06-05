package s3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeCore implements multipartCore with function fields so each test can
// configure exactly the behavior it needs.
type fakeCore struct {
	newMultipartUploadFn      func(ctx context.Context, bucket, object string, opts minio.PutObjectOptions) (string, error)
	putObjectPartFn           func(ctx context.Context, bucket, object, uploadID string, partID int, data io.Reader, size int64, opts minio.PutObjectPartOptions) (minio.ObjectPart, error)
	completeMultipartUploadFn func(ctx context.Context, bucket, object, uploadID string, parts []minio.CompletePart, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	abortMultipartUploadFn    func(ctx context.Context, bucket, object, uploadID string) error
}

func (f *fakeCore) NewMultipartUpload(ctx context.Context, bucket, object string, opts minio.PutObjectOptions) (string, error) {
	return f.newMultipartUploadFn(ctx, bucket, object, opts)
}

func (f *fakeCore) PutObjectPart(ctx context.Context, bucket, object, uploadID string, partID int, data io.Reader, size int64, opts minio.PutObjectPartOptions) (minio.ObjectPart, error) {
	return f.putObjectPartFn(ctx, bucket, object, uploadID, partID, data, size, opts)
}

func (f *fakeCore) CompleteMultipartUpload(ctx context.Context, bucket, object, uploadID string, parts []minio.CompletePart, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return f.completeMultipartUploadFn(ctx, bucket, object, uploadID, parts, opts)
}

func (f *fakeCore) AbortMultipartUpload(ctx context.Context, bucket, object, uploadID string) error {
	return f.abortMultipartUploadFn(ctx, bucket, object, uploadID)
}

// fakeGetter implements objectGetter with a function field.
type fakeGetter struct {
	getObjectFn func(ctx context.Context, bucket, object string, opts minio.GetObjectOptions) (readerStat, error)
}

func (g *fakeGetter) GetObject(ctx context.Context, bucket, object string, opts minio.GetObjectOptions) (readerStat, error) {
	return g.getObjectFn(ctx, bucket, object, opts)
}

// fakeReaderStat implements readerStat backed by a bytes.Reader; it records
// whether Close was called so tests can assert the close-on-stat-error path.
type fakeReaderStat struct {
	reader *bytes.Reader
	size   int64
	statFn func() (minio.ObjectInfo, error)
	closed bool
}

func (f *fakeReaderStat) Read(p []byte) (int, error)      { return f.reader.Read(p) }
func (f *fakeReaderStat) Close() error                    { f.closed = true; return nil }
func (f *fakeReaderStat) Stat() (minio.ObjectInfo, error) { return f.statFn() }

// newTestAdapter builds an S3Adapter with the given fakes and a fixed bucket name.
func newTestAdapter(core multipartCore, getter objectGetter) *S3Adapter {
	return &S3Adapter{core: core, getter: getter, bucket: "test-bucket"}
}

// ---------- InitiateMultipartUpload ----------

func TestInitiateMultipartUpload_Happy(t *testing.T) {
	t.Parallel()
	var gotBucket, gotObject, gotContentType string
	core := &fakeCore{
		newMultipartUploadFn: func(_ context.Context, bucket, object string, opts minio.PutObjectOptions) (string, error) {
			gotBucket = bucket
			gotObject = object
			gotContentType = opts.ContentType
			return "upload-id-1", nil
		},
	}
	a := newTestAdapter(core, nil)

	id, err := a.InitiateMultipartUpload(context.Background(), "file-abc", "application/octet-stream")

	require.NoError(t, err)
	assert.Equal(t, "upload-id-1", id)
	assert.Equal(t, "test-bucket", gotBucket)
	assert.Equal(t, "file-abc", gotObject)
	assert.Equal(t, "application/octet-stream", gotContentType)
}

func TestInitiateMultipartUpload_Error(t *testing.T) {
	t.Parallel()
	boom := errors.New("s3 unavailable")
	core := &fakeCore{
		newMultipartUploadFn: func(_ context.Context, _, _ string, _ minio.PutObjectOptions) (string, error) {
			return "", boom
		},
	}
	a := newTestAdapter(core, nil)

	id, err := a.InitiateMultipartUpload(context.Background(), "file-abc", "application/octet-stream")

	assert.ErrorIs(t, err, boom)
	assert.Empty(t, id)
}

// ---------- UploadPart ----------

func TestUploadPart_Happy(t *testing.T) {
	t.Parallel()
	payload := []byte("chunk-data")
	var gotPartID int
	var gotSize int64
	var gotData []byte

	core := &fakeCore{
		putObjectPartFn: func(_ context.Context, _, _, _ string, partID int, data io.Reader, size int64, _ minio.PutObjectPartOptions) (minio.ObjectPart, error) {
			gotPartID = partID
			gotSize = size
			gotData, _ = io.ReadAll(data)
			return minio.ObjectPart{ETag: "etag-part-1"}, nil
		},
	}
	a := newTestAdapter(core, nil)

	etag, err := a.UploadPart(context.Background(), "file-abc", "upload-id-1", 1, payload)

	require.NoError(t, err)
	assert.Equal(t, "etag-part-1", etag)
	assert.Equal(t, 1, gotPartID)
	assert.Equal(t, int64(len(payload)), gotSize)
	assert.Equal(t, payload, gotData)
}

func TestUploadPart_Error(t *testing.T) {
	t.Parallel()
	boom := errors.New("part upload failed")
	core := &fakeCore{
		putObjectPartFn: func(_ context.Context, _, _, _ string, _ int, _ io.Reader, _ int64, _ minio.PutObjectPartOptions) (minio.ObjectPart, error) {
			return minio.ObjectPart{}, boom
		},
	}
	a := newTestAdapter(core, nil)

	etag, err := a.UploadPart(context.Background(), "file-abc", "upload-id-1", 1, []byte("data"))

	assert.ErrorIs(t, err, boom)
	assert.Empty(t, etag)
}

// ---------- CompleteMultipartUpload ----------

func TestCompleteMultipartUpload_MapsPartsInOrder(t *testing.T) {
	t.Parallel()
	var gotParts []minio.CompletePart
	core := &fakeCore{
		completeMultipartUploadFn: func(_ context.Context, _, _, _ string, parts []minio.CompletePart, _ minio.PutObjectOptions) (minio.UploadInfo, error) {
			gotParts = parts
			return minio.UploadInfo{}, nil
		},
	}
	a := newTestAdapter(core, nil)

	input := []contracts.CompletedPart{
		{PartNumber: 1, ETag: "etag-1"},
		{PartNumber: 2, ETag: "etag-2"},
		{PartNumber: 3, ETag: "etag-3"},
	}
	err := a.CompleteMultipartUpload(context.Background(), "file-abc", "upload-id-1", input)

	require.NoError(t, err)
	require.Len(t, gotParts, 3)
	assert.Equal(t, 1, gotParts[0].PartNumber)
	assert.Equal(t, "etag-1", gotParts[0].ETag)
	assert.Equal(t, 2, gotParts[1].PartNumber)
	assert.Equal(t, "etag-2", gotParts[1].ETag)
	assert.Equal(t, 3, gotParts[2].PartNumber)
	assert.Equal(t, "etag-3", gotParts[2].ETag)
}

func TestCompleteMultipartUpload_Error(t *testing.T) {
	t.Parallel()
	boom := errors.New("complete failed")
	core := &fakeCore{
		completeMultipartUploadFn: func(_ context.Context, _, _, _ string, _ []minio.CompletePart, _ minio.PutObjectOptions) (minio.UploadInfo, error) {
			return minio.UploadInfo{}, boom
		},
	}
	a := newTestAdapter(core, nil)

	err := a.CompleteMultipartUpload(context.Background(), "file-abc", "upload-id-1", []contracts.CompletedPart{{PartNumber: 1, ETag: "e"}})

	assert.ErrorIs(t, err, boom)
}

func TestCompleteMultipartUpload_EmptyParts(t *testing.T) {
	t.Parallel()
	var gotParts []minio.CompletePart
	core := &fakeCore{
		completeMultipartUploadFn: func(_ context.Context, _, _, _ string, parts []minio.CompletePart, _ minio.PutObjectOptions) (minio.UploadInfo, error) {
			gotParts = parts
			return minio.UploadInfo{}, nil
		},
	}
	a := newTestAdapter(core, nil)

	err := a.CompleteMultipartUpload(context.Background(), "file-abc", "upload-id-1", nil)

	require.NoError(t, err)
	assert.Empty(t, gotParts)
}

// ---------- AbortMultipartUpload ----------

func TestAbortMultipartUpload_Happy(t *testing.T) {
	t.Parallel()
	var called bool
	core := &fakeCore{
		abortMultipartUploadFn: func(_ context.Context, _, _, _ string) error {
			called = true
			return nil
		},
	}
	a := newTestAdapter(core, nil)

	err := a.AbortMultipartUpload(context.Background(), "file-abc", "upload-id-1")

	require.NoError(t, err)
	assert.True(t, called)
}

func TestAbortMultipartUpload_Error(t *testing.T) {
	t.Parallel()
	boom := errors.New("abort failed")
	core := &fakeCore{
		abortMultipartUploadFn: func(_ context.Context, _, _, _ string) error { return boom },
	}
	a := newTestAdapter(core, nil)

	err := a.AbortMultipartUpload(context.Background(), "file-abc", "upload-id-1")

	assert.ErrorIs(t, err, boom)
}

// ---------- GetObject ----------

func TestGetObject_Happy(t *testing.T) {
	t.Parallel()
	content := "hello from object storage"
	rs := &fakeReaderStat{
		reader: bytes.NewReader([]byte(content)),
		statFn: func() (minio.ObjectInfo, error) {
			return minio.ObjectInfo{Size: int64(len(content))}, nil
		},
	}
	getter := &fakeGetter{
		getObjectFn: func(_ context.Context, _, _ string, _ minio.GetObjectOptions) (readerStat, error) {
			return rs, nil
		},
	}
	a := newTestAdapter(nil, getter)

	rc, size, err := a.GetObject(context.Background(), "file-abc")

	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), size)
	assert.False(t, rs.closed, "reader must NOT be closed on success")

	body, _ := io.ReadAll(rc)
	assert.Equal(t, content, string(body))
}

func TestGetObject_GetterError(t *testing.T) {
	t.Parallel()
	boom := errors.New("getter failed")
	getter := &fakeGetter{
		getObjectFn: func(_ context.Context, _, _ string, _ minio.GetObjectOptions) (readerStat, error) {
			return nil, boom
		},
	}
	a := newTestAdapter(nil, getter)

	rc, size, err := a.GetObject(context.Background(), "file-abc")

	assert.ErrorIs(t, err, boom)
	assert.Nil(t, rc)
	assert.Zero(t, size)
}

func TestGetObject_StatError_ClosesReader(t *testing.T) {
	t.Parallel()
	boom := errors.New("stat failed")
	rs := &fakeReaderStat{
		reader: bytes.NewReader(nil),
		statFn: func() (minio.ObjectInfo, error) { return minio.ObjectInfo{}, boom },
	}
	getter := &fakeGetter{
		getObjectFn: func(_ context.Context, _, _ string, _ minio.GetObjectOptions) (readerStat, error) {
			return rs, nil
		},
	}
	a := newTestAdapter(nil, getter)

	rc, size, err := a.GetObject(context.Background(), "file-abc")

	assert.ErrorIs(t, err, boom)
	assert.Nil(t, rc)
	assert.Zero(t, size)
	assert.True(t, rs.closed, "reader must be closed when Stat returns an error")
}

// ---------- NewS3Adapter constructor ----------

func TestNewS3Adapter_InvalidEndpoint(t *testing.T) {
	t.Parallel()
	// An endpoint containing a scheme causes minio.New to return an error.
	_, err := NewS3Adapter("http://invalid::endpoint", "key", "secret", "bucket", false)
	assert.Error(t, err)
}

// ---------- compile-time interface guard ----------

func TestAdapterSatisfiesPort(t *testing.T) {
	t.Parallel()
	// Verified by the var _ guard in object_storage.go; this test acts as
	// human-readable documentation that the interface contract is met.
	_ = strings.TrimSpace // keep import used
	var _ interface {
		GetObject(context.Context, string) (io.ReadCloser, int64, error)
	} = (*S3Adapter)(nil)
}
