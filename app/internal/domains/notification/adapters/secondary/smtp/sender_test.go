package smtp

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/secondary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSMTPSender_parseTemplate_FileTemplate(t *testing.T) {
	// Create a temporary template file
	tempDir := t.TempDir()
	templateFile := filepath.Join(tempDir, "test_template.html")
	templateContent := "Hello {{.Body}}!"

	err := os.WriteFile(templateFile, []byte(templateContent), 0o644)
	require.NoError(t, err)

	// Create minimal SMTP sender for testing parseTemplate
	sender := &SMTPSender{}

	// Test file template parsing
	tmpl, err := sender.parseTemplate(templateFile)
	require.NoError(t, err)
	assert.NotNil(t, tmpl)

	// Verify template can be executed
	var buf bytes.Buffer
	data := struct{ Body string }{Body: "World"}
	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)
	assert.Equal(t, "Hello World!", buf.String())
}

func TestSMTPSender_parseTemplate_InlineTemplate(t *testing.T) {
	// Create minimal SMTP sender for testing parseTemplate
	sender := &SMTPSender{}

	// Test inline template parsing
	inlineTemplate := "Hello {{.Body}}!"
	tmpl, err := sender.parseTemplate(inlineTemplate)
	require.NoError(t, err)
	assert.NotNil(t, tmpl)

	// Verify template can be executed
	var buf bytes.Buffer
	data := struct{ Body string }{Body: "World"}
	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)
	assert.Equal(t, "Hello World!", buf.String())
}

func TestSMTPSender_parseTemplate_NonExistentFile(t *testing.T) {
	// Create minimal SMTP sender for testing parseTemplate
	sender := &SMTPSender{}

	// Test parsing a non-existent file path as inline template
	nonExistentFile := "/path/that/does/not/exist/template.html"
	tmpl, err := sender.parseTemplate(nonExistentFile)

	// Should treat as inline template and successfully parse since it's valid template syntax
	assert.NoError(t, err)
	assert.NotNil(t, tmpl)
}

func TestSMTPSender_parseTemplate_InvalidInlineTemplate(t *testing.T) {
	// Create minimal SMTP sender for testing parseTemplate
	sender := &SMTPSender{}

	// Test invalid inline template syntax
	invalidTemplate := "Hello {{.Body"
	tmpl, err := sender.parseTemplate(invalidTemplate)

	assert.Error(t, err)
	assert.Nil(t, tmpl)
}

func TestSMTPSender_parseTemplate_PathDetection(t *testing.T) {
	// Create minimal SMTP sender for testing parseTemplate
	sender := &SMTPSender{}

	tests := []struct {
		name           string
		input          string
		expectFilePath bool
	}{
		{
			name:           "absolute path",
			input:          "/templates/email.html",
			expectFilePath: true,
		},
		{
			name:           "relative path with slash",
			input:          "templates/email.html",
			expectFilePath: true,
		},
		{
			name:           "inline template",
			input:          "Hello {{.Name}}!",
			expectFilePath: false,
		},
		{
			name:           "inline template with HTML",
			input:          "<html><body>Hello {{.Name}}!</body></html>",
			expectFilePath: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the behavior: file paths that don't exist will be treated as inline templates
			tmpl, err := sender.parseTemplate(tt.input)

			if tt.expectFilePath {
				// For file paths that don't exist, they're treated as inline templates
				// The non-existent path string should parse as a template successfully
				assert.NoError(t, err)
				assert.NotNil(t, tmpl)
			} else {
				// For inline templates, they should parse successfully if valid
				if tt.input == "Hello {{.Name}}!" || tt.input == "<html><body>Hello {{.Name}}!</body></html>" {
					assert.NoError(t, err)
					assert.NotNil(t, tmpl)
				}
			}
		})
	}
}

func TestSMTPSender_getSafeTemplateFunctions(t *testing.T) {
	sender := &SMTPSender{}
	funcs := sender.getSafeTemplateFunctions()

	// Test that all expected safe functions are present
	expectedFunctions := []string{
		"upper", "lower", "title", "trim", "replace",
		"html", "js", "url", "printf",
	}

	for _, funcName := range expectedFunctions {
		assert.Contains(t, funcs, funcName, "Safe function %s should be available", funcName)
		assert.NotNil(t, funcs[funcName], "Function %s should not be nil", funcName)
	}

	// Verify no dangerous functions are present
	dangerousFunctions := []string{
		"exec", "call", "env", "system", "open", "read", "write",
	}

	for _, funcName := range dangerousFunctions {
		assert.NotContains(t, funcs, funcName, "Dangerous function %s should not be available", funcName)
	}
}

func TestSMTPSender_SafeTemplateFunctions_StringManipulation(t *testing.T) {
	sender := &SMTPSender{}

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "upper function",
			template: "{{.Text | upper}}",
			data:     map[string]interface{}{"Text": "hello world"},
			expected: "HELLO WORLD",
		},
		{
			name:     "lower function",
			template: "{{.Text | lower}}",
			data:     map[string]interface{}{"Text": "HELLO WORLD"},
			expected: "hello world",
		},
		{
			name:     "title function",
			template: "{{.Text | title}}",
			data:     map[string]interface{}{"Text": "hello world"},
			expected: "Hello World",
		},
		{
			name:     "trim function",
			template: "{{.Text | trim}}",
			data:     map[string]interface{}{"Text": "  hello world  "},
			expected: "hello world",
		},
		{
			name:     "replace function",
			template: "{{replace .Text \"world\" \"universe\"}}",
			data:     map[string]interface{}{"Text": "hello world"},
			expected: "hello universe",
		},
		{
			name:     "printf function",
			template: "{{printf \"Hello %s, you have %d messages\" .Name .Count}}",
			data:     map[string]interface{}{"Name": "Alice", "Count": 5},
			expected: "Hello Alice, you have 5 messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := sender.parseTemplate(tt.template)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestSMTPSender_SafeTemplateFunctions_HTMLEscaping(t *testing.T) {
	sender := &SMTPSender{}

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "html escaping",
			template: "{{.Content | html}}",
			data:     map[string]interface{}{"Content": "<script>alert('xss')</script>"},
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "js escaping",
			template: "{{.Content | js}}",
			data:     map[string]interface{}{"Content": "alert('test')"},
			expected: "alert(\\'test\\')",
		},
		{
			name:     "url escaping",
			template: "{{.URL | url}}",
			data:     map[string]interface{}{"URL": "hello world & special chars"},
			expected: "hello+world+%26+special+chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := sender.parseTemplate(tt.template)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestSMTPSender_SafeTemplateFunctions_PreventsDangerousOperations(t *testing.T) {
	sender := &SMTPSender{}

	// Test templates that would be dangerous if unsafe functions were available
	dangerousTemplates := []struct {
		name     string
		template string
		reason   string
	}{
		{
			name:     "exec function not available",
			template: "{{exec \"ls -la\"}}",
			reason:   "exec function should not be available",
		},
		{
			name:     "undefined function",
			template: "{{dangerousFunction .Data}}",
			reason:   "undefined functions should cause parse errors",
		},
		{
			name:     "system function not available",
			template: "{{system \"rm -rf /\"}}",
			reason:   "system function should not be available",
		},
	}

	for _, tt := range dangerousTemplates {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sender.parseTemplate(tt.template)
			assert.Error(t, err, tt.reason)
			assert.Contains(t, strings.ToLower(err.Error()), "function",
				"Error should mention function issue for: %s", tt.reason)
		})
	}
}

func TestSMTPSender_SafeTemplateFunctions_FileAndInlineConsistency(t *testing.T) {
	sender := &SMTPSender{}

	// Template content that uses safe functions
	templateContent := `Hello {{.Name | upper}}! 
Your message: {{.Message | html}}
URL: {{.URL | url}}`

	data := map[string]interface{}{
		"Name":    "alice",
		"Message": "<script>alert('test')</script>",
		"URL":     "hello world",
	}

	// Test inline template
	inlineTmpl, err := sender.parseTemplate(templateContent)
	require.NoError(t, err)

	var inlineBuf bytes.Buffer
	err = inlineTmpl.Execute(&inlineBuf, data)
	require.NoError(t, err)

	// Test file template
	tempDir := t.TempDir()
	templateFile := filepath.Join(tempDir, "test_template.html")
	err = os.WriteFile(templateFile, []byte(templateContent), 0o644)
	require.NoError(t, err)

	fileTmpl, err := sender.parseTemplate(templateFile)
	require.NoError(t, err)

	var fileBuf bytes.Buffer
	err = fileTmpl.Execute(&fileBuf, data)
	require.NoError(t, err)

	// Both should produce identical output
	assert.Equal(t, inlineBuf.String(), fileBuf.String(),
		"File and inline templates should produce identical output")

	// Verify the output contains expected safe transformations
	output := inlineBuf.String()
	assert.Contains(t, output, "ALICE", "Name should be uppercased")
	assert.Contains(t, output, "&lt;script&gt;", "HTML should be escaped")
	assert.Contains(t, output, "hello+world", "URL should be escaped")
}

func TestSMTPSender_SafeTemplateFunctions_ChainedOperations(t *testing.T) {
	sender := &SMTPSender{}

	// Test chaining multiple safe functions
	template := `{{.Text | trim | lower | title}}`
	data := map[string]interface{}{"Text": "  HELLO WORLD  "}

	tmpl, err := sender.parseTemplate(template)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	// Should trim, then lowercase, then title case
	assert.Equal(t, "Hello World", buf.String())
}

func TestSMTPSender_validateTemplateContent_ValidTemplates(t *testing.T) {
	sender := &SMTPSender{}

	validTemplates := []struct {
		name     string
		template string
		reason   string
	}{
		{
			name:     "simple template",
			template: "Hello {{.Name}}!",
			reason:   "Simple templates should be valid",
		},
		{
			name:     "template with safe functions",
			template: "Hello {{.Name | upper}}! Your email: {{.Email | html}}",
			reason:   "Templates using safe functions should be valid",
		},
		{
			name:     "HTML template",
			template: "<html><body>Hello {{.Name}}</body></html>",
			reason:   "HTML templates should be valid",
		},
		{
			name:     "template with conditionals",
			template: "{{if .ShowMessage}}Message: {{.Message}}{{end}}",
			reason:   "Templates with conditionals should be valid",
		},
		{
			name:     "template with range",
			template: "{{range .Items}}Item: {{.}}{{end}}",
			reason:   "Templates with range should be valid",
		},
		{
			name:     "template with multiple variables",
			template: "From: {{.From}}, To: {{.To}}, Subject: {{.Subject}}",
			reason:   "Templates with multiple variables should be valid",
		},
		{
			name:     "empty template",
			template: "",
			reason:   "Empty templates should be valid",
		},
		{
			name:     "plain text template",
			template: "This is just plain text without any template syntax",
			reason:   "Plain text should be valid",
		},
	}

	for _, tt := range validTemplates {
		t.Run(tt.name, func(t *testing.T) {
			err := sender.validateTemplateContent(tt.template)
			assert.NoError(t, err, tt.reason)
		})
	}
}

func TestSMTPSender_validateTemplateContent_InvalidTemplates(t *testing.T) {
	sender := &SMTPSender{}

	invalidTemplates := []struct {
		name        string
		template    string
		expectedErr string
		reason      string
	}{
		{
			name:        "template injection with exec",
			template:    "Hello {{exec \"rm -rf /\"}}",
			expectedErr: "dangerous function",
			reason:      "Templates with exec function should be rejected",
		},
		{
			name:        "template injection with system calls",
			template:    "{{system \"cat /etc/passwd\"}}",
			expectedErr: "dangerous function",
			reason:      "Templates with system function should be rejected",
		},
		{
			name:        "template injection with file operations",
			template:    "{{read \"/etc/passwd\"}}",
			expectedErr: "dangerous function",
			reason:      "Templates with file read operations should be rejected",
		},
		{
			name:        "template injection with environment access",
			template:    "{{env \"SECRET_KEY\"}}",
			expectedErr: "dangerous function",
			reason:      "Templates accessing environment variables should be rejected",
		},
		{
			name:        "template with call function",
			template:    "{{call .SomeFunction}}",
			expectedErr: "dangerous function",
			reason:      "Templates with call function should be rejected",
		},
		{
			name:        "template with undefined function",
			template:    "{{customFunction .Data}}",
			expectedErr: "undefined function",
			reason:      "Templates with undefined functions should be rejected",
		},
		{
			name:        "deeply nested dangerous call",
			template:    "{{range .Items}}{{if .Valid}}{{exec .Command}}{{end}}{{end}}",
			expectedErr: "dangerous function",
			reason:      "Nested dangerous functions should be rejected",
		},
		{
			name:        "dangerous function in pipeline",
			template:    "{{.Command | exec}}",
			expectedErr: "dangerous function",
			reason:      "Dangerous functions in pipelines should be rejected",
		},
		{
			name:        "file path traversal attempt",
			template:    "{{.Data}}../../etc/passwd",
			expectedErr: "path traversal",
			reason:      "Templates with path traversal should be rejected",
		},
		{
			name:        "absolute path reference",
			template:    "Content: /etc/passwd {{.Data}}",
			expectedErr: "absolute path",
			reason:      "Templates with absolute paths should be rejected",
		},
		{
			name:        "script tag injection",
			template:    "<script>alert('xss')</script>{{.Data}}",
			expectedErr: "script tag",
			reason:      "Templates with script tags should be rejected",
		},
		{
			name:        "excessive template size",
			template:    strings.Repeat("a", 10241), // > 10KB (10240 bytes)
			expectedErr: "template too large",
			reason:      "Templates exceeding size limit should be rejected",
		},
		{
			name:        "excessive nesting depth",
			template:    strings.Repeat("{{range .Items}}", 100) + "data" + strings.Repeat("{{end}}", 100),
			expectedErr: "nesting too deep",
			reason:      "Templates with excessive nesting should be rejected",
		},
	}

	for _, tt := range invalidTemplates {
		t.Run(tt.name, func(t *testing.T) {
			err := sender.validateTemplateContent(tt.template)
			assert.Error(t, err, tt.reason)
			assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.expectedErr),
				"Error should contain expected message for: %s", tt.reason)
		})
	}
}

func TestSMTPSender_validateTemplateContent_EdgeCases(t *testing.T) {
	sender := &SMTPSender{}

	edgeCases := []struct {
		name      string
		template  string
		shouldErr bool
		reason    string
	}{
		{
			name:      "template with safe function names in content",
			template:  "Function names like exec and system are mentioned in this text {{.Data}}",
			shouldErr: false,
			reason:    "Safe content mentioning function names should be allowed",
		},
		{
			name:      "template with escaped braces",
			template:  "Use \\{\\{exec\\}\\} to show template syntax {{.Data}}",
			shouldErr: false,
			reason:    "Escaped template syntax should be allowed",
		},
		{
			name:      "template with comments",
			template:  "{{/* This is a comment */}}Hello {{.Name}}",
			shouldErr: false,
			reason:    "Templates with comments should be allowed",
		},
		{
			name:      "template with whitespace control",
			template:  "{{- .Data -}}",
			shouldErr: false,
			reason:    "Templates with whitespace control should be allowed",
		},
		{
			name:      "maximum allowed template size",
			template:  strings.Repeat("a", 10240), // exactly 10KB
			shouldErr: false,
			reason:    "Templates at size limit should be allowed",
		},
		{
			name:      "maximum allowed nesting depth",
			template:  strings.Repeat("{{range .Items}}", 10) + "data" + strings.Repeat("{{end}}", 10),
			shouldErr: false,
			reason:    "Templates at nesting limit should be allowed",
		},
	}

	for _, tt := range edgeCases {
		t.Run(tt.name, func(t *testing.T) {
			err := sender.validateTemplateContent(tt.template)
			if tt.shouldErr {
				assert.Error(t, err, tt.reason)
			} else {
				assert.NoError(t, err, tt.reason)
			}
		})
	}
}

// Mock implementations for email header security tests

// mockConfigPortSecure for email header testing
type mockConfigPortSecure struct{}

// Verify interface compliance
var (
	_ secondary.ConfigPort     = &mockConfigPortSecure{}
	_ secondary.LoggerPort     = &mockLoggerPortSecure{}
	_ secondary.ValidationPort = &mockValidationPortSecure{}
)

func (m *mockConfigPortSecure) GetServerEmail() string                 { return "test@example.com" }
func (m *mockConfigPortSecure) GetServerName() string                  { return "Test Server" }
func (m *mockConfigPortSecure) GetPasswordExchangeURL() string         { return "https://test.example.com" }
func (m *mockConfigPortSecure) GetInitialNotificationSubject() string  { return "Test Subject %s" }
func (m *mockConfigPortSecure) GetReminderNotificationSubject() string { return "Reminder Subject %d" }
func (m *mockConfigPortSecure) GetEmailTemplate() string               { return "Test template: {{.Body}}" }
func (m *mockConfigPortSecure) GetInitialNotificationBodyTemplate() string {
	return "Initial body template"
}

func (m *mockConfigPortSecure) GetReminderNotificationBodyTemplate() string {
	return "Reminder body template"
}
func (m *mockConfigPortSecure) GetReminderEmailTemplate() string   { return "Reminder email template" }
func (m *mockConfigPortSecure) GetReminderMessageContent() string  { return "Reminder message content" }
func (m *mockConfigPortSecure) ValidatePasswordExchangeURL() error { return nil }
func (m *mockConfigPortSecure) ValidateServerEmail() error         { return nil }
func (m *mockConfigPortSecure) ValidateTemplateFormats() error     { return nil }

// mockLoggerPortSecure for email header testing
type mockLoggerPortSecure struct{}

func (m *mockLoggerPortSecure) Debug() contracts.LogEvent { return &mockLogEventSecure{} }
func (m *mockLoggerPortSecure) Info() contracts.LogEvent  { return &mockLogEventSecure{} }
func (m *mockLoggerPortSecure) Warn() contracts.LogEvent  { return &mockLogEventSecure{} }
func (m *mockLoggerPortSecure) Error() contracts.LogEvent { return &mockLogEventSecure{} }

type mockLogEventSecure struct{}

func (m *mockLogEventSecure) Str(key, val string) contracts.LogEvent               { return m }
func (m *mockLogEventSecure) Err(err error) contracts.LogEvent                     { return m }
func (m *mockLogEventSecure) Int(key string, val int) contracts.LogEvent           { return m }
func (m *mockLogEventSecure) Bool(key string, val bool) contracts.LogEvent         { return m }
func (m *mockLogEventSecure) Dur(key string, val time.Duration) contracts.LogEvent { return m }
func (m *mockLogEventSecure) Float64(key string, val float64) contracts.LogEvent   { return m }
func (m *mockLogEventSecure) Msg(msg string)                                       {}

// mockValidationPortSecure for email header testing
type mockValidationPortSecure struct{}

func (m *mockValidationPortSecure) ValidateEmail(email string) error {
	return nil // Mock always returns valid
}

func (m *mockValidationPortSecure) SanitizeEmailForLogging(email string) string {
	return "sanitized@example.com"
}

func TestSMTPSender_buildSafeEmailHeaders(t *testing.T) {
	sender := &SMTPSender{
		config:     &mockConfigPortSecure{},
		logger:     &mockLoggerPortSecure{},
		validation: &mockValidationPortSecure{},
	}

	tests := []struct {
		name      string
		fromName  string
		fromEmail string
		to        string
		subject   string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "clean headers",
			fromName:  "Test Sender",
			fromEmail: "sender@example.com",
			to:        "recipient@example.com",
			subject:   "Test Subject",
			wantErr:   false,
		},
		{
			name:      "CRLF injection in from name",
			fromName:  "Test\r\nBcc: attacker@evil.com",
			fromEmail: "sender@example.com",
			to:        "recipient@example.com",
			subject:   "Test Subject",
			wantErr:   true,
			errSubstr: "invalid from name",
		},
		{
			name:      "CRLF injection in from email",
			fromName:  "Test Sender",
			fromEmail: "sender@example.com\r\nFrom: fake@evil.com",
			to:        "recipient@example.com",
			subject:   "Test Subject",
			wantErr:   true,
			errSubstr: "invalid from email",
		},
		{
			name:      "CRLF injection in to email",
			fromName:  "Test Sender",
			fromEmail: "sender@example.com",
			to:        "recipient@example.com\r\nBcc: attacker@evil.com",
			subject:   "Test Subject",
			wantErr:   true,
			errSubstr: "invalid to email",
		},
		{
			name:      "CRLF injection in subject",
			fromName:  "Test Sender",
			fromEmail: "sender@example.com",
			to:        "recipient@example.com",
			subject:   "Test Subject\r\nBcc: attacker@evil.com",
			wantErr:   true,
			errSubstr: "invalid subject",
		},
		{
			name:      "header injection attempt in subject",
			fromName:  "Test Sender",
			fromEmail: "sender@example.com",
			to:        "recipient@example.com",
			subject:   "Bcc: attacker@evil.com",
			wantErr:   true,
			errSubstr: "invalid subject",
		},
		{
			name:      "email spoofing attempt",
			fromName:  "Admin",
			fromEmail: "admin@bank.com\r\nFrom: real-admin@bank.com",
			to:        "victim@target.com",
			subject:   "Security Alert",
			wantErr:   true,
			errSubstr: "invalid from email",
		},
		{
			name:      "subject line with normal colon (legitimate)",
			fromName:  "Password Exchange",
			fromEmail: "noreply@password.exchange",
			to:        "user@example.com",
			subject:   "Password Exchange: New encrypted message",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers, err := sender.buildSafeEmailHeaders(tt.fromName, tt.fromEmail, tt.to, tt.subject)

			if (err != nil) != tt.wantErr {
				t.Errorf("buildSafeEmailHeaders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && tt.errSubstr != "" {
					if !strings.Contains(err.Error(), tt.errSubstr) {
						t.Errorf("buildSafeEmailHeaders() error = %v, want error containing %q", err, tt.errSubstr)
					}
				}
			} else {
				// For successful cases, verify headers are properly constructed
				if headers == "" {
					t.Error("buildSafeEmailHeaders() returned empty headers for valid input")
				}

				// Verify the headers contain expected content
				expectedPatterns := []string{
					"From:",
					"To:",
					"Subject:",
					"MIME-version: 1.0",
					"Content-Type: text/html",
				}

				for _, pattern := range expectedPatterns {
					if !strings.Contains(headers, pattern) {
						t.Errorf("buildSafeEmailHeaders() headers missing expected pattern: %q", pattern)
					}
				}
			}
		})
	}
}
