-- Remove the unique key on (message_id, email_address). The collapsed duplicate
-- rows are not restored; only the constraint is dropped.
ALTER TABLE `email_reminders` DROP INDEX `uq_email_reminders_message_email`;
