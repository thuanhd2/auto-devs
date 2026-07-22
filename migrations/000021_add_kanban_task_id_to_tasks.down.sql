DROP INDEX IF EXISTS idx_tasks_kanban_task_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS kanban_task_id;
