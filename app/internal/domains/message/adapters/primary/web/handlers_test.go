package web

import (
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

func (m *MockMessageService) SubmitMessage(ctx context.Context, req domain.MessageSubmissionRequest) (*domain.MessageSubmissionResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.MessageSubmissionResponse), args.Error(1)
}

func (m *MockMessageService) CheckMessageAccess(ctx context.Context, messageID string) (*domain.MessageAccessInfo, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MessageAccessInfo), args.Error(1)
}

func (m *MockMessageService) RetrieveMessage(ctx context.Context, req domain.MessageRetrievalRequest) (*domain.MessageRetrievalResponse, error) {
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
		name                string
		maxViewCountValue   string
		expectedStatusCode  int
		expectServiceCall   bool
		expectedErrorField  string
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
				})).Return(&domain.MessageSubmissionResponse{
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

// createMockTemplate creates a simple mock template for testing
func createMockTemplate() *template.Template {
	tmpl := template.New("templates")
	tmpl, _ = tmpl.New("decryption.html").Parse(`<html><body><h1>{{.Title}}</h1><p>HasPassword: {{.HasPassword}}</p></body></html>`)
	tmpl, _ = tmpl.New("404.html").Parse(`<html><body><h1>{{.Title}}</h1><p>404 Not Found</p></body></html>`)
	tmpl, _ = tmpl.New("home.html").Parse(`<html><body><h1>{{.Title}}</h1>{{range $key, $value := .Errors}}<div class="error">{{$key}}: {{$value}}</div>{{end}}</body></html>`)
	tmpl, _ = tmpl.New("confirmation.html").Parse(`<html><body><h1>{{.Title}}</h1><p>URL: {{.Url}}</p></body></html>`)
	return tmpl
}