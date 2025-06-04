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

func TestSanitizeEmailHeaderValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean string",
			input:    "Clean Value",
			expected: "Clean Value",
		},
		{
			name:     "carriage return",
			input:    "Bad\rValue",
			expected: "BadValue",
		},
		{
			name:     "line feed",
			input:    "Bad\nValue",
			expected: "BadValue",
		},
		{
			name:     "CRLF injection attempt",
			input:    "user@domain.com\r\nBcc: attacker@evil.com",
			expected: "user@domain.comBcc: attacker@evil.com",
		},
		{
			name:     "multiple CRLF",
			input:    "Header\r\nBcc: evil@bad.com\r\nFrom: fake@sender.com",
			expected: "HeaderBcc: evil@bad.comFrom: fake@sender.com",
		},
		{
			name:     "tab characters",
			input:    "Value\twith\ttabs",
			expected: "Value with tabs",
		},
		{
			name:     "vertical tab",
			input:    "Value\vwith\vvtab",
			expected: "Valuewithvtab",
		},
		{
			name:     "form feed",
			input:    "Value\fwith\fformfeed",
			expected: "Valuewithformfeed",
		},
		{
			name:     "mixed control characters",
			input:    "Mixed\r\n\t\v\fcharacters",
			expected: "Mixed characters", // \r\n removed, \t becomes space, \v\f removed
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeEmailHeaderValue(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeEmailHeaderValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateEmailHeaderValue(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "clean value",
			input:   "Clean Header Value",
			wantErr: false,
		},
		{
			name:      "carriage return injection",
			input:     "Value\rWith CR",
			wantErr:   true,
			errSubstr: "CRLF characters found",
		},
		{
			name:      "line feed injection",
			input:     "Value\nWith LF",
			wantErr:   true,
			errSubstr: "CRLF characters found",
		},
		{
			name:      "CRLF header injection",
			input:     "user@domain.com\r\nBcc: attacker@evil.com",
			wantErr:   true,
			errSubstr: "CRLF characters found",
		},
		{
			name:      "email spoofing attempt",
			input:     "user@domain.com\r\nFrom: admin@bank.com\r\nBcc: attacker@evil.com",
			wantErr:   true,
			errSubstr: "CRLF characters found",
		},
		{
			name:      "tab character",
			input:     "Value\twith tab",
			wantErr:   true,
			errSubstr: "control character found",
		},
		{
			name:      "vertical tab",
			input:     "Value\vwith vtab",
			wantErr:   true,
			errSubstr: "control character found",
		},
		{
			name:      "form feed",
			input:     "Value\fwith formfeed",
			wantErr:   true,
			errSubstr: "control character found",
		},
		{
			name:      "backspace character",
			input:     "Value\bwith backspace",
			wantErr:   true,
			errSubstr: "control character found",
		},
		{
			name:      "header-like pattern",
			input:     "Bcc: attacker@evil.com",
			wantErr:   true,
			errSubstr: "header-like pattern found",
		},
		{
			name:      "from header pattern",
			input:     "From: fake@sender.com",
			wantErr:   true,
			errSubstr: "header-like pattern found",
		},
		{
			name:      "to header pattern",
			input:     "To: victim@target.com",
			wantErr:   true,
			errSubstr: "header-like pattern found",
		},
		{
			name:      "cc header pattern",
			input:     "Cc: copy@recipient.com",
			wantErr:   true,
			errSubstr: "header-like pattern found",
		},
		{
			name:      "custom header pattern",
			input:     "X-Custom-Header: malicious value",
			wantErr:   true,
			errSubstr: "header-like pattern found",
		},
		{
			name:    "colon in email address (legitimate)",
			input:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "normal subject line",
			input:   "Password Exchange: New encrypted message",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmailHeaderValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmailHeaderValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("ValidateEmailHeaderValue() error = %v, want error containing %q", err, tt.errSubstr)
				}
			}
		})
	}
}