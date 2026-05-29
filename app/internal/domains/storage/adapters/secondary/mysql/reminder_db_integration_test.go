//go:build integration

package mysql

import (
	"database/sql"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/integration/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reminder_db_integration_test.go runs the reminder SQL against a real MySQL
// instance (via testcontainers). Unlike the sqlmock unit tests in adapter_test.go,
// these execute the actual MySQL-specific syntax (TIMESTAMPDIFF, INTERVAL ? HOUR,
// ON DUPLICATE KEY UPDATE) so the selection windows and upsert semantics that the
// reminder system relies on are verified end-to-end against the migrated schema.
//
// Gated behind the `integration` build tag; requires Docker.

// newReminderAdapter wires a MySQLAdapter onto an already-connected container DB,
// mirroring how adapter_test.go constructs the adapter for sqlmock.
func newReminderAdapter(db *sql.DB) *MySQLAdapter {
	return &MySQLAdapter{db: db, logger: noopLogger{}, validator: noopValidator{}}
}

// seedMessage inserts a row into messages with an explicit age (hours in the
// past) and returns its generated messageid. A nil email inserts SQL NULL so the
// "no recipient" exclusion path can be exercised.
func seedMessage(t *testing.T, db *sql.DB, uniqueID string, email interface{}, viewCount, ageHours int) int {
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

// seedReminder inserts a pre-existing reminder log row with a controllable
// last_reminder_sent age so interval/max-count exclusions can be tested.
func seedReminder(t *testing.T, db *sql.DB, messageID int, email string, count, lastSentAgeHours int) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO email_reminders (message_id, email_address, reminder_count, last_reminder_sent)
		 VALUES (?, ?, ?, NOW() - INTERVAL ? HOUR)`,
		messageID, email, count, lastSentAgeHours,
	)
	require.NoError(t, err)
}

func TestIntegration_GetUnviewedMessagesForReminders_Selection(t *testing.T) {
	db, _ := dbtest.StartMySQL(t)
	adapter := newReminderAdapter(db)

	// Eligible: old, unviewed, has email, no prior reminder.
	eligibleID := seedMessage(t, db, "eligible", "want@example.com", 0, 48)
	// Excluded: created too recently (1h ago, threshold is 24h).
	seedMessage(t, db, "too-recent", "skip@example.com", 0, 1)
	// Excluded: already viewed.
	seedMessage(t, db, "viewed", "skip@example.com", 1, 48)
	// Excluded: NULL recipient email.
	seedMessage(t, db, "no-email", nil, 0, 48)
	// Excluded: empty recipient email.
	seedMessage(t, db, "empty-email", "", 0, 48)

	const checkAfterHours, maxReminders, intervalHours = 24, 3, 24
	msgs, err := adapter.GetUnviewedMessagesForReminders(checkAfterHours, maxReminders, intervalHours)
	require.NoError(t, err)

	got := map[string]int{}
	for _, m := range msgs {
		got[m.UniqueID] = m.MessageID
	}
	assert.Contains(t, got, "eligible")
	assert.Equal(t, eligibleID, got["eligible"])
	assert.NotContains(t, got, "too-recent")
	assert.NotContains(t, got, "viewed")
	assert.NotContains(t, got, "no-email")
	assert.NotContains(t, got, "empty-email")

	// days_old must be computed from created (48h ≈ 2 days).
	for _, m := range msgs {
		if m.UniqueID == "eligible" {
			assert.Equal(t, 2, m.DaysOld)
		}
	}
}

func TestIntegration_GetUnviewedMessagesForReminders_MaxRemindersAndInterval(t *testing.T) {
	db, _ := dbtest.StartMySQL(t)
	adapter := newReminderAdapter(db)

	const checkAfterHours, maxReminders, intervalHours = 24, 3, 24

	// At max reminders -> excluded.
	atMaxID := seedMessage(t, db, "at-max", "atmax@example.com", 0, 48)
	seedReminder(t, db, atMaxID, "atmax@example.com", maxReminders, 48)

	// Below max but reminded too recently (interval not elapsed) -> excluded.
	recentID := seedMessage(t, db, "recent-reminder", "recent@example.com", 0, 48)
	seedReminder(t, db, recentID, "recent@example.com", 1, 1)

	// Below max and interval elapsed -> included.
	dueID := seedMessage(t, db, "due-again", "due@example.com", 0, 48)
	seedReminder(t, db, dueID, "due@example.com", 1, 48)

	msgs, err := adapter.GetUnviewedMessagesForReminders(checkAfterHours, maxReminders, intervalHours)
	require.NoError(t, err)

	got := map[string]bool{}
	for _, m := range msgs {
		got[m.UniqueID] = true
	}
	assert.True(t, got["due-again"], "message past interval and below max should be due")
	assert.False(t, got["at-max"], "message at max reminders should be excluded")
	assert.False(t, got["recent-reminder"], "message reminded within interval should be excluded")
}

func TestIntegration_LogReminderSent_UpsertSemantics(t *testing.T) {
	db, _ := dbtest.StartMySQL(t)
	adapter := newReminderAdapter(db)

	id := seedMessage(t, db, "log-target", "log@example.com", 0, 48)

	// First call inserts a row with count 1.
	require.NoError(t, adapter.LogReminderSent(id, "log@example.com"))
	history, err := adapter.GetReminderHistory(id)
	require.NoError(t, err)
	require.Len(t, history, 1)
	assert.Equal(t, 1, history[0].ReminderCount)
	assert.Equal(t, "log@example.com", history[0].EmailAddress)
	firstSent := history[0].LastReminderSent

	// Second call upserts: count increments, timestamp advances.
	require.NoError(t, adapter.LogReminderSent(id, "log@example.com"))
	history, err = adapter.GetReminderHistory(id)
	require.NoError(t, err)
	require.Len(t, history, 1, "upsert must not create a second row")
	assert.Equal(t, 2, history[0].ReminderCount)
	assert.False(t, history[0].LastReminderSent.Before(firstSent), "last_reminder_sent should not move backwards")
}

func TestIntegration_GetReminderHistory_Empty(t *testing.T) {
	db, _ := dbtest.StartMySQL(t)
	adapter := newReminderAdapter(db)

	id := seedMessage(t, db, "no-history", "none@example.com", 0, 48)

	history, err := adapter.GetReminderHistory(id)
	require.NoError(t, err)
	assert.Empty(t, history)
}
