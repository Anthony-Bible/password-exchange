//go:build integration

// Postgres twin of reminder_flow_integration_test.go: the complete reminder
// pipeline (ReminderService -> notification storage adapter -> StorageService
// -> PostgresAdapter -> Postgres 17). Full duplication of the MySQL file is
// intentional; #570 deletes the MySQL side.
package reminder_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sharedConfig "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/shared"
	notificationStorage "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/storage"
	notificationDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/postgres"
	storageDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	storageContracts "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/integration/dbtest"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPostgresPipeline(t *testing.T, dbCfg storageContracts.DatabaseConfig) (*notificationDomain.ReminderService, *capturingPublisher) {
	t.Helper()
	repo := postgres.NewPostgresAdapter(dbCfg, logging.NewLogger(), validation.NewAdapter())
	t.Cleanup(func() { _ = repo.Close() })
	storageService := storageDomain.NewStorageService(repo, logging.NewLogger(), validation.NewAdapter())
	publisher := &capturingPublisher{}
	configPort := sharedConfig.NewSharedConfigAdapter(config.PassConfig{EmailFrom: "server@password.exchange"})
	svc := notificationDomain.NewReminderService(
		notificationStorage.NewGRPCStorageAdapter(storageService),
		publisher,
		logging.NewLogger(),
		configPort,
		validation.NewAdapter(),
	)
	return svc, publisher
}

func seedMessagePG(t *testing.T, db *sql.DB, uniqueID, email string, viewCount, ageHours int) int {
	t.Helper()
	var id int
	require.NoError(t, db.QueryRow(
		`INSERT INTO messages (message, uniqueid, other_email, view_count, max_view_count, created, expires_at)
		 VALUES ('ciphertext', $1, $2, $3, 5, now() - make_interval(hours => $4::int), now() + interval '7 days')
		 RETURNING messageid`,
		uniqueID, email, viewCount, ageHours).Scan(&id))
	return id
}

// reminderRowPG returns (count, lastSent) for the message, or (0, zero) when none was logged.
func reminderRowPG(t *testing.T, db *sql.DB, uniqueID string) (int, time.Time) {
	t.Helper()
	var count int
	var lastSent time.Time
	err := db.QueryRow(
		`SELECT er.reminder_count, er.last_reminder_sent
		 FROM email_reminders er JOIN messages m ON m.messageid = er.message_id
		 WHERE m.uniqueid = $1`, uniqueID).Scan(&count, &lastSent)
	if err == sql.ErrNoRows {
		return 0, time.Time{}
	}
	require.NoError(t, err)
	return count, lastSent
}

func TestIntegrationPG_ReminderPipeline_EndToEnd(t *testing.T) {
	db, dbCfg := dbtest.StartPostgresBaseline(t)
	seedMessagePG(t, db, "pipeline-eligible", "wanted@example.com", 0, 48)
	seedMessagePG(t, db, "pipeline-too-recent", "skip@example.com", 0, 1)

	svc, publisher := newPostgresPipeline(t, dbCfg)
	cfg := notificationDomain.ReminderConfig{Enabled: true, CheckAfterHours: 24, MaxReminders: 3, Interval: 24}

	require.NoError(t, svc.ProcessReminders(context.Background(), cfg))

	recipients := publisher.recipients()
	assert.True(t, recipients["wanted@example.com"], "eligible recipient should receive a reminder")
	assert.False(t, recipients["skip@example.com"], "too-recent message should not be reminded")

	count, lastSent := reminderRowPG(t, db, "pipeline-eligible")
	assert.Equal(t, 1, count)
	assert.WithinDuration(t, time.Now(), lastSent, time.Minute)
	skipped, _ := reminderRowPG(t, db, "pipeline-too-recent")
	assert.Equal(t, 0, skipped)
}

func TestIntegrationPG_ReminderPipeline_IncrementsOnSecondRunAndNeverDuplicatesWithinInterval(t *testing.T) {
	db, dbCfg := dbtest.StartPostgresBaseline(t)
	seedMessagePG(t, db, "pipeline-repeat", "repeat@example.com", 0, 100)

	svc, publisher := newPostgresPipeline(t, dbCfg)
	cfg := notificationDomain.ReminderConfig{Enabled: true, CheckAfterHours: 24, MaxReminders: 3, Interval: 1}

	require.NoError(t, svc.ProcessReminders(context.Background(), cfg))
	count, _ := reminderRowPG(t, db, "pipeline-repeat")
	require.Equal(t, 1, count)

	// Immediately re-running must not re-send: the 1h interval has not elapsed.
	require.NoError(t, svc.ProcessReminders(context.Background(), cfg))
	count, _ = reminderRowPG(t, db, "pipeline-repeat")
	assert.Equal(t, 1, count, "no duplicate reminder within the interval")

	_, err := db.Exec(`UPDATE email_reminders er SET last_reminder_sent = now() - interval '5 hours'
		FROM messages m WHERE m.messageid = er.message_id AND m.uniqueid = $1`, "pipeline-repeat")
	require.NoError(t, err)

	require.NoError(t, svc.ProcessReminders(context.Background(), cfg))
	count, _ = reminderRowPG(t, db, "pipeline-repeat")
	assert.Equal(t, 2, count, "second eligible run should increment the reminder count")
	assert.True(t, publisher.recipients()["repeat@example.com"])
}
