-- Add init_workspace_script field to projects table
ALTER TABLE projects ADD COLUMN init_workspace_script TEXT;

-- Add comment for documentation
COMMENT ON COLUMN projects.init_workspace_script IS 'Bash script to be executed after git worktree creation during workspace initialization';