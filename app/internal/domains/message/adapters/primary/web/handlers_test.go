package web

import (
	"bytes"
	"context"
	"html/template"
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

// MockMessageService is a mock implementation of the message service
type MockMessageService struct {
	mock.Mock
}

func (m *MockMessageService) SubmitMessage(
	ctx context.Context,
	req domain.MessageSubmissionRequest,
) (*domain.MessageSubmissionResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.MessageSubmissionResponse), args.Error(1)
}

func (m *MockMessageService) CheckMessageAccess(
	ctx context.Context,
	messageID string,
) (*domain.MessageAccessInfo, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MessageAccessInfo), args.Error(1)
}

func (m *MockMessageService) RetrieveMessage(
	ctx context.Context,
	req domain.MessageRetrievalRequest,
) (*domain.MessageRetrievalResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MessageRetrievalResponse), args.Error(1)
}

func TestDisplayDecrypted_ShouldNotCallRetrieveMessage(t *testing.T) {
	// This test verifies the fix: DisplayDecrypted should NOT call RetrieveMessage
	// regardless of whether a passphrase is required or not

	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name               string
		requiresPassphrase bool
	}{
		{"NoPassphraseRequired", false},
		{"PassphraseRequired", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockMessageService)
			handler := NewMessageHandler(mockService)

			messageID := "test-message-id"
			mockService.On("CheckMessageAccess", mock.Anything, messageID).Return(&domain.MessageAccessInfo{
				Exists:             true,
				RequiresPassphrase: tc.requiresPassphrase,
			}, nil)

			// Create a test context directly without routing
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/decrypt/"+messageID+"/somekey", nil)

			// Create gin engine with mock templates
			gin.SetMode(gin.TestMode)
			engine := gin.New()
			engine.SetHTMLTemplate(createMockTemplate())

			c := gin.CreateTestContextOnly(w, engine)
			c.Request = req
			c.Params = gin.Params{
				{Key: "uuid", Value: messageID},
				{Key: "key", Value: "/somekey"},
			}

			// Call the handler directly
			handler.DisplayDecrypted(c)

			// The key assertion: RetrieveMessage should NEVER be called during GET request
			mockService.AssertNotCalled(t, "RetrieveMessage", mock.Anything, mock.Anything)

			// Verify CheckMessageAccess was called
			mockService.AssertCalled(t, "CheckMessageAccess", mock.Anything, messageID)

			mockService.AssertExpectations(t)
		})
	}
}

func TestDisplayDecrypted_MessageNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	messageID := "non-existent-message"
	mockService.On("CheckMessageAccess", mock.Anything, messageID).Return(&domain.MessageAccessInfo{
		Exists:             false,
		RequiresPassphrase: false,
	}, nil)

	// Create a test context directly
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/decrypt/"+messageID+"/somekey", nil)

	// Create gin engine with mock templates
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.SetHTMLTemplate(createMockTemplate())

	c := gin.CreateTestContextOnly(w, engine)
	c.Request = req
	c.Params = gin.Params{
		{Key: "uuid", Value: messageID},
		{Key: "key", Value: "/somekey"},
	}

	// Call the handler directly
	handler.DisplayDecrypted(c)

	// Should return 404 for non-existent message
	assert.Equal(t, http.StatusNotFound, w.Code)
	// Should not call RetrieveMessage for non-existent message
	mockService.AssertNotCalled(t, "RetrieveMessage", mock.Anything, mock.Anything)

	mockService.AssertExpectations(t)
}

func TestSubmitMessage_MaxViewCountValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name               string
		maxViewCountValue  string
		expectedStatusCode int
		expectServiceCall  bool
		expectedErrorField string
	}{
		{
			name:               "ValidMaxViewCount",
			maxViewCountValue:  "5",
			expectedStatusCode: http.StatusOK, // gin doesn't properly redirect in test mode
			expectServiceCall:  true,
		},
		{
			name:               "ValidMaxViewCountMinimum",
			maxViewCountValue:  "1",
			expectedStatusCode: http.StatusOK,
			expectServiceCall:  true,
		},
		{
			name:               "ValidMaxViewCountMaximum",
			maxViewCountValue:  "100",
			expectedStatusCode: http.StatusOK,
			expectServiceCall:  true,
		},
		{
			name:               "EmptyMaxViewCount",
			maxViewCountValue:  "",
			expectedStatusCode: http.StatusOK,
			expectServiceCall:  true,
		},
		{
			name:               "InvalidNonNumericMaxViewCount",
			maxViewCountValue:  "abc",
			expectedStatusCode: http.StatusBadRequest,
			expectServiceCall:  false,
			expectedErrorField: "max_view_count",
		},
		{
			name:               "InvalidZeroMaxViewCount",
			maxViewCountValue:  "0",
			expectedStatusCode: http.StatusBadRequest,
			expectServiceCall:  false,
			expectedErrorField: "max_view_count",
		},
		{
			name:               "InvalidNegativeMaxViewCount",
			maxViewCountValue:  "-1",
			expectedStatusCode: http.StatusBadRequest,
			expectServiceCall:  false,
			expectedErrorField: "max_view_count",
		},
		{
			name:               "InvalidTooLargeMaxViewCount",
			maxViewCountValue:  "101",
			expectedStatusCode: http.StatusBadRequest,
			expectServiceCall:  false,
			expectedErrorField: "max_view_count",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockMessageService)
			handler := NewMessageHandler(mockService)

			if tc.expectServiceCall {
				expectedMaxViewCount := 0
				if tc.maxViewCountValue != "" {
					switch tc.maxViewCountValue {
					case "1":
						expectedMaxViewCount = 1
					case "5":
						expectedMaxViewCount = 5
					case "100":
						expectedMaxViewCount = 100
					}
				}

				mockService.On("SubmitMessage", mock.Anything, mock.MatchedBy(func(req domain.MessageSubmissionRequest) bool {
					return req.MaxViewCount == expectedMaxViewCount
				})).
					Return(&domain.MessageSubmissionResponse{
						MessageID:  "test-id",
						DecryptURL: "http://example.com/decrypt/test-id/key",
					}, nil)
			}

			// Create form data
			formData := url.Values{}
			formData.Set("content", "test message")
			if tc.maxViewCountValue != "" {
				formData.Set("max_view_count", tc.maxViewCountValue)
			}

			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/submit", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Create gin engine with mock templates
			engine := gin.New()
			engine.SetHTMLTemplate(createMockTemplate())

			c := gin.CreateTestContextOnly(w, engine)
			c.Request = req

			// Call the handler
			handler.SubmitMessage(c)

			// Verify response
			assert.Equal(t, tc.expectedStatusCode, w.Code)

			if tc.expectServiceCall {
				mockService.AssertCalled(t, "SubmitMessage", mock.Anything, mock.Anything)
			} else {
				mockService.AssertNotCalled(t, "SubmitMessage", mock.Anything, mock.Anything)

				// For validation errors, check that the response contains error information
				if tc.expectedErrorField != "" {
					responseBody := w.Body.String()
					assert.Contains(t, responseBody, "view count", "Response should contain view count error message")
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHTMLEndpoints_DefaultBrowserBehaviorReturnsHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	mockService.On("CheckMessageAccess", mock.Anything, "test-message-id").Return(&domain.MessageAccessInfo{
		Exists:             true,
		RequiresPassphrase: false,
	}, nil)

	router := gin.New()
	router.SetHTMLTemplate(createMockTemplate())
	router.GET("/", handler.Home)
	router.GET("/confirmation", handler.Confirmation)
	router.GET("/decrypt/:uuid/*key", handler.DisplayDecrypted)

	testCases := []struct {
		name         string
		path         string
		acceptHeader string
	}{
		{name: "HomeWithoutAcceptHeader", path: "/"},
		{name: "HomeWithBrowserAcceptHeader", path: "/", acceptHeader: "text/html"},
		{name: "ConfirmationWithBrowserAcceptHeader", path: "/confirmation?content=https://password.exchange/test", acceptHeader: "text/html"},
		{name: "DisplayDecryptedWithBrowserAcceptHeader", path: "/decrypt/test-message-id/somekey", acceptHeader: "text/html"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.path, nil)
			if tc.acceptHeader != "" {
				req.Header.Set("Accept", tc.acceptHeader)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/html"))
			assert.Equal(t, "Accept", w.Header().Get("Vary"))
		})
	}

	mockService.AssertExpectations(t)
}

func TestHTMLEndpoints_AcceptNegotiationContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	mockService.On("CheckMessageAccess", mock.Anything, "test-message-id").Return(&domain.MessageAccessInfo{
		Exists:             true,
		RequiresPassphrase: false,
	}, nil)

	router := gin.New()
	router.SetHTMLTemplate(createMockTemplate())
	router.GET("/", handler.Home)
	router.GET("/confirmation", handler.Confirmation)
	router.GET("/decrypt/:uuid/*key", handler.DisplayDecrypted)

	testCases := []struct {
		name                 string
		path                 string
		acceptHeader         string
		expectedContentType  string
		expectedBodyContains string
	}{
		{
			name:                 "MarkdownPreferredByDefaultWhenOnlyMarkdownPresent",
			path:                 "/",
			acceptHeader:         "text/markdown",
			expectedContentType:  "text/markdown",
			expectedBodyContains: "Share secrets securely.",
		},
		{
			name:                 "MarkdownSelectedWhenEqualToHTMLQuality",
			path:                 "/confirmation?content=https://password.exchange/test",
			acceptHeader:         "text/html;q=0.5, text/markdown;q=0.5",
			expectedContentType:  "text/markdown",
			expectedBodyContains: "Save this link carefully.",
		},
		{
			name:                "MarkdownNotSelectedWhenLowerThanHTMLQuality",
			path:                "/decrypt/test-message-id/somekey",
			acceptHeader:        "text/html;q=0.8, text/markdown;q=0.5",
			expectedContentType: "text/html",
		},
		{
			name:                "MarkdownQZeroRejected",
			path:                "/",
			acceptHeader:        "text/markdown;q=0",
			expectedContentType: "text/html",
		},
		{
			name:                "MarkdownQZeroPointZeroRejected",
			path:                "/",
			acceptHeader:        "text/markdown;q=0.0",
			expectedContentType: "text/html",
		},
		{
			name:                "MarkdownPositiveButLowerThanDefaultHTMLRejected",
			path:                "/",
			acceptHeader:        "text/markdown;q=0.9",
			expectedContentType: "text/html",
		},
		{
			name:                "MarkdownSelectedWhenHTMLExplicitlyZero",
			path:                "/",
			acceptHeader:        "text/html;q=0, text/markdown;q=0.1",
			expectedContentType: "text/markdown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.path, nil)
			req.Header.Set("Accept", tc.acceptHeader)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), tc.expectedContentType))
			assert.Equal(t, "Accept", w.Header().Get("Vary"))
			if tc.expectedContentType == "text/markdown" {
				assert.NotContains(t, strings.ToLower(w.Body.String()), "<html")
				if tc.expectedBodyContains != "" {
					assert.Contains(t, w.Body.String(), tc.expectedBodyContains)
				}
			}
		})
	}

	mockService.AssertExpectations(t)
}

func TestCaptureHTMLResponse_CapturesStatusAndRestoresWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.SetHTMLTemplate(createMockTemplate())

	responseRecorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/missing", nil)
	context := gin.CreateTestContextOnly(responseRecorder, engine)
	context.Request = req
	originalWriter := context.Writer

	htmlOutput, status := captureHTMLResponse(context, http.StatusNotFound, "404.html", gin.H{
		"Title": "Missing",
	})

	assert.Equal(t, http.StatusNotFound, status)
	assert.Contains(t, htmlOutput, "<h1>Missing</h1>")
	assert.Same(t, originalWriter, context.Writer)
	assert.Empty(t, responseRecorder.Body.String())
}

func TestMarkdownCaptureWriter_TracksStatusAndBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	responseRecorder := httptest.NewRecorder()
	context := gin.CreateTestContextOnly(responseRecorder, engine)

	writer := &markdownCaptureWriter{
		ResponseWriter: context.Writer,
		body:           &bytes.Buffer{},
		status:         http.StatusOK,
		size:           -1,
	}

	assert.False(t, writer.Written())
	assert.Equal(t, -1, writer.Size())

	writer.WriteHeader(http.StatusTeapot)
	assert.Equal(t, http.StatusTeapot, writer.Status())

	n, err := writer.WriteString("hello")
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.True(t, writer.Written())
	assert.Equal(t, 5, writer.Size())
	assert.Equal(t, "hello", writer.body.String())
}

func TestMarkdownCaptureWriter_StreamingMethodsDoNotLeak(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	responseRecorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(responseRecorder, engine)

	writer := &markdownCaptureWriter{
		ResponseWriter: ctx.Writer,
		body:           &bytes.Buffer{},
		status:         http.StatusOK,
		size:           -1,
	}

	assert.NotPanics(t, func() { writer.Flush() })
	assert.Empty(t, responseRecorder.Body.String(), "Flush must not push bytes to underlying writer")

	ch := writer.CloseNotify()
	assert.NotNil(t, ch)

	assert.Nil(t, writer.Pusher())

	conn, brw, err := writer.Hijack()
	assert.Nil(t, conn)
	assert.Nil(t, brw)
	assert.ErrorIs(t, err, http.ErrNotSupported)
}

func TestDisplayDecryptedMarkdown_RendersInstructions(t *testing.T) {
	testCases := []struct {
		name               string
		requiresPassphrase bool
		expectedSuffix     string
	}{
		{"PassphraseRequired", true, "- requires_passphrase: true\n"},
		{"NoPassphraseRequired", false, "- requires_passphrase: false\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := displayDecryptedMarkdown(gin.H{
				"Title":       "passwordExchange Decrypted",
				"HasPassword": tc.requiresPassphrase,
			})
			assert.Contains(t, body, "# Decrypt Message")
			assert.Contains(t, body, "POST to this URL")
			assert.True(t, strings.HasSuffix(body, tc.expectedSuffix),
				"body should end with %q, got %q", tc.expectedSuffix, body)
		})
	}
}

func TestDecryptMessageMarkdown_RendersFencedContent(t *testing.T) {
	body := decryptMessageMarkdown(gin.H{
		"DecryptedMessage": "hello\nworld",
		"ViewCount":        2,
		"MaxViewCount":     5,
	})

	assert.Contains(t, body, "# Decrypted message")
	assert.Contains(t, body, "```\nhello\nworld\n```")
	assert.Contains(t, body, "- view_count: 2")
	assert.Contains(t, body, "- max_view_count: 5")
}

func TestDecryptMessageMarkdown_WrongPassphrase(t *testing.T) {
	body := decryptMessageMarkdown(gin.H{
		"DecryptedMessage": wrongPassphraseSentinel,
	})
	assert.True(t, strings.HasPrefix(body, "# Decryption failed"))
	assert.Contains(t, body, "Wrong passphrase")
	assert.NotContains(t, body, "```")
}

func TestDecryptMessageMarkdown_EmptyContent(t *testing.T) {
	body := decryptMessageMarkdown(gin.H{"DecryptedMessage": ""})
	assert.Equal(t, "# Decrypted message\n\n(empty)\n", body)
}

func TestPickFence_AvoidsBacktickCollision(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		want    string
	}{
		{"NoBackticks", "plain text", "```"},
		{"SingleBacktick", "a `b` c", "```"},
		{"TripleBacktick", "code: ```bash\necho hi\n```", "````"},
		{"QuintupleBacktick", "five: `````", "``````"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, pickFence(tc.content))
		})
	}
}

func TestDecryptMessage_MarkdownPathReturnsBuilderOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	mockService.On("RetrieveMessage", mock.Anything, mock.Anything).
		Return(&domain.MessageRetrievalResponse{
			Content:      "supersecret",
			ViewCount:    1,
			MaxViewCount: 3,
		}, nil)

	router := gin.New()
	router.SetHTMLTemplate(createMockTemplate())
	router.POST("/decrypt/:uuid/*key", handler.DecryptMessage)

	w := httptest.NewRecorder()
	form := url.Values{}
	form.Set("passphrase", "secret")
	req, _ := http.NewRequest("POST", "/decrypt/abc/Zm9v", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/markdown")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/markdown"))
	body := w.Body.String()
	assert.Contains(t, body, "# Decrypted message")
	assert.Contains(t, body, "supersecret")
	assert.Contains(t, body, "- view_count: 1")
	assert.NotContains(t, body, "<html")
}

func TestDecryptMessage_MarkdownPathReturnsWrongPassphrase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	mockService.On("RetrieveMessage", mock.Anything, mock.Anything).
		Return(nil, domain.ErrInvalidPassphrase)

	router := gin.New()
	router.SetHTMLTemplate(createMockTemplate())
	router.POST("/decrypt/:uuid/*key", handler.DecryptMessage)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/decrypt/abc/Zm9v", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/markdown")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/markdown"))
	assert.True(t, strings.HasPrefix(w.Body.String(), "# Decryption failed"))
}

func TestDisplayDecrypted_MarkdownPathReturnsInstructions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	mockService.On("CheckMessageAccess", mock.Anything, "abc").
		Return(&domain.MessageAccessInfo{Exists: true, RequiresPassphrase: true}, nil)

	router := gin.New()
	router.SetHTMLTemplate(createMockTemplate())
	router.GET("/decrypt/:uuid/*key", handler.DisplayDecrypted)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/decrypt/abc/Zm9v", nil)
	req.Header.Set("Accept", "text/markdown")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/markdown"))
	body := w.Body.String()
	assert.Contains(t, body, "# Decrypt Message")
	assert.Contains(t, body, "requires_passphrase: true")
	assert.NotContains(t, body, "<html")
}

func TestMarkdownPath_StripsScriptsAndStyles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	tmpl := template.New("templates")
	tmpl, _ = tmpl.New("home.html").Parse(
		`<html><head><style>:root { --x: 1px; }</style></head><body>` +
			`<h1>{{.Title}}</h1><p>Share secrets securely.</p>` +
			`<script>function leak(){var x=1;}</script>` +
			`</body></html>`,
	)

	router := gin.New()
	router.SetHTMLTemplate(tmpl)
	router.GET("/", handler.Home)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/markdown")

	router.ServeHTTP(w, req)

	body := w.Body.String()
	assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/markdown"))
	assert.Contains(t, body, "Share secrets securely")
	assert.NotContains(t, body, "function leak")
	assert.NotContains(t, body, ":root")
	assert.NotContains(t, body, "--x:")
}

// createMockTemplate creates a simple mock template for testing
func createMockTemplate() *template.Template {
	tmpl := template.New("templates")
	tmpl, _ = tmpl.New("decryption.html").
		Parse(`<html><body><h1>{{.Title}}</h1><p>HasPassword: {{.HasPassword}}</p></body></html>`)
	tmpl, _ = tmpl.New("404.html").Parse(`<html><body><h1>{{.Title}}</h1><p>404 Not Found</p></body></html>`)
	tmpl, _ = tmpl.New("home.html").
		Parse(`<html><body><h1>{{.Title}}</h1><p>Share secrets securely.</p>{{range $key, $value := .Errors}}<div class="error">{{$key}}: {{$value}}</div>{{end}}</body></html>`)
	tmpl, _ = tmpl.New("confirmation.html").Parse(`<html><body><h1>{{.Title}}</h1><p>URL: {{.Url}}</p><p>Save this link carefully.</p></body></html>`)
	return tmpl
}
