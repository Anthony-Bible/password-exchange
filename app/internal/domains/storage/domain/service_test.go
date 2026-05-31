package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/logtest"
)

// MockMessageRepository is a hand-written mock implementing secondary.MessageRepository.
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) InsertMessage(message *contracts.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockMessageRepository) SelectMessageByUniqueID(uniqueID string) (*contracts.Message, error) {
	args := m.Called(uniqueID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contracts.Message), args.Error(1)
}

func (m *MockMessageRepository) IncrementViewCountAndGet(uniqueID string) (*contracts.Message, error) {
	args := m.Called(uniqueID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contracts.Message), args.Error(1)
}

func (m *MockMessageRepository) DeleteExpiredMessages() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMessageRepository) GetMessage(uniqueID string) (*contracts.Message, error) {
	args := m.Called(uniqueID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contracts.Message), args.Error(1)
}

func (m *MockMessageRepository) GetUnviewedMessagesForReminders(
	olderThanHours, maxReminders, reminderIntervalHours int,
) ([]*contracts.UnviewedMessage, error) {
	args := m.Called(olderThanHours, maxReminders, reminderIntervalHours)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contracts.UnviewedMessage), args.Error(1)
}

func (m *MockMessageRepository) LogReminderSent(messageID int, emailAddress string) error {
	args := m.Called(messageID, emailAddress)
	return args.Error(0)
}

func (m *MockMessageRepository) GetReminderHistory(messageID int) ([]*contracts.ReminderLogEntry, error) {
	args := m.Called(messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contracts.ReminderLogEntry), args.Error(1)
}

func (m *MockMessageRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMessageRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMessageRepository) CreateUploadSession(ctx context.Context, session contracts.UploadSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockMessageRepository) GetUploadSession(ctx context.Context, id string) (*contracts.UploadSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contracts.UploadSession), args.Error(1)
}

func (m *MockMessageRepository) AddCompletedPart(ctx context.Context, sessionID string, part contracts.UploadSessionPart) error {
	args := m.Called(ctx, sessionID, part)
	return args.Error(0)
}

func (m *MockMessageRepository) CompleteUploadSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockMessageRepository) DeleteUploadSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockMessageRepository) DeleteExpiredUploadSessions(ctx context.Context, asOf time.Time) ([]contracts.UploadSession, error) {
	args := m.Called(ctx, asOf)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]contracts.UploadSession), args.Error(1)
}

// MockValidationPort is a hand-written mock implementing secondary.ValidationPort.
type MockValidationPort struct {
	mock.Mock
}

func (m *MockValidationPort) ValidateEmail(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockValidationPort) SanitizeEmailForLogging(email string) string {
	args := m.Called(email)
	return args.String(0)
}

// newServiceWithMocks creates the trio of mocks plus a wired StorageService.
// It is a convenience for tests that don't need to manipulate the mocks before
// constructing the service. The logger is a logtest.Recorder, so tests that
// care can inspect which severity was emitted.
func newServiceWithMocks(t *testing.T) (
	*StorageService,
	*MockMessageRepository,
	*logtest.Recorder,
	*MockValidationPort,
) {
	t.Helper()
	repo := &MockMessageRepository{}
	logger := logtest.NewRecorder()
	validation := &MockValidationPort{}
	svc := NewStorageService(repo, logger, validation)
	require.NotNil(t, svc)
	return svc, repo, logger, validation
}

func TestNewStorageService_InjectsAllPorts(t *testing.T) {
	repo := &MockMessageRepository{}
	logger := logtest.NewRecorder()
	validation := &MockValidationPort{}

	svc := NewStorageService(repo, logger, validation)

	require.NotNil(t, svc)
	var _ secondary.MessageRepository = repo
	var _ secondary.LoggerPort = logger
	var _ secondary.ValidationPort = validation
	assert.Same(t, repo, svc.repository, "repository should be injected")
	assert.Same(t, logger, svc.logger, "logger should be injected")
	assert.Same(t, validation, svc.validation, "validation should be injected")
}

func TestStoreMessage_RejectsNilMessage(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	err := svc.StoreMessage(context.Background(), nil)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNilMessage)
	repo.AssertNotCalled(t, "InsertMessage", mock.Anything)
}

// TestStoreMessage_NilMessage_LogsAtWarnNotInfo locks in the value of the
// logtest.Recorder migration. The previous setupLenientLoggerMock handed back a
// single shared event for every severity, so a bug downgrading this Warn to
// Info (or promoting it to Error) would have passed unnoticed. The recorder
// captures the level, so the test can assert it precisely.
func TestStoreMessage_NilMessage_LogsAtWarnNotInfo(t *testing.T) {
	svc, _, logger, _ := newServiceWithMocks(t)

	_ = svc.StoreMessage(context.Background(), nil)

	if got := logger.Count("warn"); got != 1 {
		t.Errorf("expected exactly 1 warn-level log, got %d: %+v", got, logger.Entries())
	}
	if got := logger.Count("info"); got != 0 {
		t.Errorf("expected no info-level logs, got %d: %v", got, logger.Messages("info"))
	}
	if msgs := logger.Messages("warn"); len(msgs) != 1 || msgs[0] != "Attempted to store nil message" {
		t.Errorf("unexpected warn messages: %v", msgs)
	}
}

func TestStoreMessage_RejectsEmptyContent(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	err := svc.StoreMessage(context.Background(), &contracts.Message{
		Content:      "",
		UniqueID:     "abc",
		MaxViewCount: 1,
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyContent)
	repo.AssertNotCalled(t, "InsertMessage", mock.Anything)
}

func TestStoreMessage_RejectsEmptyUniqueID(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	err := svc.StoreMessage(context.Background(), &contracts.Message{
		Content:      "ciphertext",
		UniqueID:     "",
		MaxViewCount: 1,
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyUniqueID)
	repo.AssertNotCalled(t, "InsertMessage", mock.Anything)
}

func TestStoreMessage_RejectsInvalidMaxViewCount(t *testing.T) {
	tests := []struct {
		name         string
		maxViewCount int
	}{
		{name: "below minimum", maxViewCount: 0},
		{name: "negative", maxViewCount: -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, _, _ := newServiceWithMocks(t)

			err := svc.StoreMessage(context.Background(), &contracts.Message{
				Content:      "ciphertext",
				UniqueID:     "abc",
				MaxViewCount: tc.maxViewCount,
			})

			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidMaxViewCount)
			repo.AssertNotCalled(t, "InsertMessage", mock.Anything)
		})
	}
}

// The upper bound on view count is a policy owned by the message domain
// (message.AbsoluteMaxViewCount). Storage only guards against the structurally
// invalid case of a count below 1, so a high count delegates to the repository.
func TestStoreMessage_AcceptsHighMaxViewCount(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	msg := &contracts.Message{
		Content:      "ciphertext",
		UniqueID:     "abc",
		MaxViewCount: 101,
	}
	repo.On("InsertMessage", msg).Return(nil).Once()

	err := svc.StoreMessage(context.Background(), msg)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestStoreMessage_ValidInputDelegatesToRepository(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	msg := &contracts.Message{
		Content:      "ciphertext",
		UniqueID:     "abc",
		MaxViewCount: 3,
	}
	repo.On("InsertMessage", msg).Return(nil).Once()

	err := svc.StoreMessage(context.Background(), msg)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestRetrieveMessage_RejectsEmptyUniqueID(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	got, err := svc.RetrieveMessage(context.Background(), "")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyUniqueID)
	assert.Nil(t, got)
	repo.AssertNotCalled(t, "IncrementViewCountAndGet", mock.Anything)
}

func TestRetrieveMessage_SurfacesRepositoryError(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	boom := errors.New("boom")
	repo.On("IncrementViewCountAndGet", "abc").Return(nil, boom).Once()

	got, err := svc.RetrieveMessage(context.Background(), "abc")

	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
	assert.Nil(t, got)
	repo.AssertExpectations(t)
}

func TestRetrieveMessage_HappyPath(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	expected := &contracts.Message{UniqueID: "abc", ViewCount: 2, MaxViewCount: 5}
	repo.On("IncrementViewCountAndGet", "abc").Return(expected, nil).Once()

	got, err := svc.RetrieveMessage(context.Background(), "abc")

	require.NoError(t, err)
	assert.Same(t, expected, got)
	repo.AssertExpectations(t)
}

func TestGetMessage_RejectsEmptyUniqueID(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	got, err := svc.GetMessage(context.Background(), "")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyUniqueID)
	assert.Nil(t, got)
	repo.AssertNotCalled(t, "GetMessage", mock.Anything)
}

func TestGetMessage_HappyPath(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	expected := &contracts.Message{UniqueID: "abc", ViewCount: 0}
	repo.On("GetMessage", "abc").Return(expected, nil).Once()

	got, err := svc.GetMessage(context.Background(), "abc")

	require.NoError(t, err)
	assert.Same(t, expected, got)
	repo.AssertExpectations(t)
}

func TestGetMessage_SurfacesRepositoryError(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	boom := errors.New("kaboom")
	repo.On("GetMessage", "abc").Return(nil, boom).Once()

	got, err := svc.GetMessage(context.Background(), "abc")

	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
	assert.Nil(t, got)
	repo.AssertExpectations(t)
}

func TestCleanupExpiredMessages_DelegatesToRepository(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	repo.On("DeleteExpiredMessages").Return(nil).Once()

	err := svc.CleanupExpiredMessages(context.Background())

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetUnviewedMessagesForReminders_RejectsInvalidParams(t *testing.T) {
	tests := []struct {
		name                  string
		olderThanHours        int
		maxReminders          int
		reminderIntervalHours int
	}{
		{name: "zero olderThanHours", olderThanHours: 0, maxReminders: 1, reminderIntervalHours: 1},
		{name: "zero maxReminders", olderThanHours: 1, maxReminders: 0, reminderIntervalHours: 1},
		{name: "zero reminderIntervalHours", olderThanHours: 1, maxReminders: 1, reminderIntervalHours: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, _, _ := newServiceWithMocks(t)

			got, err := svc.GetUnviewedMessagesForReminders(
				context.Background(),
				tc.olderThanHours,
				tc.maxReminders,
				tc.reminderIntervalHours,
			)

			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidParameter)
			assert.Nil(t, got)
			repo.AssertNotCalled(t, "GetUnviewedMessagesForReminders",
				mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestGetUnviewedMessagesForReminders_DelegatesValidInput(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	expected := []*contracts.UnviewedMessage{
		{MessageID: 1, UniqueID: "u1", RecipientEmail: "a@b.com", DaysOld: 3},
	}
	repo.On("GetUnviewedMessagesForReminders", 24, 3, 24).Return(expected, nil).Once()

	got, err := svc.GetUnviewedMessagesForReminders(context.Background(), 24, 3, 24)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

func TestLogReminderSent_RejectsInvalidMessageID(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	err := svc.LogReminderSent(context.Background(), 0, "user@example.com")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidParameter)
	repo.AssertNotCalled(t, "LogReminderSent", mock.Anything, mock.Anything)
}

func TestLogReminderSent_RejectsEmptyEmail(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	err := svc.LogReminderSent(context.Background(), 1, "")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyEmailAddress)
	repo.AssertNotCalled(t, "LogReminderSent", mock.Anything, mock.Anything)
}

func TestLogReminderSent_UsesValidationPortForSanitization(t *testing.T) {
	svc, repo, _, validation := newServiceWithMocks(t)

	repo.On("LogReminderSent", 1, "user@example.com").Return(nil).Once()
	validation.On("SanitizeEmailForLogging", "user@example.com").
		Return("u***@example.com").Maybe()

	err := svc.LogReminderSent(context.Background(), 1, "user@example.com")

	require.NoError(t, err)
	validation.AssertCalled(t, "SanitizeEmailForLogging", "user@example.com")
	repo.AssertExpectations(t)
}

func TestLogReminderSent_HappyPath(t *testing.T) {
	svc, repo, _, validation := newServiceWithMocks(t)

	repo.On("LogReminderSent", 7, "user@example.com").Return(nil).Once()
	validation.On("SanitizeEmailForLogging", mock.Anything).Return("redacted").Maybe()

	err := svc.LogReminderSent(context.Background(), 7, "user@example.com")

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetReminderHistory_RejectsInvalidMessageID(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	got, err := svc.GetReminderHistory(context.Background(), 0)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidParameter)
	assert.Nil(t, got)
	repo.AssertNotCalled(t, "GetReminderHistory", mock.Anything)
}

func TestGetReminderHistory_HappyPath(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	expected := []*contracts.ReminderLogEntry{
		{MessageID: 1, EmailAddress: "a@b.com", ReminderCount: 2, LastReminderSent: time.Now()},
	}
	repo.On("GetReminderHistory", 1).Return(expected, nil).Once()

	got, err := svc.GetReminderHistory(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

func TestHealthCheck_HappyPathReturnsNil(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)
	repo.On("Ping", mock.Anything).Return(nil).Once()

	err := svc.HealthCheck(context.Background())

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestHealthCheck_SurfacesRepositoryPingError(t *testing.T) {
	svc, repo, _, _ := newServiceWithMocks(t)

	boom := errors.New("ping failed")
	repo.On("Ping", mock.Anything).Return(boom).Once()

	err := svc.HealthCheck(context.Background())

	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
	repo.AssertExpectations(t)
}

// --- Upload session domain service tests ---

func TestCreateUploadSession_DelegatesAndRejectsEmptySessionID(t *testing.T) {
	t.Parallel()
	svc, repo, _, _ := newServiceWithMocks(t)

	err := svc.CreateUploadSession(context.Background(), contracts.UploadSession{SessionID: ""})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidParameter)
	repo.AssertNotCalled(t, "CreateUploadSession")
}

func TestCreateUploadSession_DelegatesOnValidInput(t *testing.T) {
	t.Parallel()
	svc, repo, _, _ := newServiceWithMocks(t)

	sess := contracts.UploadSession{SessionID: "s1", FileID: "f1"}
	repo.On("CreateUploadSession", mock.Anything, sess).Return(nil).Once()

	err := svc.CreateUploadSession(context.Background(), sess)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetUploadSession_RejectsEmptyID(t *testing.T) {
	t.Parallel()
	svc, repo, _, _ := newServiceWithMocks(t)

	_, err := svc.GetUploadSession(context.Background(), "")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidParameter)
	repo.AssertNotCalled(t, "GetUploadSession")
}

func TestGetUploadSession_DelegatesOnValidID(t *testing.T) {
	t.Parallel()
	svc, repo, _, _ := newServiceWithMocks(t)

	want := &contracts.UploadSession{SessionID: "s1", FileID: "f1"}
	repo.On("GetUploadSession", mock.Anything, "s1").Return(want, nil).Once()

	got, err := svc.GetUploadSession(context.Background(), "s1")
	require.NoError(t, err)
	assert.Equal(t, want, got)
	repo.AssertExpectations(t)
}

func TestAddCompletedPart_RejectsEmptySessionID(t *testing.T) {
	t.Parallel()
	svc, repo, _, _ := newServiceWithMocks(t)

	err := svc.AddCompletedPart(context.Background(), "", contracts.UploadSessionPart{PartNumber: 1})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidParameter)
	repo.AssertNotCalled(t, "AddCompletedPart")
}

func TestAddCompletedPart_DelegatesOnValidInput(t *testing.T) {
	t.Parallel()
	svc, repo, _, _ := newServiceWithMocks(t)

	part := contracts.UploadSessionPart{PartNumber: 1, ETag: "etag"}
	repo.On("AddCompletedPart", mock.Anything, "s1", part).Return(nil).Once()

	err := svc.AddCompletedPart(context.Background(), "s1", part)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCompleteUploadSession_RejectsEmptySessionID(t *testing.T) {
	t.Parallel()
	svc, repo, _, _ := newServiceWithMocks(t)

	err := svc.CompleteUploadSession(context.Background(), "")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidParameter)
	repo.AssertNotCalled(t, "CompleteUploadSession")
}

func TestCompleteUploadSession_DelegatesOnValidInput(t *testing.T) {
	t.Parallel()
	svc, repo, _, _ := newServiceWithMocks(t)

	repo.On("CompleteUploadSession", mock.Anything, "s1").Return(nil).Once()

	err := svc.CompleteUploadSession(context.Background(), "s1")
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteUploadSession_RejectsEmptySessionID(t *testing.T) {
	t.Parallel()
	svc, repo, _, _ := newServiceWithMocks(t)

	err := svc.DeleteUploadSession(context.Background(), "")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidParameter)
	repo.AssertNotCalled(t, "DeleteUploadSession")
}

func TestDeleteUploadSession_DelegatesOnValidInput(t *testing.T) {
	t.Parallel()
	svc, repo, _, _ := newServiceWithMocks(t)

	repo.On("DeleteUploadSession", mock.Anything, "s1").Return(nil).Once()

	err := svc.DeleteUploadSession(context.Background(), "s1")
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteExpiredUploadSessions_DelegatesAsOf(t *testing.T) {
	t.Parallel()
	svc, repo, _, _ := newServiceWithMocks(t)

	asOf := time.Now().UTC()
	want := []contracts.UploadSession{{SessionID: "s1"}}
	repo.On("DeleteExpiredUploadSessions", mock.Anything, asOf).Return(want, nil).Once()

	got, err := svc.DeleteExpiredUploadSessions(context.Background(), asOf)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	repo.AssertExpectations(t)
}
