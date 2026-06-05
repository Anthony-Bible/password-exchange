package rabbitmq

import (
	"context"
	"errors"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	messagepb "github.com/Anthony-Bible/password-exchange/app/pkg/pb/message"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Compile-time interface check.
var _ secondary.NotificationServicePort = &NotificationPublisher{}

// ---------------------------------------------------------------------------
// Stubs
// ---------------------------------------------------------------------------

// stubChannel is a test double for amqpChannel. Each function field defaults to
// a sensible no-op so individual tests only override what they need to exercise.
type stubChannel struct {
	queueDeclareFn       func(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	publishWithContextFn func(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	closeFn              func() error
}

func newStubChannel() *stubChannel {
	return &stubChannel{
		queueDeclareFn: func(name string, _, _, _, _ bool, _ amqp.Table) (amqp.Queue, error) {
			return amqp.Queue{Name: name}, nil
		},
		publishWithContextFn: func(_ context.Context, _, _ string, _, _ bool, _ amqp.Publishing) error {
			return nil
		},
		closeFn: func() error { return nil },
	}
}

func (s *stubChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return s.queueDeclareFn(name, durable, autoDelete, exclusive, noWait, args)
}

func (s *stubChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return s.publishWithContextFn(ctx, exchange, key, mandatory, immediate, msg)
}

func (s *stubChannel) Close() error { return s.closeFn() }

// stubConnection is a test double for amqpConnection.
type stubConnection struct {
	closeFn func() error
}

func newStubConnection() *stubConnection {
	return &stubConnection{closeFn: func() error { return nil }}
}

func (s *stubConnection) Close() error { return s.closeFn() }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newTestPublisher constructs a NotificationPublisher with injected stubs,
// bypassing NewNotificationPublisher which requires a live broker.
func newTestPublisher(ch amqpChannel, conn amqpConnection, queueName string) *NotificationPublisher {
	return &NotificationPublisher{
		channel:    ch,
		connection: conn,
		queueName:  queueName,
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestSendMessageNotification_HappyPath verifies the full observable contract:
// queue declared with the configured name, publish targets the correct exchange
// and routing key, AMQP headers match the protocol requirements, and the body
// deserializes to a correctly field-mapped protobuf.
func TestSendMessageNotification_HappyPath(t *testing.T) {
	t.Parallel()

	const queueName = "test-notifications"

	req := domain.MessageNotificationRequest{
		SenderEmail:    "sender@example.com",
		SenderName:     "Alice",
		RecipientEmail: "recipient@example.com",
		RecipientName:  "Bob",
		MessageURL:     "https://password.exchange/m/abc123",
		AdditionalInfo: "handle with care",
	}

	var (
		capturedQueueName string
		capturedExchange  string
		capturedKey       string
		capturedMsg       amqp.Publishing
	)

	ch := newStubChannel()
	ch.queueDeclareFn = func(name string, _, _, _, _ bool, _ amqp.Table) (amqp.Queue, error) {
		capturedQueueName = name
		return amqp.Queue{Name: name}, nil
	}
	ch.publishWithContextFn = func(_ context.Context, exchange, key string, _, _ bool, msg amqp.Publishing) error {
		capturedExchange = exchange
		capturedKey = key
		capturedMsg = msg
		return nil
	}

	p := newTestPublisher(ch, newStubConnection(), queueName)

	err := p.SendMessageNotification(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, queueName, capturedQueueName, "QueueDeclare should use the configured queue name")
	assert.Equal(t, "", capturedExchange, "exchange must be the default (empty string)")
	assert.Equal(t, queueName, capturedKey, "routing key must equal the queue name")
	assert.Equal(t, "application/protobuf", capturedMsg.ContentType)
	assert.Equal(t, uint8(amqp.Persistent), capturedMsg.DeliveryMode)

	var pbMsg messagepb.Message
	require.NoError(t, proto.Unmarshal(capturedMsg.Body, &pbMsg), "published body must be valid protobuf")

	assert.Equal(t, req.SenderEmail, pbMsg.Email)
	assert.Equal(t, req.SenderName, pbMsg.FirstName)
	assert.Equal(t, req.RecipientName, pbMsg.OtherFirstName)
	assert.Equal(t, req.RecipientEmail, pbMsg.OtherEmail)
	assert.Equal(t, req.MessageURL, pbMsg.Url)
	assert.Equal(t, req.AdditionalInfo, pbMsg.Hidden)
	assert.Contains(t, pbMsg.Content, req.MessageURL)
}

// TestSendMessageNotification_QueueDeclareError verifies that a QueueDeclare
// failure is propagated and Publish is never attempted.
func TestSendMessageNotification_QueueDeclareError(t *testing.T) {
	t.Parallel()

	declareErr := errors.New("broker unreachable")
	publishCalled := false

	ch := newStubChannel()
	ch.queueDeclareFn = func(_ string, _, _, _, _ bool, _ amqp.Table) (amqp.Queue, error) {
		return amqp.Queue{}, declareErr
	}
	ch.publishWithContextFn = func(_ context.Context, _, _ string, _, _ bool, _ amqp.Publishing) error {
		publishCalled = true
		return nil
	}

	p := newTestPublisher(ch, newStubConnection(), "q")

	err := p.SendMessageNotification(context.Background(), domain.MessageNotificationRequest{
		MessageURL: "https://password.exchange/m/x",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, declareErr)
	assert.False(t, publishCalled, "Publish must not be called when QueueDeclare fails")
}

// TestSendMessageNotification_PublishError verifies that a PublishWithContext
// failure is propagated as the return value.
func TestSendMessageNotification_PublishError(t *testing.T) {
	t.Parallel()

	publishErr := errors.New("channel closed unexpectedly")

	ch := newStubChannel()
	ch.publishWithContextFn = func(_ context.Context, _, _ string, _, _ bool, _ amqp.Publishing) error {
		return publishErr
	}

	p := newTestPublisher(ch, newStubConnection(), "q")

	err := p.SendMessageNotification(context.Background(), domain.MessageNotificationRequest{
		MessageURL: "https://password.exchange/m/x",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, publishErr)
}

// TestClose_CallsChannelAndConnectionClose verifies that Close() invokes Close
// on both the channel and the connection.
func TestClose_CallsChannelAndConnectionClose(t *testing.T) {
	t.Parallel()

	channelClosed := false
	connClosed := false

	ch := newStubChannel()
	ch.closeFn = func() error {
		channelClosed = true
		return nil
	}

	conn := newStubConnection()
	conn.closeFn = func() error {
		connClosed = true
		return nil
	}

	p := newTestPublisher(ch, conn, "q")

	assert.NoError(t, p.Close())
	assert.True(t, channelClosed, "channel.Close() must be called")
	assert.True(t, connClosed, "connection.Close() must be called")
}

// TestClose_NilChannelAndConnection guards against nil-dereference panics on a
// zero-value NotificationPublisher.
func TestClose_NilChannelAndConnection(t *testing.T) {
	t.Parallel()

	p := &NotificationPublisher{queueName: "q"}

	assert.NotPanics(t, func() {
		assert.NoError(t, p.Close())
	})
}

// TestSendMessageNotification_TableDriven exercises distinct request field
// combinations to confirm the proto mapping holds across non-trivial and
// zero-value inputs.
func TestSendMessageNotification_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  domain.MessageNotificationRequest
	}{
		{
			name: "all fields populated",
			req: domain.MessageNotificationRequest{
				SenderEmail:    "alice@example.com",
				SenderName:     "Alice",
				RecipientEmail: "bob@example.com",
				RecipientName:  "Bob",
				MessageURL:     "https://password.exchange/m/1",
				AdditionalInfo: "secret hint",
			},
		},
		{
			name: "empty additional info",
			req: domain.MessageNotificationRequest{
				SenderEmail:    "x@example.com",
				SenderName:     "X",
				RecipientEmail: "y@example.com",
				RecipientName:  "Y",
				MessageURL:     "https://password.exchange/m/2",
				AdditionalInfo: "",
			},
		},
		{
			name: "unicode names",
			req: domain.MessageNotificationRequest{
				SenderEmail:    "sender@example.com",
				SenderName:     "日向 葵",
				RecipientEmail: "recipient@example.com",
				RecipientName:  "Björn",
				MessageURL:     "https://password.exchange/m/unicode",
				AdditionalInfo: "こんにちは",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var capturedBody []byte

			ch := newStubChannel()
			ch.publishWithContextFn = func(_ context.Context, _, _ string, _, _ bool, msg amqp.Publishing) error {
				capturedBody = msg.Body
				return nil
			}

			p := newTestPublisher(ch, newStubConnection(), "notifications")

			err := p.SendMessageNotification(context.Background(), tc.req)
			require.NoError(t, err)

			var pbMsg messagepb.Message
			require.NoError(t, proto.Unmarshal(capturedBody, &pbMsg))

			assert.Equal(t, tc.req.SenderEmail, pbMsg.Email)
			assert.Equal(t, tc.req.SenderName, pbMsg.FirstName)
			assert.Equal(t, tc.req.RecipientName, pbMsg.OtherFirstName)
			assert.Equal(t, tc.req.RecipientEmail, pbMsg.OtherEmail)
			assert.Equal(t, tc.req.MessageURL, pbMsg.Url)
			assert.Equal(t, tc.req.AdditionalInfo, pbMsg.Hidden)
			assert.Contains(t, pbMsg.Content, tc.req.MessageURL)
		})
	}
}
