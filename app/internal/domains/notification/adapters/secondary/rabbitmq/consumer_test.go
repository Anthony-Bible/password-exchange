package rabbitmq

import (
	"context"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	pb "github.com/Anthony-Bible/password-exchange/app/pkg/pb/message"
	"github.com/golang/protobuf/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessageHandler implements domain.MessageHandler for testing
type MockMessageHandler struct {
	mock.Mock
}

func (m *MockMessageHandler) HandleMessage(ctx context.Context, msg domain.QueueMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func TestHandleMessage_ValidProtobufMessage(t *testing.T) {
	// Arrange
	consumer := NewRabbitMQConsumer()
	mockHandler := &MockMessageHandler{}

	// Create a valid protobuf message
	pbMsg := &pb.Message{
		Email:          "test@example.com",
		FirstName:      "John",
		OtherFirstName: "Jane",
		OtherLastName:  "Doe",
		OtherEmail:     "jane.doe@example.com",
		UniqueId:       "unique123",
		Content:        "Test password message",
		Url:            "https://password.exchange/view/unique123",
		Hidden:         "false",
		Captcha:        "captcha123",
	}

	// Marshal to protobuf bytes
	pbBytes, err := proto.Marshal(pbMsg)
	assert.NoError(t, err)

	// Create AMQP delivery
	delivery := amqp.Delivery{
		Body: pbBytes,
	}

	// Expected domain message after conversion
	expectedMsg := domain.QueueMessage{
		Email:          "test@example.com",
		FirstName:      "John",
		OtherFirstName: "Jane",
		OtherLastName:  "Doe",
		OtherEmail:     "jane.doe@example.com",
		UniqueID:       "unique123",
		Content:        "Test password message",
		URL:            "https://password.exchange/view/unique123",
		Hidden:         "false",
		Captcha:        "captcha123",
	}

	mockHandler.On("HandleMessage", mock.Anything, expectedMsg).Return(nil)

	// Act
	success := consumer.handleMessage(context.Background(), delivery, mockHandler, 1)

	// Assert
	assert.True(t, success)
	mockHandler.AssertExpectations(t)
}

func TestHandleMessage_EmptyBody(t *testing.T) {
	// Arrange
	consumer := NewRabbitMQConsumer()
	mockHandler := &MockMessageHandler{}

	delivery := amqp.Delivery{
		Body: nil,
	}

	// Act
	success := consumer.handleMessage(context.Background(), delivery, mockHandler, 1)

	// Assert
	assert.False(t, success)
	mockHandler.AssertNotCalled(t, "HandleMessage")
}

func TestHandleMessage_InvalidProtobufData(t *testing.T) {
	// Arrange
	consumer := NewRabbitMQConsumer()
	mockHandler := &MockMessageHandler{}

	// Invalid protobuf data
	delivery := amqp.Delivery{
		Body: []byte("invalid protobuf data"),
	}

	// Act
	success := consumer.handleMessage(context.Background(), delivery, mockHandler, 1)

	// Assert
	assert.False(t, success)
	mockHandler.AssertNotCalled(t, "HandleMessage")
}

func TestHandleMessage_HandlerError(t *testing.T) {
	// Arrange
	consumer := NewRabbitMQConsumer()
	mockHandler := &MockMessageHandler{}

	// Create valid protobuf message
	pbMsg := &pb.Message{
		Email:     "test@example.com",
		OtherEmail: "recipient@example.com",
		UniqueId:  "unique123",
	}

	pbBytes, err := proto.Marshal(pbMsg)
	assert.NoError(t, err)

	delivery := amqp.Delivery{
		Body: pbBytes,
	}

	expectedMsg := domain.QueueMessage{
		Email:     "test@example.com",
		OtherEmail: "recipient@example.com",
		UniqueID:  "unique123",
	}

	// Mock handler to return error
	mockHandler.On("HandleMessage", mock.Anything, expectedMsg).Return(assert.AnError)

	// Act
	success := consumer.handleMessage(context.Background(), delivery, mockHandler, 1)

	// Assert
	assert.False(t, success)
	mockHandler.AssertExpectations(t)
}

func TestHandleMessage_ProtobufFieldMapping(t *testing.T) {
	// Test each protobuf field is correctly mapped to domain message
	testCases := []struct {
		name     string
		pbMsg    *pb.Message
		expected domain.QueueMessage
	}{
		{
			name: "All fields populated",
			pbMsg: &pb.Message{
				Email:          "sender@example.com",
				FirstName:      "Alice",
				OtherFirstName: "Bob",
				OtherLastName:  "Smith",
				OtherEmail:     "bob.smith@example.com",
				UniqueId:       "abc123",
				Content:        "Secret password content",
				Url:            "https://password.exchange/view/abc123",
				Hidden:         "true",
				Captcha:        "cap456",
			},
			expected: domain.QueueMessage{
				Email:          "sender@example.com",
				FirstName:      "Alice",
				OtherFirstName: "Bob",
				OtherLastName:  "Smith",
				OtherEmail:     "bob.smith@example.com",
				UniqueID:       "abc123",
				Content:        "Secret password content",
				URL:            "https://password.exchange/view/abc123",
				Hidden:         "true",
				Captcha:        "cap456",
			},
		},
		{
			name: "Minimal fields",
			pbMsg: &pb.Message{
				OtherEmail: "minimal@example.com",
				UniqueId:   "min123",
			},
			expected: domain.QueueMessage{
				OtherEmail: "minimal@example.com",
				UniqueID:   "min123",
			},
		},
		{
			name: "Empty strings preserved",
			pbMsg: &pb.Message{
				Email:          "",
				FirstName:      "",
				OtherFirstName: "",
				OtherLastName:  "",
				OtherEmail:     "empty@example.com",
				UniqueId:       "empty123",
				Content:        "",
				Url:            "",
				Hidden:         "false",
				Captcha:        "",
			},
			expected: domain.QueueMessage{
				Email:          "",
				FirstName:      "",
				OtherFirstName: "",
				OtherLastName:  "",
				OtherEmail:     "empty@example.com",
				UniqueID:       "empty123",
				Content:        "",
				URL:            "",
				Hidden:         "false",
				Captcha:        "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			consumer := NewRabbitMQConsumer()
			mockHandler := &MockMessageHandler{}

			pbBytes, err := proto.Marshal(tc.pbMsg)
			assert.NoError(t, err)

			delivery := amqp.Delivery{
				Body: pbBytes,
			}

			mockHandler.On("HandleMessage", mock.Anything, tc.expected).Return(nil)

			// Act
			success := consumer.handleMessage(context.Background(), delivery, mockHandler, 1)

			// Assert
			assert.True(t, success)
			mockHandler.AssertExpectations(t)
		})
	}
}

func TestHandleMessage_ProtobufBinaryData(t *testing.T) {
	// Test handling of binary protobuf data with various encodings
	consumer := NewRabbitMQConsumer()
	mockHandler := &MockMessageHandler{}

	// Create message with special characters and Unicode
	pbMsg := &pb.Message{
		Email:     "test@example.com",
		Content:   "Password with special chars: !@#$%^&*()_+ and Unicode: 你好",
		OtherEmail: "recipient@测试.com",
		UniqueId:  "unicode-123",
	}

	pbBytes, err := proto.Marshal(pbMsg)
	assert.NoError(t, err)

	delivery := amqp.Delivery{
		Body: pbBytes,
	}

	expectedMsg := domain.QueueMessage{
		Email:     "test@example.com",
		Content:   "Password with special chars: !@#$%^&*()_+ and Unicode: 你好",
		OtherEmail: "recipient@测试.com",
		UniqueID:  "unicode-123",
	}

	mockHandler.On("HandleMessage", mock.Anything, expectedMsg).Return(nil)

	// Act
	success := consumer.handleMessage(context.Background(), delivery, mockHandler, 1)

	// Assert
	assert.True(t, success)
	mockHandler.AssertExpectations(t)
}

func TestProtobufMarshalUnmarshal_RoundTrip(t *testing.T) {
	// Test protobuf round-trip marshaling/unmarshaling
	originalMsg := &pb.Message{
		Email:          "roundtrip@example.com",
		FirstName:      "Round",
		OtherFirstName: "Trip",
		OtherLastName:  "Test",
		OtherEmail:     "trip@example.com",
		UniqueId:       "round123",
		Content:        "Round trip test content",
		Url:            "https://password.exchange/view/round123",
		Hidden:         "true",
		Captcha:        "round456",
	}

	// Marshal to bytes
	pbBytes, err := proto.Marshal(originalMsg)
	assert.NoError(t, err)
	assert.NotEmpty(t, pbBytes)

	// Unmarshal back to message
	var unmarshaledMsg pb.Message
	err = proto.Unmarshal(pbBytes, &unmarshaledMsg)
	assert.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, originalMsg.Email, unmarshaledMsg.Email)
	assert.Equal(t, originalMsg.FirstName, unmarshaledMsg.FirstName)
	assert.Equal(t, originalMsg.OtherFirstName, unmarshaledMsg.OtherFirstName)
	assert.Equal(t, originalMsg.OtherLastName, unmarshaledMsg.OtherLastName)
	assert.Equal(t, originalMsg.OtherEmail, unmarshaledMsg.OtherEmail)
	assert.Equal(t, originalMsg.UniqueId, unmarshaledMsg.UniqueId)
	assert.Equal(t, originalMsg.Content, unmarshaledMsg.Content)
	assert.Equal(t, originalMsg.Url, unmarshaledMsg.Url)
	assert.Equal(t, originalMsg.Hidden, unmarshaledMsg.Hidden)
	assert.Equal(t, originalMsg.Captcha, unmarshaledMsg.Captcha)
}

func TestProtobufMessage_FieldValidation(t *testing.T) {
	// Test protobuf message field validation and edge cases
	testCases := []struct {
		name        string
		setupMsg    func() *pb.Message
		shouldError bool
	}{
		{
			name: "Valid message with all fields",
			setupMsg: func() *pb.Message {
				return &pb.Message{
					Email:     "valid@example.com",
					OtherEmail: "other@example.com",
					UniqueId:  "valid123",
				}
			},
			shouldError: false,
		},
		{
			name: "Message with very long strings",
			setupMsg: func() *pb.Message {
				longString := make([]byte, 10000)
				for i := range longString {
					longString[i] = 'a'
				}
				return &pb.Message{
					Email:   "long@example.com",
					Content: string(longString),
					UniqueId: "long123",
				}
			},
			shouldError: false,
		},
		{
			name: "Empty message",
			setupMsg: func() *pb.Message {
				return &pb.Message{}
			},
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := tc.setupMsg()

			// Test marshaling
			pbBytes, err := proto.Marshal(msg)
			if tc.shouldError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Test unmarshaling
			var unmarshaledMsg pb.Message
			err = proto.Unmarshal(pbBytes, &unmarshaledMsg)
			assert.NoError(t, err)
		})
	}
}