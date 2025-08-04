-- Migration Down: Revert Task Management Enhancements

-- Drop triggers first
DROP TRIGGER IF EXISTS update_project_statistics_trigger ON tasks;

DROP TRIGGER IF EXISTS task_audit_log_trigger ON tasks;

DROP TRIGGER IF EXISTS check_circular_dependencies_trigger ON task_dependencies;

-- Drop functions
DROP FUNCTION IF EXISTS update_task_statistics ();

DROP FUNCTION IF EXISTS create_task_audit_log ();

DROP FUNCTION IF EXISTS check_circular_dependencies ();

-- Drop tables
DROP TABLE IF EXISTS task_attachments;

DROP TABLE IF EXISTS task_comments;

DROP TABLE IF EXISTS task_dependencies;

DROP TABLE IF EXISTS task_templates;

DROP TABLE IF EXISTS task_audit_logs;

-- Drop indexes
DROP INDEX IF EXISTS idx_tasks_tags_gin;

DROP INDEX IF EXISTS idx_tasks_actual_hours;

DROP INDEX IF EXISTS idx_tasks_estimated_hours;

DROP INDEX IF EXISTS idx_tasks_due_date;

DROP INDEX IF EXISTS idx_tasks_assigned_to;

DROP INDEX IF EXISTS idx_tasks_is_template;

DROP INDEX IF EXISTS idx_tasks_is_archived;

DROP INDEX IF EXISTS idx_tasks_parent_task_id;

DROP INDEX IF EXISTS idx_tasks_priority;

-- Drop columns from tasks table
ALTER TABLE tasks DROP COLUMN IF EXISTS due_date;

ALTER TABLE tasks DROP COLUMN IF EXISTS assigned_to;

ALTER TABLE tasks DROP COLUMN IF EXISTS template_id;

ALTER TABLE tasks DROP COLUMN IF EXISTS is_template;

ALTER TABLE tasks DROP COLUMN IF EXISTS is_archived;

ALTER TABLE tasks DROP COLUMN IF EXISTS parent_task_id;

ALTER TABLE tasks DROP COLUMN IF EXISTS tags;

ALTER TABLE tasks DROP COLUMN IF EXISTS actual_hours;

ALTER TABLE tasks DROP COLUMN IF EXISTS estimated_hours;

ALTER TABLE tasks DROP COLUMN IF EXISTS priority;