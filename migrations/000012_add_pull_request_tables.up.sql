-- Create pull_requests table
CREATE TABLE IF NOT EXISTS pull_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL,
    github_pr_number INTEGER NOT NULL,
    repository VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED', 'CLOSED')),
    head_branch VARCHAR(255) NOT NULL,
    base_branch VARCHAR(255) NOT NULL DEFAULT 'main',
    github_url VARCHAR(500),
    merge_commit_sha VARCHAR(40),
    merged_at TIMESTAMP,
    closed_at TIMESTAMP,
    created_by VARCHAR(255),
    merged_by VARCHAR(255),
    reviewers JSONB DEFAULT '[]',
    labels JSONB DEFAULT '[]',
    assignees JSONB DEFAULT '[]',
    is_draft BOOLEAN DEFAULT FALSE,
    mergeable BOOLEAN,
    mergeable_state VARCHAR(50),
    additions INTEGER,
    deletions INTEGER,
    changed_files INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Create indexes for pull_requests
CREATE INDEX IF NOT EXISTS idx_pull_requests_task_id ON pull_requests(task_id);
CREATE INDEX IF NOT EXISTS idx_pull_requests_repository ON pull_requests(repository);
CREATE INDEX IF NOT EXISTS idx_pull_requests_status ON pull_requests(status);
CREATE INDEX IF NOT EXISTS idx_pull_requests_github_pr_number ON pull_requests(github_pr_number);
CREATE INDEX IF NOT EXISTS idx_pull_requests_created_at ON pull_requests(created_at);
CREATE INDEX IF NOT EXISTS idx_pull_requests_updated_at ON pull_requests(updated_at);
CREATE INDEX IF NOT EXISTS idx_pull_requests_deleted_at ON pull_requests(deleted_at) WHERE deleted_at IS NOT NULL;

-- Unique constraint to prevent duplicate PRs for the same repository and PR number
CREATE UNIQUE INDEX IF NOT EXISTS idx_pull_requests_unique_repo_pr 
ON pull_requests(repository, github_pr_number) 
WHERE deleted_at IS NULL;

-- Create pull_request_comments table
CREATE TABLE IF NOT EXISTS pull_request_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pull_request_id UUID NOT NULL,
    github_id BIGINT UNIQUE,
    author VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    file_path VARCHAR(500),
    line INTEGER,
    is_resolved BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(id) ON DELETE CASCADE
);

-- Create indexes for pull_request_comments
CREATE INDEX IF NOT EXISTS idx_pr_comments_pull_request_id ON pull_request_comments(pull_request_id);
CREATE INDEX IF NOT EXISTS idx_pr_comments_github_id ON pull_request_comments(github_id);
CREATE INDEX IF NOT EXISTS idx_pr_comments_author ON pull_request_comments(author);
CREATE INDEX IF NOT EXISTS idx_pr_comments_created_at ON pull_request_comments(created_at);
CREATE INDEX IF NOT EXISTS idx_pr_comments_deleted_at ON pull_request_comments(deleted_at) WHERE deleted_at IS NOT NULL;

-- Create pull_request_reviews table
CREATE TABLE IF NOT EXISTS pull_request_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pull_request_id UUID NOT NULL,
    github_id BIGINT UNIQUE,
    reviewer VARCHAR(255) NOT NULL,
    state VARCHAR(50) NOT NULL CHECK (state IN ('APPROVED', 'CHANGES_REQUESTED', 'COMMENTED', 'DISMISSED', 'PENDING')),
    body TEXT,
    submitted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(id) ON DELETE CASCADE
);

-- Create indexes for pull_request_reviews
CREATE INDEX IF NOT EXISTS idx_pr_reviews_pull_request_id ON pull_request_reviews(pull_request_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviews_github_id ON pull_request_reviews(github_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviews_reviewer ON pull_request_reviews(reviewer);
CREATE INDEX IF NOT EXISTS idx_pr_reviews_state ON pull_request_reviews(state);
CREATE INDEX IF NOT EXISTS idx_pr_reviews_created_at ON pull_request_reviews(created_at);
CREATE INDEX IF NOT EXISTS idx_pr_reviews_deleted_at ON pull_request_reviews(deleted_at) WHERE deleted_at IS NOT NULL;

-- Create pull_request_checks table for CI/CD status
CREATE TABLE IF NOT EXISTS pull_request_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pull_request_id UUID NOT NULL,
    check_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('PENDING', 'SUCCESS', 'FAILURE', 'ERROR', 'CANCELLED')),
    conclusion VARCHAR(50) CHECK (conclusion IN ('SUCCESS', 'FAILURE', 'NEUTRAL', 'CANCELLED', 'TIMED_OUT', 'ACTION_REQUIRED')),
    details_url VARCHAR(500),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(id) ON DELETE CASCADE
);

-- Create indexes for pull_request_checks
CREATE INDEX IF NOT EXISTS idx_pr_checks_pull_request_id ON pull_request_checks(pull_request_id);
CREATE INDEX IF NOT EXISTS idx_pr_checks_check_name ON pull_request_checks(check_name);
CREATE INDEX IF NOT EXISTS idx_pr_checks_status ON pull_request_checks(status);
CREATE INDEX IF NOT EXISTS idx_pr_checks_conclusion ON pull_request_checks(conclusion);
CREATE INDEX IF NOT EXISTS idx_pr_checks_created_at ON pull_request_checks(created_at);
CREATE INDEX IF NOT EXISTS idx_pr_checks_deleted_at ON pull_request_checks(deleted_at) WHERE deleted_at IS NOT NULL;

-- Create trigger to automatically update updated_at timestamp for pull_requests
CREATE OR REPLACE FUNCTION update_pull_requests_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_pull_requests_updated_at
    BEFORE UPDATE ON pull_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_pull_requests_updated_at();

-- Create trigger to automatically update updated_at timestamp for pull_request_comments
CREATE OR REPLACE FUNCTION update_pull_request_comments_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_pull_request_comments_updated_at
    BEFORE UPDATE ON pull_request_comments
    FOR EACH ROW
    EXECUTE FUNCTION update_pull_request_comments_updated_at();

-- Create trigger to automatically update updated_at timestamp for pull_request_reviews
CREATE OR REPLACE FUNCTION update_pull_request_reviews_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_pull_request_reviews_updated_at
    BEFORE UPDATE ON pull_request_reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_pull_request_reviews_updated_at();

-- Create trigger to automatically update updated_at timestamp for pull_request_checks
CREATE OR REPLACE FUNCTION update_pull_request_checks_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_pull_request_checks_updated_at
    BEFORE UPDATE ON pull_request_checks
    FOR EACH ROW
    EXECUTE FUNCTION update_pull_request_checks_updated_at();