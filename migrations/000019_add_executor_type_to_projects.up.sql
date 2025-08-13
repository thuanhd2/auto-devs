-- Add executor_type field to projects table
ALTER TABLE projects ADD COLUMN executor_type VARCHAR(50) NOT NULL DEFAULT 'claude-code';

-- Add check constraint to ensure only valid values
ALTER TABLE projects ADD CONSTRAINT chk_executor_type CHECK (executor_type IN ('claude-code', 'fake-code'));

-- Add comment for documentation
COMMENT ON COLUMN projects.executor_type IS 'Type of AI executor to use for planning and implementation tasks';