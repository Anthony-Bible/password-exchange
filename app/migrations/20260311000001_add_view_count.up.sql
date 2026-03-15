-- Migration to add view_count column to messages table
-- This migration adds view counting functionality for automatic deletion after 5 views

-- Add the view_count column with default value 0
ALTER TABLE messages 
ADD COLUMN view_count INT NOT NULL DEFAULT 0;

-- Add an index on view_count for performance
CREATE INDEX idx_messages_view_count ON messages(view_count);

-- Update existing messages to have view_count = 0 (this is redundant due to DEFAULT but explicit)
UPDATE messages SET view_count = 0 WHERE view_count IS NULL;