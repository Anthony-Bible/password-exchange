package domain

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

// MockNotificationSender mocks the NotificationSender interface
type MockNotificationSender struct {
	mock.Mock
}

func (m *MockNotificationSender) SendNotification(ctx context.Context, req NotificationRequest) (*NotificationResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*NotificationResponse), args.Error(1)
}

// MockNotificationPublisher mocks the NotificationPublisher interface
type MockNotificationPublisher struct {
	mock.Mock
}

func (m *MockNotificationPublisher) PublishNotification(ctx context.Context, req NotificationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// MockQueueConsumer mocks the QueueConsumer interface
type MockQueueConsumer struct {
	mock.Mock
}

func (m *MockQueueConsumer) StartConsuming(ctx context.Context, queueConn QueueConnection, handler MessageHandler, concurrency int) error {
	args := m.Called(ctx, queueConn, handler, concurrency)
	return args.Error(0)
}

func (m *MockQueueConsumer) Connect(ctx context.Context, queueConn QueueConnection) error {
	args := m.Called(ctx, queueConn)
	return args.Error(0)
}

func (m *MockQueueConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockTemplateRenderer mocks the TemplateRenderer interface
type MockTemplateRenderer struct {
	mock.Mock
}

func (m *MockTemplateRenderer) RenderTemplate(ctx context.Context, templateName string, data NotificationTemplateData) (string, error) {
	args := m.Called(ctx, templateName, data)
	return args.String(0), args.Error(1)
}

// MockStorageRepository mocks the StorageRepository interface
type MockStorageRepository struct {
	mock.Mock
}

func (m *MockStorageRepository) GetUnviewedMessagesForReminders(ctx context.Context, checkAfterHours, maxReminders int) ([]*UnviewedMessage, error) {
	args := m.Called(ctx, checkAfterHours, maxReminders)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*UnviewedMessage), args.Error(1)
}

func (m *MockStorageRepository) GetReminderHistory(ctx context.Context, messageID int) ([]*ReminderLogEntry, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ReminderLogEntry), args.Error(1)
}

func (m *MockStorageRepository) LogReminderSent(ctx context.Context, messageID int, recipientEmail string) error {
	args := m.Called(ctx, messageID, recipientEmail)
	return args.Error(0)
}

// Test NewNotificationService constructor
func TestNewNotificationService(t *testing.T) {
	// Arrange
	mockEmailSender := &MockNotificationSender{}
	mockQueueConsumer := &MockQueueConsumer{}
	mockTemplateRenderer := &MockTemplateRenderer{}
	mockStorageRepo := &MockStorageRepository{}

	// Act
	service := NewNotificationService(mockEmailSender, mockQueueConsumer, mockTemplateRenderer, mockStorageRepo, &MockNotificationPublisher{})

	// Assert
	assert.NotNil(t, service)
	assert.Equal(t, mockEmailSender, service.emailSender)
	assert.Equal(t, mockQueueConsumer, service.queueConsumer)
	assert.Equal(t, mockTemplateRenderer, service.templateRenderer)
	assert.NotNil(t, service.reminderService)
}

// Test SendNotification with valid request
func TestSendNotification_ValidRequest_Success(t *testing.T) {
	// Arrange
	mockEmailSender := &MockNotificationSender{}
	mockQueueConsumer := &MockQueueConsumer{}
	mockTemplateRenderer := &MockTemplateRenderer{}
	mockStorageRepo := &MockStorageRepository{}

	service := NewNotificationService(mockEmailSender, mockQueueConsumer, mockTemplateRenderer, mockStorageRepo, &MockNotificationPublisher{})
	
	ctx := context.Background()
	req := NotificationRequest{
		To:      "test@example.com",
		From:    "sender@example.com",
		Subject: "Test Subject",
	}
	
	expectedResponse := &NotificationResponse{
		Success:   true,
		MessageID: "msg-123",
	}

	mockEmailSender.On("SendNotification", ctx, req).Return(expectedResponse, nil)

	// Act
	response, err := service.SendNotification(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
	mockEmailSender.AssertExpectations(t)
}

// Test SendNotification with invalid request
func TestSendNotification_InvalidRequest_ReturnsError(t *testing.T) {
	// Arrange
	mockEmailSender := &MockNotificationSender{}
	mockQueueConsumer := &MockQueueConsumer{}
	mockTemplateRenderer := &MockTemplateRenderer{}
	mockStorageRepo := &MockStorageRepository{}

	service := NewNotificationService(mockEmailSender, mockQueueConsumer, mockTemplateRenderer, mockStorageRepo, &MockNotificationPublisher{})
	
	ctx := context.Background()
	req := NotificationRequest{
		To:      "", // Invalid: empty recipient
		From:    "sender@example.com",
		Subject: "Test Subject",
	}

	// Act
	response, err := service.SendNotification(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "recipient email is required")
	mockEmailSender.AssertNotCalled(t, "SendNotification")
}

// Test SendNotification when email sending fails
func TestSendNotification_EmailSendingFails_ReturnsError(t *testing.T) {
	// Arrange
	mockEmailSender := &MockNotificationSender{}
	mockQueueConsumer := &MockQueueConsumer{}
	mockTemplateRenderer := &MockTemplateRenderer{}
	mockStorageRepo := &MockStorageRepository{}

	service := NewNotificationService(mockEmailSender, mockQueueConsumer, mockTemplateRenderer, mockStorageRepo, &MockNotificationPublisher{})
	
	ctx := context.Background()
	req := NotificationRequest{
		To:      "test@example.com",
		From:    "sender@example.com",
		Subject: "Test Subject",
	}
	
	sendError := errors.New("email sending failed")
	mockEmailSender.On("SendNotification", ctx, req).Return(nil, sendError)

	// Act
	response, err := service.SendNotification(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "email sending failed")
	mockEmailSender.AssertExpectations(t)
}

// Test validateNotificationRequest
func TestValidateNotificationRequest(t *testing.T) {
	service := &NotificationService{}

	tests := []struct {
		name    string
		req     NotificationRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: NotificationRequest{
				To:      "test@example.com",
				From:    "sender@example.com",
				Subject: "Test Subject",
			},
			wantErr: false,
		},
		{
			name: "empty recipient",
			req: NotificationRequest{
				To:      "",
				From:    "sender@example.com",
				Subject: "Test Subject",
			},
			wantErr: true,
			errMsg:  "recipient email is required",
		},
		{
			name: "empty sender",
			req: NotificationRequest{
				To:      "test@example.com",
				From:    "",
				Subject: "Test Subject",
			},
			wantErr: true,
			errMsg:  "sender email is required",
		},
		{
			name: "empty subject",
			req: NotificationRequest{
				To:      "test@example.com",
				From:    "sender@example.com",
				Subject: "",
			},
			wantErr: true,
			errMsg:  "subject is required",
		},
		{
			name: "invalid recipient email",
			req: NotificationRequest{
				To:      "invalid-email",
				From:    "sender@example.com",
				Subject: "Test Subject",
			},
			wantErr: true,
			errMsg:  "invalid recipient email",
		},
		{
			name: "invalid sender email",
			req: NotificationRequest{
				To:      "test@example.com",
				From:    "invalid-email",
				Subject: "Test Subject",
			},
			wantErr: true,
			errMsg:  "invalid sender email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateNotificationRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test createNotificationRequest
func TestCreateNotificationRequest(t *testing.T) {
	// Arrange
	service := &NotificationService{}
	queueMsg := QueueMessage{
		FirstName:      "John",
		OtherFirstName: "Jane",
		OtherEmail:     "jane@example.com",
		Content:        "Test message content",
		URL:            "https://example.com/decrypt/123",
		Hidden:         "password123",
	}

	// Act
	notificationReq := service.createNotificationRequest(queueMsg)

	// Assert
	assert.Equal(t, "jane@example.com", notificationReq.To)
	assert.Equal(t, "server@password.exchange", notificationReq.From)
	assert.Equal(t, "Password Exchange", notificationReq.FromName)
	assert.Equal(t, "Encrypted Message from Password Exchange from John", notificationReq.Subject)
	assert.Equal(t, "Test message content", notificationReq.MessageContent)
	assert.Equal(t, "John", notificationReq.SenderName)
	assert.Equal(t, "Jane", notificationReq.RecipientName)
	assert.Equal(t, "https://example.com/decrypt/123", notificationReq.MessageURL)
	assert.Equal(t, "password123", notificationReq.Hidden)
}

// Test HandleMessage success
func TestHandleMessage_Success(t *testing.T) {
	// Arrange
	mockEmailSender := &MockNotificationSender{}
	mockQueueConsumer := &MockQueueConsumer{}
	mockTemplateRenderer := &MockTemplateRenderer{}
	mockStorageRepo := &MockStorageRepository{}

	service := NewNotificationService(mockEmailSender, mockQueueConsumer, mockTemplateRenderer, mockStorageRepo, &MockNotificationPublisher{})
	
	ctx := context.Background()
	queueMsg := QueueMessage{
		FirstName:      "John",
		OtherFirstName: "Jane",
		OtherEmail:     "jane@example.com",
		Content:        "Test message content",
		URL:            "https://example.com/decrypt/123",
		Hidden:         "password123",
	}

	expectedResponse := &NotificationResponse{
		Success:   true,
		MessageID: "msg-123",
	}

	mockEmailSender.On("SendNotification", ctx, mock.AnythingOfType("NotificationRequest")).Return(expectedResponse, nil)

	// Act
	err := service.HandleMessage(ctx, queueMsg)

	// Assert
	assert.NoError(t, err)
	mockEmailSender.AssertExpectations(t)
}

// Test HandleMessage failure
func TestHandleMessage_SendNotificationFails_ReturnsError(t *testing.T) {
	// Arrange
	mockEmailSender := &MockNotificationSender{}
	mockQueueConsumer := &MockQueueConsumer{}
	mockTemplateRenderer := &MockTemplateRenderer{}
	mockStorageRepo := &MockStorageRepository{}

	service := NewNotificationService(mockEmailSender, mockQueueConsumer, mockTemplateRenderer, mockStorageRepo, &MockNotificationPublisher{})
	
	ctx := context.Background()
	queueMsg := QueueMessage{
		FirstName:      "John",
		OtherFirstName: "Jane",
		OtherEmail:     "jane@example.com",
		Content:        "Test message content",
		URL:            "https://example.com/decrypt/123",
		Hidden:         "password123",
	}

	sendError := errors.New("email sending failed")
	mockEmailSender.On("SendNotification", ctx, mock.AnythingOfType("NotificationRequest")).Return(nil, sendError)

	// Act
	err := service.HandleMessage(ctx, queueMsg)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email sending failed")
	mockEmailSender.AssertExpectations(t)
}