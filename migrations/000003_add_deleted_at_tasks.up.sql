-- Add deleted_at column to tasks table for soft delete
ALTER TABLE tasks ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- Create index on deleted_at for better performance
CREATE INDEX idx_tasks_deleted_at ON tasks (deleted_at);