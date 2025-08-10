-- Rollback migration: 000014_add_missing_columns_to_execution_logs

-- Drop indexes
DROP INDEX IF EXISTS idx_execution_logs_metadata;

DROP INDEX IF EXISTS idx_execution_logs_created_at;

-- Drop columns
ALTER TABLE execution_logs DROP COLUMN IF EXISTS metadata;

ALTER TABLE execution_logs DROP COLUMN IF EXISTS created_at;