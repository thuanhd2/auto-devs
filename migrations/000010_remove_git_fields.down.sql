-- Rollback: Add back Git-related fields to projects table
ALTER TABLE projects
ADD COLUMN repo_url VARCHAR(500),
ADD COLUMN main_branch VARCHAR(100) DEFAULT 'main',
ADD COLUMN git_auth_method VARCHAR(20),
ADD COLUMN git_enabled BOOLEAN DEFAULT false;

-- Recreate indexes
CREATE INDEX idx_projects_git_enabled ON projects (git_enabled);

CREATE INDEX idx_projects_git_auth_method ON projects (git_auth_method)
WHERE
    git_auth_method IS NOT NULL;

-- Migrate data back from repository_url to repo_url if needed
UPDATE projects
SET
    repo_url = repository_url
WHERE
    repository_url IS NOT NULL
    AND repo_url IS NULL;