-- Add Git-related fields to projects table
ALTER TABLE projects
ADD COLUMN repository_url VARCHAR(500),
ADD COLUMN main_branch VARCHAR(100) DEFAULT 'main',
ADD COLUMN worktree_base_path VARCHAR(500),
ADD COLUMN git_auth_method VARCHAR(20),
ADD COLUMN git_enabled BOOLEAN DEFAULT false;

-- Create indexes for Git-related queries
CREATE INDEX idx_projects_git_enabled ON projects (git_enabled);

CREATE INDEX idx_projects_repository_url ON projects (repository_url)
WHERE
    repository_url IS NOT NULL;

CREATE INDEX idx_projects_git_auth_method ON projects (git_auth_method)
WHERE
    git_auth_method IS NOT NULL;

-- Migrate existing repo_url data to repository_url for backward compatibility
UPDATE projects
SET
    repository_url = repo_url
WHERE
    repo_url IS NOT NULL
    AND repository_url IS NULL;

-- Add comments for documentation
COMMENT ON COLUMN projects.repository_url IS 'Git repository URL (HTTPS or SSH)';

COMMENT ON COLUMN projects.main_branch IS 'Default branch name for the repository';

COMMENT ON COLUMN projects.worktree_base_path IS 'Base path for Git worktree operations';

COMMENT ON COLUMN projects.git_auth_method IS 'Authentication method: ssh or https';

COMMENT ON COLUMN projects.git_enabled IS 'Whether Git integration is enabled for this project';