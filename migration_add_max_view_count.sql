-- Migration to add max_view_count column to messages table
-- This allows configurable maximum view counts per message

ALTER TABLE messages 
ADD COLUMN max_view_count INT DEFAULT 5 NOT NULL 
COMMENT 'Maximum number of times this message can be viewed before deletion';

-- Update existing messages to have the default max view count of 5
UPDATE messages 
SET max_view_count = 5 
WHERE max_view_count IS NULL;

-- Add index for performance if needed
CREATE INDEX idx_messages_max_view_count ON messages(max_view_count);