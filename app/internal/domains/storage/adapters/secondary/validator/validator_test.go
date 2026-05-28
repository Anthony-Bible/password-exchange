package validator

import (
	"strings"
	"testing"
)

func TestNewValidationAdapter_ReturnsNonNil(t *testing.T) {
	adapter := NewValidationAdapter()
	if adapter == nil {
		t.Fatal("NewValidationAdapter() returned nil, want non-nil")
	}
}

func TestValidateEmail_AcceptsValidEmail(t *testing.T) {
	adapter := NewValidationAdapter()
	if err := adapter.ValidateEmail("user@example.com"); err != nil {
		t.Errorf("ValidateEmail(%q) returned error %v, want nil", "user@example.com", err)
	}
}

func TestValidateEmail_RejectsInvalidEmail(t *testing.T) {
	adapter := NewValidationAdapter()
	if err := adapter.ValidateEmail("not-an-email"); err == nil {
		t.Errorf("ValidateEmail(%q) returned nil, want non-nil error", "not-an-email")
	}
}

func TestValidateEmail_RejectsEmptyEmail(t *testing.T) {
	adapter := NewValidationAdapter()
	if err := adapter.ValidateEmail(""); err == nil {
		t.Errorf("ValidateEmail(%q) returned nil, want non-nil error", "")
	}
}

func TestSanitizeEmailForLogging_MasksEmail(t *testing.T) {
	adapter := NewValidationAdapter()
	const input = "user@example.com"
	got := adapter.SanitizeEmailForLogging(input)

	t.Run("non-empty", func(t *testing.T) {
		if got == "" {
			t.Errorf("SanitizeEmailForLogging(%q) returned empty string, want non-empty", input)
		}
	})

	t.Run("not equal to input", func(t *testing.T) {
		if got == input {
			t.Errorf("SanitizeEmailForLogging(%q) returned %q, want different from input", input, got)
		}
	})

	t.Run("domain preserved", func(t *testing.T) {
		if !strings.Contains(got, "@example.com") {
			t.Errorf("SanitizeEmailForLogging(%q) = %q, want contains %q", input, got, "@example.com")
		}
	})
}

func TestSanitizeEmailForLogging_HandlesEmpty(t *testing.T) {
	adapter := NewValidationAdapter()
	got := adapter.SanitizeEmailForLogging("")
	const want = "[EMPTY_EMAIL]"
	if got != want {
		t.Errorf("SanitizeEmailForLogging(%q) = %q, want %q", "", got, want)
	}
}
