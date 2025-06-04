package validation

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
)

var (
	ErrInvalidEmail = errors.New("invalid email address format")
	ErrEmptyEmail   = errors.New("email address cannot be empty")
)

// Email validation constants
const (
	// Maximum email length according to RFC 5321
	MaxEmailLength = 320
	// Email pattern for additional validation
	emailDomainPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
)

var emailRegex = regexp.MustCompile(emailDomainPattern)

// ValidateEmail performs comprehensive email validation
func ValidateEmail(email string) error {
	// Check for empty email
	if strings.TrimSpace(email) == "" {
		return ErrEmptyEmail
	}

	// Check email length
	if len(email) > MaxEmailLength {
		return fmt.Errorf("%w: email too long (%d chars, max %d)", ErrInvalidEmail, len(email), MaxEmailLength)
	}

	// Normalize email (trim whitespace and convert to lowercase)
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	// Use Go's built-in email validation (RFC 5322)
	_, err := mail.ParseAddress(normalizedEmail)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEmail, err)
	}

	// Additional regex validation for stricter format checking
	if !emailRegex.MatchString(normalizedEmail) {
		return fmt.Errorf("%w: does not match expected pattern", ErrInvalidEmail)
	}

	// Check for common suspicious patterns
	if err := checkSuspiciousEmailPatterns(normalizedEmail); err != nil {
		return err
	}

	return nil
}

// checkSuspiciousEmailPatterns checks for potentially problematic email patterns
func checkSuspiciousEmailPatterns(email string) error {
	// Check for multiple consecutive dots
	if strings.Contains(email, "..") {
		return fmt.Errorf("%w: contains consecutive dots", ErrInvalidEmail)
	}

	// Check for starting or ending with dot
	localPart := strings.Split(email, "@")[0]
	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return fmt.Errorf("%w: local part cannot start or end with dot", ErrInvalidEmail)
	}

	// Check for valid domain part
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return fmt.Errorf("%w: invalid email structure", ErrInvalidEmail)
	}

	domain := parts[1]
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return fmt.Errorf("%w: domain cannot start or end with dot", ErrInvalidEmail)
	}

	return nil
}

// SanitizeEmailForLogging sanitizes email addresses for safe logging
func SanitizeEmailForLogging(email string) string {
	if email == "" {
		return "[EMPTY_EMAIL]"
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "[INVALID_EMAIL_FORMAT]"
	}

	localPart := parts[0]
	domain := parts[1]

	// Mask the local part but keep first char and length info
	if len(localPart) <= 1 {
		return fmt.Sprintf("%s***@%s", localPart, domain)
	} else if len(localPart) <= 3 {
		return fmt.Sprintf("%s**@%s", string(localPart[0]), domain)
	} else {
		return fmt.Sprintf("%s***%s@%s", string(localPart[0]), string(localPart[len(localPart)-1]), domain)
	}
}

// NormalizeEmail normalizes an email address (lowercase, trim whitespace)
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// SanitizeEmailHeaderValue sanitizes email header values to prevent CRLF injection
// Removes carriage return (\r) and line feed (\n) characters to prevent header injection attacks
func SanitizeEmailHeaderValue(value string) string {
	// Remove all CR and LF characters to prevent CRLF injection
	sanitized := strings.ReplaceAll(value, "\r", "")
	sanitized = strings.ReplaceAll(sanitized, "\n", "")
	
	// Also remove other potentially dangerous control characters
	sanitized = strings.ReplaceAll(sanitized, "\t", " ") // Replace tabs with spaces
	sanitized = strings.ReplaceAll(sanitized, "\v", "")  // Remove vertical tabs
	sanitized = strings.ReplaceAll(sanitized, "\f", "")  // Remove form feeds
	
	return sanitized
}

// ValidateEmailHeaderValue validates that an email header value is safe from injection
func ValidateEmailHeaderValue(value string) error {
	// Check for CRLF injection attempts
	if strings.Contains(value, "\r") || strings.Contains(value, "\n") {
		return fmt.Errorf("email header injection detected: CRLF characters found")
	}
	
	// Check for other control characters that could be dangerous
	controlChars := []rune{'\t', '\v', '\f', '\b'}
	for _, char := range controlChars {
		if strings.ContainsRune(value, char) {
			return fmt.Errorf("email header injection detected: control character found")
		}
	}
	
	// Check for suspicious header-like patterns that could indicate injection
	// Look for patterns that start with typical email headers followed by colon and space
	// This is more restrictive to avoid false positives with legitimate content like "Company: Product"
	suspiciousHeaders := []string{
		"bcc:", "cc:", "to:", "from:", "reply-to:", "return-path:", 
		"x-", "content-", "mime-", "message-id:", "date:", "received:",
	}
	
	lowerValue := strings.ToLower(value)
	for _, header := range suspiciousHeaders {
		if strings.HasPrefix(lowerValue, header) {
			return fmt.Errorf("email header injection detected: header-like pattern found")
		}
	}
	
	return nil
}