-- Remove Git-related fields from projects table
-- This migration removes the following columns:
-- - repo_url (replaced by repository_url)
-- - main_branch
-- - git_auth_method
-- - git_enabled

-- Drop indexes first
DROP INDEX IF EXISTS idx_projects_git_enabled;

DROP INDEX IF EXISTS idx_projects_git_auth_method;

-- Drop columns
ALTER TABLE projects
DROP COLUMN IF EXISTS repo_url,
DROP COLUMN IF EXISTS main_branch,
DROP COLUMN IF EXISTS git_auth_method,
DROP COLUMN IF EXISTS git_enabled;