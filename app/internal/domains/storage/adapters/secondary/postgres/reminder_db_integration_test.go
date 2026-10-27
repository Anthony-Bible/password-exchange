//go:build integration

package postgres

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// seedMessage inserts a message aged ageSecs seconds. A nil email stores NULL.
func seedMessage(tb testingTB, db *sql.DB, uniqueID string, email any, viewCount int, ageSecs float64) int {
	tb.Helper()
	var id int
	require.NoError(tb, db.QueryRow(
		`INSERT INTO messages (message, uniqueid, other_email, view_count, max_view_count, created, expires_at)
		 VALUES ('ciphertext', $1, $2, $3, 5, now() - make_interval(secs => $4::double precision), now() + interval '7 days')
		 RETURNING messageid`,
		uniqueID, email, viewCount, ageSecs).Scan(&id))
	return id
}

func seedReminder(tb testingTB, db *sql.DB, messageID int, email string, count int, lastSentAgeSecs float64) {
	tb.Helper()
	_, err := db.Exec(
		`INSERT INTO email_reminders (message_id, email_address, reminder_count, last_reminder_sent)
		 VALUES ($1, $2, $3, now() - make_interval(secs => $4::double precision))`,
		messageID, email, count, lastSentAgeSecs)
	require.NoError(tb, err)
}

const hour = 3600.0

func TestIntegration_GetUnviewedMessagesForReminders_Selection(t *testing.T) {
	db, repo := startAdapter(t)
	eligible := seedMessage(t, db, "eligible", "want@example.com", 0, 48*hour)
	seedMessage(t, db, "too-recent", "skip@example.com", 0, 1*hour)
	seedMessage(t, db, "viewed", "skip@example.com", 1, 48*hour)
	seedMessage(t, db, "no-email", nil, 0, 48*hour)
	seedMessage(t, db, "empty-email", "", 0, 48*hour)

	msgs, err := repo.GetUnviewedMessagesForReminders(24, 3, 24)

	require.NoError(t, err)
	require.Len(t, msgs, 1)
	assert.Equal(t, "eligible", msgs[0].UniqueID)
	assert.Equal(t, eligible, msgs[0].MessageID)
	assert.Equal(t, "want@example.com", msgs[0].RecipientEmail)
	assert.Equal(t, 2, msgs[0].DaysOld, "48h old is exactly 2 full days")
}

func TestIntegration_GetUnviewedMessagesForReminders_DaysOldIsFloorOfFullDays(t *testing.T) {
	db, repo := startAdapter(t)
	seedMessage(t, db, "almost-two", "a@example.com", 0, 47*hour+3000)

	msgs, err := repo.GetUnviewedMessagesForReminders(1, 3, 1)

	require.NoError(t, err)
	require.Len(t, msgs, 1)
	assert.Equal(t, 1, msgs[0].DaysOld)
}

func TestIntegration_GetUnviewedMessagesForReminders_MaxRemindersAndInterval(t *testing.T) {
	db, repo := startAdapter(t)
	dueAgain := seedMessage(t, db, "due-again", "a@example.com", 0, 100*hour)
	seedReminder(t, db, dueAgain, "a@example.com", 1, 30*hour)
	atMax := seedMessage(t, db, "at-max", "b@example.com", 0, 100*hour)
	seedReminder(t, db, atMax, "b@example.com", 3, 30*hour)
	recent := seedMessage(t, db, "recent-reminder", "c@example.com", 0, 100*hour)
	seedReminder(t, db, recent, "c@example.com", 1, 1*hour)

	msgs, err := repo.GetUnviewedMessagesForReminders(24, 3, 24)

	require.NoError(t, err)
	got := map[string]bool{}
	for _, m := range msgs {
		got[m.UniqueID] = true
	}
	assert.Equal(t, map[string]bool{"due-again": true}, got)
}

// Upsert: k calls -> exactly one row with reminder_count = k; distinct emails -> distinct rows.
func TestIntegration_Property_LogReminderSentUpsertsCount(t *testing.T) {
	db, repo := startAdapter(t)
	rapid.Check(t, func(rt *rapid.T) {
		resetTables(rt, db)
		id := seedMessage(rt, db, "up", "x@example.com", 0, 48*hour)
		k := rapid.IntRange(1, 8).Draw(rt, "k")
		extra := rapid.IntRange(0, 3).Draw(rt, "extraEmails")

		for i := 0; i < k; i++ {
			require.NoError(rt, repo.LogReminderSent(id, "x@example.com"))
		}
		for e := 0; e < extra; e++ {
			require.NoError(rt, repo.LogReminderSent(id, fmt.Sprintf("other%d@example.com", e)))
		}

		history, err := repo.GetReminderHistory(id)
		require.NoError(rt, err)
		require.Len(rt, history, 1+extra, "distinct emails get distinct rows")
		for _, h := range history {
			if h.EmailAddress == "x@example.com" {
				assert.Equal(rt, k, h.ReminderCount)
			} else {
				assert.Equal(rt, 1, h.ReminderCount)
			}
		}
	})
}

func TestIntegration_GetReminderHistory_Empty(t *testing.T) {
	db, repo := startAdapter(t)
	id := seedMessage(t, db, "no-history", "none@example.com", 0, 48*hour)
	history, err := repo.GetReminderHistory(id)
	require.NoError(t, err)
	assert.Empty(t, history)
}

// PBT reminder window. A message is returned iff
//
//	view_count = 0 AND other_email non-empty AND age > olderThanHours AND
//	(no reminder OR (count < max AND last-sent age > intervalHours)),
//
// DaysOld = floor(age / 24h), and each message appears at most once.
func TestIntegration_Property_ReminderWindowSelection(t *testing.T) {
	db, repo := startAdapter(t)
	type seeded struct {
		uid         string
		age         float64
		view        int
		email       any
		hasReminder bool
		count       int
		lastAge     float64
	}
	// delta is the signed distance from a threshold, |delta| in [2s, 120s].
	delta := func(rt *rapid.T, label string) float64 {
		mag := float64(rapid.IntRange(2, 120).Draw(rt, label+"Mag"))
		if rapid.Bool().Draw(rt, label+"Sign") {
			return mag
		}
		return -mag
	}
	rapid.Check(t, func(rt *rapid.T) {
		resetTables(rt, db)
		olderThan := rapid.IntRange(1, 168).Draw(rt, "olderThanHours")
		interval := rapid.IntRange(1, 72).Draw(rt, "intervalHours")
		maxRem := rapid.IntRange(1, 5).Draw(rt, "maxReminders")
		n := rapid.IntRange(1, 6).Draw(rt, "n")

		var all []seeded
		for i := 0; i < n; i++ {
			s := seeded{uid: fmt.Sprintf("m%d", i)}
			s.age = float64(olderThan)*hour + delta(rt, "age")
			s.view = rapid.SampledFrom([]int{0, 0, 0, 1, 2}).Draw(rt, "view")
			switch rapid.IntRange(0, 3).Draw(rt, "emailKind") {
			case 0:
				s.email = nil
			case 1:
				s.email = ""
			default:
				s.email = "to@example.com"
			}
			s.hasReminder = rapid.Bool().Draw(rt, "hasReminder")
			s.count = rapid.IntRange(0, 6).Draw(rt, "count")
			s.lastAge = float64(interval)*hour + delta(rt, "last")
			id := seedMessage(rt, db, s.uid, s.email, s.view, s.age)
			if s.hasReminder {
				seedReminder(rt, db, id, "r@example.com", s.count, s.lastAge)
			}
			all = append(all, s)
		}

		got, err := repo.GetUnviewedMessagesForReminders(olderThan, maxRem, interval)
		require.NoError(rt, err)

		returned := map[string]int{}
		for _, m := range got {
			returned[m.UniqueID]++
		}
		for _, s := range all {
			emailOK := s.email != nil && s.email != ""
			ageOK := s.age > float64(olderThan)*hour
			reminderOK := !s.hasReminder || (s.count < maxRem && s.lastAge > float64(interval)*hour)
			want := s.view == 0 && emailOK && ageOK && reminderOK

			assert.Equal(rt, want, returned[s.uid] == 1, "message %+v (olderThan=%d interval=%d max=%d)", s, olderThan, interval, maxRem)
			assert.LessOrEqual(rt, returned[s.uid], 1, "message %s returned more than once", s.uid)
		}
		for _, m := range got {
			for _, s := range all {
				if s.uid == m.UniqueID {
					assert.Equal(rt, int(time.Duration(s.age*float64(time.Second))/(24*time.Hour)), m.DaysOld, "DaysOld = floor(age/24h)")
				}
			}
		}
	})
}
