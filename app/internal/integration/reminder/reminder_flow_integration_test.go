//go:build integration

// Package reminder_test exercises the complete reminder pipeline against a real
// database: notification ReminderService -> notification GRPCStorageAdapter ->
// storage StorageService -> MySQLAdapter -> MySQL. This mirrors exactly how
// cmd/reminder wires the system at runtime, so it verifies that configuration
// flows through, that eligible messages are selected by the real SQL, that
// notifications are published, and that reminder state is persisted (and
// incremented on subsequent runs).
//
// Gated behind the `integration` build tag; requires Docker.
package reminder_test

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	notificationLogger "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/logger"
	sharedConfig "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/shared"
	notificationStorage "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/storage"
	notificationValidator "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/validator"
	notificationDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	notificationContracts "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
	storageLogger "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/logger"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/mysql"
	storageValidator "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/validator"
	storageDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	storageContracts "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/integration/dbtest"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capturingPublisher records the notifications the reminder service publishes,
// standing in for the RabbitMQ publisher used in production.
type capturingPublisher struct {
	mu       sync.Mutex
	captured []notificationContracts.NotificationRequest
}

func (p *capturingPublisher) PublishNotification(_ context.Context, req notificationContracts.NotificationRequest) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.captured = append(p.captured, req)
	return nil
}

func (p *capturingPublisher) recipients() map[string]bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := map[string]bool{}
	for _, r := range p.captured {
		out[r.To] = true
	}
	return out
}

// newPipeline wires the reminder pipeline exactly like cmd/reminder.go, using
// the real adapters but a capturing publisher in place of RabbitMQ.
func newPipeline(dbCfg storageContracts.DatabaseConfig) (*notificationDomain.ReminderService, *capturingPublisher) {
	storageRepo := mysql.NewMySQLAdapter(dbCfg, storageLogger.NewAdapter(), storageValidator.NewValidationAdapter())
	storageService := storageDomain.NewStorageService(storageRepo, storageLogger.NewAdapter(), storageValidator.NewValidationAdapter())
	notifStorageAdapter := notificationStorage.NewGRPCStorageAdapter(storageService)
	publisher := &capturingPublisher{}
	configPort := sharedConfig.NewSharedConfigAdapter(config.PassConfig{EmailFrom: "server@password.exchange"})

	svc := notificationDomain.NewReminderService(
		notifStorageAdapter,
		publisher,
		notificationLogger.NewAdapter(),
		configPort,
		notificationValidator.NewValidationAdapter(),
	)
	return svc, publisher
}

// seedMessage inserts a message with an explicit age (hours in the past) and
// returns its generated messageid.
func seedMessage(t *testing.T, db *sql.DB, uniqueID, email string, viewCount, ageHours int) int {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO messages (message, uniqueid, other_email, view_count, max_view_count, created, expires_at)
		 VALUES (?, ?, ?, ?, 5, NOW() - INTERVAL ? HOUR, NOW() + INTERVAL 7 DAY)`,
		"ciphertext", uniqueID, email, viewCount, ageHours,
	)
	require.NoError(t, err)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return int(id)
}

// reminderRow returns the reminder_count and last_reminder_sent for the message
// with the given uniqueid, or (0, "") if no reminder has been logged.
func reminderRow(t *testing.T, db *sql.DB, uniqueID string) (int, string) {
	t.Helper()
	var count int
	var lastSent string
	err := db.QueryRow(
		`SELECT er.reminder_count, er.last_reminder_sent
		 FROM email_reminders er JOIN messages m ON m.messageid = er.message_id
		 WHERE m.uniqueid = ?`, uniqueID,
	).Scan(&count, &lastSent)
	if err == sql.ErrNoRows {
		return 0, ""
	}
	require.NoError(t, err)
	return count, lastSent
}

func TestIntegration_ReminderPipeline_EndToEnd(t *testing.T) {
	db, dbCfg := dbtest.StartMySQL(t)

	// Seed: one eligible message and one too-recent message.
	seedMessage(t, db, "pipeline-eligible", "wanted@example.com", 0, 48)
	seedMessage(t, db, "pipeline-too-recent", "skip@example.com", 0, 1)

	svc, publisher := newPipeline(dbCfg)
	reminderCfg := notificationDomain.ReminderConfig{Enabled: true, CheckAfterHours: 24, MaxReminders: 3, Interval: 24}

	// First run: eligible message reminded, too-recent skipped.
	require.NoError(t, svc.ProcessReminders(context.Background(), reminderCfg))

	recipients := publisher.recipients()
	assert.True(t, recipients["wanted@example.com"], "eligible recipient should receive a reminder")
	assert.False(t, recipients["skip@example.com"], "too-recent message should not be reminded")

	// The reminder must be persisted in email_reminders with count 1.
	count, lastSent := reminderRow(t, db, "pipeline-eligible")
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, lastSent)

	// The too-recent message must have no reminder row.
	skippedCount, _ := reminderRow(t, db, "pipeline-too-recent")
	assert.Equal(t, 0, skippedCount)
}

func TestIntegration_ReminderPipeline_IncrementsOnSecondRun(t *testing.T) {
	db, dbCfg := dbtest.StartMySQL(t)

	seedMessage(t, db, "pipeline-repeat", "repeat@example.com", 0, 100)

	svc, _ := newPipeline(dbCfg)
	reminderCfg := notificationDomain.ReminderConfig{Enabled: true, CheckAfterHours: 24, MaxReminders: 3, Interval: 1}

	require.NoError(t, svc.ProcessReminders(context.Background(), reminderCfg))
	count, _ := reminderRow(t, db, "pipeline-repeat")
	require.Equal(t, 1, count)

	// Age the existing reminder so the interval has elapsed for the second run.
	_, err := db.Exec(`UPDATE email_reminders er
		JOIN messages m ON m.messageid = er.message_id
		SET er.last_reminder_sent = NOW() - INTERVAL 5 HOUR
		WHERE m.uniqueid = ?`, "pipeline-repeat")
	require.NoError(t, err)

	require.NoError(t, svc.ProcessReminders(context.Background(), reminderCfg))
	count, _ = reminderRow(t, db, "pipeline-repeat")
	assert.Equal(t, 2, count, "second eligible run should increment the reminder count")
}
