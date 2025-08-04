-- Drop triggers
DROP TRIGGER IF EXISTS update_executions_updated_at ON executions;
DROP TRIGGER IF EXISTS update_processes_updated_at ON processes;

-- Drop indexes
DROP INDEX IF EXISTS idx_executions_task_id;
DROP INDEX IF EXISTS idx_executions_status;
DROP INDEX IF EXISTS idx_executions_process_id;
DROP INDEX IF EXISTS idx_executions_started_at;

DROP INDEX IF EXISTS idx_processes_execution_id;
DROP INDEX IF EXISTS idx_processes_process_id;
DROP INDEX IF EXISTS idx_processes_status;

DROP INDEX IF EXISTS idx_execution_logs_execution_id;
DROP INDEX IF EXISTS idx_execution_logs_log_level;
DROP INDEX IF EXISTS idx_execution_logs_timestamp;
DROP INDEX IF EXISTS idx_execution_logs_source;

-- Drop check constraints
ALTER TABLE executions DROP CONSTRAINT IF EXISTS valid_execution_status;
ALTER TABLE processes DROP CONSTRAINT IF EXISTS valid_process_status;
ALTER TABLE execution_logs DROP CONSTRAINT IF EXISTS valid_log_level;

-- Drop tables (in reverse order due to foreign key dependencies)
DROP TABLE IF EXISTS execution_logs;
DROP TABLE IF EXISTS processes;
DROP TABLE IF EXISTS executions;