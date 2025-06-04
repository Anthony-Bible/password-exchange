package shared

import (
	"strings"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/validation"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
)

func TestSharedConfigAdapter_GetEmailTemplate(t *testing.T) {
	cfg := config.PassConfig{}
	adapter := NewSharedConfigAdapter(cfg)

	expected := "/templates/email_template.html"
	result := adapter.GetEmailTemplate()

	if result != expected {
		t.Errorf("GetEmailTemplate() = %q, want %q", result, expected)
	}
}

func TestSharedConfigAdapter_GetServerEmail(t *testing.T) {
	tests := []struct {
		name      string
		emailFrom string
		expected  string
	}{
		{
			name:      "with configured email",
			emailFrom: "test@example.com",
			expected:  "test@example.com",
		},
		{
			name:      "with empty email",
			emailFrom: "",
			expected:  "server@password.exchange",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.PassConfig{
				EmailFrom: tt.emailFrom,
			}
			adapter := NewSharedConfigAdapter(cfg)

			result := adapter.GetServerEmail()
			if result != tt.expected {
				t.Errorf("GetServerEmail() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSharedConfigAdapter_GetServerName(t *testing.T) {
	cfg := config.PassConfig{}
	adapter := NewSharedConfigAdapter(cfg)

	expected := "Password Exchange"
	result := adapter.GetServerName()

	if result != expected {
		t.Errorf("GetServerName() = %q, want %q", result, expected)
	}
}

func TestSharedConfigAdapter_GetPasswordExchangeURL(t *testing.T) {
	tests := []struct {
		name     string
		prodHost string
		expected string
	}{
		{
			name:     "with configured prod host",
			prodHost: "custom.example.com",
			expected: "https://custom.example.com",
		},
		{
			name:     "with empty prod host",
			prodHost: "",
			expected: "https://password.exchange",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.PassConfig{
				ProdHost: tt.prodHost,
			}
			adapter := NewSharedConfigAdapter(cfg)

			result := adapter.GetPasswordExchangeURL()
			if result != tt.expected {
				t.Errorf("GetPasswordExchangeURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSharedConfigAdapter_GetInitialNotificationSubject(t *testing.T) {
	cfg := config.PassConfig{}
	adapter := NewSharedConfigAdapter(cfg)

	expected := "Encrypted Message from Password Exchange from %s"
	result := adapter.GetInitialNotificationSubject()

	if result != expected {
		t.Errorf("GetInitialNotificationSubject() = %q, want %q", result, expected)
	}
}

func TestSharedConfigAdapter_GetReminderNotificationSubject(t *testing.T) {
	cfg := config.PassConfig{}
	adapter := NewSharedConfigAdapter(cfg)

	expected := "Reminder: You have an unviewed encrypted message (Reminder #%d)"
	result := adapter.GetReminderNotificationSubject()

	if result != expected {
		t.Errorf("GetReminderNotificationSubject() = %q, want %q", result, expected)
	}
}

func TestSharedConfigAdapter_GetInitialNotificationBodyTemplate(t *testing.T) {
	cfg := config.PassConfig{}
	adapter := NewSharedConfigAdapter(cfg)

	expected := "Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about"
	result := adapter.GetInitialNotificationBodyTemplate()

	if result != expected {
		t.Errorf("GetInitialNotificationBodyTemplate() = %q, want %q", result, expected)
	}
}

func TestSharedConfigAdapter_GetReminderNotificationBodyTemplate(t *testing.T) {
	cfg := config.PassConfig{}
	adapter := NewSharedConfigAdapter(cfg)

	expected := ""
	result := adapter.GetReminderNotificationBodyTemplate()

	if result != expected {
		t.Errorf("GetReminderNotificationBodyTemplate() = %q, want %q", result, expected)
	}
}

func TestSharedConfigAdapter_GetReminderEmailTemplate(t *testing.T) {
	cfg := config.PassConfig{}
	adapter := NewSharedConfigAdapter(cfg)

	expected := "/templates/reminder_email_template.html"
	result := adapter.GetReminderEmailTemplate()

	if result != expected {
		t.Errorf("GetReminderEmailTemplate() = %q, want %q", result, expected)
	}
}

func TestSharedConfigAdapter_GetReminderMessageContent(t *testing.T) {
	cfg := config.PassConfig{}
	adapter := NewSharedConfigAdapter(cfg)

	expected := "Please check your original email for the secure decrypt link. For security reasons, the decrypt link cannot be included in reminder emails. If you cannot find the original email, please contact the sender to resend the message."
	result := adapter.GetReminderMessageContent()

	if result != expected {
		t.Errorf("GetReminderMessageContent() = %q, want %q", result, expected)
	}
}

func TestSharedConfigAdapter_ImplementsConfigPort(t *testing.T) {
	cfg := config.PassConfig{}
	adapter := NewSharedConfigAdapter(cfg)

	// This test ensures the adapter properly implements the ConfigPort interface
	// by calling all interface methods
	_ = adapter.GetEmailTemplate()
	_ = adapter.GetServerEmail()
	_ = adapter.GetServerName()
	_ = adapter.GetPasswordExchangeURL()
	_ = adapter.GetInitialNotificationSubject()
	_ = adapter.GetReminderNotificationSubject()
	_ = adapter.GetInitialNotificationBodyTemplate()
	_ = adapter.GetReminderNotificationBodyTemplate()
	_ = adapter.GetReminderEmailTemplate()
	_ = adapter.GetReminderMessageContent()
}

// TDD: Tests for configuration validation - these should fail initially
func TestSharedConfigAdapter_ValidatePasswordExchangeURL(t *testing.T) {
	tests := []struct {
		name     string
		prodHost string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid HTTPS URL",
			prodHost: "password.exchange",
			wantErr:  false,
		},
		{
			name:     "valid subdomain",
			prodHost: "dev.password.exchange",
			wantErr:  false,
		},
		{
			name:     "valid complete HTTPS URL",
			prodHost: "https://password.exchange", 
			wantErr:  false, // Now valid since validator handles complete URLs
		},
		{
			name:     "invalid - HTTP protocol not allowed",
			prodHost: "http://password.exchange",
			wantErr:  true,
			errMsg:   "URL must use HTTPS protocol",
		},
		{
			name:     "invalid - contains path",
			prodHost: "password.exchange/path",
			wantErr:  true,
			errMsg:   "URL should not contain path",
		},
		{
			name:     "invalid - empty host",
			prodHost: "",
			wantErr:  false, // Should use default, no validation error
		},
		{
			name:     "invalid - contains port",
			prodHost: "password.exchange:8080",
			wantErr:  true,
			errMsg:   "URL should not contain port",
		},
		{
			name:     "invalid - malformed domain",
			prodHost: "invalid..domain",
			wantErr:  true,
			errMsg:   "invalid domain format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.PassConfig{
				ProdHost: tt.prodHost,
			}
			adapter := NewSharedConfigAdapter(cfg)

			// This method doesn't exist yet - we need to implement it
			err := adapter.ValidatePasswordExchangeURL()
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePasswordExchangeURL() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePasswordExchangeURL() error = %v, want to contain %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePasswordExchangeURL() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestSharedConfigAdapter_ValidateServerEmail(t *testing.T) {
	tests := []struct {
		name      string
		emailFrom string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid email",
			emailFrom: "test@example.com",
			wantErr:   false,
		},
		{
			name:      "valid email with subdomain",
			emailFrom: "noreply@mail.password.exchange",
			wantErr:   false,
		},
		{
			name:      "empty email - uses default",
			emailFrom: "",
			wantErr:   false, // Should use default, no validation error
		},
		{
			name:      "invalid - missing @ symbol",
			emailFrom: "invalidemail.com",
			wantErr:   true,
			errMsg:    "invalid email address format",
		},
		{
			name:      "invalid - missing domain",
			emailFrom: "test@",
			wantErr:   true,
			errMsg:    "invalid email address format",
		},
		{
			name:      "invalid - missing local part",
			emailFrom: "@example.com",
			wantErr:   true,
			errMsg:    "invalid email address format",
		},
		{
			name:      "invalid - too long",
			emailFrom: strings.Repeat("a", 310) + "@example.com", // Creates 322 char email (310 + 12)
			wantErr:   true,
			errMsg:    "email too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.PassConfig{
				EmailFrom: tt.emailFrom,
			}
			adapter := NewSharedConfigAdapter(cfg)

			// This method doesn't exist yet - we need to implement it
			err := adapter.ValidateServerEmail()
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateServerEmail() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateServerEmail() error = %v, want to contain %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateServerEmail() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestSharedConfigAdapter_ValidateTemplateFormats(t *testing.T) {
	cfg := config.PassConfig{}
	adapter := NewSharedConfigAdapter(cfg)

	// Test valid templates (should pass)
	err := adapter.ValidateTemplateFormats()
	if err != nil {
		t.Errorf("ValidateTemplateFormats() with valid templates unexpected error = %v", err)
	}
}

// testConfigAdapter is a test implementation for testing template validation errors
type testConfigAdapter struct {
	initialSubject   string
	reminderSubject  string
	bodyTemplate     string
}

func (t *testConfigAdapter) GetEmailTemplate() string { return "/templates/email_template.html" }
func (t *testConfigAdapter) GetServerEmail() string { return "server@password.exchange" }
func (t *testConfigAdapter) GetServerName() string { return "Password Exchange" }
func (t *testConfigAdapter) GetPasswordExchangeURL() string { return "https://password.exchange" }
func (t *testConfigAdapter) GetInitialNotificationSubject() string { return t.initialSubject }
func (t *testConfigAdapter) GetReminderNotificationSubject() string { return t.reminderSubject }
func (t *testConfigAdapter) GetInitialNotificationBodyTemplate() string { return t.bodyTemplate }
func (t *testConfigAdapter) GetReminderNotificationBodyTemplate() string { return "" }
func (t *testConfigAdapter) GetReminderEmailTemplate() string { return "/templates/reminder_email_template.html" }
func (t *testConfigAdapter) GetReminderMessageContent() string { return "test content" }
func (t *testConfigAdapter) ValidatePasswordExchangeURL() error { return nil }
func (t *testConfigAdapter) ValidateServerEmail() error { return nil }

func (t *testConfigAdapter) ValidateTemplateFormats() error {
	validator := validation.NewConfigValidator()
	return validator.ValidateTemplateFormats(
		t.initialSubject,
		t.reminderSubject,
		t.bodyTemplate,
	)
}

func TestTemplateValidation_InvalidTemplates(t *testing.T) {
	tests := []struct {
		name            string
		initialSubject  string
		reminderSubject string
		bodyTemplate    string
		wantErr         bool
		errMsg          string
	}{
		{
			name:            "invalid initial subject - no placeholders",
			initialSubject:  "No placeholders here",
			reminderSubject: "Reminder: You have an unviewed encrypted message (Reminder #%d)",
			bodyTemplate:    "Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about",
			wantErr:         true,
			errMsg:          "template placeholder mismatch",
		},
		{
			name:            "invalid reminder subject - wrong type",
			initialSubject:  "Encrypted Message from Password Exchange from %s",
			reminderSubject: "Reminder: You have an unviewed encrypted message (Reminder #%s)", // Should be %d
			bodyTemplate:    "Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about",
			wantErr:         true,
			errMsg:          "template placeholder mismatch",
		},
		{
			name:            "invalid body template - missing placeholders",
			initialSubject:  "Encrypted Message from Password Exchange from %s",
			reminderSubject: "Reminder: You have an unviewed encrypted message (Reminder #%d)",
			bodyTemplate:    "Hi %s, \n %s used our service", // Missing 2 placeholders
			wantErr:         true,
			errMsg:          "template placeholder mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &testConfigAdapter{
				initialSubject:  tt.initialSubject,
				reminderSubject: tt.reminderSubject,
				bodyTemplate:    tt.bodyTemplate,
			}

			err := adapter.ValidateTemplateFormats()
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateTemplateFormats() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateTemplateFormats() error = %v, want to contain %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTemplateFormats() unexpected error = %v", err)
				}
			}
		})
	}
}