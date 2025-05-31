package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	mockStorageRepo := &MockStorageRepository{}
	mockNotificationPublisher := &MockNotificationPublisher{}
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher)

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
	mockStorageRepo := &MockStorageRepository{}
	mockNotificationPublisher := &MockNotificationPublisher{}
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher)

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
	mockStorageRepo := &MockStorageRepository{}
	mockNotificationPublisher := &MockNotificationPublisher{}
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher)

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
	mockStorageRepo := &MockStorageRepository{}
	mockNotificationPublisher := &MockNotificationPublisher{}
	service := NewReminderService(mockStorageRepo, mockNotificationPublisher)

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