package smtp

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
)

// TestHTMLMessageNotEscaped verifies that HTML content in the Message field
// is rendered without escaping when using template.HTML type.
func TestHTMLMessageNotEscaped(t *testing.T) {
	// Create a simple template that uses the Message field
	tmpl := template.Must(template.New("test").Parse(`
<div style="margin: 0 0 20px; font-size: 14px;">
    {{.Message}}
</div>
`))

	// Test data with HTML content in the Message field
	data := contracts.NotificationTemplateData{
		Message:       template.HTML(`<a href="https://dev.password.exchange/decrypt/xhtPdzuMlrcmua3X1T0HLnROzJjqolvX7bRHEBiGaSE=/J0JeYH5_cubEPWQXeVWgN78drZ02CuLu6SDH1kgmx2A=">here</a>`),
		SenderName:    "Test Sender",
		RecipientName: "Test Recipient",
		MessageURL:    "https://dev.password.exchange/decrypt/test",
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	// Verify the href is NOT escaped
	if bytes.Contains([]byte(output), []byte("&lt;a href")) {
		t.Error("HTML anchor tag was escaped - this should not happen with template.HTML")
	}

	// Verify the URL is NOT escaped
	if bytes.Contains([]byte(output), []byte("&#34;")) {
		t.Error("Quotes were escaped - this should not happen with template.HTML")
	}

	// Verify the href link is present and correct
	expectedHref := `https://dev.password.exchange/decrypt/xhtPdzuMlrcmua3X1T0HLnROzJjqolvX7bRHEBiGaSE=/J0JeYH5_cubEPWQXeVWgN78drZ02CuLu6SDH1kgmx2A=`
	if !bytes.Contains([]byte(output), []byte(expectedHref)) {
		t.Errorf("Expected href URL not found in output: %s", expectedHref)
	}

	// Verify the link text is present
	if !bytes.Contains([]byte(output), []byte(">here</a>")) {
		t.Error("Expected link text 'here' not found in output")
	}

	t.Logf("✓ HTML Message correctly rendered without escaping:\n%s", output)
}
