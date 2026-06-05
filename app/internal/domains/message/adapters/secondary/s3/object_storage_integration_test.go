//go:build integration

package s3_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	s3adapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/secondary/s3"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/integration/garagetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// s3MinPartSize is the minimum non-final part size enforced by S3 (5 MiB).
	s3MinPartSize = 5 * 1024 * 1024
)

// sharedAdapter is created once in TestMain and reused by all integration tests.
// Each test uses a unique fileID so there is no cross-test state.
var sharedAdapter *s3adapter.S3Adapter

func TestMain(m *testing.M) {
	endpoint, accessKey, secretKey, bucket, teardown := garagetest.MustStart()

	var err error
	sharedAdapter, err = s3adapter.NewS3Adapter(endpoint, accessKey, secretKey, bucket, false)
	if err != nil {
		log.Fatalf("garagetest: create S3 adapter: %v", err)
	}

	code := m.Run()
	teardown()
	os.Exit(code)
}

func uniqueFileID(t *testing.T, suffix string) string {
	t.Helper()
	name := strings.NewReplacer("/", "-", " ", "-").Replace(t.Name())
	return fmt.Sprintf("%s-%s", name, suffix)
}

func sha256hex(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

func TestIntegration_MultipartRoundTrip(t *testing.T) {
	ctx := context.Background()

	fileID := uniqueFileID(t, "roundtrip")
	uploadID, err := sharedAdapter.InitiateMultipartUpload(ctx, fileID, "application/octet-stream")
	require.NoError(t, err, "initiate multipart upload")

	// Part 1: exactly 5 MiB (S3 minimum for non-final parts).
	part1Data := bytes.Repeat([]byte("A"), s3MinPartSize)
	// Part 2: ~1 KiB final part.
	part2Data := bytes.Repeat([]byte("B"), 1024)

	etag1, err := sharedAdapter.UploadPart(ctx, fileID, uploadID, 1, part1Data)
	require.NoError(t, err, "upload part 1")
	assert.NotEmpty(t, etag1)

	etag2, err := sharedAdapter.UploadPart(ctx, fileID, uploadID, 2, part2Data)
	require.NoError(t, err, "upload part 2")
	assert.NotEmpty(t, etag2)

	err = sharedAdapter.CompleteMultipartUpload(ctx, fileID, uploadID, []contracts.CompletedPart{
		{PartNumber: 1, ETag: etag1},
		{PartNumber: 2, ETag: etag2},
	})
	require.NoError(t, err, "complete multipart upload")

	expected := append(part1Data, part2Data...)
	expectedHash := sha256hex(expected)
	expectedSize := int64(len(expected))

	rc, size, err := sharedAdapter.GetObject(ctx, fileID)
	require.NoError(t, err, "get object")
	defer rc.Close()

	assert.Equal(t, expectedSize, size)

	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, expectedHash, sha256hex(got), "content hash mismatch")
}

func TestIntegration_AbortDiscardsUpload(t *testing.T) {
	ctx := context.Background()

	fileID := uniqueFileID(t, "abort")
	uploadID, err := sharedAdapter.InitiateMultipartUpload(ctx, fileID, "application/octet-stream")
	require.NoError(t, err)

	// Upload one part so there is staged data to discard.
	partData := bytes.Repeat([]byte("C"), s3MinPartSize)
	_, err = sharedAdapter.UploadPart(ctx, fileID, uploadID, 1, partData)
	require.NoError(t, err)

	err = sharedAdapter.AbortMultipartUpload(ctx, fileID, uploadID)
	require.NoError(t, err, "abort multipart upload")

	// The object must not exist after an abort.
	_, _, err = sharedAdapter.GetObject(ctx, fileID)
	assert.Error(t, err, "GetObject after abort should error (object does not exist)")
}

func TestIntegration_GetObjectMissing(t *testing.T) {
	_, _, err := sharedAdapter.GetObject(context.Background(), uniqueFileID(t, "missing"))
	assert.Error(t, err, "GetObject for a non-existent key should return an error")
}

func TestIntegration_CompleteWithBadETag(t *testing.T) {
	ctx := context.Background()

	fileID := uniqueFileID(t, "badeTag")
	uploadID, err := sharedAdapter.InitiateMultipartUpload(ctx, fileID, "application/octet-stream")
	require.NoError(t, err)

	partData := bytes.Repeat([]byte("D"), s3MinPartSize)
	_, err = sharedAdapter.UploadPart(ctx, fileID, uploadID, 1, partData)
	require.NoError(t, err)

	// Deliberately supply a wrong ETag — S3 must reject the completion.
	err = sharedAdapter.CompleteMultipartUpload(ctx, fileID, uploadID, []contracts.CompletedPart{
		{PartNumber: 1, ETag: "\"not-a-real-etag\""},
	})
	assert.Error(t, err, "CompleteMultipartUpload with a bad ETag should error")
}
