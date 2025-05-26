-- Migration: Add email_reminders table for daily reminder email tracking
-- This table tracks reminder attempts for unviewed messages

CREATE TABLE email_reminders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    message_id INT NOT NULL,
    email_address VARCHAR(255) NOT NULL,
    reminder_count INT DEFAULT 0,
    last_reminder_sent TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(messageid) ON DELETE CASCADE,
    INDEX idx_message_id (message_id),
    INDEX idx_email_address (email_address),
    INDEX idx_last_reminder_sent (last_reminder_sent)
);

-- Add index for efficient queries of unviewed messages for reminders
CREATE INDEX idx_messages_reminder_lookup ON messages(view_count, created, other_email);