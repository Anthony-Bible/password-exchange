package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessageService is a mock implementation of MessageServicePort
type MockMessageService struct {
	mock.Mock
}

func (m *MockMessageService) SubmitMessage(ctx context.Context, req domain.MessageSubmissionRequest) (*domain.MessageSubmissionResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.MessageSubmissionResponse), args.Error(1)
}

func (m *MockMessageService) CheckMessageAccess(ctx context.Context, messageID string) (*domain.MessageAccessInfo, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).(*domain.MessageAccessInfo), args.Error(1)
}

func (m *MockMessageService) RetrieveMessage(ctx context.Context, req domain.MessageRetrievalRequest) (*domain.MessageRetrievalResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.MessageRetrievalResponse), args.Error(1)
}

func setupTestRouter(mockService *MockMessageService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return setupRouter(NewMessageAPIHandler(mockService))
}

func TestSubmitMessage_Success(t *testing.T) {
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	// Setup mock expectations
	expectedDomainReq := domain.MessageSubmissionRequest{
		Content:          "Test message",
		SenderName:       "John Doe",
		SenderEmail:      "john@example.com",
		RecipientName:    "Jane Doe",
		RecipientEmail:   "jane@example.com",
		Passphrase:       "test123",
		AdditionalInfo:   "Additional info",
		SendNotification: true,
		SkipEmail:        false,
	}

	expectedResponse := &domain.MessageSubmissionResponse{
		MessageID:  "test-message-id",
		DecryptURL: "https://example.com/decrypt/test-message-id/key123",
		Success:    true,
	}

	mockService.On("SubmitMessage", mock.Anything, expectedDomainReq).Return(expectedResponse, nil)

	// Prepare request
	requestBody := models.MessageSubmissionRequest{
		Content: "Test message",
		Sender: &models.Sender{
			Name:  "John Doe",
			Email: "john@example.com",
		},
		Recipient: &models.Recipient{
			Name:  "Jane Doe",
			Email: "jane@example.com",
		},
		Passphrase:       "test123",
		AdditionalInfo:   "Additional info",
		SendNotification: true,
		AntiSpamAnswer:   "blue",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/messages", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.MessageSubmissionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-message-id", response.MessageID)
	assert.Equal(t, "https://example.com/decrypt/test-message-id/key123", response.DecryptURL)
	assert.True(t, response.NotificationSent)

	mockService.AssertExpectations(t)
}

func TestSubmitMessage_ValidationError(t *testing.T) {
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	// Request with missing required fields for notification
	requestBody := models.MessageSubmissionRequest{
		Content:          "Test message",
		SendNotification: true,
		AntiSpamAnswer:   "blue",
		// Missing sender and recipient
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/messages", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse models.StandardErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, models.ErrorCodeValidationFailed, errorResponse.Error)
	assert.Contains(t, errorResponse.Details, "validation_errors")
}

func TestGetMessageInfo_Success(t *testing.T) {
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	// Setup mock expectations
	expectedAccessInfo := &domain.MessageAccessInfo{
		MessageID:          "test-message-id",
		Exists:             true,
		RequiresPassphrase: true,
	}

	mockService.On("CheckMessageAccess", mock.Anything, "test-message-id").Return(expectedAccessInfo, nil)

	req, _ := http.NewRequest("GET", "/api/v1/messages/test-message-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.MessageAccessInfoResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-message-id", response.MessageID)
	assert.True(t, response.Exists)
	assert.True(t, response.RequiresPassphrase)

	mockService.AssertExpectations(t)
}

func TestGetMessageInfo_NotFound(t *testing.T) {
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	// Setup mock expectations
	expectedAccessInfo := &domain.MessageAccessInfo{
		MessageID:          "test-message-id",
		Exists:             false,
		RequiresPassphrase: false,
	}

	mockService.On("CheckMessageAccess", mock.Anything, "test-message-id").Return(expectedAccessInfo, nil)

	req, _ := http.NewRequest("GET", "/api/v1/messages/test-message-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResponse models.StandardErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMessageNotFound, errorResponse.Error)

	mockService.AssertExpectations(t)
}

func TestHealthCheck(t *testing.T) {
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "1.0.0", response.Version)
	assert.Contains(t, response.Services, "database")
	assert.Contains(t, response.Services, "encryption")
	assert.Contains(t, response.Services, "email")
}

func TestAPIInfo(t *testing.T) {
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIInfoResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", response.Version)
	assert.Contains(t, response.Endpoints, "submit")
	assert.Contains(t, response.Endpoints, "access")
	assert.Contains(t, response.Endpoints, "decrypt")
	assert.True(t, response.Features["emailNotifications"])
	assert.True(t, response.Features["passphraseProtection"])
	assert.True(t, response.Features["antiSpamProtection"])
}