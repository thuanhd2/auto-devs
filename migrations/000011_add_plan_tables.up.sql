-- Create plans table
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT valid_plan_status CHECK (
        status IN ('DRAFT', 'REVIEWING', 'APPROVED', 'REJECTED')
    )
);

-- Create plan_versions table for version tracking
CREATE TABLE plan_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans (id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for performance
CREATE INDEX idx_plans_task_id ON plans (task_id);
CREATE INDEX idx_plans_status ON plans (status);
CREATE INDEX idx_plans_created_at ON plans (created_at);
CREATE INDEX idx_plans_deleted_at ON plans (deleted_at);

CREATE INDEX idx_plan_versions_plan_id ON plan_versions (plan_id);
CREATE INDEX idx_plan_versions_version ON plan_versions (plan_id, version);
CREATE INDEX idx_plan_versions_created_at ON plan_versions (created_at);
CREATE INDEX idx_plan_versions_deleted_at ON plan_versions (deleted_at);

-- Create partial unique indexes to ensure constraints
CREATE UNIQUE INDEX idx_plans_unique_task_id ON plans (task_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_plan_versions_unique_version ON plan_versions (plan_id, version) WHERE deleted_at IS NULL;

-- Create full-text search index on plan content
CREATE INDEX idx_plans_content_fts ON plans USING gin(to_tsvector('english', content));

-- Add trigger to automatically update updated_at on plans
CREATE OR REPLACE FUNCTION update_plans_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_plans_updated_at
    BEFORE UPDATE ON plans
    FOR EACH ROW
    EXECUTE FUNCTION update_plans_updated_at();

-- Add trigger to create initial version when plan is created
CREATE OR REPLACE FUNCTION create_initial_plan_version()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO plan_versions (plan_id, version, content, created_by)
    VALUES (NEW.id, 1, NEW.content, 'system');
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER create_initial_plan_version
    AFTER INSERT ON plans
    FOR EACH ROW
    EXECUTE FUNCTION create_initial_plan_version();