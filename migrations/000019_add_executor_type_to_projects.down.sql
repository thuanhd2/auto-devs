-- Remove executor_type field from projects table
ALTER TABLE projects DROP CONSTRAINT IF EXISTS chk_executor_type;
ALTER TABLE projects DROP COLUMN executor_type;