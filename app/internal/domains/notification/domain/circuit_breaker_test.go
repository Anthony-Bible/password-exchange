package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
)

// Test CircuitBreaker CanExecute when closed
func TestCircuitBreaker_CanExecute_Closed_AllowsExecution(t *testing.T) {
	cb := &CircuitBreaker{state: CircuitBreakerClosed}
	err := cb.CanExecute()
	assert.NoError(t, err)
}

// Test CircuitBreaker CanExecute when open
func TestCircuitBreaker_CanExecute_Open_PreventsExecution(t *testing.T) {
	cb := &CircuitBreaker{
		state:           CircuitBreakerOpen,
		lastFailureTime: time.Now(),
	}
	err := cb.CanExecute()
	assert.Error(t, err)
	assert.Equal(t, ErrCircuitBreakerOpen, err)
}

// Test CircuitBreaker transitions from open to half-open after timeout
func TestCircuitBreaker_CanExecute_OpenTimeoutExpired_TransitionsToHalfOpen(t *testing.T) {
	cb := &CircuitBreaker{
		state:           CircuitBreakerOpen,
		lastFailureTime: time.Now().Add(-CircuitBreakerTimeout - time.Second),
	}
	err := cb.CanExecute()
	assert.NoError(t, err)
	assert.Equal(t, CircuitBreakerHalfOpen, cb.state)
}

// Test CircuitBreaker CanExecute when half-open
func TestCircuitBreaker_CanExecute_HalfOpen_AllowsExecution(t *testing.T) {
	cb := &CircuitBreaker{state: CircuitBreakerHalfOpen}
	err := cb.CanExecute()
	assert.NoError(t, err)
}

// Test CircuitBreaker RecordSuccess resets state
func TestCircuitBreaker_RecordSuccess_ResetsFaiilureCountAndCloses(t *testing.T) {
	cb := &CircuitBreaker{
		state:        CircuitBreakerHalfOpen,
		failureCount: 3,
	}
	cb.RecordSuccess()
	assert.Equal(t, 0, cb.failureCount)
	assert.Equal(t, CircuitBreakerClosed, cb.state)
}

// Test CircuitBreaker RecordFailure below threshold
func TestCircuitBreaker_RecordFailure_BelowThreshold_IncrementsCount(t *testing.T) {
	cb := &CircuitBreaker{
		state:        CircuitBreakerClosed,
		failureCount: 2,
	}
	cb.RecordFailure()
	assert.Equal(t, 3, cb.failureCount)
	assert.Equal(t, CircuitBreakerClosed, cb.state)
	assert.True(t, cb.lastFailureTime.After(time.Now().Add(-time.Second)))
}

// Test CircuitBreaker RecordFailure opens circuit at threshold
func TestCircuitBreaker_RecordFailure_AtThreshold_OpensCircuit(t *testing.T) {
	cb := &CircuitBreaker{
		state:        CircuitBreakerClosed,
		failureCount: CircuitBreakerThreshold - 1,
	}
	cb.RecordFailure()
	assert.Equal(t, CircuitBreakerThreshold, cb.failureCount)
	assert.Equal(t, CircuitBreakerOpen, cb.state)
	assert.True(t, cb.lastFailureTime.After(time.Now().Add(-time.Second)))
}

// Test retryWithBackoff success on first attempt
func TestRetryWithBackoff_SuccessFirstAttempt_NoRetry(t *testing.T) {
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	callCount := 0
	operation := func() error {
		callCount++
		return nil
	}

	err := service.retryWithBackoff(ctx, operation, "test_operation")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

// Test retryWithBackoff with eventual success
func TestRetryWithBackoff_EventualSuccess_RetriesAndSucceeds(t *testing.T) {
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	callCount := 0
	operation := func() error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := service.retryWithBackoff(ctx, operation, "test_operation")
	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

// Test retryWithBackoff max retries exceeded
func TestRetryWithBackoff_MaxRetriesExceeded_ReturnsError(t *testing.T) {
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx := context.Background()
	callCount := 0
	operation := func() error {
		callCount++
		return errors.New("persistent failure")
	}

	err := service.retryWithBackoff(ctx, operation, "test_operation")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum retries exceeded")
	assert.Equal(t, MaxRetries, callCount)
}

// Test retryWithBackoff context cancellation
func TestRetryWithBackoff_ContextCancelled_ReturnsContextError(t *testing.T) {
	mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation := createTestMocks()
	
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher, mockLogger, mockConfig, mockValidation)

	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0
	operation := func() error {
		callCount++
		if callCount == 1 {
			cancel()
		}
		return errors.New("failure")
	}

	err := service.retryWithBackoff(ctx, operation, "test_operation")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// Test CircuitBreaker logging when transitioning to OPEN state
func TestCircuitBreaker_RecordFailure_LogsWhenTransitioningToOpen(t *testing.T) {
	// Arrange - create circuit breaker with logger
	mockLogger := &MockLoggerPort{}
	mockLogEvent := &MockLogEvent{}
	
	// Set up expectations for logging when transitioning to OPEN state
	mockLogger.On("Error").Return(mockLogEvent)
	mockLogEvent.On("Str", "state", "OPEN").Return(mockLogEvent)
	mockLogEvent.On("Str", "reason", "threshold_exceeded").Return(mockLogEvent)
	mockLogEvent.On("Int", "failures", CircuitBreakerThreshold).Return(mockLogEvent)
	mockLogEvent.On("Msg", "Circuit breaker transitioned to OPEN state due to repeated failures").Return()
	
	cb := &CircuitBreaker{
		state:        CircuitBreakerClosed,
		failureCount: CircuitBreakerThreshold - 1,
		logger:       mockLogger,
	}

	// Act - record failure that should trigger state transition
	cb.RecordFailure()

	// Assert
	assert.Equal(t, CircuitBreakerOpen, cb.state)
	assert.Equal(t, CircuitBreakerThreshold, cb.failureCount)
	mockLogger.AssertExpectations(t)
	mockLogEvent.AssertExpectations(t)
}

// Test CircuitBreaker logging when transitioning from OPEN to HALF_OPEN
func TestCircuitBreaker_CanExecute_LogsWhenTransitioningToHalfOpen(t *testing.T) {
	// Arrange - create circuit breaker with logger in OPEN state with expired timeout
	mockLogger := &MockLoggerPort{}
	mockLogEvent := &MockLogEvent{}
	
	// Set up expectations for logging when transitioning to HALF_OPEN state
	mockLogger.On("Info").Return(mockLogEvent)
	mockLogEvent.On("Str", "state", "HALF_OPEN").Return(mockLogEvent)
	mockLogEvent.On("Str", "reason", "timeout_expired").Return(mockLogEvent)
	mockLogEvent.On("Dur", "timeout", CircuitBreakerTimeout).Return(mockLogEvent)
	mockLogEvent.On("Msg", "Circuit breaker transitioned to HALF_OPEN state after timeout").Return()
	
	cb := &CircuitBreaker{
		state:           CircuitBreakerOpen,
		lastFailureTime: time.Now().Add(-CircuitBreakerTimeout - time.Second),
		logger:          mockLogger,
	}

	// Act - call CanExecute which should trigger state transition
	err := cb.CanExecute()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, CircuitBreakerHalfOpen, cb.state)
	mockLogger.AssertExpectations(t)
	mockLogEvent.AssertExpectations(t)
}

// Test CircuitBreaker logging when transitioning from HALF_OPEN to CLOSED
func TestCircuitBreaker_RecordSuccess_LogsWhenTransitioningToClosed(t *testing.T) {
	// Arrange - create circuit breaker with logger in HALF_OPEN state
	mockLogger := &MockLoggerPort{}
	mockLogEvent := &MockLogEvent{}
	
	// Set up expectations for logging when transitioning to CLOSED state
	mockLogger.On("Info").Return(mockLogEvent)
	mockLogEvent.On("Str", "state", "CLOSED").Return(mockLogEvent)
	mockLogEvent.On("Str", "reason", "operation_succeeded").Return(mockLogEvent)
	mockLogEvent.On("Msg", "Circuit breaker transitioned to CLOSED state after successful operation").Return()
	
	cb := &CircuitBreaker{
		state:        CircuitBreakerHalfOpen,
		failureCount: 3,
		logger:       mockLogger,
	}

	// Act - record success which should transition to closed
	cb.RecordSuccess()

	// Assert
	assert.Equal(t, CircuitBreakerClosed, cb.state)
	assert.Equal(t, 0, cb.failureCount)
	mockLogger.AssertExpectations(t)
	mockLogEvent.AssertExpectations(t)
}

// Test CircuitBreaker with nil logger doesn't panic
func TestCircuitBreaker_NilLogger_DoesNotPanic(t *testing.T) {
	// Arrange - create circuit breaker without logger
	cb := &CircuitBreaker{
		state:        CircuitBreakerClosed,
		failureCount: CircuitBreakerThreshold - 1,
		logger:       nil,
	}

	// Act & Assert - operations should not panic with nil logger
	assert.NotPanics(t, func() {
		cb.RecordFailure()
	})
	
	assert.NotPanics(t, func() {
		cb.RecordSuccess()
	})
	
	assert.NotPanics(t, func() {
		cb.CanExecute()
	})
}