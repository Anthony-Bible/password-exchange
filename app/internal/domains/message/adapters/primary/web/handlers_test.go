package web

import (
	"context"
	"net/http"
	"net/http/httptest"
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
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "uuid", Value: messageID},
				{Key: "key", Value: "/somekey"},
			}

			// Mock the HTML method to prevent template loading
			c.Header("Content-Type", "text/html; charset=utf-8")
			
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
	c, _ := gin.CreateTestContext(w)
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