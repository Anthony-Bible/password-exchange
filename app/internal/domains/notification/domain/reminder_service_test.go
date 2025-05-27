package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test NewReminderService constructor
func TestNewReminderService(t *testing.T) {
	// Arrange
	mockStorageRepo := &MockStorageRepository{}
	mockEmailSender := &MockNotificationSender{}

	// Act
	service := NewReminderService(mockStorageRepo, mockEmailSender)

	// Assert
	assert.NotNil(t, service)
	assert.Equal(t, mockStorageRepo, service.storageRepo)
	assert.Equal(t, mockEmailSender, service.emailSender)
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
	mockStorageRepo := &MockStorageRepository{}
	mockEmailSender := &MockNotificationSender{}
	service := NewReminderService(mockStorageRepo, mockEmailSender)

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
	mockEmailSender.AssertNotCalled(t, "SendNotification")
}

// Test ProcessReminders with invalid config
func TestProcessReminders_InvalidConfig_ReturnsError(t *testing.T) {
	// Arrange
	mockStorageRepo := &MockStorageRepository{}
	mockEmailSender := &MockNotificationSender{}
	service := NewReminderService(mockStorageRepo, mockEmailSender)

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
	mockStorageRepo := &MockStorageRepository{}
	mockEmailSender := &MockNotificationSender{}
	service := NewReminderService(mockStorageRepo, mockEmailSender)

	ctx := context.Background()
	config := ReminderConfig{
		Enabled:         true,
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        24,
	}

	// Mock storage to return empty list
	mockStorageRepo.On("GetUnviewedMessagesForReminders", ctx, 24, 3).Return([]*UnviewedMessage{}, nil)

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	assert.NoError(t, err)
	mockStorageRepo.AssertExpectations(t)
	mockEmailSender.AssertNotCalled(t, "SendNotification")
}

// Test ProcessReminders with storage error
func TestProcessReminders_StorageError_ReturnsError(t *testing.T) {
	// Arrange
	mockStorageRepo := &MockStorageRepository{}
	mockEmailSender := &MockNotificationSender{}
	service := NewReminderService(mockStorageRepo, mockEmailSender)

	ctx := context.Background()
	config := ReminderConfig{
		Enabled:         true,
		CheckAfterHours: 24,
		MaxReminders:    3,
		Interval:        24,
	}

	storageError := errors.New("database connection failed")
	mockStorageRepo.On("GetUnviewedMessagesForReminders", ctx, 24, 3).Return(nil, storageError)

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
	mockStorageRepo := &MockStorageRepository{}
	mockEmailSender := &MockNotificationSender{}
	service := NewReminderService(mockStorageRepo, mockEmailSender)

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
	mockStorageRepo.On("GetUnviewedMessagesForReminders", ctx, 24, 3).Return(messages, nil)
	mockStorageRepo.On("GetReminderHistory", ctx, 123).Return([]*ReminderLogEntry{}, nil)
	mockStorageRepo.On("LogReminderSent", ctx, 123, "test@example.com").Return(nil)

	// Mock email sending
	expectedResponse := &NotificationResponse{
		Success:   true,
		MessageID: "msg-456",
	}
	mockEmailSender.On("SendNotification", ctx, mock.AnythingOfType("NotificationRequest")).Return(expectedResponse, nil)

	// Act
	err := service.ProcessReminders(ctx, config)

	// Assert
	assert.NoError(t, err)
	mockStorageRepo.AssertExpectations(t)
	mockEmailSender.AssertExpectations(t)
}

// Test ProcessMessageReminder with valid request
func TestProcessMessageReminder_ValidRequest_Success(t *testing.T) {
	// Arrange
	mockStorageRepo := &MockStorageRepository{}
	mockEmailSender := &MockNotificationSender{}
	service := NewReminderService(mockStorageRepo, mockEmailSender)

	ctx := context.Background()
	req := ReminderRequest{
		MessageID:      123,
		UniqueID:       "abc123",
		RecipientEmail: "test@example.com",
		DaysOld:        2,
		DecryptionURL:  "https://password.exchange/decrypt/abc123",
	}

	// Mock storage calls
	mockStorageRepo.On("GetReminderHistory", ctx, 123).Return([]*ReminderLogEntry{}, nil)
	mockStorageRepo.On("LogReminderSent", ctx, 123, "test@example.com").Return(nil)

	// Mock email sending
	expectedResponse := &NotificationResponse{
		Success:   true,
		MessageID: "msg-456",
	}
	mockEmailSender.On("SendNotification", ctx, mock.AnythingOfType("NotificationRequest")).Return(expectedResponse, nil)

	// Act
	err := service.ProcessMessageReminder(ctx, req)

	// Assert
	assert.NoError(t, err)
	mockStorageRepo.AssertExpectations(t)
	mockEmailSender.AssertExpectations(t)
}

// Test ProcessMessageReminder with invalid email
func TestProcessMessageReminder_InvalidEmail_ReturnsError(t *testing.T) {
	// Arrange
	mockStorageRepo := &MockStorageRepository{}
	mockEmailSender := &MockNotificationSender{}
	service := NewReminderService(mockStorageRepo, mockEmailSender)

	ctx := context.Background()
	req := ReminderRequest{
		MessageID:      123,
		UniqueID:       "abc123",
		RecipientEmail: "invalid-email", // Invalid email format
		DaysOld:        2,
		DecryptionURL:  "https://password.exchange/decrypt/abc123",
	}

	// Act
	err := service.ProcessMessageReminder(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid recipient email address")
	mockStorageRepo.AssertNotCalled(t, "GetReminderHistory")
	mockEmailSender.AssertNotCalled(t, "SendNotification")
}

// Test ProcessMessageReminder with storage error getting history
func TestProcessMessageReminder_HistoryStorageError_ReturnsError(t *testing.T) {
	// Arrange
	mockStorageRepo := &MockStorageRepository{}
	mockEmailSender := &MockNotificationSender{}
	service := NewReminderService(mockStorageRepo, mockEmailSender)

	ctx := context.Background()
	req := ReminderRequest{
		MessageID:      123,
		UniqueID:       "abc123",
		RecipientEmail: "test@example.com",
		DaysOld:        2,
		DecryptionURL:  "https://password.exchange/decrypt/abc123",
	}

	historyError := errors.New("database timeout")
	mockStorageRepo.On("GetReminderHistory", ctx, 123).Return(nil, historyError)

	// Act
	err := service.ProcessMessageReminder(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get reminder history")
	assert.Contains(t, err.Error(), "database timeout")
	mockStorageRepo.AssertExpectations(t)
	mockEmailSender.AssertNotCalled(t, "SendNotification")
}

// Test ProcessMessageReminder with email sending error
func TestProcessMessageReminder_EmailSendingError_ReturnsError(t *testing.T) {
	// Arrange
	mockStorageRepo := &MockStorageRepository{}
	mockEmailSender := &MockNotificationSender{}
	service := NewReminderService(mockStorageRepo, mockEmailSender)

	ctx := context.Background()
	req := ReminderRequest{
		MessageID:      123,
		UniqueID:       "abc123",
		RecipientEmail: "test@example.com",
		DaysOld:        2,
		DecryptionURL:  "https://password.exchange/decrypt/abc123",
	}

	// Mock storage calls
	mockStorageRepo.On("GetReminderHistory", ctx, 123).Return([]*ReminderLogEntry{}, nil)

	// Mock email sending failure
	emailError := errors.New("SMTP server unavailable")
	mockEmailSender.On("SendNotification", ctx, mock.AnythingOfType("NotificationRequest")).Return(nil, emailError)

	// Act
	err := service.ProcessMessageReminder(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send reminder email")
	assert.Contains(t, err.Error(), "SMTP server unavailable")
	mockStorageRepo.AssertExpectations(t)
	mockEmailSender.AssertExpectations(t)
}

// Test ProcessMessageReminder with existing reminder history
func TestProcessMessageReminder_ExistingHistory_IncrementsReminderNumber(t *testing.T) {
	// Arrange
	mockStorageRepo := &MockStorageRepository{}
	mockEmailSender := &MockNotificationSender{}
	service := NewReminderService(mockStorageRepo, mockEmailSender)

	ctx := context.Background()
	req := ReminderRequest{
		MessageID:      123,
		UniqueID:       "abc123",
		RecipientEmail: "test@example.com",
		DaysOld:        2,
		DecryptionURL:  "https://password.exchange/decrypt/abc123",
	}

	// Mock existing reminder history
	existingHistory := []*ReminderLogEntry{
		{
			MessageID:      123,
			RecipientEmail: "test@example.com",
			ReminderCount:  2, // Already sent 2 reminders
			SentAt:         time.Now().Add(-24 * time.Hour),
		},
	}

	// Mock storage calls
	mockStorageRepo.On("GetReminderHistory", ctx, 123).Return(existingHistory, nil)
	mockStorageRepo.On("LogReminderSent", ctx, 123, "test@example.com").Return(nil)

	// Mock email sending - expect reminder #3
	expectedResponse := &NotificationResponse{
		Success:   true,
		MessageID: "msg-456",
	}
	mockEmailSender.On("SendNotification", ctx, mock.MatchedBy(func(req NotificationRequest) bool {
		return req.Subject == "Reminder: You have an unviewed encrypted message (Reminder #3)"
	})).Return(expectedResponse, nil)

	// Act
	err := service.ProcessMessageReminder(ctx, req)

	// Assert
	assert.NoError(t, err)
	mockStorageRepo.AssertExpectations(t)
	mockEmailSender.AssertExpectations(t)
}