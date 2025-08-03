-- Remove deleted_at column from projects table
DROP INDEX IF EXISTS idx_projects_deleted_at;

ALTER TABLE projects DROP COLUMN IF EXISTS deleted_at;