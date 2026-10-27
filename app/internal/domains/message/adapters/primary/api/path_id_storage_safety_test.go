package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"pgregory.net/rapid"
)

// Binding decision (epic #562): input Postgres would reject (NUL, invalid
// UTF-8) is refused at the API layer, never forwarded to the service/DB. This
// includes IDs taken from URL path params, which Gin percent-decodes. On
// MySQL these requests were a 404; on Postgres they would surface as 5xx.

// badPathIDs are percent-encoded so they survive URL construction and are
// decoded into NUL / invalid UTF-8 by the server before reaching handlers.
var badPathIDs = map[string]string{
	"NUL":           "abc%00def",
	"invalid UTF-8": "abc%FFdef",
	"lone 0xC0":     "%C0%80",
}

func assertClientRejection(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusNotFound}, w.Code,
		"storage-unsafe path ID must be a 400/404, never 5xx; body=%s", w.Body.String())
}

func TestMessageRoutes_RejectStorageUnsafePathIDs(t *testing.T) {
	routes := []struct {
		method, suffix, body string
	}{
		{http.MethodGet, "", ""},
		{http.MethodPost, "/decrypt", `{"passphrase":"x"}`},
		{http.MethodPost, "/notify", `{"recipientEmail":"a@example.com","decryptUrl":"https://example.com/decrypt/x/y"}`},
	}
	for _, r := range routes {
		for name, id := range badPathIDs {
			t.Run(r.method+r.suffix+"/"+name, func(t *testing.T) {
				service := new(MockMessageService)
				router := setupTestRouter(service)

				req, _ := http.NewRequest(r.method, "/api/v1/messages/"+id+r.suffix, strings.NewReader(r.body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assertClientRejection(t, w)
				service.AssertNotCalled(t, "CheckMessageAccess", mock.Anything, mock.Anything)
				service.AssertNotCalled(t, "RetrieveMessage", mock.Anything, mock.Anything)
				service.AssertNotCalled(t, "NotifyMessage", mock.Anything, mock.Anything)
			})
		}
	}
}

func TestFileRoutes_RejectStorageUnsafePathIDs(t *testing.T) {
	for name, id := range badPathIDs {
		t.Run("download/"+name, func(t *testing.T) {
			service := new(MockFileService)
			router := setupFileTestRouter(service)

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/files/"+id, nil)
			req.Header.Set("X-File-Key", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assertClientRejection(t, w)
			service.AssertNotCalled(t, "DownloadFile", mock.Anything, mock.Anything)
		})
		t.Run("chunks fileID/"+name, func(t *testing.T) {
			service := new(MockFileService)
			router := setupFileTestRouter(service)

			w := postChunk(router, id, "valid-session")

			assertClientRejection(t, w)
			service.AssertNotCalled(t, "UploadChunk", mock.Anything, mock.Anything)
		})
		// The upload session ID arrives in the multipart form body and is
		// looked up by GetUploadSession, so it needs the same guard.
		t.Run("chunks sessionID/"+name, func(t *testing.T) {
			service := new(MockFileService)
			router := setupFileTestRouter(service)

			raw, _ := url.PathUnescape(id)
			w := postChunk(router, "valid-file", raw)

			assertClientRejection(t, w)
			service.AssertNotCalled(t, "UploadChunk", mock.Anything, mock.Anything)
		})
	}
}

// postChunk sends an otherwise well-formed chunk upload. fileIDEscaped is
// placed in the URL verbatim; sessionID goes in the form body.
func postChunk(router http.Handler, fileIDEscaped, sessionID string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("sessionID", sessionID)
	_ = mw.WriteField("chunkIndex", "1")
	_ = mw.WriteField("totalChunks", "1")
	fw, _ := mw.CreateFormFile("data", "chunk.bin")
	_, _ = fw.Write([]byte("data"))
	_ = mw.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/files/"+fileIDEscaped+"/chunks", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// Property: any percent-encoded byte sequence that decodes to NUL-containing
// or non-UTF-8 text is never forwarded to the service and never yields 5xx.
func TestMessageGet_Property_UnsafeIDsNeverReachServiceNorReturn5xx(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		prefix := rapid.StringMatching(`[a-z0-9]{0,8}`).Draw(rt, "prefix")
		bad := rapid.SampledFrom([]string{"%00", "%FF", "%FE", "%C0", "%80", "%C3%28"}).Draw(rt, "bad")
		suffix := rapid.StringMatching(`[a-z0-9]{0,8}`).Draw(rt, "suffix")

		service := new(MockMessageService)
		router := setupTestRouter(service)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/messages/"+prefix+bad+suffix, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code >= 500 {
			rt.Fatalf("got %d for unsafe id", w.Code)
		}
		if len(service.Calls) != 0 {
			rt.Fatalf("service reached with unsafe id: %v", service.Calls)
		}
	})
}
