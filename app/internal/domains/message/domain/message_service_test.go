package domain

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/contracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const notificationAsyncReturnThreshold = 1 * time.Second

// --- Mocks ---

type mockEncryptionService struct{ mock.Mock }

func (m *mockEncryptionService) GenerateKey(ctx context.Context, length int32) ([]byte, error) {
	args := m.Called(ctx, length)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockEncryptionService) Encrypt(ctx context.Context, plaintext []string, key []byte) ([]string, error) {
	args := m.Called(ctx, plaintext, key)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockEncryptionService) Decrypt(ctx context.Context, ciphertext []string, key []byte) ([]string, error) {
	args := m.Called(ctx, ciphertext, key)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockEncryptionService) GenerateID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *mockEncryptionService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type mockStorageService struct{ mock.Mock }

func (m *mockStorageService) StoreMessage(ctx context.Context, req MessageStorageRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockStorageService) RetrieveMessage(
	ctx context.Context,
	req MessageRetrievalStorageRequest,
) (*MessageStorageResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*MessageStorageResponse), args.Error(1)
}

func (m *mockStorageService) GetMessage(
	ctx context.Context,
	req MessageRetrievalStorageRequest,
) (*MessageStorageResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*MessageStorageResponse), args.Error(1)
}

func (m *mockStorageService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type mockNotificationService struct{ mock.Mock }

func (m *mockNotificationService) SendMessageNotification(ctx context.Context, req MessageNotificationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// blockingNotificationService simulates a notification sender that blocks until released.
type blockingNotificationService struct {
	mock.Mock
	release chan struct{}
	called  chan struct{}
}

func (m *blockingNotificationService) SendMessageNotification(ctx context.Context, req MessageNotificationRequest) error {
	args := m.Called(ctx, req)
	select {
	case m.called <- struct{}{}:
	default:
	}
	<-m.release
	return args.Error(0)
}

type mockPasswordHasher struct{ mock.Mock }

func (m *mockPasswordHasher) Hash(ctx context.Context, password string) (string, error) {
	args := m.Called(ctx, password)
	return args.String(0), args.Error(1)
}

func (m *mockPasswordHasher) Verify(ctx context.Context, password, hash string) (bool, error) {
	args := m.Called(ctx, password, hash)
	return args.Bool(0), args.Error(1)
}

type mockURLBuilder struct{ mock.Mock }

func (m *mockURLBuilder) BuildDecryptURL(messageID string, encryptionKey []byte) string {
	args := m.Called(messageID, encryptionKey)
	return args.String(0)
}

type mockTurnstileValidator struct{ mock.Mock }

func (m *mockTurnstileValidator) ValidateToken(ctx context.Context, token string, remoteIP string) (bool, error) {
	args := m.Called(ctx, token, remoteIP)
	return args.Bool(0), args.Error(1)
}

type mockLogger struct{ mock.Mock }

func (m *mockLogger) Debug() contracts.LogEvent {
	args := m.Called()
	return args.Get(0).(contracts.LogEvent)
}

func (m *mockLogger) Info() contracts.LogEvent {
	args := m.Called()
	return args.Get(0).(contracts.LogEvent)
}

func (m *mockLogger) Warn() contracts.LogEvent {
	args := m.Called()
	return args.Get(0).(contracts.LogEvent)
}

func (m *mockLogger) Error() contracts.LogEvent {
	args := m.Called()
	return args.Get(0).(contracts.LogEvent)
}

type mockLogEvent struct{ mock.Mock }

func (m *mockLogEvent) Err(err error) contracts.LogEvent {
	args := m.Called(err)
	return args.Get(0).(contracts.LogEvent)
}

func (m *mockLogEvent) Str(key, val string) contracts.LogEvent {
	args := m.Called(key, val)
	return args.Get(0).(contracts.LogEvent)
}

func (m *mockLogEvent) Int(key string, val int) contracts.LogEvent {
	args := m.Called(key, val)
	return args.Get(0).(contracts.LogEvent)
}

func (m *mockLogEvent) Bool(key string, val bool) contracts.LogEvent {
	args := m.Called(key, val)
	return args.Get(0).(contracts.LogEvent)
}

func (m *mockLogEvent) Dur(key string, val time.Duration) contracts.LogEvent {
	args := m.Called(key, val)
	return args.Get(0).(contracts.LogEvent)
}

func (m *mockLogEvent) Float64(key string, val float64) contracts.LogEvent {
	args := m.Called(key, val)
	return args.Get(0).(contracts.LogEvent)
}

func (m *mockLogEvent) Msg(msg string) {
	m.Called(msg)
}

// Int32, Int64, Interface and Msgf round out the shared LogEvent interface.
// The message domain never calls them, so they are simple pass-throughs.
func (m *mockLogEvent) Int32(string, int32) contracts.LogEvent   { return m }
func (m *mockLogEvent) Int64(string, int64) contracts.LogEvent   { return m }
func (m *mockLogEvent) Interface(string, any) contracts.LogEvent { return m }
func (m *mockLogEvent) Msgf(string, ...any)                      {}

type mockConfig struct{ mock.Mock }

func (m *mockConfig) GetDefaultMaxViewCount() int {
	args := m.Called()
	return args.Int(0)
}

type mockValidation struct{ mock.Mock }

func (m *mockValidation) SanitizeEmailForLogging(email string) string {
	args := m.Called(email)
	return args.String(0)
}

// setupLenientLoggerMock sets up lenient expectations for a logger mock that will match any log calls
func setupLenientLoggerMock(l *mockLogger) *mockLogEvent {
	ev := &mockLogEvent{}
	l.On("Debug").Return(ev).Maybe()
	l.On("Info").Return(ev).Maybe()
	l.On("Warn").Return(ev).Maybe()
	l.On("Error").Return(ev).Maybe()

	ev.On("Err", mock.Anything).Return(ev).Maybe()
	ev.On("Str", mock.Anything, mock.Anything).Return(ev).Maybe()
	ev.On("Int", mock.Anything, mock.Anything).Return(ev).Maybe()
	ev.On("Bool", mock.Anything, mock.Anything).Return(ev).Maybe()
	ev.On("Dur", mock.Anything, mock.Anything).Return(ev).Maybe()
	ev.On("Float64", mock.Anything, mock.Anything).Return(ev).Maybe()
	ev.On("Msg", mock.Anything).Return().Maybe()

	return ev
}

// setupTestMocks sets up standard expectations for all mocks
func setupTestMocks(
	enc *mockEncryptionService,
	stor *mockStorageService,
	notif *mockNotificationService,
	hasher *mockPasswordHasher,
	urlb *mockURLBuilder,
	turnstile *mockTurnstileValidator,
	logger *mockLogger,
	config *mockConfig,
	validation *mockValidation,
) {
	setupLenientLoggerMock(logger)
	validation.On("SanitizeEmailForLogging", mock.Anything).Return("sanitized-email@example.com").Maybe()
	config.On("GetDefaultMaxViewCount").Return(5).Maybe()
}

func createTestMocks() (
	*mockEncryptionService,
	*mockStorageService,
	*mockNotificationService,
	*mockPasswordHasher,
	*mockURLBuilder,
	*mockTurnstileValidator,
	*mockLogger,
	*mockConfig,
	*mockValidation,
) {
	return new(mockEncryptionService),
		new(mockStorageService),
		new(mockNotificationService),
		new(mockPasswordHasher),
		new(mockURLBuilder),
		new(mockTurnstileValidator),
		new(mockLogger),
		new(mockConfig),
		new(mockValidation)
}

// --- Tests ---

func TestRetrieveMessage_PropagatesExpiresAt(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	fixedExpiry := time.Date(2030, 3, 7, 12, 0, 0, 0, time.UTC)
	encodedContent := base64.URLEncoding.EncodeToString([]byte("secret"))
	storageResp := &MessageStorageResponse{
		MessageID:        "msg-1",
		EncryptedContent: "ciphertext",
		HasPassphrase:    false,
		ViewCount:        1,
		MaxViewCount:     5,
		ExpiresAt:        &fixedExpiry,
	}

	stor.On("GetMessage", mock.Anything, MessageRetrievalStorageRequest{MessageID: "msg-1"}).
		Return(storageResp, nil)
	stor.On("RetrieveMessage", mock.Anything, MessageRetrievalStorageRequest{MessageID: "msg-1"}).
		Return(storageResp, nil)
	enc.On("Decrypt", mock.Anything, []string{"ciphertext"}, []byte("key")).
		Return([]string{encodedContent}, nil)

	resp, err := svc.RetrieveMessage(context.Background(), MessageRetrievalRequest{
		MessageID:     "msg-1",
		DecryptionKey: []byte("key"),
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp.ExpiresAt, "ExpiresAt must be propagated from storage to retrieval response")
	assert.Equal(t, fixedExpiry.Unix(), resp.ExpiresAt.Unix())

	stor.AssertExpectations(t)
	enc.AssertExpectations(t)
}

func TestRetrieveMessage_NilExpiresAtPropagated(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	encodedContent := base64.URLEncoding.EncodeToString([]byte("secret"))
	storageResp := &MessageStorageResponse{
		MessageID:        "msg-legacy",
		EncryptedContent: "ciphertext",
		HasPassphrase:    false,
		ViewCount:        1,
		MaxViewCount:     5,
		ExpiresAt:        nil,
	}

	stor.On("GetMessage", mock.Anything, MessageRetrievalStorageRequest{MessageID: "msg-legacy"}).
		Return(storageResp, nil)
	stor.On("RetrieveMessage", mock.Anything, MessageRetrievalStorageRequest{MessageID: "msg-legacy"}).
		Return(storageResp, nil)
	enc.On("Decrypt", mock.Anything, []string{"ciphertext"}, []byte("key")).
		Return([]string{encodedContent}, nil)

	resp, err := svc.RetrieveMessage(context.Background(), MessageRetrievalRequest{
		MessageID:     "msg-legacy",
		DecryptionKey: []byte("key"),
	})

	assert.NoError(t, err)
	assert.Nil(t, resp.ExpiresAt, "nil ExpiresAt from storage must remain nil in retrieval response")

	stor.AssertExpectations(t)
	enc.AssertExpectations(t)
}

func TestSubmitMessage_CustomExpirationHours(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	enc.On("GenerateKey", mock.Anything, int32(32)).Return([]byte("key12345678901234567890123456789"), nil)
	enc.On("Encrypt", mock.Anything, mock.Anything, mock.Anything).Return([]string{"ciphertext"}, nil)
	enc.On("GenerateID", mock.Anything).Return("msg-custom-ttl", nil)
	stor.On("StoreMessage", mock.Anything, mock.MatchedBy(func(req MessageStorageRequest) bool {
		// ExpiresAt must be set and roughly match 48h from now (within a 5-minute tolerance)
		if req.ExpiresAt == nil {
			return false
		}
		expected := time.Now().Add(48 * time.Hour)
		diff := req.ExpiresAt.Sub(expected)
		if diff < 0 {
			diff = -diff
		}
		return diff < 5*time.Minute
	})).Return(nil)
	urlb.On("BuildDecryptURL", "msg-custom-ttl", mock.Anything).Return("https://example.com/decrypt/msg-custom-ttl")

	before := time.Now()
	resp, err := svc.SubmitMessage(context.Background(), MessageSubmissionRequest{
		Content:         "secret",
		ExpirationHours: 48,
	})
	after := time.Now()

	assert.NoError(t, err)
	assert.NotNil(t, resp.ExpiresAt)
	// ExpiresAt should be approximately 48 hours from now
	expectedMin := before.Add(48 * time.Hour)
	expectedMax := after.Add(48 * time.Hour)
	assert.True(t, !resp.ExpiresAt.Before(expectedMin), "ExpiresAt should be >= 48h from before")
	assert.True(t, !resp.ExpiresAt.After(expectedMax), "ExpiresAt should be <= 48h from after")

	stor.AssertExpectations(t)
}

func TestSubmitMessage_DefaultExpirationWhenZero(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	enc.On("GenerateKey", mock.Anything, int32(32)).Return([]byte("key12345678901234567890123456789"), nil)
	enc.On("Encrypt", mock.Anything, mock.Anything, mock.Anything).Return([]string{"ciphertext"}, nil)
	enc.On("GenerateID", mock.Anything).Return("msg-default-ttl", nil)
	stor.On("StoreMessage", mock.Anything, mock.MatchedBy(func(req MessageStorageRequest) bool {
		if req.ExpiresAt == nil {
			return false
		}
		expected := time.Now().Add(DefaultMessageTTL)
		diff := req.ExpiresAt.Sub(expected)
		if diff < 0 {
			diff = -diff
		}
		return diff < 5*time.Minute
	})).Return(nil)
	urlb.On("BuildDecryptURL", "msg-default-ttl", mock.Anything).Return("https://example.com/decrypt/msg-default-ttl")

	resp, err := svc.SubmitMessage(context.Background(), MessageSubmissionRequest{
		Content:         "secret",
		ExpirationHours: 0, // zero → use default
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp.ExpiresAt)
	// ExpiresAt should be approximately DefaultMessageTTL from now
	expected := time.Now().Add(DefaultMessageTTL)
	diff := resp.ExpiresAt.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	assert.Less(t, diff, 5*time.Minute, "ExpiresAt should be within 5 minutes of default TTL")

	stor.AssertExpectations(t)
}

func TestSubmitMessage_ExpirationHoursValidation(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	tests := []struct {
		name  string
		hours int
	}{
		{"negative hours", -1},
		{"exceeds max", MaxExpirationHours + 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.SubmitMessage(context.Background(), MessageSubmissionRequest{
				Content:         "secret",
				ExpirationHours: tc.hours,
			})
			assert.Error(t, err, "should reject ExpirationHours=%d", tc.hours)
		})
	}

	stor.AssertNotCalled(t, "StoreMessage", mock.Anything, mock.Anything)
}

func TestSubmitMessage_DoesNotWaitForNotificationPublish(t *testing.T) {
	enc := new(mockEncryptionService)
	stor := new(mockStorageService)
	notif := &blockingNotificationService{
		release: make(chan struct{}),
		called:  make(chan struct{}, 1),
	}
	hasher := new(mockPasswordHasher)
	urlb := new(mockURLBuilder)
	turnstile := new(mockTurnstileValidator)
	logger := new(mockLogger)
	config := new(mockConfig)
	validation := new(mockValidation)

	setupLenientLoggerMock(logger)
	validation.On("SanitizeEmailForLogging", mock.Anything).Return("sanitized-email@example.com").Maybe()
	config.On("GetDefaultMaxViewCount").Return(5).Maybe()

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	turnstile.On("ValidateToken", mock.Anything, "turnstile-token", "").Return(true, nil)
	enc.On("GenerateKey", mock.Anything, int32(32)).Return([]byte("key12345678901234567890123456789"), nil)
	enc.On("Encrypt", mock.Anything, []string{"secret"}, []byte("key12345678901234567890123456789")).Return([]string{"ciphertext"}, nil)
	enc.On("GenerateID", mock.Anything).Return("msg-notify-async", nil)
	stor.On("StoreMessage", mock.Anything, mock.MatchedBy(func(req MessageStorageRequest) bool {
		return req.MessageID == "msg-notify-async" && req.RecipientEmail == "recipient@example.com"
	})).Return(nil)
	urlb.On("BuildDecryptURL", "msg-notify-async", []byte("key12345678901234567890123456789")).Return("https://example.com/decrypt/msg-notify-async")
	notif.On("SendMessageNotification", mock.Anything, mock.Anything).Return(nil)

	done := make(chan struct{})
	var resp *MessageSubmissionResponse
	var err error

	go func() {
		resp, err = svc.SubmitMessage(context.Background(), MessageSubmissionRequest{
			Content:          "secret",
			SendNotification: true,
			TurnstileToken:   "turnstile-token",
			SenderName:       "Sender",
			SenderEmail:      "sender@example.com",
			RecipientName:    "Recipient",
			RecipientEmail:   "recipient@example.com",
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(notificationAsyncReturnThreshold):
		t.Fatalf(
			"SubmitMessage should return within %v even when notification publish is blocked",
			notificationAsyncReturnThreshold,
		)
	}

	close(notif.release)

	select {
	case <-notif.called:
	case <-time.After(1 * time.Second):
		t.Fatal("expected notification send to be invoked")
	}

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "msg-notify-async", resp.MessageID)

	turnstile.AssertExpectations(t)
	enc.AssertExpectations(t)
	stor.AssertExpectations(t)
	notif.AssertExpectations(t)
}
