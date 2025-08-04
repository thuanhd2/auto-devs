-- Drop triggers
DROP TRIGGER IF EXISTS update_worktrees_updated_at ON worktrees;

-- Drop indexes
DROP INDEX IF EXISTS idx_worktrees_task_id;

DROP INDEX IF EXISTS idx_worktrees_project_id;

DROP INDEX IF EXISTS idx_worktrees_status;

DROP INDEX IF EXISTS idx_worktrees_branch_name;

DROP INDEX IF EXISTS idx_tasks_branch_name;

DROP INDEX IF EXISTS idx_tasks_git_status;

DROP INDEX IF EXISTS idx_tasks_worktree_path;

-- Drop check constraints
ALTER TABLE worktrees
DROP CONSTRAINT IF EXISTS valid_worktree_status;

ALTER TABLE tasks DROP CONSTRAINT IF EXISTS valid_git_status;

-- Drop Git-related columns from tasks table
ALTER TABLE tasks
DROP COLUMN IF EXISTS worktree_path,
DROP COLUMN IF EXISTS git_status;

-- Drop worktrees table
DROP TABLE IF EXISTS worktrees;