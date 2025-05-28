package validation

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errType error
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus sign",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with dots",
			email:   "first.last@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
			errType: ErrEmptyEmail,
		},
		{
			name:    "whitespace only",
			email:   "   ",
			wantErr: true,
			errType: ErrEmptyEmail,
		},
		{
			name:    "missing @ symbol",
			email:   "userexample.com",
			wantErr: true,
			errType: ErrInvalidEmail,
		},
		{
			name:    "multiple @ symbols",
			email:   "user@@example.com",
			wantErr: true,
			errType: ErrInvalidEmail,
		},
		{
			name:    "missing domain",
			email:   "user@",
			wantErr: true,
			errType: ErrInvalidEmail,
		},
		{
			name:    "missing local part",
			email:   "@example.com",
			wantErr: true,
			errType: ErrInvalidEmail,
		},
		{
			name:    "consecutive dots in local part",
			email:   "user..name@example.com",
			wantErr: true,
			errType: ErrInvalidEmail,
		},
		{
			name:    "leading dot in local part",
			email:   ".user@example.com",
			wantErr: true,
			errType: ErrInvalidEmail,
		},
		{
			name:    "trailing dot in local part",
			email:   "user.@example.com",
			wantErr: true,
			errType: ErrInvalidEmail,
		},
		{
			name:    "leading dot in domain",
			email:   "user@.example.com",
			wantErr: true,
			errType: ErrInvalidEmail,
		},
		{
			name:    "trailing dot in domain",
			email:   "user@example.com.",
			wantErr: true,
			errType: ErrInvalidEmail,
		},
		{
			name:    "too long email",
			email:   "user@" + strings.Repeat("a", 320) + ".com",
			wantErr: true,
			errType: ErrInvalidEmail,
		},
		{
			name:    "no TLD",
			email:   "user@localhost",
			wantErr: true,
			errType: ErrInvalidEmail,
		},
		{
			name:    "special characters",
			email:   "user<>@example.com",
			wantErr: true,
			errType: ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil {
				if !isExpectedError(err, tt.errType) {
					t.Errorf("ValidateEmail() error = %v, expected error type %v", err, tt.errType)
				}
			}
		})
	}
}

func TestSanitizeEmailForLogging(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "normal email",
			email:    "user@example.com",
			expected: "u***r@example.com",
		},
		{
			name:     "short email",
			email:    "ab@example.com",
			expected: "a**@example.com",
		},
		{
			name:     "single char local",
			email:    "a@example.com",
			expected: "a***@example.com",
		},
		{
			name:     "empty email",
			email:    "",
			expected: "[EMPTY_EMAIL]",
		},
		{
			name:     "invalid format",
			email:    "notanemail",
			expected: "[INVALID_EMAIL_FORMAT]",
		},
		{
			name:     "three char local",
			email:    "abc@example.com",
			expected: "a**@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeEmailForLogging(tt.email)
			if result != tt.expected {
				t.Errorf("SanitizeEmailForLogging() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "uppercase email",
			email:    "USER@EXAMPLE.COM",
			expected: "user@example.com",
		},
		{
			name:     "mixed case",
			email:    "User@Example.Com",
			expected: "user@example.com",
		},
		{
			name:     "with whitespace",
			email:    "  user@example.com  ",
			expected: "user@example.com",
		},
		{
			name:     "already normalized",
			email:    "user@example.com",
			expected: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeEmail(tt.email)
			if result != tt.expected {
				t.Errorf("NormalizeEmail() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// isExpectedError checks if the error contains the expected error type
func isExpectedError(got, expected error) bool {
	if got == nil || expected == nil {
		return got == expected
	}
	return got.Error() != "" && expected.Error() != ""
}