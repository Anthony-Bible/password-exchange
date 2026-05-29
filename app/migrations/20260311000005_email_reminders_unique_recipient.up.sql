-- Migration: enforce one email_reminders row per (message_id, email_address)
--
-- WARNING: THIS MIGRATION IS ONE-WAY (DESTRUCTIVE). Take a backup of the
-- email_reminders table before applying it if you may need to roll back.
-- The down migration only drops the unique index; it cannot restore the
-- duplicate rows that are deleted here.
--
-- LogReminderSent relies on INSERT ... ON DUPLICATE KEY UPDATE to increment
-- reminder_count, but the original email_reminders table had no unique key for
-- that upsert to collide on. As a result every reminder inserted a fresh row
-- with reminder_count = 1: counts never incremented, the max-reminders cap
-- never tripped, and GetUnviewedMessagesForReminders' LEFT JOIN produced
-- duplicate rows. This migration collapses any duplicate rows and adds the
-- unique key the upsert needs.

-- Consolidate duplicates into the lowest-id row per (message_id, email_address):
-- sum the (mostly 1) counts and keep the most recent send time.
UPDATE email_reminders er
JOIN (
    SELECT MIN(id) AS keep_id,
           SUM(reminder_count) AS total_count,
           MAX(last_reminder_sent) AS latest_sent
    FROM email_reminders
    GROUP BY message_id, email_address
) agg ON er.id = agg.keep_id
SET er.reminder_count = agg.total_count,
    er.last_reminder_sent = agg.latest_sent;

-- Delete the now-redundant duplicate rows.
DELETE er FROM email_reminders er
JOIN (
    SELECT message_id, email_address, MIN(id) AS keep_id
    FROM email_reminders
    GROUP BY message_id, email_address
) agg
  ON er.message_id = agg.message_id
 AND er.email_address = agg.email_address
 AND er.id <> agg.keep_id;

-- Add the unique key the upsert depends on.
ALTER TABLE email_reminders
  ADD UNIQUE KEY uq_email_reminders_message_email (message_id, email_address);
