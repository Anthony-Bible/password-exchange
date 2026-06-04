// Package batch provides a shared result type for batch operations that may
// experience partial failures. Callers record per-item outcomes and inspect
// the aggregate for reporting and exit-code decisions.
package batch

import (
	sherr "github.com/Anthony-Bible/password-exchange/app/internal/shared/errors"
)

// BatchError captures one failed item in a batch along with its categorized
// error, derived from the wrapped sentinel chain via sherr.Category.
type BatchError struct {
	ItemID   string
	Err      error
	Category sherr.Category
}

// BatchResult aggregates the outcome of a batch operation.
type BatchResult struct {
	TotalProcessed int
	SuccessCount   int
	FailureCount   int
	Errors         []BatchError
}

// NewBatchResult returns an empty BatchResult ready for use.
func NewBatchResult() *BatchResult {
	return &BatchResult{}
}

// RecordSuccess increments the success counter for the given item.
func (r *BatchResult) RecordSuccess(itemID string) {
	r.TotalProcessed++
	r.SuccessCount++
}

// RecordFailure appends a categorized failure for the given item.
func (r *BatchResult) RecordFailure(itemID string, err error) {
	r.TotalProcessed++
	r.FailureCount++
	r.Errors = append(r.Errors, BatchError{
		ItemID:   itemID,
		Err:      err,
		Category: sherr.CategoryOf(err),
	})
}

// HasFailures reports whether any item failed.
func (r *BatchResult) HasFailures() bool {
	return r.FailureCount > 0
}

// HasOperationalFailures reports whether any failure was retryable/operational,
// which is the signal monitoring should treat as a system-health concern.
func (r *BatchResult) HasOperationalFailures() bool {
	for _, e := range r.Errors {
		if e.Category == sherr.CategoryOperational {
			return true
		}
	}
	return false
}
