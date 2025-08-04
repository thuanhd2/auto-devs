

-- Drop indexes
DROP INDEX IF EXISTS idx_task_status_histories_deleted_at;
DROP INDEX IF EXISTS idx_task_status_histories_created_at;
DROP INDEX IF EXISTS idx_task_status_histories_to_status;
DROP INDEX IF EXISTS idx_task_status_histories_task_id;

-- Drop table
DROP TABLE IF EXISTS task_status_histories;