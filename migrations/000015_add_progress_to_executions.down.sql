-- Remove progress and result columns from executions table
DROP INDEX IF EXISTS idx_executions_progress;

DROP INDEX IF EXISTS idx_executions_result;

ALTER TABLE executions DROP COLUMN IF EXISTS progress;

ALTER TABLE executions DROP COLUMN IF EXISTS result;