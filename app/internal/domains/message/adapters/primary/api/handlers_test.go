package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/middleware"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessageService is a mock implementation of MessageServicePort
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
	return args.Get(0).(*domain.MessageAccessInfo), args.Error(1)
}

func (m *MockMessageService) RetrieveMessage(
	ctx context.Context,
	req domain.MessageRetrievalRequest,
) (*domain.MessageRetrievalResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.MessageRetrievalResponse), args.Error(1)
}

func setupTestRouter(mockService *MockMessageService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Create minimal metrics setup for testing
	registry := prometheus.NewRegistry()
	metrics := middleware.NewPrometheusMetrics(registry)

	return setupRouter(NewMessageAPIHandler(mockService), metrics, registry)
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
		Captcha:          "blue",
		SendNotification: true,
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
		Content: "Test message",
		Recipient: &models.Recipient{
			Name: "Jane Smith",
		},
		SendNotification: true,
		AntiSpamAnswer:   "blue",
		// Missing sender
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
	assert.Contains(t, errorResponse.Details, "sender")
	assert.Contains(t, errorResponse.Details, "recipient.email")
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

func TestSubmitMessage_WithMaxViewCount(t *testing.T) {
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	// Setup mock expectations - should receive the MaxViewCount in domain request
	expectedDomainReq := domain.MessageSubmissionRequest{
		Content:          "Test message with custom view count",
		SenderName:       "John Doe",
		SenderEmail:      "john@example.com",
		RecipientName:    "Jane Doe",
		RecipientEmail:   "jane@example.com",
		SendNotification: true,
		Captcha:          "blue",
		MaxViewCount:     25, // Custom view count
	}

	expectedResponse := &domain.MessageSubmissionResponse{
		MessageID:  "test-message-id",
		DecryptURL: "https://example.com/decrypt/test-message-id/key123",
		Success:    true,
	}

	mockService.On("SubmitMessage", mock.Anything, expectedDomainReq).Return(expectedResponse, nil)

	// Prepare request with maxViewCount
	requestBody := models.MessageSubmissionRequest{
		Content: "Test message with custom view count",
		Sender: &models.Sender{
			Name:  "John Doe",
			Email: "john@example.com",
		},
		Recipient: &models.Recipient{
			Name:  "Jane Doe",
			Email: "jane@example.com",
		},
		SendNotification: true,
		AntiSpamAnswer:   "blue",
		MaxViewCount:     25, // Custom view count
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/messages", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetMessageInfo_NilExpiresAtIsNullInResponse(t *testing.T) {
	// When domain returns nil ExpiresAt (legacy data), the API must return "expiresAt": null,
	// NOT a fabricated time.Now()+TTL which would be semantically wrong.
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	expectedAccessInfo := &domain.MessageAccessInfo{
		MessageID:          "test-message-id",
		Exists:             true,
		RequiresPassphrase: false,
		ExpiresAt:          nil, // legacy message — no expiry stored in DB
	}

	mockService.On("CheckMessageAccess", mock.Anything, "test-message-id").Return(expectedAccessInfo, nil)

	req, _ := http.NewRequest("GET", "/api/v1/messages/test-message-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check raw JSON — expiresAt must be null, not a fabricated timestamp
	var raw map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &raw)
	assert.NoError(t, err)
	expiresAt, hasKey := raw["expiresAt"]
	assert.True(t, hasKey, "expiresAt key must be present in response")
	assert.Nil(t, expiresAt, "expiresAt must be null when domain has no expiry, not a fabricated time")

	mockService.AssertExpectations(t)
}

func TestSubmitMessage_NilExpiresAtIsNullInResponse(t *testing.T) {
	// When domain returns nil ExpiresAt on submission, the API must return "expiresAt": null,
	// NOT time.Now()+TTL which is dead-code fallback that could mislead the caller.
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	expectedDomainReq := domain.MessageSubmissionRequest{
		Content:          "Test message",
		SenderName:       "John Doe",
		SenderEmail:      "john@example.com",
		RecipientName:    "Jane Doe",
		RecipientEmail:   "jane@example.com",
		SendNotification: true,
		Captcha:          "blue",
	}
	expectedResponse := &domain.MessageSubmissionResponse{
		MessageID:  "test-message-id",
		DecryptURL: "https://example.com/decrypt/test-message-id/key123",
		ExpiresAt:  nil, // domain somehow returned no expiry
		Success:    true,
	}

	mockService.On("SubmitMessage", mock.Anything, expectedDomainReq).Return(expectedResponse, nil)

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
		SendNotification: true,
		AntiSpamAnswer:   "blue",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/messages", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Check raw JSON — expiresAt must be null, not a fabricated timestamp
	var raw map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &raw)
	assert.NoError(t, err)
	expiresAt, hasKey := raw["expiresAt"]
	assert.True(t, hasKey, "expiresAt key must be present in response")
	assert.Nil(t, expiresAt, "expiresAt must be null when domain returns nil, not a fabricated time")

	mockService.AssertExpectations(t)
}

func TestSubmitMessage_MaxViewCountValidation(t *testing.T) {
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	testCases := []struct {
		name           string
		maxViewCount   int
		expectedStatus int
	}{
		{"ValidLow", 1, http.StatusCreated},
		{"ValidMid", 50, http.StatusCreated},
		{"ValidHigh", 100, http.StatusCreated},
		{"InvalidZero", 0, http.StatusCreated}, // 0 should be valid (use default)
		{"InvalidNegative", -1, http.StatusBadRequest},
		{"InvalidTooHigh", 101, http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Only set up mock expectation for valid cases
			if tc.expectedStatus == http.StatusCreated {
				expectedDomainReq := domain.MessageSubmissionRequest{
					Content:          "Test message",
					SenderName:       "John Doe",
					SenderEmail:      "john@example.com",
					RecipientName:    "Jane Doe",
					RecipientEmail:   "jane@example.com",
					SendNotification: true,
					Captcha:          "blue",
					MaxViewCount:     tc.maxViewCount,
				}

				expectedResponse := &domain.MessageSubmissionResponse{
					MessageID:  "test-message-id",
					DecryptURL: "https://example.com/decrypt/test-message-id/key123",
					Success:    true,
				}

				mockService.On("SubmitMessage", mock.Anything, expectedDomainReq).Return(expectedResponse, nil).Once()
			}

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
				SendNotification: true,
				AntiSpamAnswer:   "blue",
				MaxViewCount:     tc.maxViewCount,
			}

			body, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/messages", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Correlation-ID", "test-correlation-id")

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, tc.expectedStatus, w.Code, "Test case: %s", tc.name)
		})
	}
}

func TestGetMessageInfo_UsesRealExpiresAt(t *testing.T) {
	// Verify that ExpiresAt in the response comes from the domain (DB), not time.Now()
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	fixedExpiry := time.Date(2030, 6, 15, 12, 0, 0, 0, time.UTC)
	expectedAccessInfo := &domain.MessageAccessInfo{
		MessageID:          "test-message-id",
		Exists:             true,
		RequiresPassphrase: false,
		ExpiresAt:          &fixedExpiry,
	}

	mockService.On("CheckMessageAccess", mock.Anything, "test-message-id").Return(expectedAccessInfo, nil)

	req, _ := http.NewRequest("GET", "/api/v1/messages/test-message-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.MessageAccessInfoResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(
		t,
		fixedExpiry.Unix(),
		response.ExpiresAt.Unix(),
		"ExpiresAt should match the DB value, not be recalculated",
	)
}

func TestSubmitMessage_UsesRealExpiresAt(t *testing.T) {
	// Verify that ExpiresAt in the submit response comes from the domain, not time.Now()
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	fixedExpiry := time.Date(2030, 6, 15, 12, 0, 0, 0, time.UTC)
	expectedDomainReq := domain.MessageSubmissionRequest{
		Content:          "Test message",
		SenderName:       "John Doe",
		SenderEmail:      "john@example.com",
		RecipientName:    "Jane Doe",
		RecipientEmail:   "jane@example.com",
		SendNotification: true,
		Captcha:          "blue",
	}
	expectedResponse := &domain.MessageSubmissionResponse{
		MessageID:  "test-message-id",
		DecryptURL: "https://example.com/decrypt/test-message-id/key123",
		ExpiresAt:  &fixedExpiry,
		Success:    true,
	}

	mockService.On("SubmitMessage", mock.Anything, expectedDomainReq).Return(expectedResponse, nil)

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
		SendNotification: true,
		AntiSpamAnswer:   "blue",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/messages", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.MessageSubmissionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(
		t,
		fixedExpiry.Unix(),
		response.ExpiresAt.Unix(),
		"ExpiresAt should match the domain value, not be recalculated",
	)

	mockService.AssertExpectations(t)
}
