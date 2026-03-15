DROP INDEX `idx_messages_expires_at` ON `messages`;
ALTER TABLE `messages` DROP COLUMN `expires_at`;
