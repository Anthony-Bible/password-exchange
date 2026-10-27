package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Binding decision (epic #562): storage-unsafe input (NUL, invalid UTF-8) is
// rejected before reaching the service/DB. The web decrypt routes take the
// message ID from a percent-decoded path param; on Postgres an unguarded ID
// would become a 5xx, where MySQL returned the not-found page.

func TestDecryptRoutes_RejectStorageUnsafeMessageIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	bad := map[string]string{
		"NUL":           "abc%00def",
		"invalid UTF-8": "abc%FFdef",
	}
	for name, id := range bad {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			t.Run(method+"/"+name, func(t *testing.T) {
				service := new(MockMessageService)
				handler := NewMessageHandler(service)
				router := gin.New()
				router.SetHTMLTemplate(createMockTemplate())
				router.GET("/decrypt/:uuid/*key", handler.DisplayDecrypted)
				router.POST("/decrypt/:uuid/*key", handler.DecryptMessage)

				req, _ := http.NewRequest(method, "/decrypt/"+id+"/Zm9v", strings.NewReader(""))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Contains(t, []int{http.StatusBadRequest, http.StatusNotFound}, w.Code, w.Body.String())
				service.AssertNotCalled(t, "CheckMessageAccess", mock.Anything, mock.Anything)
				service.AssertNotCalled(t, "RetrieveMessage", mock.Anything, mock.Anything)
			})
		}
	}
}
