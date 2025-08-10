-- Remove line field from execution_logs table
DROP INDEX IF EXISTS idx_execution_logs_execution_id_line;

ALTER TABLE execution_logs DROP COLUMN IF EXISTS line;

-- Remove updated_at field and trigger from execution_logs table
DROP TRIGGER IF EXISTS trigger_update_execution_logs_updated_at ON execution_logs;

DROP FUNCTION IF EXISTS update_execution_logs_updated_at ();

ALTER TABLE execution_logs DROP COLUMN IF EXISTS updated_at;