package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// Binding decision (epic #562, open question 4): strings that Postgres would
// reject (over varchar(255), NUL, invalid UTF-8) are rejected at the API layer
// with a 400, and never reach the file service (and so never the database).

func initiateBody(filename, contentType, messageID string) []byte {
	// Hand-built JSON so NUL can be sent as \u0000 and raw invalid bytes can be injected.
	esc := func(s string) string {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		var b strings.Builder
		for _, r := range s {
			if r < 0x20 {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
		return b.String()
	}
	return []byte(`{"filename":"` + esc(filename) + `","contentType":"` + esc(contentType) +
		`","totalSize":10,"chunkSize":4,"messageID":"` + esc(messageID) + `"}`)
}

func postInitiate(router http.Handler, body []byte) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/files/initiate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func okInitiateResponse() *domain.InitiateUploadResponse {
	return &domain.InitiateUploadResponse{FileID: "f", SessionID: "s", EncryptionKey: make([]byte, 32)}
}

func TestInitiateUploadHandler_RejectsStorageUnsafeStrings(t *testing.T) {
	cases := map[string][3]string{
		"filename 256 chars":     {strings.Repeat("a", 256), "text/plain", "msg"},
		"content type 256 chars": {"a.txt", strings.Repeat("c", 256), "msg"},
		"message id 256 chars":   {"a.txt", "text/plain", strings.Repeat("m", 256)},
		"filename 256 4-byte":    {strings.Repeat("😀", 256), "text/plain", "msg"},
		"filename NUL":           {"a\x00.txt", "text/plain", "msg"},
		"content type NUL":       {"a.txt", "text/\x00plain", "msg"},
		"message id NUL":         {"a.txt", "text/plain", "m\x00sg"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			service := new(MockFileService)
			router := setupFileTestRouter(service)

			w := postInitiate(router, initiateBody(c[0], c[1], c[2]))

			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), "validation_failed")
			service.AssertNotCalled(t, "InitiateUpload", mock.Anything, mock.Anything)
		})
	}
}

// The limit is counted in characters (Postgres varchar), not bytes: 255
// four-byte runes are 1020 bytes and must still be accepted.
func TestInitiateUploadHandler_AcceptsExactly255CharactersEvenWhenMultibyte(t *testing.T) {
	for name, s := range map[string]string{"ascii": strings.Repeat("a", 255), "4-byte runes": strings.Repeat("😀", 255)} {
		t.Run(name, func(t *testing.T) {
			service := new(MockFileService)
			service.On("InitiateUpload", mock.Anything, mock.Anything).Return(okInitiateResponse(), nil).Once()
			router := setupFileTestRouter(service)

			w := postInitiate(router, initiateBody(s, s, s))

			assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())
			service.AssertExpectations(t)
		})
	}
}

// Property: the request is accepted iff the filename has at most 255 characters.
func TestInitiateUploadHandler_Property_FilenameAcceptedIffWithin255Characters(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		name := rapid.StringOfN(rapid.Rune().Filter(func(r rune) bool { return r != 0 }), 1, 300, -1).
			Filter(func(s string) bool { return s != "" }).Draw(rt, "filename")
		service := new(MockFileService)
		service.On("InitiateUpload", mock.Anything, mock.Anything).Return(okInitiateResponse(), nil).Maybe()
		router := setupFileTestRouter(service)

		w := postInitiate(router, initiateBody(name, "text/plain", "msg"))

		if utf8.RuneCountInString(name) <= 255 {
			assert.Equal(rt, http.StatusCreated, w.Code)
		} else {
			assert.Equal(rt, http.StatusBadRequest, w.Code)
		}
	})
}

// Whatever the client sends (including raw invalid UTF-8 bytes), nothing the
// file service receives may violate a Postgres constraint.
func TestInitiateUploadHandler_Property_ServiceNeverReceivesStorageUnsafeStrings(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		raw := rapid.StringN(0, 300, -1).Draw(rt, "rawFilename")
		var mu sync.Mutex
		var seen []domain.InitiateUploadRequest
		service := new(MockFileService)
		service.On("InitiateUpload", mock.Anything, mock.MatchedBy(func(r domain.InitiateUploadRequest) bool {
			mu.Lock()
			defer mu.Unlock()
			seen = append(seen, r)
			return true
		})).Return(okInitiateResponse(), nil).Maybe()
		router := setupFileTestRouter(service)

		// Inject raw bytes unescaped to simulate a client that is not valid UTF-8.
		body := []byte(`{"filename":"` + strings.NewReplacer(`\`, "_", `"`, "_", "\n", "_", "\r", "_", "\t", "_").Replace(raw) +
			`","contentType":"text/plain","totalSize":10,"chunkSize":4,"messageID":"msg"}`)
		postInitiate(router, body)

		mu.Lock()
		defer mu.Unlock()
		for _, r := range seen {
			require.True(rt, utf8.ValidString(r.Filename), "invalid UTF-8 reached the service: %q", r.Filename)
			assert.NotContains(rt, r.Filename, "\x00")
			assert.LessOrEqual(rt, utf8.RuneCountInString(r.Filename), 255)
		}
	})
}
