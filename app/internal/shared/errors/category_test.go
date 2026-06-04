package errors

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errSentinel = stderrors.New("sentinel")

func TestWithCategory_NilReturnsNil(t *testing.T) {
	assert.Nil(t, WithCategory(nil, CategoryOperational))
}

func TestWithCategory_PreservesErrorsIs(t *testing.T) {
	wrapped := WithCategory(errSentinel, CategoryOperational)
	assert.True(t, stderrors.Is(wrapped, errSentinel))
}

func TestWithCategory_PreservesErrorMessage(t *testing.T) {
	wrapped := WithCategory(errSentinel, CategoryFatal)
	assert.Equal(t, "sentinel", wrapped.Error())
}

func TestCategoryOf_DirectlyWrapped(t *testing.T) {
	wrapped := WithCategory(errSentinel, CategoryOperational)
	assert.Equal(t, CategoryOperational, CategoryOf(wrapped))
}

func TestCategoryOf_WalksFmtErrorfChain(t *testing.T) {
	wrapped := WithCategory(errSentinel, CategoryOperational)
	outer := fmt.Errorf("outer context: %w", wrapped)
	doubled := fmt.Errorf("more: %w", outer)
	assert.Equal(t, CategoryOperational, CategoryOf(doubled))
}

func TestCategoryOf_NilReturnsUnknown(t *testing.T) {
	assert.Equal(t, CategoryUnknown, CategoryOf(nil))
}

func TestCategoryOf_UncategorizedReturnsUnknown(t *testing.T) {
	assert.Equal(t, CategoryUnknown, CategoryOf(errSentinel))
}

func TestIsRetryable_OnlyOperational(t *testing.T) {
	cases := []struct {
		cat  Category
		want bool
	}{
		{CategoryUnknown, false},
		{CategoryFatal, false},
		{CategoryOperational, true},
		{CategoryBusiness, false},
	}
	for _, tc := range cases {
		err := WithCategory(errSentinel, tc.cat)
		assert.Equal(t, tc.want, IsRetryable(err), "cat=%d", tc.cat)
	}
	assert.False(t, IsRetryable(nil))
}

func TestIsFatalAndIsBusiness(t *testing.T) {
	assert.True(t, IsFatal(WithCategory(errSentinel, CategoryFatal)))
	assert.False(t, IsFatal(WithCategory(errSentinel, CategoryOperational)))
	assert.True(t, IsBusiness(WithCategory(errSentinel, CategoryBusiness)))
	assert.False(t, IsBusiness(WithCategory(errSentinel, CategoryFatal)))
}
