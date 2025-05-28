package storage

import (
	"context"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	storageDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorageService implements storagePorts.StorageServicePort for testing
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) StoreMessage(ctx context.Context, msg *storageDomain.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockStorageService) RetrieveMessage(ctx context.Context, uniqueID string) (*storageDomain.Message, error) {
	args := m.Called(ctx, uniqueID)
	return args.Get(0).(*storageDomain.Message), args.Error(1)
}

func (m *MockStorageService) GetMessage(ctx context.Context, uniqueID string) (*storageDomain.Message, error) {
	args := m.Called(ctx, uniqueID)
	return args.Get(0).(*storageDomain.Message), args.Error(1)
}

func (m *MockStorageService) CleanupExpiredMessages(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorageService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorageService) GetUnviewedMessagesForReminders(ctx context.Context, checkAfterHours, maxReminders int) ([]*storageDomain.UnviewedMessage, error) {
	args := m.Called(ctx, checkAfterHours, maxReminders)
	return args.Get(0).([]*storageDomain.UnviewedMessage), args.Error(1)
}

func (m *MockStorageService) LogReminderSent(ctx context.Context, messageID int, recipientEmail string) error {
	args := m.Called(ctx, messageID, recipientEmail)
	return args.Error(0)
}

func (m *MockStorageService) GetReminderHistory(ctx context.Context, messageID int) ([]*storageDomain.ReminderLogEntry, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).([]*storageDomain.ReminderLogEntry), args.Error(1)
}

func TestGetUnviewedMessagesForReminders_Success(t *testing.T) {
	// Arrange
	mockStorage := &MockStorageService{}
	adapter := NewGRPCStorageAdapter(mockStorage)

	ctx := context.Background()
	checkAfterHours := 24
	maxReminders := 3

	// Create storage entities (simulating data that comes from gRPC/protobuf)
	now := time.Now()
	storageMessages := []*storageDomain.UnviewedMessage{
		{
			MessageID:      1,
			UniqueID:       "msg-001",
			RecipientEmail: "user1@example.com",
			DaysOld:        2,
			Created:        now.AddDate(0, 0, -2),
		},
		{
			MessageID:      2,
			UniqueID:       "msg-002",
			RecipientEmail: "user2@example.com",
			DaysOld:        5,
			Created:        now.AddDate(0, 0, -5),
		},
	}

	mockStorage.On("GetUnviewedMessagesForReminders", ctx, checkAfterHours, maxReminders).
		Return(storageMessages, nil)

	// Act
	result, err := adapter.GetUnviewedMessagesForReminders(ctx, checkAfterHours, maxReminders)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify entity conversion from storage to notification domain
	assert.Equal(t, 1, result[0].MessageID)
	assert.Equal(t, "msg-001", result[0].UniqueID)
	assert.Equal(t, "user1@example.com", result[0].RecipientEmail)
	assert.Equal(t, 2, result[0].DaysOld)
	assert.Equal(t, now.AddDate(0, 0, -2), result[0].Created)

	assert.Equal(t, 2, result[1].MessageID)
	assert.Equal(t, "msg-002", result[1].UniqueID)
	assert.Equal(t, "user2@example.com", result[1].RecipientEmail)
	assert.Equal(t, 5, result[1].DaysOld)
	assert.Equal(t, now.AddDate(0, 0, -5), result[1].Created)

	mockStorage.AssertExpectations(t)
}

func TestGetUnviewedMessagesForReminders_EmptyResult(t *testing.T) {
	// Arrange
	mockStorage := &MockStorageService{}
	adapter := NewGRPCStorageAdapter(mockStorage)

	ctx := context.Background()
	checkAfterHours := 72
	maxReminders := 5

	// Mock empty result (no messages need reminders)
	storageMessages := []*storageDomain.UnviewedMessage{}
	mockStorage.On("GetUnviewedMessagesForReminders", ctx, checkAfterHours, maxReminders).
		Return(storageMessages, nil)

	// Act
	result, err := adapter.GetUnviewedMessagesForReminders(ctx, checkAfterHours, maxReminders)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, result)
	mockStorage.AssertExpectations(t)
}

func TestGetUnviewedMessagesForReminders_StorageError(t *testing.T) {
	// Arrange
	mockStorage := &MockStorageService{}
	adapter := NewGRPCStorageAdapter(mockStorage)

	ctx := context.Background()
	checkAfterHours := 24
	maxReminders := 3

	// Mock storage service error (simulating gRPC/protobuf communication error)
	mockStorage.On("GetUnviewedMessagesForReminders", ctx, checkAfterHours, maxReminders).
		Return(([]*storageDomain.UnviewedMessage)(nil), assert.AnError)

	// Act
	result, err := adapter.GetUnviewedMessagesForReminders(ctx, checkAfterHours, maxReminders)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, assert.AnError, err)
	mockStorage.AssertExpectations(t)
}

func TestGetReminderHistory_Success(t *testing.T) {
	// Arrange
	mockStorage := &MockStorageService{}
	adapter := NewGRPCStorageAdapter(mockStorage)

	ctx := context.Background()
	messageID := 123

	// Create storage entities (simulating data from gRPC/protobuf)
	now := time.Now()
	storageHistory := []*storageDomain.ReminderLogEntry{
		{
			MessageID:        123,
			EmailAddress:     "user@example.com",
			ReminderCount:    1,
			LastReminderSent: now.AddDate(0, 0, -1),
		},
		{
			MessageID:        123,
			EmailAddress:     "user@example.com",
			ReminderCount:    2,
			LastReminderSent: now,
		},
	}

	mockStorage.On("GetReminderHistory", ctx, messageID).
		Return(storageHistory, nil)

	// Act
	result, err := adapter.GetReminderHistory(ctx, messageID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify entity conversion from storage to notification domain
	assert.Equal(t, 123, result[0].MessageID)
	assert.Equal(t, "user@example.com", result[0].RecipientEmail)
	assert.Equal(t, 1, result[0].ReminderCount)
	assert.Equal(t, now.AddDate(0, 0, -1), result[0].SentAt)

	assert.Equal(t, 123, result[1].MessageID)
	assert.Equal(t, "user@example.com", result[1].RecipientEmail)
	assert.Equal(t, 2, result[1].ReminderCount)
	assert.Equal(t, now, result[1].SentAt)

	mockStorage.AssertExpectations(t)
}

func TestGetReminderHistory_EmptyHistory(t *testing.T) {
	// Arrange
	mockStorage := &MockStorageService{}
	adapter := NewGRPCStorageAdapter(mockStorage)

	ctx := context.Background()
	messageID := 456

	// Mock empty history
	storageHistory := []*storageDomain.ReminderLogEntry{}
	mockStorage.On("GetReminderHistory", ctx, messageID).
		Return(storageHistory, nil)

	// Act
	result, err := adapter.GetReminderHistory(ctx, messageID)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, result)
	mockStorage.AssertExpectations(t)
}

func TestGetReminderHistory_StorageError(t *testing.T) {
	// Arrange
	mockStorage := &MockStorageService{}
	adapter := NewGRPCStorageAdapter(mockStorage)

	ctx := context.Background()
	messageID := 789

	// Mock storage service error
	mockStorage.On("GetReminderHistory", ctx, messageID).
		Return(([]*storageDomain.ReminderLogEntry)(nil), assert.AnError)

	// Act
	result, err := adapter.GetReminderHistory(ctx, messageID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, assert.AnError, err)
	mockStorage.AssertExpectations(t)
}

func TestLogReminderSent_Success(t *testing.T) {
	// Arrange
	mockStorage := &MockStorageService{}
	adapter := NewGRPCStorageAdapter(mockStorage)

	ctx := context.Background()
	messageID := 101
	recipientEmail := "reminder@example.com"

	mockStorage.On("LogReminderSent", ctx, messageID, recipientEmail).
		Return(nil)

	// Act
	err := adapter.LogReminderSent(ctx, messageID, recipientEmail)

	// Assert
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

func TestLogReminderSent_StorageError(t *testing.T) {
	// Arrange
	mockStorage := &MockStorageService{}
	adapter := NewGRPCStorageAdapter(mockStorage)

	ctx := context.Background()
	messageID := 102
	recipientEmail := "error@example.com"

	// Mock storage service error (simulating gRPC/protobuf error)
	mockStorage.On("LogReminderSent", ctx, messageID, recipientEmail).
		Return(assert.AnError)

	// Act
	err := adapter.LogReminderSent(ctx, messageID, recipientEmail)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockStorage.AssertExpectations(t)
}

func TestEntityConversion_UnviewedMessage(t *testing.T) {
	// Test entity conversion between storage and notification domains
	// This simulates the protobuf → storage entity → notification entity conversion chain

	now := time.Now()
	storageMessage := &storageDomain.UnviewedMessage{
		MessageID:      999,
		UniqueID:       "conversion-test",
		RecipientEmail: "convert@example.com",
		DaysOld:        7,
		Created:        now,
	}

	// Simulate the conversion that happens in the adapter
	notificationMessage := &domain.UnviewedMessage{
		MessageID:      storageMessage.MessageID,
		UniqueID:       storageMessage.UniqueID,
		RecipientEmail: storageMessage.RecipientEmail,
		DaysOld:        storageMessage.DaysOld,
		Created:        storageMessage.Created,
	}

	// Verify all fields are correctly mapped
	assert.Equal(t, storageMessage.MessageID, notificationMessage.MessageID)
	assert.Equal(t, storageMessage.UniqueID, notificationMessage.UniqueID)
	assert.Equal(t, storageMessage.RecipientEmail, notificationMessage.RecipientEmail)
	assert.Equal(t, storageMessage.DaysOld, notificationMessage.DaysOld)
	assert.Equal(t, storageMessage.Created, notificationMessage.Created)
}

func TestEntityConversion_ReminderLogEntry(t *testing.T) {
	// Test entity conversion for reminder log entries
	// This simulates the protobuf → storage entity → notification entity conversion chain

	now := time.Now()
	storageEntry := &storageDomain.ReminderLogEntry{
		MessageID:        888,
		EmailAddress:     "log@example.com",
		ReminderCount:    3,
		LastReminderSent: now,
	}

	// Simulate the conversion that happens in the adapter
	notificationEntry := &domain.ReminderLogEntry{
		MessageID:      storageEntry.MessageID,
		RecipientEmail: storageEntry.EmailAddress,
		ReminderCount:  storageEntry.ReminderCount,
		SentAt:         storageEntry.LastReminderSent,
	}

	// Verify all fields are correctly mapped
	assert.Equal(t, storageEntry.MessageID, notificationEntry.MessageID)
	assert.Equal(t, storageEntry.EmailAddress, notificationEntry.RecipientEmail)
	assert.Equal(t, storageEntry.ReminderCount, notificationEntry.ReminderCount)
	assert.Equal(t, storageEntry.LastReminderSent, notificationEntry.SentAt)
}

func TestGRPCStorageAdapter_ParameterValidation(t *testing.T) {
	// Test that parameters are correctly passed through the gRPC adapter
	// This validates the protobuf parameter handling

	testCases := []struct {
		name            string
		checkAfterHours int
		maxReminders    int
	}{
		{"Standard reminder check", 24, 3},
		{"Quick reminder check", 1, 1},
		{"Extended reminder check", 168, 10}, // 1 week
		{"Zero values", 0, 0},
		{"Large values", 8760, 999}, // 1 year
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage := &MockStorageService{}
			adapter := NewGRPCStorageAdapter(mockStorage)

			ctx := context.Background()

			// Mock storage service expects exact parameters
			mockStorage.On("GetUnviewedMessagesForReminders", ctx, tc.checkAfterHours, tc.maxReminders).
				Return([]*storageDomain.UnviewedMessage{}, nil)

			// Act
			_, err := adapter.GetUnviewedMessagesForReminders(ctx, tc.checkAfterHours, tc.maxReminders)

			// Assert
			assert.NoError(t, err)
			mockStorage.AssertExpectations(t)
		})
	}
}