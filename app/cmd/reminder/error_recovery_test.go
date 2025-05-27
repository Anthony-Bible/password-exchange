package reminder

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorageService for testing
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) GetUnviewedMessagesForReminders(ctx context.Context, checkAfterHours, maxReminders int) ([]*domain.UnviewedMessage, error) {
	args := m.Called(ctx, checkAfterHours, maxReminders)
	return args.Get(0).([]*domain.UnviewedMessage), args.Error(1)
}

func (m *MockStorageService) GetReminderHistory(ctx context.Context, messageID int) ([]*domain.ReminderLogEntry, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).([]*domain.ReminderLogEntry), args.Error(1)
}

func (m *MockStorageService) LogReminderSent(ctx context.Context, messageID int, email string) error {
	args := m.Called(ctx, messageID, email)
	return args.Error(0)
}

// Additional required methods for StorageServicePort interface
func (m *MockStorageService) StoreMessage(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockStorageService) RetrieveMessage(ctx context.Context, uniqueID string) (*domain.Message, error) {
	args := m.Called(ctx, uniqueID)
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *MockStorageService) GetMessage(ctx context.Context, uniqueID string) (*domain.Message, error) {
	args := m.Called(ctx, uniqueID)
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *MockStorageService) CleanupExpiredMessages(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorageService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestRetryWithBackoff_Success(t *testing.T) {
	cfg := &config.Config{
		Reminder: config.ReminderConfig{
			Enabled:           true,
			CheckAfterHours:   24,
			MaxReminders:      3,
			ReminderInterval:  24,
		},
	}
	
	mockStorage := &MockStorageService{}
	processor := NewReminderProcessor(mockStorage, cfg)
	
	// Test successful operation on first attempt
	callCount := 0
	operation := func() error {
		callCount++
		return nil
	}
	
	ctx := context.Background()
	err := processor.RetryWithBackoff(ctx, operation, "test_operation")
	
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRetryWithBackoff_SuccessAfterRetries(t *testing.T) {
	cfg := &config.Config{
		Reminder: config.ReminderConfig{
			Enabled:           true,
			CheckAfterHours:   24,
			MaxReminders:      3,
			ReminderInterval:  24,
		},
	}
	
	mockStorage := &MockStorageService{}
	processor := NewReminderProcessor(mockStorage, cfg)
	
	// Test successful operation after 2 failures
	callCount := 0
	operation := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary failure")
		}
		return nil
	}
	
	ctx := context.Background()
	err := processor.RetryWithBackoff(ctx, operation, "test_operation")
	
	assert.NoError(t, err)
	assert.Equal(t, 3, callCount)
}

func TestRetryWithBackoff_MaxRetriesExceeded(t *testing.T) {
	cfg := &config.Config{
		Reminder: config.ReminderConfig{
			Enabled:           true,
			CheckAfterHours:   24,
			MaxReminders:      3,
			ReminderInterval:  24,
		},
	}
	
	mockStorage := &MockStorageService{}
	processor := NewReminderProcessor(mockStorage, cfg)
	
	// Test operation that always fails
	callCount := 0
	operation := func() error {
		callCount++
		return errors.New("persistent failure")
	}
	
	ctx := context.Background()
	err := processor.RetryWithBackoff(ctx, operation, "test_operation")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum retries exceeded")
	assert.Equal(t, MaxRetries, callCount)
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cb := &CircuitBreaker{
		state: CircuitBreakerClosed,
	}
	
	// Initially closed, should allow execution
	err := cb.CanExecute()
	assert.NoError(t, err)
	
	// Record failures to open circuit breaker
	for i := 0; i < CircuitBreakerThreshold; i++ {
		cb.RecordFailure()
	}
	
	// Should be open now
	assert.Equal(t, CircuitBreakerOpen, cb.state)
	err = cb.CanExecute()
	assert.Equal(t, ErrCircuitBreakerOpen, err)
	
	// Wait for timeout to transition to half-open
	cb.lastFailureTime = time.Now().Add(-CircuitBreakerTimeout - time.Second)
	err = cb.CanExecute()
	assert.NoError(t, err)
	assert.Equal(t, CircuitBreakerHalfOpen, cb.state)
	
	// Record success to close circuit breaker
	cb.RecordSuccess()
	assert.Equal(t, CircuitBreakerClosed, cb.state)
	assert.Equal(t, 0, cb.failureCount)
}

func TestCircuitBreaker_PreventExecution(t *testing.T) {
	cfg := &config.Config{
		Reminder: config.ReminderConfig{
			Enabled:           true,
			CheckAfterHours:   24,
			MaxReminders:      3,
			ReminderInterval:  24,
		},
	}
	
	mockStorage := &MockStorageService{}
	processor := NewReminderProcessor(mockStorage, cfg)
	
	// Open the circuit breaker by recording failures
	for i := 0; i < CircuitBreakerThreshold; i++ {
		processor.circuitBreaker.RecordFailure()
	}
	
	// Try to execute operation with open circuit breaker
	callCount := 0
	operation := func() error {
		callCount++
		return nil
	}
	
	ctx := context.Background()
	err := processor.RetryWithBackoff(ctx, operation, "test_operation")
	
	assert.Error(t, err)
	assert.Equal(t, ErrCircuitBreakerOpen, err)
	assert.Equal(t, 0, callCount) // Operation should not be called
}

func TestRetryWithBackoff_ContextCancellation(t *testing.T) {
	cfg := &config.Config{
		Reminder: config.ReminderConfig{
			Enabled:           true,
			CheckAfterHours:   24,
			MaxReminders:      3,
			ReminderInterval:  24,
		},
	}
	
	mockStorage := &MockStorageService{}
	processor := NewReminderProcessor(mockStorage, cfg)
	
	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	// Operation that always fails
	operation := func() error {
		return errors.New("failure")
	}
	
	start := time.Now()
	err := processor.RetryWithBackoff(ctx, operation, "test_operation")
	duration := time.Since(start)
	
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Less(t, duration, 200*time.Millisecond) // Should fail quickly due to context timeout
}

func TestProcessReminders_GracefulDegradation(t *testing.T) {
	cfg := &config.Config{
		Reminder: config.ReminderConfig{
			Enabled:           true,
			CheckAfterHours:   24,
			MaxReminders:      3,
			ReminderInterval:  24,
		},
	}
	
	mockStorage := &MockStorageService{}
	processor := NewReminderProcessor(mockStorage, cfg)
	
	// Mock messages
	messages := []*domain.UnviewedMessage{
		{MessageID: 1, RecipientEmail: "test1@example.com"},
		{MessageID: 2, RecipientEmail: "test2@example.com"},
		{MessageID: 3, RecipientEmail: "test3@example.com"},
	}
	
	// Mock successful retrieval of messages
	mockStorage.On("GetUnviewedMessagesForReminders", mock.Anything, 24, 3).Return(messages, nil)
	
	// Mock reminder history - first call succeeds, second fails, third succeeds
	mockStorage.On("GetReminderHistory", mock.Anything, 1).Return([]*domain.ReminderLogEntry{}, nil)
	mockStorage.On("GetReminderHistory", mock.Anything, 2).Return([]*domain.ReminderLogEntry(nil), errors.New("database error"))
	mockStorage.On("GetReminderHistory", mock.Anything, 3).Return([]*domain.ReminderLogEntry{}, nil)
	
	// Mock logging reminders
	mockStorage.On("LogReminderSent", mock.Anything, 1, "test1@example.com").Return(nil)
	mockStorage.On("LogReminderSent", mock.Anything, 3, "test3@example.com").Return(nil)
	
	ctx := context.Background()
	err := processor.ProcessReminders(ctx)
	
	// Should succeed with graceful degradation (2 out of 3 messages processed)
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

func TestProcessReminders_TotalFailure(t *testing.T) {
	cfg := &config.Config{
		Reminder: config.ReminderConfig{
			Enabled:           true,
			CheckAfterHours:   24,
			MaxReminders:      3,
			ReminderInterval:  24,
		},
	}
	
	mockStorage := &MockStorageService{}
	processor := NewReminderProcessor(mockStorage, cfg)
	
	// Mock messages
	messages := []*domain.UnviewedMessage{
		{MessageID: 1, RecipientEmail: "test1@example.com"},
		{MessageID: 2, RecipientEmail: "test2@example.com"},
	}
	
	// Mock successful retrieval of messages
	mockStorage.On("GetUnviewedMessagesForReminders", mock.Anything, 24, 3).Return(messages, nil)
	
	// Mock all operations to fail
	mockStorage.On("GetReminderHistory", mock.Anything, mock.Anything).Return([]*domain.ReminderLogEntry(nil), errors.New("database error"))
	
	ctx := context.Background()
	err := processor.ProcessReminders(ctx)
	
	// Should fail because no messages were processed
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to process any of 2 reminder messages")
	mockStorage.AssertExpectations(t)
}