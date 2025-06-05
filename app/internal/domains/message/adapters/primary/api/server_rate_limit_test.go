package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServerRateLimiting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("message submission rate limiting", func(t *testing.T) {
		mockService := &MockMessageService{}
		server := NewServer(mockService)
		router := server.GetRouter()

		// Mock successful message submission
		mockResponse := &domain.MessageSubmissionResponse{
			MessageID:  "test-id",
			DecryptURL: "https://example.com/decrypt/test-id/key",
			Success:    true,
			Error:      nil,
		}

		// Set up mock to expect up to 10 calls (the rate limit) with flexible matching
		mockService.On("SubmitMessage", mock.AnythingOfType("*context.timerCtx"), mock.MatchedBy(func(req domain.MessageSubmissionRequest) bool {
			return req.Content == "test message" && req.SendNotification == false && req.RecipientName == "Jane Smith"
		})).Return(mockResponse, nil).Times(10)

		requestBody := map[string]interface{}{
			"content":          "test message",
			"recipient": map[string]interface{}{
				"name": "Jane Smith",
			},
			"sendNotification": false,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		// Test that 10 requests succeed (within rate limit)
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Forwarded-For", "192.168.1.1") // Consistent IP
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusCreated, w.Code, "request %d should succeed", i+1)
		}

		// Test that 11th request is rate limited
		req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code, "11th request should be rate limited")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "rate_limit_exceeded", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("message access rate limiting", func(t *testing.T) {
		mockService := &MockMessageService{}
		server := NewServer(mockService)
		router := server.GetRouter()

		// Mock successful message access
		mockResponse := &domain.MessageAccessInfo{
			MessageID:          "test-id",
			Exists:             true,
			RequiresPassphrase: false,
		}

		// Set up mock to expect up to 100 calls (the rate limit)
		mockService.On("CheckMessageAccess", mock.AnythingOfType("*context.timerCtx"), "test-id").Return(mockResponse, nil).Times(100)

		// Test that 100 requests succeed (within rate limit)
		for i := 0; i < 100; i++ {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/test-id?key=test-key", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.2") // Different IP from previous test
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
		}

		// Test that 101st request is rate limited
		req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/test-id?key=test-key", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.2")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code, "101st request should be rate limited")

		mockService.AssertExpectations(t)
	})

	t.Run("message decrypt rate limiting", func(t *testing.T) {
		mockService := &MockMessageService{}
		server := NewServer(mockService)
		router := server.GetRouter()

		// Mock successful message decryption
		mockResponse := &domain.MessageRetrievalResponse{
			MessageID: "test-id",
			Content:   "decrypted content",
			ViewCount: 1,
			Success:   true,
		}

		// Set up mock to expect up to 20 calls (the rate limit) with flexible matching
		mockService.On("RetrieveMessage", mock.AnythingOfType("*context.timerCtx"), mock.MatchedBy(func(req domain.MessageRetrievalRequest) bool {
			return req.MessageID == "test-id"
		})).Return(mockResponse, nil).Times(20)

		requestBody := map[string]interface{}{
			"decryptionKey": "test-key",
		}
		bodyBytes, _ := json.Marshal(requestBody)

		// Test that 20 requests succeed (within rate limit)
		for i := 0; i < 20; i++ {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/test-id/decrypt", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Forwarded-For", "192.168.1.3") // Different IP from previous tests
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
		}

		// Test that 21st request is rate limited
		req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/test-id/decrypt", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "192.168.1.3")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code, "21st request should be rate limited")

		mockService.AssertExpectations(t)
	})

	t.Run("health check rate limiting", func(t *testing.T) {
		mockService := &MockMessageService{}
		server := NewServer(mockService)
		router := server.GetRouter()

		// Test that 300 requests succeed (within rate limit)
		for i := 0; i < 300; i++ {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.4") // Different IP from previous tests
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
		}

		// Test that 301st request is rate limited
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.4")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code, "301st request should be rate limited")
	})

	t.Run("different IPs have separate rate limits", func(t *testing.T) {
		mockService := &MockMessageService{}
		server := NewServer(mockService)
		router := server.GetRouter()

		// Mock message submission responses
		mockResponse := &domain.MessageSubmissionResponse{
			MessageID:  "test-id",
			DecryptURL: "https://example.com/decrypt/test-id/key",
			Success:    true,
		}

		// Each IP should be able to make 10 requests - flexible matching for 3 IPs × 10 requests
		mockService.On("SubmitMessage", mock.AnythingOfType("*context.timerCtx"), mock.MatchedBy(func(req domain.MessageSubmissionRequest) bool {
			return req.Content == "test message" && req.SendNotification == false
		})).Return(mockResponse, nil).Times(30) // 3 IPs × 10 requests

		requestBody := map[string]interface{}{
			"content":          "test message",
			"sendNotification": false,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		ips := []string{"192.168.1.10", "192.168.1.11", "192.168.1.12"}

		for _, ip := range ips {
			// Each IP should be able to make 10 successful requests
			for i := 0; i < 10; i++ {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Forwarded-For", ip)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)
				assert.Equal(t, http.StatusCreated, w.Code, "request %d from IP %s should succeed", i+1, ip)
			}

			// 11th request from same IP should be rate limited
			req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Forwarded-For", ip)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusTooManyRequests, w.Code, "11th request from IP %s should be rate limited", ip)
		}

		mockService.AssertExpectations(t)
	})

	t.Run("rate limit error response format", func(t *testing.T) {
		mockService := &MockMessageService{}
		server := NewServer(mockService)
		router := server.GetRouter()

		// Mock message submission to reach rate limit
		mockResponse := &domain.MessageSubmissionResponse{
			MessageID:  "test-id",
			DecryptURL: "https://example.com/decrypt/test-id/key",
			Success:    true,
		}

		mockService.On("SubmitMessage", mock.AnythingOfType("*context.timerCtx"), mock.MatchedBy(func(req domain.MessageSubmissionRequest) bool {
			return req.Content == "test message" && req.SendNotification == false
		})).Return(mockResponse, nil).Times(10)

		requestBody := map[string]interface{}{
			"content":          "test message",
			"sendNotification": false,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		// Make 10 requests to reach the limit
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Forwarded-For", "192.168.1.5")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}

		// Make the rate-limited request
		req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "192.168.1.5")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify error response format
		assert.Equal(t, "rate_limit_exceeded", response["error"])
		assert.Equal(t, "Rate limit exceeded. Please try again later.", response["message"])
		assert.Equal(t, "/api/v1/messages", response["path"])
		assert.NotEmpty(t, response["timestamp"])
		assert.NotEmpty(t, response["correlation_id"])

		mockService.AssertExpectations(t)
	})
}