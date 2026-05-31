-- Add file_upload_sessions table to persist chunked upload session metadata
-- across web-service restarts and replicas. Completed sessions are retained
-- so finalized files remain downloadable; incomplete sessions are swept by
-- the session-cleanup goroutine once their expires_at has elapsed.

CREATE TABLE IF NOT EXISTS `file_upload_sessions` (
  `session_id`      VARCHAR(255) NOT NULL,
  `file_id`         VARCHAR(255) NOT NULL,
  `message_id`      VARCHAR(255) NOT NULL DEFAULT '',
  `upload_id`       VARCHAR(255) NOT NULL DEFAULT '',
  `filename`        VARCHAR(255) NOT NULL DEFAULT '',
  `content_type`    VARCHAR(255) NOT NULL DEFAULT '',
  `total_size`      BIGINT       NOT NULL DEFAULT 0,
  `total_chunks`    INT          NOT NULL DEFAULT 0,
  `status`          VARCHAR(50)  NOT NULL DEFAULT 'active',
  `encryption_key`  BLOB         NULL,
  `completed_parts` JSON         NOT NULL,
  `created_at`      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `expires_at`      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`session_id`),
  UNIQUE KEY `uq_file_id` (`file_id`),
  INDEX `idx_expires_at` (`expires_at`),
  INDEX `idx_status_expires_at` (`status`, `expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='Upload session metadata for chunked encrypted file uploads';
