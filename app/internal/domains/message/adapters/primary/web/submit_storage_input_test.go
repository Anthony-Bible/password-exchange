package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Binding decision (epic #562): over-length, NUL and invalid UTF-8 input is
// rejected at the API layer with a 4xx before reaching the service/DB. The
// HTML form route must honour this for every stored form field, otherwise
// Postgres turns the input into a 5xx.

func postSubmitForm(t *testing.T, service *MockMessageService, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler := NewMessageHandler(service)
	engine := gin.New()
	engine.SetHTMLTemplate(createMockTemplate())
	engine.POST("/", handler.SubmitMessage)

	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func TestSubmitMessage_RejectsStorageUnsafeOrOverlongFormFields(t *testing.T) {
	fields := []string{"firstname", "email", "other_firstname", "other_email", "other_lastname", "other_information"}
	bad := map[string]string{
		"NUL":                "a\x00@b.com",
		"invalid UTF-8":      "a\xff@b.com",
		"256-char over-long": strings.Repeat("a", 256-len("@b.com")+1) + "@b.com",
	}
	// Only other_email is length-capped with certainty (varchar(255)); the
	// remaining fields must at least reject NUL and invalid UTF-8.
	for _, field := range fields {
		for name, value := range bad {
			if strings.HasPrefix(name, "256") && field != "other_email" {
				continue
			}
			t.Run(field+"/"+name, func(t *testing.T) {
				service := new(MockMessageService)
				service.On("GetDefaultMaxViewCount").Return(5).Maybe()
				service.On("SubmitMessage", mock.Anything, mock.Anything).Return((*domain.MessageSubmissionResponse)(nil), assert.AnError).Maybe()

				form := url.Values{}
				form.Set("content", "test message")
				form.Set("enableEmail", "on")
				form.Set(field, value)

				w := postSubmitForm(t, service, form)

				assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
				service.AssertNotCalled(t, "SubmitMessage", mock.Anything, mock.Anything)
			})
		}
	}
}

func TestSubmitMessage_AcceptsRecipientEmailAtExactly255Characters(t *testing.T) {
	service := new(MockMessageService)
	service.On("GetDefaultMaxViewCount").Return(5).Maybe()
	service.On("SubmitMessage", mock.Anything, mock.Anything).Return((*domain.MessageSubmissionResponse)(nil), assert.AnError).Maybe()

	email := strings.Repeat("a", 255-len("@b.com")) + "@b.com"
	assert.Len(t, email, 255)
	form := url.Values{}
	form.Set("content", "test message")
	form.Set("other_email", email)

	w := postSubmitForm(t, service, form)

	assert.NotEqual(t, http.StatusBadRequest, w.Code, "255-char email is within the limit and must not be rejected as invalid input")
	service.AssertCalled(t, "SubmitMessage", mock.Anything, mock.Anything)
}

func TestSubmitMessage_CountsRunesNotBytesForRecipientEmail(t *testing.T) {
	service := new(MockMessageService)
	service.On("GetDefaultMaxViewCount").Return(5).Maybe()
	service.On("SubmitMessage", mock.Anything, mock.Anything).Return((*domain.MessageSubmissionResponse)(nil), assert.AnError).Maybe()

	// 250 multibyte runes + "@b.com" (6) = 256 runes? use 249 -> 255 runes, >255 bytes.
	email := strings.Repeat("é", 249) + "@b.com"
	form := url.Values{}
	form.Set("content", "test message")
	form.Set("other_email", email)

	w := postSubmitForm(t, service, form)

	assert.NotEqual(t, http.StatusBadRequest, w.Code, "length cap is 255 characters (runes), not bytes")
	service.AssertCalled(t, "SubmitMessage", mock.Anything, mock.Anything)
}
