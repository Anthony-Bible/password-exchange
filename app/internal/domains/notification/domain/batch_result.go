// Package domain – reminder batch reporting.
//
// BatchResult records per-item outcomes from ProcessReminders so callers can
// distinguish total failure from partial success, and infrastructure failures
// from invalid-input failures. Lives in this package because the reminder
// pipeline is the only batch processor in the app today; promote to a shared
// package only when a second consumer appears.
package domain

import stderrors "errors"

// BatchError captures one failed item in a reminder batch.
type BatchError struct {
	ItemID string
	Err    error
}

// BatchResult aggregates the outcome of a reminder batch.
type BatchResult struct {
	TotalProcessed int
	SuccessCount   int
	FailureCount   int
	Errors         []BatchError
}

// NewBatchResult returns an empty BatchResult.
func NewBatchResult() *BatchResult { return &BatchResult{} }

// RecordSuccess increments the success counter.
func (r *BatchResult) RecordSuccess(itemID string) {
	r.TotalProcessed++
	r.SuccessCount++
}

// RecordFailure appends a failure entry.
func (r *BatchResult) RecordFailure(itemID string, err error) {
	r.TotalProcessed++
	r.FailureCount++
	r.Errors = append(r.Errors, BatchError{ItemID: itemID, Err: err})
}

// HasFailures reports whether any item failed.
func (r *BatchResult) HasFailures() bool { return r.FailureCount > 0 }

// HasInfraFailures reports whether any failure looks like a transient
// infrastructure problem (queue, SMTP, storage). Useful for alerting:
// validation failures are not a system health concern, infra failures are.
func (r *BatchResult) HasInfraFailures() bool {
	for _, e := range r.Errors {
		if isInfraError(e.Err) {
			return true
		}
	}
	return false
}

// nonRetryableSentinels are domain errors where retrying is guaranteed to
// keep failing — validation problems and missing-config conditions. Anything
// not in this list is treated as transient and eligible for retry.
var nonRetryableSentinels = []error{
	ErrInvalidEmailAddress,
	ErrInvalidNotificationRequest,
	ErrMessageUnmarshalFailed,
	ErrEmptyMessageBody,
	ErrTemplateNotFound,
	ErrInvalidCheckAfterHours,
	ErrInvalidMaxReminders,
	ErrInvalidReminderInterval,
}

// isRetryable reports whether retryWithBackoff should attempt err again.
func isRetryable(err error) bool {
	for _, s := range nonRetryableSentinels {
		if stderrors.Is(err, s) {
			return false
		}
	}
	return true
}

// infraSentinels are domain errors that indicate transient infrastructure
// problems worth surfacing to monitoring.
var infraSentinels = []error{
	ErrEmailSendFailed,
	ErrQueueConnectionFailed,
	ErrQueueConsumeFailed,
	ErrTemplateRenderFailed,
	ErrSMTPAuthFailed,
}

func isInfraError(err error) bool {
	for _, s := range infraSentinels {
		if stderrors.Is(err, s) {
			return true
		}
	}
	return false
}
