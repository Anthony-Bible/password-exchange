package domain

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

type mockNotificationService struct{ mock.Mock }

func (m *mockNotificationService) SendMessageNotification(ctx context.Context, req MessageNotificationRequest) error {
	args := m.Called(ctx, req)
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

// --- Tests ---

func TestRetrieveMessage_PropagatesExpiresAt(t *testing.T) {
	// RetrieveMessage must carry ExpiresAt from storage through to the response.
	enc := new(mockEncryptionService)
	stor := new(mockStorageService)
	notif := new(mockNotificationService)
	hasher := new(mockPasswordHasher)
	urlb := new(mockURLBuilder)
	turnstile := new(mockTurnstileValidator)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile)

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
	// Legacy messages with nil ExpiresAt in storage must yield nil ExpiresAt in the response.
	enc := new(mockEncryptionService)
	stor := new(mockStorageService)
	notif := new(mockNotificationService)
	hasher := new(mockPasswordHasher)
	urlb := new(mockURLBuilder)
	turnstile := new(mockTurnstileValidator)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile)

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
	// SubmitMessage must use the provided ExpirationHours when set.
	enc := new(mockEncryptionService)
	stor := new(mockStorageService)
	notif := new(mockNotificationService)
	hasher := new(mockPasswordHasher)
	urlb := new(mockURLBuilder)
	turnstile := new(mockTurnstileValidator)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile)

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
	// When ExpirationHours is zero, the default TTL (7 days) must be applied.
	enc := new(mockEncryptionService)
	stor := new(mockStorageService)
	notif := new(mockNotificationService)
	hasher := new(mockPasswordHasher)
	urlb := new(mockURLBuilder)
	turnstile := new(mockTurnstileValidator)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile)

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
	// ExpirationHours outside the allowed range must be rejected.
	enc := new(mockEncryptionService)
	stor := new(mockStorageService)
	notif := new(mockNotificationService)
	hasher := new(mockPasswordHasher)
	urlb := new(mockURLBuilder)
	turnstile := new(mockTurnstileValidator)

	svc := NewMessageService(enc, stor, notif, hasher, urlb, turnstile)

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
