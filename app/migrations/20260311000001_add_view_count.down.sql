DROP INDEX `idx_messages_view_count` ON `messages`;
ALTER TABLE `messages` DROP COLUMN `view_count`;
