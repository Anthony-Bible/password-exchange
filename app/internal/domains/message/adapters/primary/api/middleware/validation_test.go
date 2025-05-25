package middleware

import (
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/stretchr/testify/assert"
)

func TestValidateMessageSubmission(t *testing.T) {
	tests := []struct {
		name           string
		request        *models.MessageSubmissionRequest
		expectErrors   bool
		expectedFields []string
	}{
		{
			name: "valid basic message without notification",
			request: &models.MessageSubmissionRequest{
				Content:          "Test message",
				SendNotification: false,
			},
			expectErrors: false,
		},
		{
			name: "valid message with notification",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			},
			expectErrors: false,
		},
		{
			name: "missing content",
			request: &models.MessageSubmissionRequest{
				Content:          "",
				SendNotification: false,
			},
			expectErrors:   true,
			expectedFields: []string{"content"},
		},
		{
			name: "content too long",
			request: &models.MessageSubmissionRequest{
				Content:          string(make([]byte, 10001)), // Exceeds max length
				SendNotification: false,
			},
			expectErrors:   true,
			expectedFields: []string{"content"},
		},
		{
			name: "passphrase too long",
			request: &models.MessageSubmissionRequest{
				Content:          "Test message",
				Passphrase:       string(make([]byte, 501)), // Exceeds max length
				SendNotification: false,
			},
			expectErrors:   true,
			expectedFields: []string{"passphrase"},
		},
		{
			name: "notification enabled but missing sender",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			},
			expectErrors:   true,
			expectedFields: []string{"sender"},
		},
		{
			name: "notification enabled but missing recipient",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			},
			expectErrors:   true,
			expectedFields: []string{"recipient"},
		},
		{
			name: "notification enabled but sender missing name",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			},
			expectErrors:   true,
			expectedFields: []string{"sender.name"},
		},
		{
			name: "notification enabled but sender missing email",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			},
			expectErrors:   true,
			expectedFields: []string{"sender.email"},
		},
		{
			name: "notification enabled but invalid sender email",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "invalid-email",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			},
			expectErrors:   true,
			expectedFields: []string{"sender.email"},
		},
		{
			name: "notification enabled but recipient missing name",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			},
			expectErrors:   true,
			expectedFields: []string{"recipient.name"},
		},
		{
			name: "notification enabled but recipient missing email",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "",
				},
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			},
			expectErrors:   true,
			expectedFields: []string{"recipient.email"},
		},
		{
			name: "notification enabled but invalid recipient email",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "invalid-email",
				},
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			},
			expectErrors:   true,
			expectedFields: []string{"recipient.email"},
		},
		{
			name: "notification enabled but missing anti-spam answer",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "",
			},
			expectErrors:   true,
			expectedFields: []string{"antiSpamAnswer"},
		},
		{
			name: "notification enabled but wrong anti-spam answer",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "red",
			},
			expectErrors:   true,
			expectedFields: []string{"antiSpamAnswer"},
		},
		{
			name: "notification enabled with case-insensitive anti-spam answer",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "BLUE",
			},
			expectErrors: false,
		},
		{
			name: "notification enabled with whitespace in anti-spam answer",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "  blue  ",
			},
			expectErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateMessageSubmission(tt.request)

			if tt.expectErrors {
				assert.NotNil(t, errors, "Expected validation errors but got none")
				if errors != nil {
					for _, field := range tt.expectedFields {
						assert.Contains(t, errors, field, "Expected error for field %s", field)
					}
				}
			} else {
				assert.Nil(t, errors, "Expected no validation errors but got: %v", errors)
			}
		})
	}
}

func TestValidateStruct(t *testing.T) {
	tests := []struct {
		name           string
		input          interface{}
		expectErrors   bool
		expectedFields []string
	}{
		{
			name: "valid sender",
			input: &models.Sender{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			expectErrors: false,
		},
		{
			name: "sender missing name",
			input: &models.Sender{
				Name:  "",
				Email: "john@example.com",
			},
			expectErrors:   true,
			expectedFields: []string{"name"},
		},
		{
			name: "sender missing email",
			input: &models.Sender{
				Name:  "John Doe",
				Email: "",
			},
			expectErrors:   true,
			expectedFields: []string{"email"},
		},
		{
			name: "sender invalid email",
			input: &models.Sender{
				Name:  "John Doe",
				Email: "invalid-email",
			},
			expectErrors:   true,
			expectedFields: []string{"email"},
		},
		{
			name: "sender name too long",
			input: &models.Sender{
				Name:  string(make([]byte, 101)), // Exceeds max length
				Email: "john@example.com",
			},
			expectErrors:   true,
			expectedFields: []string{"name"},
		},
		{
			name: "valid recipient",
			input: &models.Recipient{
				Name:  "Jane Smith",
				Email: "jane@example.com",
			},
			expectErrors: false,
		},
		{
			name: "recipient missing name",
			input: &models.Recipient{
				Name:  "",
				Email: "jane@example.com",
			},
			expectErrors:   true,
			expectedFields: []string{"name"},
		},
		{
			name: "recipient missing email",
			input: &models.Recipient{
				Name:  "Jane Smith",
				Email: "",
			},
			expectErrors:   true,
			expectedFields: []string{"email"},
		},
		{
			name: "recipient invalid email",
			input: &models.Recipient{
				Name:  "Jane Smith",
				Email: "invalid-email",
			},
			expectErrors:   true,
			expectedFields: []string{"email"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateStruct(tt.input)

			if tt.expectErrors {
				assert.NotNil(t, errors, "Expected validation errors but got none")
				if errors != nil {
					for _, field := range tt.expectedFields {
						assert.Contains(t, errors, field, "Expected error for field %s", field)
					}
				}
			} else {
				assert.Nil(t, errors, "Expected no validation errors but got: %v", errors)
			}
		})
	}
}

func TestAntiSpamValidation(t *testing.T) {
	tests := []struct {
		name     string
		answer   string
		expected bool
	}{
		{"correct answer lowercase", "blue", true},
		{"correct answer uppercase", "BLUE", true},
		{"correct answer mixed case", "Blue", true},
		{"correct answer with whitespace", "  blue  ", true},
		{"wrong answer", "red", false},
		{"wrong answer", "green", false},
		{"empty answer", "", false},
		{"partial answer", "blu", false},
		{"answer with extra text", "blue color", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic directly through a message submission request
			req := &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   tt.answer,
			}

			errors := ValidateMessageSubmission(req)

			if tt.expected {
				// Should not have anti-spam error
				if errors != nil {
					assert.NotContains(t, errors, "antiSpamAnswer", "Should not have anti-spam error for valid answer: %s", tt.answer)
				}
			} else {
				// Should have anti-spam error
				assert.NotNil(t, errors, "Expected validation errors for invalid anti-spam answer: %s", tt.answer)
				if errors != nil {
					assert.Contains(t, errors, "antiSpamAnswer", "Expected anti-spam error for invalid answer: %s", tt.answer)
				}
			}
		})
	}
}