-- Postgres baseline rollback.
--
-- WARNING: THIS MIGRATION IS DESTRUCTIVE. It drops every application table and
-- all data in them. Take a backup before applying it if you may need the data.

DROP TABLE IF EXISTS email_reminders, file_upload_sessions, messages;
