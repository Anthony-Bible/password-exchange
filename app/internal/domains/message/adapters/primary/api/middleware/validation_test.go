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
				QuestionID:       intPtr(0),
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
				QuestionID:       intPtr(0),
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
				QuestionID:       intPtr(0),
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
		name       string
		questionId *int
		answer     string
		expected   bool
	}{
		// Question 0: What color is the sky?
		{"question 0 correct answer lowercase", intPtr(0), "blue", true},
		{"question 0 correct answer uppercase", intPtr(0), "BLUE", true},
		{"question 0 correct answer mixed case", intPtr(0), "Blue", true},
		{"question 0 correct answer with whitespace", intPtr(0), "  blue  ", true},
		{"question 0 wrong answer", intPtr(0), "red", false},
		{"question 0 wrong answer green", intPtr(0), "green", false},

		// Question 1: What is 2 + 2?
		{"question 1 correct numeric", intPtr(1), "4", true},
		{"question 1 correct word", intPtr(1), "four", true},
		{"question 1 wrong answer", intPtr(1), "5", false},

		// Question 2: How many days are in a week?
		{"question 2 correct numeric", intPtr(2), "7", true},
		{"question 2 correct word", intPtr(2), "seven", true},
		{"question 2 wrong answer", intPtr(2), "6", false},

		// Question 3: What animal says meow?
		{"question 3 correct singular", intPtr(3), "cat", true},
		{"question 3 correct plural", intPtr(3), "cats", true},
		{"question 3 wrong answer", intPtr(3), "dog", false},

		// Question 4: What do you use to write?
		{"question 4 correct", intPtr(4), "pen", true},
		{"question 4 wrong answer", intPtr(4), "car", false},

		// Question 5: How many legs does a dog have?
		{"question 5 correct numeric", intPtr(5), "4", true},
		{"question 5 correct word", intPtr(5), "four", true},
		{"question 5 wrong answer", intPtr(5), "3", false},

		// Edge cases
		{"empty answer", intPtr(0), "", false},
		{"partial answer", intPtr(0), "blu", false},
		{"answer with extra text", intPtr(0), "blue color", false},
		{"invalid question id", intPtr(99), "blue", false},
		{"nil question id defaults to 0", nil, "blue", true},
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
				QuestionID:       tt.questionId,
			}

			errors := ValidateMessageSubmission(req)

			if tt.expected {
				// Should not have anti-spam error
				if errors != nil {
					assert.NotContains(
						t,
						errors,
						"antiSpamAnswer",
						"Should not have anti-spam error for valid answer: %s",
						tt.answer,
					)
				}
			} else {
				// Should have anti-spam error
				assert.NotNil(t, errors, "Expected validation errors for invalid anti-spam answer: %s", tt.answer)
				if errors != nil {
					assert.Contains(
						t,
						errors,
						"antiSpamAnswer",
						"Expected anti-spam error for invalid answer: %s",
						tt.answer,
					)
				}
			}
		})
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
