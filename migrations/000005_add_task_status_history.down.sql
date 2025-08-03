-- Drop trigger first
DROP TRIGGER IF EXISTS track_task_status_change_trigger ON tasks;

-- Drop function
DROP FUNCTION IF EXISTS track_task_status_change();

-- Drop indexes
DROP INDEX IF EXISTS idx_task_status_history_deleted_at;
DROP INDEX IF EXISTS idx_task_status_history_created_at;
DROP INDEX IF EXISTS idx_task_status_history_to_status;
DROP INDEX IF EXISTS idx_task_status_history_task_id;

-- Drop table
DROP TABLE IF EXISTS task_status_history;