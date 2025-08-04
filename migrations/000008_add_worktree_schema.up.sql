-- Create worktrees table for Git worktree tracking
CREATE TABLE worktrees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    task_id UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    branch_name VARCHAR(255) NOT NULL,
    worktree_path TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'creating',
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        deleted_at TIMESTAMP
    WITH
        TIME ZONE NULL,
        UNIQUE (task_id),
        UNIQUE (worktree_path)
);

-- Add Git-related fields to tasks table
ALTER TABLE tasks
ADD COLUMN worktree_path TEXT,
ADD COLUMN git_status VARCHAR(50) DEFAULT 'none';

-- Create indexes for better performance
CREATE INDEX idx_worktrees_task_id ON worktrees (task_id);

CREATE INDEX idx_worktrees_project_id ON worktrees (project_id);

CREATE INDEX idx_worktrees_status ON worktrees (status);

CREATE INDEX idx_worktrees_branch_name ON worktrees (branch_name);

CREATE INDEX idx_tasks_branch_name ON tasks (branch_name)
WHERE
    branch_name IS NOT NULL;

CREATE INDEX idx_tasks_git_status ON tasks (git_status);

CREATE INDEX idx_tasks_worktree_path ON tasks (worktree_path)
WHERE
    worktree_path IS NOT NULL;

-- Add check constraints for valid status values
ALTER TABLE worktrees
ADD CONSTRAINT valid_worktree_status CHECK (
    status IN (
        'creating',
        'active',
        'completed',
        'cleaning',
        'error'
    )
);

ALTER TABLE tasks
ADD CONSTRAINT valid_git_status CHECK (
    git_status IN (
        'none',
        'creating',
        'active',
        'completed',
        'cleaning',
        'error'
    )
);

-- Create trigger for auto-updating updated_at column on worktrees
CREATE TRIGGER update_worktrees_updated_at 
    BEFORE UPDATE ON worktrees 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE worktrees IS 'Git worktree tracking for tasks';

COMMENT ON COLUMN worktrees.task_id IS 'Reference to the task this worktree belongs to';

COMMENT ON COLUMN worktrees.project_id IS 'Reference to the project this worktree belongs to';

COMMENT ON COLUMN worktrees.branch_name IS 'Git branch name for this worktree';

COMMENT ON COLUMN worktrees.worktree_path IS 'File system path to the worktree directory';

COMMENT ON COLUMN worktrees.status IS 'Current status of the worktree: creating, active, completed, cleaning, error';

COMMENT ON COLUMN worktrees.deleted_at IS 'Soft delete timestamp';

COMMENT ON COLUMN tasks.worktree_path IS 'File system path to the task worktree directory';

COMMENT ON COLUMN tasks.git_status IS 'Git integration status for this task: none, creating, active, completed, cleaning, error';