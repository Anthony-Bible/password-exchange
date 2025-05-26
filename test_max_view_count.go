package main

import (
	"context"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorageService for testing
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) StoreMessage(ctx context.Context, req domain.MessageStorageRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockStorageService) RetrieveMessage(ctx context.Context, req domain.MessageRetrievalStorageRequest) (*domain.MessageStorageResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.MessageStorageResponse), args.Error(1)
}

func (m *MockStorageService) GetMessage(ctx context.Context, req domain.MessageRetrievalStorageRequest) (*domain.MessageStorageResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.MessageStorageResponse), args.Error(1)
}

// MockEncryptionService for testing
type MockEncryptionService struct {
	mock.Mock
}

func (m *MockEncryptionService) GenerateKey(ctx context.Context, length int32) ([]byte, error) {
	args := m.Called(ctx, length)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncryptionService) Encrypt(ctx context.Context, plaintext []string, key []byte) ([]string, error) {
	args := m.Called(ctx, plaintext, key)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockEncryptionService) Decrypt(ctx context.Context, ciphertext []string, key []byte) ([]string, error) {
	args := m.Called(ctx, ciphertext, key)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockEncryptionService) GenerateID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// MockNotificationService for testing
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendMessageNotification(ctx context.Context, req domain.MessageNotificationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// MockPasswordHasher for testing
type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) Hash(ctx context.Context, password string) (string, error) {
	args := m.Called(ctx, password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) Verify(ctx context.Context, password, hash string) (bool, error) {
	args := m.Called(ctx, password, hash)
	return args.Bool(0), args.Error(1)
}

// MockURLBuilder for testing
type MockURLBuilder struct {
	mock.Mock
}

func (m *MockURLBuilder) BuildDecryptURL(messageID string, encryptionKey []byte) string {
	args := m.Called(messageID, encryptionKey)
	return args.String(0)
}

// TestMaxViewCount_ConfigDefault tests that default max view count is used from config
func TestMaxViewCount_ConfigDefault(t *testing.T) {
	// Set up config with default max view count
	config.Config.DefaultMaxViewCount = 10

	// Set up mocks
	mockStorage := new(MockStorageService)
	mockEncryption := new(MockEncryptionService)
	mockNotification := new(MockNotificationService)
	mockHasher := new(MockPasswordHasher)
	mockURLBuilder := new(MockURLBuilder)

	// Set up expectations
	mockEncryption.On("GenerateKey", mock.Anything, int32(32)).Return([]byte("test-key"), nil)
	mockEncryption.On("Encrypt", mock.Anything, []string{"test content"}, []byte("test-key")).Return([]string{"encrypted"}, nil)
	mockEncryption.On("GenerateID", mock.Anything).Return("test-id", nil)
	mockURLBuilder.On("BuildDecryptURL", "test-id", []byte("test-key")).Return("http://test.com/decrypt")
	
	// Expect StoreMessage to be called with MaxViewCount = 10 (from config)
	mockStorage.On("StoreMessage", mock.Anything, mock.MatchedBy(func(req domain.MessageStorageRequest) bool {
		return req.MaxViewCount == 10
	})).Return(nil)

	// Create service
	service := domain.NewMessageService(mockEncryption, mockStorage, mockNotification, mockHasher, mockURLBuilder)

	// Create request without MaxViewCount (should use config default)
	req := domain.MessageSubmissionRequest{
		Content:     "test content",
		SenderName:  "Test User",
		SenderEmail: "test@example.com",
		SkipEmail:   true,
		MaxViewCount: 0, // Not set - should use config default
	}

	// Submit message
	_, err := service.SubmitMessage(context.Background(), req)

	// Verify
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

// TestMaxViewCount_CustomValue tests that custom max view count is used when provided
func TestMaxViewCount_CustomValue(t *testing.T) {
	// Set up config with default max view count
	config.Config.DefaultMaxViewCount = 5

	// Set up mocks
	mockStorage := new(MockStorageService)
	mockEncryption := new(MockEncryptionService)
	mockNotification := new(MockNotificationService)
	mockHasher := new(MockPasswordHasher)
	mockURLBuilder := new(MockURLBuilder)

	// Set up expectations
	mockEncryption.On("GenerateKey", mock.Anything, int32(32)).Return([]byte("test-key"), nil)
	mockEncryption.On("Encrypt", mock.Anything, []string{"test content"}, []byte("test-key")).Return([]string{"encrypted"}, nil)
	mockEncryption.On("GenerateID", mock.Anything).Return("test-id", nil)
	mockURLBuilder.On("BuildDecryptURL", "test-id", []byte("test-key")).Return("http://test.com/decrypt")
	
	// Expect StoreMessage to be called with MaxViewCount = 25 (custom value)
	mockStorage.On("StoreMessage", mock.Anything, mock.MatchedBy(func(req domain.MessageStorageRequest) bool {
		return req.MaxViewCount == 25
	})).Return(nil)

	// Create service
	service := domain.NewMessageService(mockEncryption, mockStorage, mockNotification, mockHasher, mockURLBuilder)

	// Create request with custom MaxViewCount
	req := domain.MessageSubmissionRequest{
		Content:      "test content",
		SenderName:   "Test User", 
		SenderEmail:  "test@example.com",
		SkipEmail:    true,
		MaxViewCount: 25, // Custom value - should override config default
	}

	// Submit message
	_, err := service.SubmitMessage(context.Background(), req)

	// Verify
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

// TestMaxViewCount_ValidationFailure tests that invalid max view count is rejected
func TestMaxViewCount_ValidationFailure(t *testing.T) {
	// Set up mocks (won't be called due to validation failure)
	mockStorage := new(MockStorageService)
	mockEncryption := new(MockEncryptionService)
	mockNotification := new(MockNotificationService)
	mockHasher := new(MockPasswordHasher)
	mockURLBuilder := new(MockURLBuilder)

	// Create service
	service := domain.NewMessageService(mockEncryption, mockStorage, mockNotification, mockHasher, mockURLBuilder)

	// Test cases for invalid max view counts
	testCases := []struct {
		name         string
		maxViewCount int
	}{
		{"TooLow", -1},
		{"Zero", 0}, // Note: 0 should use default, but let's test explicit validation
		{"TooHigh", 101},
		{"WayTooHigh", 1000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := domain.MessageSubmissionRequest{
				Content:      "test content",
				SenderName:   "Test User",
				SenderEmail:  "test@example.com", 
				SkipEmail:    true,
				MaxViewCount: tc.maxViewCount,
			}

			// This should fail validation if MaxViewCount is not 0 but outside 1-100 range
			if tc.maxViewCount != 0 {
				_, err := service.SubmitMessage(context.Background(), req)
				assert.Error(t, err, "Expected validation error for MaxViewCount=%d", tc.maxViewCount)
				assert.Contains(t, err.Error(), "max view count must be between 1 and 100")
			}
		})
	}
}