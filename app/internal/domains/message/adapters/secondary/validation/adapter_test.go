package validation

import (
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/stretchr/testify/assert"
)

// Compile-time guard: NewAdapter must satisfy secondary.ValidationPort.
var _ secondary.ValidationPort = NewAdapter()

func TestNewAdapter_ReturnsNonNil(t *testing.T) {
	t.Parallel()
	a := NewAdapter()
	assert.NotNil(t, a)
}

func TestSanitizeEmailForLogging(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string returns empty email sentinel",
			input: "",
			want:  "[EMPTY_EMAIL]",
		},
		{
			name:  "no at-sign returns invalid format sentinel",
			input: "notanemail",
			want:  "[INVALID_EMAIL_FORMAT]",
		},
		{
			name:  "multiple at-signs returns invalid format sentinel",
			input: "a@b@c",
			want:  "[INVALID_EMAIL_FORMAT]",
		},
		{
			name:  "one-char local part preserves single char with stars",
			input: "x@example.com",
			want:  "x***@example.com",
		},
		{
			name:  "two-char local part shows first char with double stars",
			input: "ab@example.com",
			want:  "a**@example.com",
		},
		{
			name:  "three-char local part shows first char with double stars",
			input: "abc@example.com",
			want:  "a**@example.com",
		},
		{
			name:  "four-char local part shows first and last chars around stars",
			input: "abcd@example.com",
			want:  "a***d@example.com",
		},
		{
			name:  "five-char local part (first) shows first and last chars around stars",
			input: "first@example.com",
			want:  "f***t@example.com",
		},
		{
			name:  "long local part shows first and last chars around stars",
			input: "anthony@example.com",
			want:  "a***y@example.com",
		},
		{
			name:  "very long local part shows first and last chars around stars",
			input: "verylongemail@example.com",
			want:  "v***l@example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := NewAdapter()
			got := a.SanitizeEmailForLogging(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}
