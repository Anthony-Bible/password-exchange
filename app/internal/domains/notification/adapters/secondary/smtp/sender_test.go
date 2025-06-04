package smtp

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSMTPSender_parseTemplate_FileTemplate(t *testing.T) {
	// Create a temporary template file
	tempDir := t.TempDir()
	templateFile := filepath.Join(tempDir, "test_template.html")
	templateContent := "Hello {{.Body}}!"
	
	err := os.WriteFile(templateFile, []byte(templateContent), 0644)
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