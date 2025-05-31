package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// createTestMocks creates all required mocks for testing
func createTestMocks() (*MockStorageRepository, *MockNotificationPublisher, *MockLoggerPort, *MockConfigPort, *MockValidationPort) {
	mockStorageRepo := &MockStorageRepository{}
	mockNotificationPublisher := &MockNotificationPublisher{}
	mockLogger := &MockLoggerPort{}
	mockConfig := &MockConfigPort{}
	mockValidation := &MockValidationPort{}
	
	// Set up default mock expectations for config
	mockConfig.On("GetServerEmail").Return("test@example.com")
	mockConfig.On("GetServerName").Return("Test Server")
	mockConfig.On("GetEmailTemplate").Return("/templates/email_template.html")
	mockConfig.On("GetPasswordExchangeURL").Return("https://test.password.exchange")
	
	// No default storage expectations - set up as needed in individual tests
	
	// Set up default mock expectations for validation (can be overridden in specific tests)
	mockValidation.On("ValidateEmail", mock.AnythingOfType("string")).Return(nil)
	mockValidation.On("SanitizeEmailForLogging", mock.AnythingOfType("string")).Return("sanitized@example.com")
	
	return mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation
}

// Test NewReminderService constructor
func TestNewReminderService(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()

	// Act
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	// Assert
	assert.NotNil(t, service)
	assert.Equal(t, mockStorageRepo, service.storageRepo)
	assert.Equal(t, mockNotificationPublisher, service.notificationPublisher)
	assert.NotNil(t, service.circuitBreaker)
	assert.Equal(t, CircuitBreakerClosed, service.circuitBreaker.state)
}

// Test validateReminderConfig with valid configuration
func TestValidateReminderConfig_ValidConfig_NoError(t *testing.T) {
	// Arrange
	service := &ReminderService{}
	config := ReminderConfig{
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        24,
	}

	// Act
	err := service.validateReminderConfig(config)

	// Assert
	assert.NoError(t, err)
}

// Test validateReminderConfig with invalid CheckAfterHours
func TestValidateReminderConfig_InvalidCheckAfterHours_ReturnsError(t *testing.T) {
	service := &ReminderService{}

	tests := []struct {
		name            string
		checkAfterHours int
		wantErr         error
	}{
		{
			name:            "too low",
			checkAfterHours: 0,
			wantErr:         ErrInvalidCheckAfterHours,
		},
		{
			name:            "too high",
			checkAfterHours: 9000,
			wantErr:         ErrInvalidCheckAfterHours,
		},
		{
			name:            "minimum valid",
			checkAfterHours: MinCheckAfterHours,
			wantErr:         nil,
		},
		{
			name:            "maximum valid",
			checkAfterHours: MaxCheckAfterHours,
			wantErr:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ReminderConfig{
				CheckAfterHours: tt.checkAfterHours,
				MaxReminders:    3,
				Interval:        24,
			}

			err := service.validateReminderConfig(config)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test validateReminderConfig with invalid MaxReminders
func TestValidateReminderConfig_InvalidMaxReminders_ReturnsError(t *testing.T) {
	service := &ReminderService{}

	tests := []struct {
		name         string
		maxReminders int
		wantErr      error
	}{
		{
			name:         "too low",
			maxReminders: 0,
			wantErr:      ErrInvalidMaxReminders,
		},
		{
			name:         "too high",
			maxReminders: 15,
			wantErr:      ErrInvalidMaxReminders,
		},
		{
			name:         "minimum valid",
			maxReminders: MinMaxReminders,
			wantErr:      nil,
		},
		{
			name:         "maximum valid",
			maxReminders: MaxMaxReminders,
			wantErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ReminderConfig{
				CheckAfterHours: 24,
				MaxReminders:    tt.maxReminders,
				Interval:        24,
			}

			err := service.validateReminderConfig(config)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test validateReminderConfig with invalid Interval
func TestValidateReminderConfig_InvalidInterval_ReturnsError(t *testing.T) {
	service := &ReminderService{}

	tests := []struct {
		name     string
		interval int
		wantErr  error
	}{
		{
			name:     "too low",
			interval: 0,
			wantErr:  ErrInvalidReminderInterval,
		},
		{
			name:     "too high",
			interval: 800,
			wantErr:  ErrInvalidReminderInterval,
		},
		{
			name:     "minimum valid",
			interval: MinReminderInterval,
			wantErr:  nil,
		},
		{
			name:     "maximum valid",
			interval: MaxReminderInterval,
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ReminderConfig{
				CheckAfterHours: 24,
				MaxReminders:    3,
				Interval:        tt.interval,
			}

			err := service.validateReminderConfig(config)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test ProcessReminders with disabled config
func TestProcessReminders_DisabledConfig_ReturnsEarly(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	config := ReminderConfig{
		Enabled:         false,
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        24,
	}

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	assert.NoError(t, err)
	// Verify no storage calls were made
	mockStorageRepo.AssertNotCalled(t, "GetUnviewedMessagesForReminders")
	mockNotificationPublisher.AssertNotCalled(t, "PublishNotification")
}

// Test ProcessReminders with invalid config
func TestProcessReminders_InvalidConfig_ReturnsError(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	config := ReminderConfig{
		Enabled:         true,
		CheckAfterHours: 0, // Invalid
		MaxReminders:    3,
		Interval:        24,
	}

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checkAfterHours must be between 1 and 8760 hours")
	mockStorageRepo.AssertNotCalled(t, "GetUnviewedMessagesForReminders")
}

// Test ProcessReminders with no messages to process
func TestProcessReminders_NoMessages_ReturnsSuccess(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	config := ReminderConfig{
		Enabled:         true,
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        24,
	}

	// Mock storage to return empty list
	mockStorageRepo.On("GetUnviewedMessagesForReminders", ctx, 24, 3, 24).Return([]*UnviewedMessage{}, nil)

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	assert.NoError(t, err)
	mockStorageRepo.AssertExpectations(t)
	mockNotificationPublisher.AssertNotCalled(t, "PublishNotification")
}

// Test ProcessReminders with storage error
func TestProcessReminders_StorageError_ReturnsError(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	config := ReminderConfig{
		Enabled:         true,
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        24,
	}

	storageError := errors.New("database connection failed")
	mockStorageRepo.On("GetUnviewedMessagesForReminders", ctx, 24, 3, 24).Return(nil, storageError)

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get unviewed messages")
	assert.Contains(t, err.Error(), "database connection failed")
	mockStorageRepo.AssertExpectations(t)
}

// Test ProcessReminders with successful message processing
func TestProcessReminders_SuccessfulProcessing_ReturnsSuccess(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	config := ReminderConfig{
		Enabled:         true,
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        24,
	}

	messages := []*UnviewedMessage{
		{
			MessageID:      123,
			UniqueID:       "abc123",
			RecipientEmail: "test@example.com",
			DaysOld:        2,
			Created:        time.Now().Add(-48 * time.Hour),
		},
	}

	// Mock storage calls
	mockStorageRepo.On("GetUnviewedMessagesForReminders", ctx, 24, 3, 24).Return(messages, nil)
	mockStorageRepo.On("GetReminderHistory", ctx, 123).Return([]*ReminderLogEntry{}, nil)
	mockStorageRepo.On("LogReminderSent", ctx, 123, "test@example.com").Return(nil)

	// Mock email sending
	mockNotificationPublisher.On("PublishNotification", ctx, mock.AnythingOfType("NotificationRequest")).Return(nil)

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	assert.NoError(t, err)
	mockStorageRepo.AssertExpectations(t)
	mockNotificationPublisher.AssertExpectations(t)
}

// Test ProcessReminders with context cancellation
func TestProcessReminders_ContextCancelled_ReturnsError(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := ReminderConfig{
		Enabled:         true,
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        24,
	}

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// Test ProcessReminders with timeout during storage operation
func TestProcessReminders_StorageTimeout_HandledGracefully(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	config := ReminderConfig{
		Enabled:         true,
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        24,
	}

	// Mock storage to simulate slow operation
	mockStorageRepo.On("GetUnviewedMessagesForReminders", ctx, 24, 3, 24).Return(nil, context.DeadlineExceeded)

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get unviewed messages")
	mockStorageRepo.AssertExpectations(t)
}

// Test ProcessMessageReminder with malformed recipient email (edge cases)
func TestProcessMessageReminder_EdgeCaseEmails_Validation(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	
	// Override validation to return errors for invalid emails
	mockValidation.ExpectedCalls = nil // Clear default expectations
	mockValidation.On("ValidateEmail", mock.AnythingOfType("string")).Return(errors.New("invalid email"))
	
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()

	testCases := []struct {
		name  string
		email string
	}{
		{"empty email", ""},
		{"email with spaces", "test @example.com"},
		{"email without @", "testexample.com"},
		{"email without domain", "test@"},
		{"email with multiple @", "test@@example.com"},
		{"very long email", "a" + string(make([]byte, 320)) + "@example.com"},
		{"email with special chars", "test<script>@example.com"},
		{"email with null byte", "test\x00@example.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := ReminderRequest{
				MessageID:      123,
				UniqueID:       "abc123",
				RecipientEmail: tc.email,
				DaysOld:        2,
				DecryptionURL:  "", // Empty - template references original email
			}

			// Act
			err := service.ProcessMessageReminder(ctx, req)

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid recipient email address")
			mockStorageRepo.AssertNotCalled(t, "GetReminderHistory")
			mockNotificationPublisher.AssertNotCalled(t, "PublishNotification")
		})
	}
}

// Test ProcessMessageReminder with empty decryption URL - new behavior references original email
func TestProcessMessageReminder_EmptyURL_StillProcesses(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	req := ReminderRequest{
		MessageID:      123,
		UniqueID:       "abc123",
		RecipientEmail: "test@example.com",
		DaysOld:        2,
		DecryptionURL:  "", // Empty - template references original email
	}

	// Mock storage calls
	mockStorageRepo.On("GetReminderHistory", ctx, 123).Return([]*ReminderLogEntry{}, nil)
	mockStorageRepo.On("LogReminderSent", ctx, 123, "test@example.com").Return(nil)

	// Mock email sending - should work with empty URL (template references original email)
	mockNotificationPublisher.On("PublishNotification", ctx, mock.AnythingOfType("NotificationRequest")).Return(nil)

	// Act
	err := service.ProcessMessageReminder(ctx, req)

	// Assert
	// Empty URLs are acceptable - template now references original email
	assert.NoError(t, err)
	mockStorageRepo.AssertExpectations(t)
	mockNotificationPublisher.AssertExpectations(t)
}

// Test ProcessReminders with storage logging failure - continues with other messages
func TestProcessReminders_LoggingFailure_ContinuesProcessing(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	config := ReminderConfig{
		Enabled:         true,
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        24,
	}

	messages := []*UnviewedMessage{
		{
			MessageID:      123,
			UniqueID:       "abc123",
			RecipientEmail: "test@example.com",
			DaysOld:        2,
			Created:        time.Now().Add(-48 * time.Hour),
		},
		{
			MessageID:      124,
			UniqueID:       "def456",
			RecipientEmail: "valid@example.com",
			DaysOld:        1,
			Created:        time.Now().Add(-24 * time.Hour),
		},
	}

	// Mock storage calls - first message logging fails, second succeeds
	mockStorageRepo.On("GetUnviewedMessagesForReminders", ctx, 24, 3, 24).Return(messages, nil)
	
	// First message - logging fails after retry attempts
	mockStorageRepo.On("GetReminderHistory", ctx, 123).Return([]*ReminderLogEntry{}, nil)
	mockStorageRepo.On("LogReminderSent", ctx, 123, "test@example.com").Return(errors.New("logging database unavailable"))
	
	// Second message - succeeds
	mockStorageRepo.On("GetReminderHistory", ctx, 124).Return([]*ReminderLogEntry{}, nil)
	mockStorageRepo.On("LogReminderSent", ctx, 124, "valid@example.com").Return(nil)

	// Mock successful email sending for both
	mockNotificationPublisher.On("PublishNotification", ctx, mock.AnythingOfType("NotificationRequest")).Return(nil)

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	// Should succeed because at least one message was processed successfully
	assert.NoError(t, err)
	mockStorageRepo.AssertExpectations(t)
	mockNotificationPublisher.AssertExpectations(t)
}

// Test ProcessReminders with circuit breaker activation
func TestProcessReminders_CircuitBreakerOpen_StopsProcessing(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	config := ReminderConfig{
		Enabled:         true,
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        24,
	}

	// Force circuit breaker to open state
	service.circuitBreaker.state = CircuitBreakerOpen
	service.circuitBreaker.lastFailureTime = time.Now()

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
	mockStorageRepo.AssertNotCalled(t, "GetUnviewedMessagesForReminders")
}

// Test ProcessReminders with mixed success and failure scenarios
func TestProcessReminders_MixedResults_ContinuesProcessing(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	config := ReminderConfig{
		Enabled:         true,
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        24,
	}

	messages := []*UnviewedMessage{
		{
			MessageID:      123,
			UniqueID:       "abc123",
			RecipientEmail: "valid@example.com",
			DaysOld:        2,
			Created:        time.Now().Add(-48 * time.Hour),
		},
		{
			MessageID:      124,
			UniqueID:       "def456",
			RecipientEmail: "invalid-email", // This will fail validation
			DaysOld:        3,
			Created:        time.Now().Add(-72 * time.Hour),
		},
		{
			MessageID:      125,
			UniqueID:       "ghi789",
			RecipientEmail: "another@example.com",
			DaysOld:        1,
			Created:        time.Now().Add(-24 * time.Hour),
		},
	}

	// Override validation to fail for "invalid-email"
	mockValidation.ExpectedCalls = nil // Clear default expectations
	mockValidation.On("ValidateEmail", "valid@example.com").Return(nil)
	mockValidation.On("ValidateEmail", "invalid-email").Return(errors.New("invalid email format"))
	mockValidation.On("ValidateEmail", "another@example.com").Return(nil)
	mockValidation.On("SanitizeEmailForLogging", mock.AnythingOfType("string")).Return("sanitized@example.com")
	
	// Mock storage calls
	mockStorageRepo.On("GetUnviewedMessagesForReminders", ctx, 24, 3, 24).Return(messages, nil)
	
	// Valid email processing (message 123)
	mockStorageRepo.On("GetReminderHistory", ctx, 123).Return([]*ReminderLogEntry{}, nil)
	mockStorageRepo.On("LogReminderSent", ctx, 123, "valid@example.com").Return(nil)
	
	// Invalid email processing (message 124) - won't reach storage calls
	
	// Valid email processing (message 125)
	mockStorageRepo.On("GetReminderHistory", ctx, 125).Return([]*ReminderLogEntry{}, nil)
	mockStorageRepo.On("LogReminderSent", ctx, 125, "another@example.com").Return(nil)

	// Mock email sending for valid emails only
	mockNotificationPublisher.On("PublishNotification", ctx, mock.MatchedBy(func(req NotificationRequest) bool {
		return req.To == "valid@example.com" || req.To == "another@example.com"
	})).Return(nil)

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	// Should succeed overall despite individual failures
	assert.NoError(t, err)
	mockStorageRepo.AssertExpectations(t)
	mockNotificationPublisher.AssertExpectations(t)
}

// Test ProcessMessageReminder with zero MessageID
func TestProcessMessageReminder_ZeroMessageID_Error(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	req := ReminderRequest{
		MessageID:      0, // Invalid
		UniqueID:       "abc123",
		RecipientEmail: "test@example.com",
		DaysOld:        2,
		DecryptionURL:  "", // Empty - template references original email
	}

	// Act
	err := service.ProcessMessageReminder(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "messageID must be greater than 0")
	mockStorageRepo.AssertNotCalled(t, "GetReminderHistory")
}

// Test ProcessMessageReminder with empty UniqueID
func TestProcessMessageReminder_EmptyUniqueID_Error(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	req := ReminderRequest{
		MessageID:      123,
		UniqueID:       "", // Invalid
		RecipientEmail: "test@example.com",
		DaysOld:        2,
		DecryptionURL:  "", // Empty - template references original email
	}

	// Act
	err := service.ProcessMessageReminder(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uniqueID cannot be empty")
	mockStorageRepo.AssertNotCalled(t, "GetReminderHistory")
}

// Test ProcessMessageReminder with negative DaysOld
func TestProcessMessageReminder_NegativeDaysOld_Error(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	req := ReminderRequest{
		MessageID:      123,
		UniqueID:       "abc123",
		RecipientEmail: "test@example.com",
		DaysOld:        -1, // Invalid
		DecryptionURL:  "", // Empty - template references original email
	}

	// Act
	err := service.ProcessMessageReminder(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "daysOld must be non-negative")
	mockStorageRepo.AssertNotCalled(t, "GetReminderHistory")
}

// Test reminder interval logic - messages should only be sent after interval has passed
func TestProcessReminders_ReminderInterval_RespectedCorrectly(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	config := ReminderConfig{
		Enabled:         true,
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        48, // 48 hour interval between reminders
	}

	// The storage adapter should receive the interval parameter (48 hours)
	// and only return messages where last_reminder_sent is either NULL
	// or more than 48 hours ago
	messages := []*UnviewedMessage{
		{
			MessageID:      123,
			UniqueID:       "abc123",
			RecipientEmail: "test@example.com",
			DaysOld:        3,
			Created:        time.Now().Add(-72 * time.Hour),
		},
	}

	// Mock storage calls with the interval parameter
	mockStorageRepo.On("GetUnviewedMessagesForReminders", ctx, 24, 3, 48).Return(messages, nil)
	mockStorageRepo.On("GetReminderHistory", ctx, 123).Return([]*ReminderLogEntry{
		{
			MessageID:      123,
			RecipientEmail: "test@example.com",
			ReminderCount:  1,
			SentAt:         time.Now().Add(-49 * time.Hour), // Last reminder sent 49 hours ago
		},
	}, nil)
	mockStorageRepo.On("LogReminderSent", ctx, 123, "test@example.com").Return(nil)

	// Mock email sending
	mockNotificationPublisher.On("PublishNotification", ctx, mock.AnythingOfType("NotificationRequest")).Return(nil)

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	assert.NoError(t, err)
	mockStorageRepo.AssertExpectations(t)
	mockNotificationPublisher.AssertExpectations(t)
	
	// Verify that the storage adapter was called with the correct interval parameter
	mockStorageRepo.AssertCalled(t, "GetUnviewedMessagesForReminders", ctx, 24, 3, 48)
}

// Test ProcessMessageReminder with valid request
func TestProcessMessageReminder_ValidRequest_Success(t *testing.T) {
	// Arrange
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	req := ReminderRequest{
		MessageID:      123,
		UniqueID:       "abc123",
		RecipientEmail: "test@example.com",
		DaysOld:        2,
		DecryptionURL:  "", // Empty - template references original email
	}

	// Mock storage calls
	mockStorageRepo.On("GetReminderHistory", ctx, 123).Return([]*ReminderLogEntry{}, nil)
	mockStorageRepo.On("LogReminderSent", ctx, 123, "test@example.com").Return(nil)

	// Mock email sending
	mockNotificationPublisher.On("PublishNotification", ctx, mock.AnythingOfType("NotificationRequest")).Return(nil)

	// Act
	err := service.ProcessMessageReminder(ctx, req)

	// Assert
	assert.NoError(t, err)
	mockStorageRepo.AssertExpectations(t)
	mockNotificationPublisher.AssertExpectations(t)
}