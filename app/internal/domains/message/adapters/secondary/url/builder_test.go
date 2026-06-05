package url

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time proof that *URLBuilder satisfies the port interface.
var _ secondary.URLBuilderPort = NewURLBuilder("")

func TestNewURLBuilder_ReturnsNonNil(t *testing.T) {
	t.Parallel()
	b := NewURLBuilder("https://example.com/")
	require.NotNil(t, b)
}

func TestBuildDecryptURL_ExactFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		baseURL   string
		messageID string
		key       []byte
		wantURL   string
	}{
		{
			name:      "base URL with trailing slash produces well-formed URL",
			baseURL:   "https://example.com/",
			messageID: "abc123",
			key:       []byte("key"),
			wantURL:   "https://example.com/decrypt/abc123/" + base64.URLEncoding.EncodeToString([]byte("key")),
		},
		{
			// No trailing slash: fmt.Sprintf concatenates directly, producing
			// "https://example.comdecrypt/..." — documents current (intentional) behavior.
			name:      "base URL WITHOUT trailing slash concatenates directly — no slash inserted",
			baseURL:   "https://example.com",
			messageID: "abc123",
			key:       []byte("key"),
			wantURL:   "https://example.comdecrypt/abc123/" + base64.URLEncoding.EncodeToString([]byte("key")),
		},
		{
			name:      "empty key bytes produce empty base64 segment",
			baseURL:   "https://example.com/",
			messageID: "msg-999",
			key:       []byte{},
			wantURL:   "https://example.com/decrypt/msg-999/",
		},
		{
			name:      "empty message ID is preserved verbatim",
			baseURL:   "https://example.com/",
			messageID: "",
			key:       []byte("k"),
			wantURL:   "https://example.com/decrypt//" + base64.URLEncoding.EncodeToString([]byte("k")),
		},
		{
			name:      "message ID with special characters is preserved verbatim",
			baseURL:   "https://example.com/",
			messageID: "abc-123_XYZ",
			key:       []byte("secretkey"),
			wantURL:   "https://example.com/decrypt/abc-123_XYZ/" + base64.URLEncoding.EncodeToString([]byte("secretkey")),
		},
		{
			name:      "empty base URL and empty message ID",
			baseURL:   "",
			messageID: "",
			key:       []byte("k"),
			wantURL:   "decrypt//" + base64.URLEncoding.EncodeToString([]byte("k")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b := NewURLBuilder(tc.baseURL)
			got := b.BuildDecryptURL(tc.messageID, tc.key)
			assert.Equal(t, tc.wantURL, got)
		})
	}
}

// TestBuildDecryptURL_KeyIsBase64URLEncoded verifies URL-safe base64 is used:
// 0xfb 0xff encodes to "+/" in standard base64 and "-_" in URL-safe base64.
func TestBuildDecryptURL_KeyIsBase64URLEncoded(t *testing.T) {
	t.Parallel()

	key := []byte{0xfb, 0xff}
	b := NewURLBuilder("https://example.com/")
	got := b.BuildDecryptURL("msg", key)

	// Must use URL-safe characters.
	assert.Contains(t, got, "-", "expected URL-safe '-' from base64url encoding")
	assert.Contains(t, got, "_", "expected URL-safe '_' from base64url encoding")

	// Must NOT contain the standard base64 equivalents.
	// Strip the scheme and host so we only inspect the path/key segment.
	pathPart := strings.TrimPrefix(got, "https://example.com/decrypt/msg/")
	assert.NotContains(t, pathPart, "+", "standard base64 '+' must not appear in URL-safe encoded key")
	assert.NotContains(t, pathPart, "/", "standard base64 '/' must not appear in URL-safe encoded key")

	// Cross-check: decoding the extracted key segment must round-trip back to the original bytes.
	decoded, err := base64.URLEncoding.DecodeString(pathPart)
	require.NoError(t, err, "the key segment must be valid base64url")
	assert.Equal(t, key, decoded)
}

// TestBuildDecryptURL_KeySegmentIsDecodable asserts round-trip decodability for varied key contents.
func TestBuildDecryptURL_KeySegmentIsDecodable(t *testing.T) {
	t.Parallel()

	keys := [][]byte{
		{0x00},
		{0xff, 0xfe, 0xfd},
		[]byte("hello world"),
		make([]byte, 32), // all-zero 256-bit key
	}

	for _, key := range keys {
		t.Run(base64.URLEncoding.EncodeToString(key), func(t *testing.T) {
			t.Parallel()
			b := NewURLBuilder("https://example.com/")
			got := b.BuildDecryptURL("msg", key)

			// Extract the key segment: last path component after the final '/'.
			lastSlash := strings.LastIndex(got, "/")
			require.Greater(t, lastSlash, 0, "URL must contain at least one '/'")
			segment := got[lastSlash+1:]

			decoded, err := base64.URLEncoding.DecodeString(segment)
			require.NoError(t, err)
			assert.Equal(t, key, decoded)
		})
	}
}

func TestBuildE2EDecryptURL_ExactFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		baseURL   string
		messageID string
		e2eKey    string
		wantURL   string
	}{
		{
			name:      "non-empty e2eKey appends hash fragment",
			baseURL:   "https://example.com/",
			messageID: "abc123",
			e2eKey:    "someBase64UrlKey",
			wantURL:   "https://example.com/decrypt/abc123#key=someBase64UrlKey",
		},
		{
			name:      "empty e2eKey produces no fragment",
			baseURL:   "https://example.com/",
			messageID: "abc123",
			e2eKey:    "",
			wantURL:   "https://example.com/decrypt/abc123",
		},
		{
			name:      "base URL without trailing slash concatenates directly",
			baseURL:   "https://example.com",
			messageID: "abc123",
			e2eKey:    "",
			wantURL:   "https://example.comdecrypt/abc123",
		},
		{
			name:      "base URL without trailing slash with e2eKey appends fragment",
			baseURL:   "https://example.com",
			messageID: "abc123",
			e2eKey:    "myKey",
			wantURL:   "https://example.comdecrypt/abc123#key=myKey",
		},
		{
			name:      "empty message ID is preserved",
			baseURL:   "https://example.com/",
			messageID: "",
			e2eKey:    "k",
			wantURL:   "https://example.com/decrypt/#key=k",
		},
		{
			name:      "empty message ID with no e2eKey",
			baseURL:   "https://example.com/",
			messageID: "",
			e2eKey:    "",
			wantURL:   "https://example.com/decrypt/",
		},
		{
			name:      "e2eKey with equals sign and URL-safe chars is not further encoded",
			baseURL:   "https://example.com/",
			messageID: "xyz",
			e2eKey:    "abc-_==",
			wantURL:   "https://example.com/decrypt/xyz#key=abc-_==",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b := NewURLBuilder(tc.baseURL)
			got := b.BuildE2EDecryptURL(tc.messageID, tc.e2eKey)
			assert.Equal(t, tc.wantURL, got)
		})
	}
}

// TestBuildE2EDecryptURL_FragmentPrefixIsHashKey verifies the separator is "#key=", not "?key=" or "#" alone.
func TestBuildE2EDecryptURL_FragmentPrefixIsHashKey(t *testing.T) {
	t.Parallel()
	b := NewURLBuilder("https://example.com/")
	got := b.BuildE2EDecryptURL("msg-1", "THEKEY")

	assert.True(t, strings.Contains(got, "#key=THEKEY"),
		"fragment must be exactly '#key=<e2eKey>', got: %s", got)
	assert.False(t, strings.Contains(got, "?key="),
		"fragment separator must be '#', not '?', got: %s", got)
}

// TestBuildE2EDecryptURL_NoFragmentWhenKeyIsEmpty guards that an empty e2eKey produces no '#' in the URL.
func TestBuildE2EDecryptURL_NoFragmentWhenKeyIsEmpty(t *testing.T) {
	t.Parallel()
	b := NewURLBuilder("https://example.com/")
	got := b.BuildE2EDecryptURL("msg-1", "")

	assert.False(t, strings.Contains(got, "#"),
		"URL must not contain a fragment when e2eKey is empty, got: %s", got)
}
