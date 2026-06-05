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

func (m *mockURLBuilder) BuildE2EDecryptURL(messageID string, e2eKeyBase64Url string) string {
	args := m.Called(messageID, e2eKeyBase64Url)
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

// setupLenientLoggerMock sets up lenient expectations for a logger mock that will match any log calls.
func setupLenientLoggerMock(l *mockLogger) {
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
}

// setupTestMocks sets up standard expectations for all mocks.
func setupTestMocks(
	_ *mockEncryptionService,
	_ *mockStorageService,
	_ *mockNotificationService,
	_ *mockPasswordHasher,
	_ *mockURLBuilder,
	_ *mockTurnstileValidator,
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
	assert.False(t, resp.ExpiresAt.Before(expectedMin), "ExpiresAt should be >= 48h from before")
	assert.False(t, resp.ExpiresAt.After(expectedMax), "ExpiresAt should be <= 48h from after")

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

func TestSubmitMessage_ClientEncrypted_UsesE2EURLAndSkipsServerEncryption(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	req := MessageSubmissionRequest{
		Content:           "client-side-ciphertext",
		IsClientEncrypted: true,
	}

	enc.On("GenerateID", mock.Anything).Return("msg-client-e2e", nil)

	stor.On("StoreMessage", mock.Anything, mock.MatchedBy(func(storeReq MessageStorageRequest) bool {
		return storeReq.IsClientEncrypted && storeReq.Content == "client-side-ciphertext"
	})).Return(nil)

	urlb.On("BuildE2EDecryptURL", "msg-client-e2e", "").Return("https://example.com/decrypt/msg-client-e2e").Maybe()

	resp, err := svc.SubmitMessage(context.Background(), req)
	assert.NoError(t, err)
	assert.Empty(t, resp.Key, "key must be empty for client-side encrypted messages")
	urlb.AssertCalled(t, "BuildE2EDecryptURL", "msg-client-e2e", "")
	urlb.AssertNotCalled(t, "BuildDecryptURL", "msg-client-e2e", mock.Anything)
	enc.AssertNotCalled(t, "GenerateKey", mock.Anything, int32(32))
	enc.AssertNotCalled(t, "Encrypt", mock.Anything, []string{"client-side-ciphertext"}, mock.Anything)
}

func TestRetrieveMessage_ClientEncrypted_ReturnsCiphertextWithoutServerDecrypt(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	storageResp := &MessageStorageResponse{
		MessageID:         "msg-client-cipher",
		EncryptedContent:  "client-side-ciphertext",
		IsClientEncrypted: true,
		HasPassphrase:     false,
		ViewCount:         1,
		MaxViewCount:      5,
	}

	stor.On("GetMessage", mock.Anything, MessageRetrievalStorageRequest{MessageID: "msg-client-cipher"}).
		Return(storageResp, nil)
	stor.On("RetrieveMessage", mock.Anything, MessageRetrievalStorageRequest{MessageID: "msg-client-cipher"}).
		Return(storageResp, nil)
	enc.On("Decrypt", mock.Anything, []string{"client-side-ciphertext"}, mock.Anything).
		Return([]string{"unexpected-server-decrypt"}, nil)

	resp, err := svc.RetrieveMessage(context.Background(), MessageRetrievalRequest{
		MessageID: "msg-client-cipher",
	})
	assert.NoError(t, err)
	assert.Equal(t, "client-side-ciphertext", resp.Content)
	enc.AssertNotCalled(t, "Decrypt", mock.Anything, []string{"client-side-ciphertext"}, mock.Anything)
}

func TestNotifyMessage_Success(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	storageResp := &MessageStorageResponse{MessageID: "msg-notify-1"}
	stor.On("GetMessage", mock.Anything, MessageRetrievalStorageRequest{MessageID: "msg-notify-1"}).
		Return(storageResp, nil)
	turnstile.On("ValidateToken", mock.Anything, "valid-token", "").Return(true, nil)
	notif.On("SendMessageNotification", mock.Anything, mock.MatchedBy(func(r MessageNotificationRequest) bool {
		return r.MessageURL == "https://example.com/decrypt/msg-notify-1#key=abc&fid=f1&fk=k1" &&
			r.RecipientEmail == "recipient@example.com"
	})).Return(nil)

	err := svc.NotifyMessage(context.Background(), MessageNotifyRequest{
		MessageID:      "msg-notify-1",
		ShareURL:       "https://example.com/decrypt/msg-notify-1#key=abc&fid=f1&fk=k1",
		SenderName:     "Alice",
		SenderEmail:    "alice@example.com",
		RecipientName:  "Bob",
		RecipientEmail: "recipient@example.com",
		TurnstileToken: "valid-token",
	})

	assert.NoError(t, err)
	notif.AssertExpectations(t)
	turnstile.AssertExpectations(t)
}

func TestNotifyMessage_MissingShareURL(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	err := svc.NotifyMessage(context.Background(), MessageNotifyRequest{
		MessageID:      "msg-notify-2",
		RecipientEmail: "recipient@example.com",
		TurnstileToken: "valid-token",
	})

	assert.ErrorIs(t, err, ErrInvalidMessageRequest)
	notif.AssertNotCalled(t, "SendMessageNotification", mock.Anything, mock.Anything)
}

func TestNotifyMessage_MissingRecipientEmail(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	err := svc.NotifyMessage(context.Background(), MessageNotifyRequest{
		MessageID:      "msg-notify-3",
		ShareURL:       "https://example.com/decrypt/msg-notify-3",
		TurnstileToken: "valid-token",
	})

	assert.ErrorIs(t, err, ErrInvalidMessageRequest)
	notif.AssertNotCalled(t, "SendMessageNotification", mock.Anything, mock.Anything)
}

func TestNotifyMessage_MissingTurnstileToken(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	err := svc.NotifyMessage(context.Background(), MessageNotifyRequest{
		MessageID:      "msg-notify-4",
		ShareURL:       "https://example.com/decrypt/msg-notify-4",
		RecipientEmail: "recipient@example.com",
	})

	assert.ErrorIs(t, err, ErrInvalidMessageRequest)
	notif.AssertNotCalled(t, "SendMessageNotification", mock.Anything, mock.Anything)
}

func TestNotifyMessage_InvalidTurnstileToken(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	turnstile.On("ValidateToken", mock.Anything, "bad-token", "").Return(false, nil)

	err := svc.NotifyMessage(context.Background(), MessageNotifyRequest{
		MessageID:      "msg-notify-5",
		ShareURL:       "https://example.com/decrypt/msg-notify-5",
		RecipientEmail: "recipient@example.com",
		TurnstileToken: "bad-token",
	})

	assert.ErrorIs(t, err, ErrInvalidMessageRequest)
	notif.AssertNotCalled(t, "SendMessageNotification", mock.Anything, mock.Anything)
}

// TestGetDefaultMaxViewCount_ReturnsConfigValue verifies that the service
// delegates to the config port and returns whatever value it provides.
func TestGetDefaultMaxViewCount_ReturnsConfigValue(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupLenientLoggerMock(logger)
	validation.On("SanitizeEmailForLogging", mock.Anything).Return("sanitized-email@example.com").Maybe()
	config.On("GetDefaultMaxViewCount").Return(12)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	got := svc.GetDefaultMaxViewCount()
	assert.Equal(t, 12, got, "GetDefaultMaxViewCount must return the value from the config port")
	config.AssertCalled(t, "GetDefaultMaxViewCount")
}

// TestGetDefaultMaxViewCount_FallsBackToConstantWhenConfigReturnsZero verifies
// that when the config port returns a non-positive value the service falls back
// to the DefaultMaxViewCount domain constant (5) rather than propagating an
// invalid zero or negative value to callers.
func TestGetDefaultMaxViewCount_FallsBackToConstantWhenConfigReturnsZero(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupLenientLoggerMock(logger)
	validation.On("SanitizeEmailForLogging", mock.Anything).Return("sanitized-email@example.com").Maybe()
	config.On("GetDefaultMaxViewCount").Return(0)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	got := svc.GetDefaultMaxViewCount()
	assert.Equal(t, DefaultMaxViewCount, got,
		"GetDefaultMaxViewCount must fall back to DefaultMaxViewCount constant when config returns 0")
}

// TestGetDefaultMaxViewCount_FallsBackToConstantWhenConfigReturnsNegative
// is the negative-input variant of the fallback contract.
func TestGetDefaultMaxViewCount_FallsBackToConstantWhenConfigReturnsNegative(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupLenientLoggerMock(logger)
	validation.On("SanitizeEmailForLogging", mock.Anything).Return("sanitized-email@example.com").Maybe()
	config.On("GetDefaultMaxViewCount").Return(-3)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	got := svc.GetDefaultMaxViewCount()
	assert.Equal(t, DefaultMaxViewCount, got,
		"GetDefaultMaxViewCount must fall back to DefaultMaxViewCount constant when config returns a negative value")
}

func TestNotifyMessage_MessageNotFound(t *testing.T) {
	enc, stor, notif, hasher, urlb, turnstile, logger, config, validation := createTestMocks()
	setupTestMocks(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile, logger, config, validation)

	turnstile.On("ValidateToken", mock.Anything, "valid-token", "").Return(true, nil)
	stor.On("GetMessage", mock.Anything, MessageRetrievalStorageRequest{MessageID: "missing-msg"}).
		Return((*MessageStorageResponse)(nil), ErrMessageNotFound)

	err := svc.NotifyMessage(context.Background(), MessageNotifyRequest{
		MessageID:      "missing-msg",
		ShareURL:       "https://example.com/decrypt/missing-msg",
		RecipientEmail: "recipient@example.com",
		TurnstileToken: "valid-token",
	})

	assert.ErrorIs(t, err, ErrMessageNotFound)
	notif.AssertNotCalled(t, "SendMessageNotification", mock.Anything, mock.Anything)
}
