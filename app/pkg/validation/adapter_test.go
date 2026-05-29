package validation

import (
	"strings"
	"testing"
)

func TestAdapter_ValidateEmail_AcceptsValidEmail(t *testing.T) {
	if err := NewAdapter().ValidateEmail("user@example.com"); err != nil {
		t.Errorf("ValidateEmail(%q) returned error %v, want nil", "user@example.com", err)
	}
}

func TestAdapter_ValidateEmail_RejectsInvalidEmail(t *testing.T) {
	if err := NewAdapter().ValidateEmail("not-an-email"); err == nil {
		t.Errorf("ValidateEmail(%q) returned nil, want non-nil error", "not-an-email")
	}
}

func TestAdapter_ValidateEmail_RejectsEmptyEmail(t *testing.T) {
	if err := NewAdapter().ValidateEmail(""); err == nil {
		t.Errorf("ValidateEmail(%q) returned nil, want non-nil error", "")
	}
}

func TestAdapter_SanitizeEmailForLogging_MasksEmail(t *testing.T) {
	const input = "user@example.com"
	got := NewAdapter().SanitizeEmailForLogging(input)

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

func TestAdapter_SanitizeEmailForLogging_HandlesEmpty(t *testing.T) {
	got := NewAdapter().SanitizeEmailForLogging("")
	const want = "[EMPTY_EMAIL]"
	if got != want {
		t.Errorf("SanitizeEmailForLogging(%q) = %q, want %q", "", got, want)
	}
}
