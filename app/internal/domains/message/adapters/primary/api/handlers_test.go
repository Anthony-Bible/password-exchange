package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/middleware"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockMessageService is a mock implementation of MessageServicePort.
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

func (m *MockMessageService) NotifyMessage(
	ctx context.Context,
	req domain.MessageNotifyRequest,
) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockMessageService) GetDefaultMaxViewCount() int {
	args := m.Called()
	return args.Int(0)
}

// healthCheckFn is a tiny test-only function-typed implementation of the
// HealthCheck contract for both secondary ports. Tests inject a closure that
// returns the desired error/timing behaviour without standing up a full mock.
type healthCheckFn func(ctx context.Context) error

// stubEncryptionPort satisfies secondary.EncryptionServicePort. Only
// HealthCheck is exercised by Readyz tests; the other methods panic if
// called so misuse fails loudly.
type stubEncryptionPort struct {
	healthCheck healthCheckFn
}

func (s *stubEncryptionPort) GenerateKey(context.Context, int32) ([]byte, error) {
	panic("stubEncryptionPort.GenerateKey not implemented")
}

func (s *stubEncryptionPort) Encrypt(context.Context, []string, []byte) ([]string, error) {
	panic("stubEncryptionPort.Encrypt not implemented")
}

func (s *stubEncryptionPort) Decrypt(context.Context, []string, []byte) ([]string, error) {
	panic("stubEncryptionPort.Decrypt not implemented")
}

func (s *stubEncryptionPort) GenerateID(context.Context) (string, error) {
	panic("stubEncryptionPort.GenerateID not implemented")
}

func (s *stubEncryptionPort) HealthCheck(ctx context.Context) error {
	if s.healthCheck == nil {
		return nil
	}
	return s.healthCheck(ctx)
}

// stubStoragePort satisfies secondary.StorageServicePort with the same
// "only HealthCheck is real" pattern.
type stubStoragePort struct {
	healthCheck healthCheckFn
}

func (s *stubStoragePort) StoreMessage(context.Context, contracts.MessageStorageRequest) error {
	panic("stubStoragePort.StoreMessage not implemented")
}

func (s *stubStoragePort) RetrieveMessage(
	context.Context,
	contracts.MessageRetrievalStorageRequest,
) (*contracts.MessageStorageResponse, error) {
	panic("stubStoragePort.RetrieveMessage not implemented")
}

func (s *stubStoragePort) GetMessage(
	context.Context,
	contracts.MessageRetrievalStorageRequest,
) (*contracts.MessageStorageResponse, error) {
	panic("stubStoragePort.GetMessage not implemented")
}

func (s *stubStoragePort) HealthCheck(ctx context.Context) error {
	if s.healthCheck == nil {
		return nil
	}
	return s.healthCheck(ctx)
}

func setupTestRouter(mockService *MockMessageService) *gin.Engine {
	return setupTestRouterWithProbes(mockService, &stubEncryptionPort{}, &stubStoragePort{})
}

func setupTestRouterWithProbes(
	mockService *MockMessageService,
	enc *stubEncryptionPort,
	stor *stubStoragePort,
) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Create minimal metrics setup for testing
	registry := prometheus.NewRegistry()
	metrics := middleware.NewPrometheusMetrics(registry)

	return setupRouter(NewMessageAPIHandler(mockService, enc, stor), metrics, registry, nil)
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
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(jsonBody))
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

func TestSubmitMessage_ClientEncrypted_PassesFlagAndHidesKey(t *testing.T) {
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	mockService.On("SubmitMessage", mock.Anything, mock.MatchedBy(func(req domain.MessageSubmissionRequest) bool {
		return req.IsClientEncrypted
	})).Return(&domain.MessageSubmissionResponse{
		MessageID:         "msg-client-e2e",
		DecryptURL:        "https://example.com/decrypt/msg-client-e2e",
		Key:               "",
		IsClientEncrypted: true,
		Success:           true,
	}, nil)

	req, _ := http.NewRequest(
		http.MethodPost,
		"/api/v1/messages",
		bytes.NewBufferString(`{"content":"MTIzNDU2Nzg5MDEy.Y2lwaGVydGV4dA","isClientEncrypted":true}`),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Empty(t, response["key"], "key must be empty when isClientEncrypted=true")
	assert.Equal(t, true, response["isClientEncrypted"], "response must include isClientEncrypted=true")
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
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(jsonBody))
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

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/messages/test-message-id", nil)
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

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/messages/test-message-id", nil)
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

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "1.0.0", response.Version)
	assert.Equal(t, "healthy", response.Services["database"])
	assert.Equal(t, "healthy", response.Services["encryption"])
	assert.NotContains(t, response.Services, "email")
}

func TestHealthCheck_DegradedWhenStorageUnhealthy(t *testing.T) {
	mockService := new(MockMessageService)
	enc := &stubEncryptionPort{healthCheck: func(context.Context) error { return nil }}
	stor := &stubStoragePort{healthCheck: func(context.Context) error { return errors.New("storage down") }}
	router := setupTestRouterWithProbes(mockService, enc, stor)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthCheckResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, "degraded", response.Status)
	assert.Equal(t, "unhealthy", response.Services["database"])
	assert.Equal(t, "healthy", response.Services["encryption"])
}

func TestHealthCheck_DegradedWhenEncryptionUnhealthy(t *testing.T) {
	mockService := new(MockMessageService)
	enc := &stubEncryptionPort{healthCheck: func(context.Context) error { return errors.New("encryption down") }}
	stor := &stubStoragePort{healthCheck: func(context.Context) error { return nil }}
	router := setupTestRouterWithProbes(mockService, enc, stor)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthCheckResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, "degraded", response.Status)
	assert.Equal(t, "healthy", response.Services["database"])
	assert.Equal(t, "unhealthy", response.Services["encryption"])
}

func TestHealthCheck_UnhealthyWhenAllDown(t *testing.T) {
	mockService := new(MockMessageService)
	enc := &stubEncryptionPort{healthCheck: func(context.Context) error { return errors.New("encryption down") }}
	stor := &stubStoragePort{healthCheck: func(context.Context) error { return errors.New("storage down") }}
	router := setupTestRouterWithProbes(mockService, enc, stor)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthCheckResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, "unhealthy", response.Status)
	assert.Equal(t, "unhealthy", response.Services["database"])
	assert.Equal(t, "unhealthy", response.Services["encryption"])
}

func TestHealthCheck_AbortsViaContextTimeout(t *testing.T) {
	// A stuck dependency must not hang the public health endpoint. The handler
	// attaches a deadline to the dispatched context; the stub blocks until that
	// fires and then returns ctx.Err(), so /api/v1/health must still respond
	// (200, that dep marked "unhealthy") rather than blocking indefinitely.
	mockService := new(MockMessageService)
	enc := &stubEncryptionPort{healthCheck: func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}}
	stor := &stubStoragePort{healthCheck: func(context.Context) error { return nil }}
	router := setupTestRouterWithProbes(mockService, enc, stor)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		router.ServeHTTP(w, req)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("/api/v1/health blocked past the in-handler deadline")
	}

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthCheckResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, "degraded", response.Status)
	assert.Equal(t, "unhealthy", response.Services["encryption"])
	assert.Equal(t, "healthy", response.Services["database"])
}

func TestAPIInfo(t *testing.T) {
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/info", nil)
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
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(body))
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

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/messages/test-message-id", nil)
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
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(jsonBody))
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
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(body))
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

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/messages/test-message-id", nil)
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
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(jsonBody))
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

func TestDecryptMessage_IncludesExpiresAt(t *testing.T) {
	// When domain returns ExpiresAt, it must appear in the decrypt response JSON.
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	fixedExpiry := time.Date(2030, 3, 7, 12, 0, 0, 0, time.UTC)
	mockService.On("RetrieveMessage", mock.Anything, mock.MatchedBy(func(req domain.MessageRetrievalRequest) bool {
		return req.MessageID == "test-message-id"
	})).Return(&domain.MessageRetrievalResponse{
		MessageID:    "test-message-id",
		Content:      "secret content",
		ViewCount:    1,
		MaxViewCount: 5,
		ExpiresAt:    &fixedExpiry,
		Success:      true,
	}, nil)

	body, _ := json.Marshal(models.MessageDecryptRequest{DecryptionKey: "dGVzdGtleQ=="})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/messages/test-message-id/decrypt", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.MessageDecryptResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response.ExpiresAt, "expiresAt must be present in decrypt response")
	assert.Equal(t, fixedExpiry.Unix(), response.ExpiresAt.Unix())

	mockService.AssertExpectations(t)
}

func TestLivez_AlwaysReturns200WithoutTouchingDependencies(t *testing.T) {
	mockService := new(MockMessageService)
	// Deps panic if their HealthCheck is called — proves /livez is dependency-free.
	enc := &stubEncryptionPort{healthCheck: func(context.Context) error {
		t.Fatalf("/livez must not call encryption HealthCheck")
		return nil
	}}
	stor := &stubStoragePort{healthCheck: func(context.Context) error {
		t.Fatalf("/livez must not call storage HealthCheck")
		return nil
	}}
	router := setupTestRouterWithProbes(mockService, enc, stor)

	req, _ := http.NewRequest(http.MethodGet, "/livez", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	require := require.New(t)
	require.NoError(json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal("ok", body["status"])
}

func TestReadyz_ReturnsOKWhenBothDependenciesHealthy(t *testing.T) {
	mockService := new(MockMessageService)
	enc := &stubEncryptionPort{healthCheck: func(context.Context) error { return nil }}
	stor := &stubStoragePort{healthCheck: func(context.Context) error { return nil }}
	router := setupTestRouterWithProbes(mockService, enc, stor)

	req, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReadyz_Returns503WhenEncryptionUnhealthy(t *testing.T) {
	mockService := new(MockMessageService)
	enc := &stubEncryptionPort{healthCheck: func(context.Context) error {
		return errors.New("encryption down")
	}}
	stor := &stubStoragePort{healthCheck: func(context.Context) error { return nil }}
	router := setupTestRouterWithProbes(mockService, enc, stor)

	req, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var body map[string]interface{}
	require := require.New(t)
	require.NoError(json.Unmarshal(w.Body.Bytes(), &body))
	require.Contains(body, "failed")
	failed, _ := body["failed"].([]interface{})
	require.Contains(failed, "encryption")
}

func TestReadyz_Returns503WhenStorageUnhealthy(t *testing.T) {
	mockService := new(MockMessageService)
	enc := &stubEncryptionPort{healthCheck: func(context.Context) error { return nil }}
	stor := &stubStoragePort{healthCheck: func(context.Context) error {
		return errors.New("storage down")
	}}
	router := setupTestRouterWithProbes(mockService, enc, stor)

	req, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var body map[string]interface{}
	require := require.New(t)
	require.NoError(json.Unmarshal(w.Body.Bytes(), &body))
	failed, _ := body["failed"].([]interface{})
	require.Contains(failed, "storage")
}

func TestReadyz_Returns503WhenBothUnhealthy(t *testing.T) {
	mockService := new(MockMessageService)
	enc := &stubEncryptionPort{healthCheck: func(context.Context) error {
		return errors.New("encryption down")
	}}
	stor := &stubStoragePort{healthCheck: func(context.Context) error {
		return errors.New("storage down")
	}}
	router := setupTestRouterWithProbes(mockService, enc, stor)

	req, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var body map[string]interface{}
	require := require.New(t)
	require.NoError(json.Unmarshal(w.Body.Bytes(), &body))
	failed, _ := body["failed"].([]interface{})
	require.Contains(failed, "encryption")
	require.Contains(failed, "storage")
}

func TestReadyz_AbortsViaContextTimeout(t *testing.T) {
	// A stuck dependency must not hang readiness. The handler attaches a 2s
	// deadline to the dispatched context; the stub blocks until that fires
	// and then returns ctx.Err(), so /readyz must respond 503 rather than
	// blocking indefinitely.
	mockService := new(MockMessageService)
	enc := &stubEncryptionPort{healthCheck: func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}}
	stor := &stubStoragePort{healthCheck: func(context.Context) error { return nil }}
	router := setupTestRouterWithProbes(mockService, enc, stor)

	req, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		router.ServeHTTP(w, req)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("/readyz blocked past the in-handler deadline")
	}

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestDecryptMessage_NilExpiresAtIsNullInResponse(t *testing.T) {
	// When domain returns nil ExpiresAt (legacy message), the decrypt response must have "expiresAt": null.
	mockService := new(MockMessageService)
	router := setupTestRouter(mockService)

	mockService.On("RetrieveMessage", mock.Anything, mock.MatchedBy(func(req domain.MessageRetrievalRequest) bool {
		return req.MessageID == "test-message-id"
	})).Return(&domain.MessageRetrievalResponse{
		MessageID:    "test-message-id",
		Content:      "secret content",
		ViewCount:    1,
		MaxViewCount: 5,
		ExpiresAt:    nil,
		Success:      true,
	}, nil)

	body, _ := json.Marshal(models.MessageDecryptRequest{DecryptionKey: "dGVzdGtleQ=="})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/messages/test-message-id/decrypt", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var raw map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &raw)
	assert.NoError(t, err)
	expiresAt, hasKey := raw["expiresAt"]
	assert.True(t, hasKey, "expiresAt key must be present in decrypt response")
	assert.Nil(t, expiresAt, "expiresAt must be null for legacy messages")

	mockService.AssertExpectations(t)
}
