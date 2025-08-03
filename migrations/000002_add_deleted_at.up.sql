-- Add deleted_at column to projects table for soft delete
ALTER TABLE projects ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- Create index on deleted_at for better performance
CREATE INDEX idx_projects_deleted_at ON projects (deleted_at);