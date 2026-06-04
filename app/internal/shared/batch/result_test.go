package batch

import (
	stderrors "errors"
	"fmt"
	"testing"

	sherr "github.com/Anthony-Bible/password-exchange/app/internal/shared/errors"
	"github.com/stretchr/testify/assert"
)

var errSentinel = sherr.WithCategory(stderrors.New("transient"), sherr.CategoryOperational)
var errBusiness = sherr.WithCategory(stderrors.New("bad input"), sherr.CategoryBusiness)

func TestNewBatchResult_Empty(t *testing.T) {
	r := NewBatchResult()
	assert.Equal(t, 0, r.TotalProcessed)
	assert.Equal(t, 0, r.SuccessCount)
	assert.Equal(t, 0, r.FailureCount)
	assert.False(t, r.HasFailures())
	assert.False(t, r.HasOperationalFailures())
}

func TestRecordSuccess_IncrementsCounters(t *testing.T) {
	r := NewBatchResult()
	r.RecordSuccess("item-1")
	r.RecordSuccess("item-2")
	assert.Equal(t, 2, r.TotalProcessed)
	assert.Equal(t, 2, r.SuccessCount)
	assert.Equal(t, 0, r.FailureCount)
}

func TestRecordFailure_AppendsCategorizedError(t *testing.T) {
	r := NewBatchResult()
	wrapped := fmt.Errorf("context: %w", errSentinel)
	r.RecordFailure("item-1", wrapped)

	assert.Equal(t, 1, r.TotalProcessed)
	assert.Equal(t, 1, r.FailureCount)
	assert.Equal(t, 0, r.SuccessCount)
	assert.Len(t, r.Errors, 1)
	assert.Equal(t, "item-1", r.Errors[0].ItemID)
	assert.Equal(t, sherr.CategoryOperational, r.Errors[0].Category)
	assert.True(t, stderrors.Is(r.Errors[0].Err, errSentinel))
}

func TestHasFailures(t *testing.T) {
	r := NewBatchResult()
	r.RecordSuccess("ok")
	assert.False(t, r.HasFailures())
	r.RecordFailure("bad", errBusiness)
	assert.True(t, r.HasFailures())
}

func TestHasOperationalFailures(t *testing.T) {
	r := NewBatchResult()
	r.RecordFailure("a", errBusiness)
	assert.False(t, r.HasOperationalFailures())
	r.RecordFailure("b", errSentinel)
	assert.True(t, r.HasOperationalFailures())
}
