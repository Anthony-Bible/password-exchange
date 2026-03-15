DROP INDEX `idx_messages_max_view_count` ON `messages`;
ALTER TABLE `messages` DROP COLUMN `max_view_count`;
