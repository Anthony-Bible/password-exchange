// Package errors provides shared error categorization primitives used across
// Password Exchange domains. It defines three error categories (Fatal,
// Operational, Business) and helpers that walk the standard fmt.Errorf %w chain
// via errors.As, so domain code can keep using sentinel errors and %w wrapping
// while gaining a single retry/handling vocabulary.
package errors

import (
	stderrors "errors"
)

// Category classifies an error for retry and handling decisions.
type Category int

const (
	// CategoryUnknown is the zero value, used when no category is attached.
	CategoryUnknown Category = iota
	// CategoryFatal indicates an unrecoverable condition (misconfiguration,
	// programmer error, missing dependency). Callers should not retry.
	CategoryFatal
	// CategoryOperational indicates a transient infrastructure failure
	// (network, queue, SMTP timeout). Callers may retry.
	CategoryOperational
	// CategoryBusiness indicates a domain rule violation or invalid input
	// (validation, authorization). Callers should not retry; it is not a
	// system failure.
	CategoryBusiness
)

// String returns a stable, lowercase identifier for the category, suitable
// for structured log fields and metrics labels.
func (c Category) String() string {
	switch c {
	case CategoryFatal:
		return "fatal"
	case CategoryOperational:
		return "operational"
	case CategoryBusiness:
		return "business"
	default:
		return "unknown"
	}
}

// Categorized is implemented by errors that carry a Category.
type Categorized interface {
	error
	Category() Category
}

type categoryError struct {
	err      error
	category Category
}

func (e *categoryError) Error() string      { return e.err.Error() }
func (e *categoryError) Unwrap() error      { return e.err }
func (e *categoryError) Category() Category { return e.category }

// WithCategory returns err annotated with the given category. The returned
// error implements Unwrap so errors.Is and errors.As continue to work against
// the wrapped error. Passing a nil err returns nil.
func WithCategory(err error, c Category) error {
	if err == nil {
		return nil
	}
	return &categoryError{err: err, category: c}
}

// CategoryOf walks the error chain via errors.As and returns the first
// attached Category. Returns CategoryUnknown if none is found or err is nil.
func CategoryOf(err error) Category {
	if err == nil {
		return CategoryUnknown
	}
	var c Categorized
	if stderrors.As(err, &c) {
		return c.Category()
	}
	return CategoryUnknown
}

// IsRetryable reports whether err should be retried. Only operational errors
// are retryable.
func IsRetryable(err error) bool {
	return CategoryOf(err) == CategoryOperational
}

// IsFatal reports whether err is categorized as fatal.
func IsFatal(err error) bool {
	return CategoryOf(err) == CategoryFatal
}

// IsBusiness reports whether err is categorized as a business/validation error.
func IsBusiness(err error) bool {
	return CategoryOf(err) == CategoryBusiness
}
