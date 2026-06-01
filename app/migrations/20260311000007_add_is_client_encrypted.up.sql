ALTER TABLE `messages`
ADD COLUMN `is_client_encrypted` TINYINT(1) NOT NULL DEFAULT 0 AFTER `uniqueid`;
