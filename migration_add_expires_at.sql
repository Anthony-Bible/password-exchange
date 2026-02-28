-- Add expires_at column to messages table
-- Messages expire 7 days after creation by default

ALTER TABLE messages
  ADD COLUMN expires_at TIMESTAMP NOT NULL
  DEFAULT (CURRENT_TIMESTAMP + INTERVAL 7 DAY)
  COMMENT 'Expires 7 days after creation';

-- Backfill existing rows that have the default zero value
UPDATE messages SET expires_at = created + INTERVAL 7 DAY
  WHERE expires_at = '0000-00-00 00:00:00';

CREATE INDEX idx_messages_expires_at ON messages(expires_at);
