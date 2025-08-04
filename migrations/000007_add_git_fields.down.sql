-- Drop indexes
DROP INDEX IF EXISTS idx_projects_git_enabled;

DROP INDEX IF EXISTS idx_projects_repository_url;

DROP INDEX IF EXISTS idx_projects_git_auth_method;

-- Drop Git-related columns from projects table
ALTER TABLE projects
DROP COLUMN IF EXISTS repository_url,
DROP COLUMN IF EXISTS main_branch,
DROP COLUMN IF EXISTS worktree_base_path,
DROP COLUMN IF EXISTS git_auth_method,
DROP COLUMN IF EXISTS git_enabled;