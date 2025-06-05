package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEmailNotificationFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                    string
		request                 *models.MessageSubmissionRequest
		mockServiceResponse     *domain.MessageSubmissionResponse
		mockServiceError        error
		expectedStatusCode      int
		expectedNotificationSent bool
		expectServiceCalled     bool
	}{
		{
			name: "successful message submission with email notification",
			request: &models.MessageSubmissionRequest{
				Content: "Test secret message",
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
			mockServiceResponse: &domain.MessageSubmissionResponse{
				MessageID:  "test-message-id",
				DecryptURL: "https://example.com/decrypt/test-message-id/key",
				Success:    true,
			},
			mockServiceError:        nil,
			expectedStatusCode:      http.StatusCreated,
			expectedNotificationSent: true,
			expectServiceCalled:     true,
		},
		{
			name: "successful message submission without email notification",
			request: &models.MessageSubmissionRequest{
				Content:          "Test secret message",
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: false,
			},
			mockServiceResponse: &domain.MessageSubmissionResponse{
				MessageID:  "test-message-id",
				DecryptURL: "https://example.com/decrypt/test-message-id/key",
				Success:    true,
			},
			mockServiceError:        nil,
			expectedStatusCode:      http.StatusCreated,
			expectedNotificationSent: false,
			expectServiceCalled:     true,
		},
		{
			name: "email notification enabled but service fails",
			request: &models.MessageSubmissionRequest{
				Content: "Test secret message",
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
			mockServiceResponse: &domain.MessageSubmissionResponse{
				MessageID:  "test-message-id",
				DecryptURL: "https://example.com/decrypt/test-message-id/key",
				Success:    false, // Email sending failed
			},
			mockServiceError:        nil,
			expectedStatusCode:      http.StatusCreated,
			expectedNotificationSent: false, // Email failed
			expectServiceCalled:     true,
		},
		{
			name: "missing sender information when notification enabled",
			request: &models.MessageSubmissionRequest{
				Content: "Test secret message",
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			},
			mockServiceResponse:     nil,
			mockServiceError:        nil,
			expectedStatusCode:      http.StatusBadRequest,
			expectedNotificationSent: false,
			expectServiceCalled:     false, // Validation should fail before service is called
		},
		{
			name: "missing recipient information when notification enabled",
			request: &models.MessageSubmissionRequest{
				Content: "Test secret message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			},
			mockServiceResponse:     nil,
			mockServiceError:        nil,
			expectedStatusCode:      http.StatusBadRequest,
			expectedNotificationSent: false,
			expectServiceCalled:     false,
		},
		{
			name: "invalid anti-spam answer when notification enabled",
			request: &models.MessageSubmissionRequest{
				Content: "Test secret message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "red", // Wrong answer
			},
			mockServiceResponse:     nil,
			mockServiceError:        nil,
			expectedStatusCode:      http.StatusBadRequest,
			expectedNotificationSent: false,
			expectServiceCalled:     false,
		},
		{
			name: "missing anti-spam answer when notification enabled",
			request: &models.MessageSubmissionRequest{
				Content: "Test secret message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				SendNotification: true,
				AntiSpamAnswer:   "", // Missing answer
			},
			mockServiceResponse:     nil,
			mockServiceError:        nil,
			expectedStatusCode:      http.StatusBadRequest,
			expectedNotificationSent: false,
			expectServiceCalled:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockMessageService)

			if tt.expectServiceCalled {
				// Set up expectations for the service call
				mockService.On("SubmitMessage", mock.Anything, mock.MatchedBy(func(req domain.MessageSubmissionRequest) bool {
					// Verify the domain request is correctly mapped
					assert.Equal(t, tt.request.Content, req.Content)
					assert.Equal(t, tt.request.SendNotification, req.SendNotification)

					if tt.request.Sender != nil {
						assert.Equal(t, tt.request.Sender.Name, req.SenderName)
						assert.Equal(t, tt.request.Sender.Email, req.SenderEmail)
					}

					if tt.request.Recipient != nil {
						assert.Equal(t, tt.request.Recipient.Name, req.RecipientName)
						assert.Equal(t, tt.request.Recipient.Email, req.RecipientEmail)
					}

					return true
				})).Return(tt.mockServiceResponse, tt.mockServiceError)
			}

			// Create handler
			handler := NewMessageAPIHandler(mockService)

			// Create request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create recorder and router
			recorder := httptest.NewRecorder()
			router := gin.New()
			router.POST("/api/v1/messages", handler.SubmitMessage)

			// Execute request
			router.ServeHTTP(recorder, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatusCode, recorder.Code)

			if tt.expectedStatusCode == http.StatusCreated {
				// Parse successful response
				var response models.MessageSubmissionResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				// Assert response fields
				assert.Equal(t, tt.mockServiceResponse.MessageID, response.MessageID)
				assert.Equal(t, tt.expectedNotificationSent, response.NotificationSent)
				assert.NotEmpty(t, response.DecryptURL)
				assert.NotEmpty(t, response.WebURL)
				assert.False(t, response.ExpiresAt.IsZero())
			} else if tt.expectedStatusCode == http.StatusBadRequest {
				// Parse error response
				var errorResponse models.StandardErrorResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
				require.NoError(t, err)

				// Assert error response
				assert.Equal(t, "validation_failed", errorResponse.Error)
				assert.Equal(t, "Request validation failed", errorResponse.Message)
				assert.NotNil(t, errorResponse.Details)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestEmailValidationFlow(t *testing.T) {
	tests := []struct {
		name                 string
		senderEmail          string
		recipientEmail       string
		expectEmailErrors    bool
		expectedErrorFields  []string
	}{
		{
			name:             "valid email addresses",
			senderEmail:      "john@example.com",
			recipientEmail:   "jane@example.com",
			expectEmailErrors: false,
		},
		{
			name:                "invalid sender email format",
			senderEmail:         "invalid-email",
			recipientEmail:      "jane@example.com",
			expectEmailErrors:   true,
			expectedErrorFields: []string{"sender.email"},
		},
		{
			name:                "invalid recipient email format",
			senderEmail:         "john@example.com",
			recipientEmail:      "invalid-email",
			expectEmailErrors:   true,
			expectedErrorFields: []string{"recipient.email"},
		},
		{
			name:                "both emails invalid",
			senderEmail:         "invalid-sender",
			recipientEmail:      "invalid-recipient",
			expectEmailErrors:   true,
			expectedErrorFields: []string{"sender.email", "recipient.email"},
		},
		{
			name:                "email with special characters",
			senderEmail:         "test+tag@example.com",
			recipientEmail:      "user.name@subdomain.example.org",
			expectEmailErrors:   false,
		},
		{
			name:                "email with international domain",
			senderEmail:         "user@example.co.uk",
			recipientEmail:      "test@example.de",
			expectEmailErrors:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &models.MessageSubmissionRequest{
				Content: "Test message",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: tt.senderEmail,
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: tt.recipientEmail,
				},
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			}

			// Create mock service (won't be called for validation errors)
			mockService := new(MockMessageService)

			if !tt.expectEmailErrors {
				// If no email errors expected, set up successful service response
				mockService.On("SubmitMessage", mock.Anything, mock.Anything).Return(&domain.MessageSubmissionResponse{
					MessageID:  "test-id",
					DecryptURL: "https://example.com/decrypt/test-id/key",
					Success:    true,
				}, nil)
			}

			// Create handler
			handler := NewMessageAPIHandler(mockService)

			// Create request
			reqBody, err := json.Marshal(request)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create recorder and router
			recorder := httptest.NewRecorder()
			router := gin.New()
			router.POST("/api/v1/messages", handler.SubmitMessage)

			// Execute request
			router.ServeHTTP(recorder, req)

			if tt.expectEmailErrors {
				// Should return validation error
				assert.Equal(t, http.StatusBadRequest, recorder.Code)

				var errorResponse models.StandardErrorResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
				require.NoError(t, err)

				assert.Equal(t, "validation_failed", errorResponse.Error)
				assert.NotNil(t, errorResponse.Details)

				// Check for expected error fields
				for _, field := range tt.expectedErrorFields {
					assert.Contains(t, errorResponse.Details, field, "Expected error for field: %s", field)
				}
			} else {
				// Should succeed
				assert.Equal(t, http.StatusCreated, recorder.Code)

				var response models.MessageSubmissionResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.NotEmpty(t, response.MessageID)
				assert.True(t, response.NotificationSent)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestConditionalValidationFlow(t *testing.T) {
	// Test that validation rules change based on SendNotification flag
	tests := []struct {
		name               string
		sendNotification   bool
		includeSender      bool
		includeRecipient   bool
		includeAntiSpam    bool
		expectSuccess      bool
	}{
		{
			name:             "notification disabled - minimal data required",
			sendNotification: false,
			includeSender:    false,
			includeRecipient: true, // Recipient name now always required
			includeAntiSpam:  false,
			expectSuccess:    true,
		},
		{
			name:             "notification disabled - extra data ignored",
			sendNotification: false,
			includeSender:    true,
			includeRecipient: true,
			includeAntiSpam:  false, // Anti-spam not required when notification disabled
			expectSuccess:    true,
		},
		{
			name:             "notification enabled - all data required and provided",
			sendNotification: true,
			includeSender:    true,
			includeRecipient: true,
			includeAntiSpam:  true,
			expectSuccess:    true,
		},
		{
			name:             "notification enabled - missing sender",
			sendNotification: true,
			includeSender:    false,
			includeRecipient: true,
			includeAntiSpam:  true,
			expectSuccess:    false,
		},
		{
			name:             "notification enabled - missing recipient",
			sendNotification: true,
			includeSender:    true,
			includeRecipient: false,
			includeAntiSpam:  true,
			expectSuccess:    false,
		},
		{
			name:             "notification enabled - missing anti-spam",
			sendNotification: true,
			includeSender:    true,
			includeRecipient: true,
			includeAntiSpam:  false,
			expectSuccess:    false,
		},
		{
			name:             "missing recipient - optional when notifications disabled",
			sendNotification: false,
			includeSender:    false,
			includeRecipient: false,
			includeAntiSpam:  false,
			expectSuccess:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &models.MessageSubmissionRequest{
				Content:          "Test message",
				SendNotification: tt.sendNotification,
			}

			if tt.includeSender {
				request.Sender = &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				}
			}

			if tt.includeRecipient {
				request.Recipient = &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				}
			}

			if tt.includeAntiSpam {
				request.AntiSpamAnswer = "blue"
			}

			// Create mock service
			mockService := new(MockMessageService)

			if tt.expectSuccess {
				mockService.On("SubmitMessage", mock.Anything, mock.Anything).Return(&domain.MessageSubmissionResponse{
					MessageID:  "test-id",
					DecryptURL: "https://example.com/decrypt/test-id/key",
					Success:    true,
				}, nil)
			}

			// Create handler
			handler := NewMessageAPIHandler(mockService)

			// Create request
			reqBody, err := json.Marshal(request)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create recorder and router
			recorder := httptest.NewRecorder()
			router := gin.New()
			router.POST("/api/v1/messages", handler.SubmitMessage)

			// Execute request
			router.ServeHTTP(recorder, req)

			if tt.expectSuccess {
				assert.Equal(t, http.StatusCreated, recorder.Code)
			} else {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}