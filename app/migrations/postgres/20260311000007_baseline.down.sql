-- WARNING: DESTRUCTIVE. This drops every application table and ALL of its
-- data (messages, email reminders, file upload sessions). It cannot be undone.
-- Only run it against a database you are prepared to wipe.
DROP TABLE IF EXISTS email_reminders, file_upload_sessions, messages;
