package grpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// reminder_grpc_integration_test.go exercises the reminder RPCs end-to-end across
// a real gRPC transport: a genuine DbServiceServer (GRPCServer) is registered on
// an in-memory bufconn listener and driven by a genuine DbServiceClient. Unlike
// the unit tests, this verifies the full request → server → service-port →
// response path including protobuf<->Go translation and gRPC error propagation.
// The storage service is a configurable stub so the test stays fast and
// hermetic while still covering the wire contract reminder workers depend on.

// newReminderTestClient stands up the storage gRPC server over bufconn backed by
// the supplied stub and returns a connected DbServiceClient. All resources are
// torn down via t.Cleanup.
func newReminderTestClient(t *testing.T, svc *stubStorageService) database.DbServiceClient {
	t.Helper()

	adapter := NewGRPCServer(svc, "", &recordingLogger{}, &stubValidator{})

	lis := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	database.RegisterDbServiceServer(grpcServer, adapter)

	serveErrCh := make(chan error, 1)
	go func() { serveErrCh <- grpcServer.Serve(lis) }()
	t.Cleanup(func() {
		grpcServer.Stop()
		_ = lis.Close()
		if err := <-serveErrCh; err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("gRPC Serve returned unexpected error: %v", err)
		}
	})

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return database.NewDbServiceClient(conn)
}

// TestGRPC_GetUnviewedMessagesForReminders_RoundTrip verifies that request
// parameters reach the storage service intact and that domain results are
// translated into the protobuf response (including the "2006-01-02 15:04:05"
// timestamp format and int32 narrowing) that reminder workers parse.
func TestGRPC_GetUnviewedMessagesForReminders_RoundTrip(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 5, 20, 13, 45, 30, 0, time.UTC)
	svc := &stubStorageService{
		unviewedMessages: []*contracts.UnviewedMessage{
			{MessageID: 7, UniqueID: "uuid-7", RecipientEmail: "a@example.com", Created: created, DaysOld: 9},
			{MessageID: 8, UniqueID: "uuid-8", RecipientEmail: "b@example.com", Created: created, DaysOld: 9},
		},
	}
	client := newReminderTestClient(t, svc)

	resp, err := client.GetUnviewedMessagesForReminders(context.Background(), &database.GetUnviewedMessagesRequest{
		OlderThanHours:        24,
		MaxReminders:          3,
		ReminderIntervalHours: 12,
	})
	require.NoError(t, err)

	// Parameters survived the protobuf decode on the server side.
	assert.Equal(t, 24, svc.gotOlderThanHours)
	assert.Equal(t, 3, svc.gotMaxReminders)
	assert.Equal(t, 12, svc.gotIntervalHours)

	require.Len(t, resp.GetMessages(), 2)
	first := resp.GetMessages()[0]
	assert.Equal(t, int32(7), first.GetMessageId())
	assert.Equal(t, "uuid-7", first.GetUniqueId())
	assert.Equal(t, "a@example.com", first.GetRecipientEmail())
	assert.Equal(t, int32(9), first.GetDaysOld())
	assert.Equal(t, "2026-05-20 13:45:30", first.GetCreated())
}

// TestGRPC_GetUnviewedMessagesForReminders_Empty verifies the empty result is
// returned as an empty list rather than an error, so a quiet reminder run is a
// no-op for the caller.
func TestGRPC_GetUnviewedMessagesForReminders_Empty(t *testing.T) {
	t.Parallel()

	client := newReminderTestClient(t, &stubStorageService{})

	resp, err := client.GetUnviewedMessagesForReminders(context.Background(), &database.GetUnviewedMessagesRequest{
		OlderThanHours: 24, MaxReminders: 3, ReminderIntervalHours: 12,
	})
	require.NoError(t, err)
	assert.Empty(t, resp.GetMessages())
}

// TestGRPC_GetUnviewedMessagesForReminders_Error verifies a storage error
// surfaces to the client as a non-nil gRPC error instead of a silent empty
// response, so reminder workers can retry rather than treat an outage as "no
// messages".
func TestGRPC_GetUnviewedMessagesForReminders_Error(t *testing.T) {
	t.Parallel()

	svc := &stubStorageService{unviewedErr: errors.New("db unavailable")}
	client := newReminderTestClient(t, svc)

	_, err := client.GetUnviewedMessagesForReminders(context.Background(), &database.GetUnviewedMessagesRequest{
		OlderThanHours: 24, MaxReminders: 3, ReminderIntervalHours: 12,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db unavailable")
}

// TestGRPC_LogReminderSent_RoundTrip verifies the message ID and email reach the
// storage service unchanged across the wire.
func TestGRPC_LogReminderSent_RoundTrip(t *testing.T) {
	t.Parallel()

	svc := &stubStorageService{}
	client := newReminderTestClient(t, svc)

	_, err := client.LogReminderSent(context.Background(), &database.LogReminderRequest{
		MessageId:    42,
		EmailAddress: "recipient@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, 42, svc.gotLogMessageID)
	assert.Equal(t, "recipient@example.com", svc.gotLogEmail)
}

// TestGRPC_LogReminderSent_Error verifies a storage failure to persist the
// reminder log surfaces as a gRPC error.
func TestGRPC_LogReminderSent_Error(t *testing.T) {
	t.Parallel()

	svc := &stubStorageService{logReminderErr: errors.New("insert failed")}
	client := newReminderTestClient(t, svc)

	_, err := client.LogReminderSent(context.Background(), &database.LogReminderRequest{
		MessageId: 1, EmailAddress: "x@example.com",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert failed")
}

// TestGRPC_GetReminderHistory_RoundTrip verifies reminder history entries are
// translated into the protobuf response, including the reminder count and the
// formatted last-sent timestamp.
func TestGRPC_GetReminderHistory_RoundTrip(t *testing.T) {
	t.Parallel()

	lastSent := time.Date(2026, 5, 25, 8, 0, 0, 0, time.UTC)
	svc := &stubStorageService{
		reminderHistory: []*contracts.ReminderLogEntry{
			{MessageID: 5, EmailAddress: "c@example.com", ReminderCount: 2, LastReminderSent: lastSent},
		},
	}
	client := newReminderTestClient(t, svc)

	resp, err := client.GetReminderHistory(context.Background(), &database.GetReminderHistoryRequest{MessageId: 5})
	require.NoError(t, err)
	assert.Equal(t, 5, svc.gotHistoryMessageID)

	require.Len(t, resp.GetEntries(), 1)
	entry := resp.GetEntries()[0]
	assert.Equal(t, int32(5), entry.GetMessageId())
	assert.Equal(t, "c@example.com", entry.GetEmailAddress())
	assert.Equal(t, int32(2), entry.GetReminderCount())
	assert.Equal(t, "2026-05-25 08:00:00", entry.GetLastReminderSent())
}

// TestGRPC_GetReminderHistory_Empty verifies a message with no reminder history
// yields an empty list, not an error.
func TestGRPC_GetReminderHistory_Empty(t *testing.T) {
	t.Parallel()

	client := newReminderTestClient(t, &stubStorageService{})

	resp, err := client.GetReminderHistory(context.Background(), &database.GetReminderHistoryRequest{MessageId: 99})
	require.NoError(t, err)
	assert.Empty(t, resp.GetEntries())
}

// TestGRPC_GetReminderHistory_Error verifies a storage error surfaces to the client.
func TestGRPC_GetReminderHistory_Error(t *testing.T) {
	t.Parallel()

	svc := &stubStorageService{reminderHistErr: errors.New("query failed")}
	client := newReminderTestClient(t, svc)

	_, err := client.GetReminderHistory(context.Background(), &database.GetReminderHistoryRequest{MessageId: 5})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
}
