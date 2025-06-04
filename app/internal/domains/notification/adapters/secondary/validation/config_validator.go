package validation

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
)

// ConfigValidator provides shared validation logic for ConfigPort implementations
type ConfigValidator struct{}

// NewConfigValidator creates a new config validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// ValidatePasswordExchangeURL validates a password exchange URL.
// For URLs that are already complete (with protocol), validates the full URL.
// For host-only strings, validates the domain format.
func (cv *ConfigValidator) ValidatePasswordExchangeURL(urlString string) error {
	if urlString == "" {
		return nil // Empty is allowed (will use default)
	}

	// Check if it's a complete URL or just a host
	if strings.HasPrefix(urlString, "http://") || strings.HasPrefix(urlString, "https://") {
		// Complete URL - validate format and require HTTPS
		parsedURL, err := url.Parse(urlString)
		if err != nil {
			return fmt.Errorf("invalid URL format: %w", err)
		}

		if parsedURL.Scheme != "https" {
			return errors.New("URL must use HTTPS protocol")
		}

		if parsedURL.Host == "" {
			return errors.New("URL must have a valid hostname")
		}

		if strings.Contains(parsedURL.Host, "..") {
			return errors.New("invalid domain format")
		}
	} else {
		// Host-only string - validate as domain
		if strings.Contains(urlString, "/") {
			return errors.New("URL should not contain path")
		}

		if strings.Contains(urlString, ":") {
			return errors.New("URL should not contain port")
		}

		// Validate domain format using URL parsing
		testURL := "https://" + urlString
		parsedURL, err := url.Parse(testURL)
		if err != nil {
			return fmt.Errorf("invalid domain format: %w", err)
		}

		if strings.Contains(parsedURL.Host, "..") {
			return errors.New("invalid domain format")
		}
	}

	return nil
}

// ValidateServerEmail validates a server email address.
func (cv *ConfigValidator) ValidateServerEmail(email string) error {
	if email == "" {
		return nil // Empty is allowed (will use default)
	}

	// Use existing email validation utility
	return validation.ValidateEmail(email)
}

// ValidateTemplateFormats validates template strings for proper formatting.
func (cv *ConfigValidator) ValidateTemplateFormats(initialSubject, reminderSubject, bodyTemplate string) error {
	// Validate initial notification subject - should have one %s placeholder
	if err := cv.validateStringTemplate(initialSubject, 1, "s"); err != nil {
		return fmt.Errorf("initial notification subject: %w", err)
	}

	// Validate reminder notification subject - should have one %d placeholder  
	if err := cv.validateStringTemplate(reminderSubject, 1, "d"); err != nil {
		return fmt.Errorf("reminder notification subject: %w", err)
	}

	// Validate initial notification body template - should have multiple %s placeholders
	if err := cv.validateStringTemplate(bodyTemplate, 4, "s"); err != nil {
		return fmt.Errorf("initial notification body template: %w", err)
	}

	return nil
}

// validateStringTemplate validates that a template string has the expected number and type of placeholders
func (cv *ConfigValidator) validateStringTemplate(template string, expectedCount int, expectedType string) error {
	// Create pattern to match format specifiers
	pattern := fmt.Sprintf(`%%[#\-\+ 0]*[*]?[*]?%s`, expectedType)
	re := regexp.MustCompile(pattern)
	
	matches := re.FindAllString(template, -1)
	actualCount := len(matches)
	
	if actualCount != expectedCount {
		return fmt.Errorf("template placeholder mismatch: expected %d %%%s placeholders, got %d", 
			expectedCount, expectedType, actualCount)
	}
	
	return nil
}