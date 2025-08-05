-- Create plans table for AI implementation plans
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
    version INTEGER NOT NULL DEFAULT 1,
    
    -- JSONB fields for complex data
    steps JSONB DEFAULT '[]'::jsonb,
    context JSONB DEFAULT '{}'::jsonb,
    
    -- Metadata
    created_by VARCHAR(255),
    approved_by VARCHAR(255),
    approved_at TIMESTAMP WITH TIME ZONE,
    rejected_by VARCHAR(255),
    rejected_at TIMESTAMP WITH TIME ZONE,
    
    -- Standard timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Create plan_versions table for plan versioning
CREATE TABLE plan_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL,
    
    -- JSONB fields for versioned data
    steps JSONB DEFAULT '[]'::jsonb,
    context JSONB DEFAULT '{}'::jsonb,
    
    -- Versioning metadata
    created_by VARCHAR(255),
    change_log TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX idx_plans_task_id ON plans(task_id);
CREATE INDEX idx_plans_status ON plans(status);
CREATE INDEX idx_plans_created_by ON plans(created_by);
CREATE INDEX idx_plans_approved_by ON plans(approved_by);
CREATE INDEX idx_plans_rejected_by ON plans(rejected_by);
CREATE INDEX idx_plans_created_at ON plans(created_at);
CREATE INDEX idx_plans_updated_at ON plans(updated_at);
CREATE INDEX idx_plans_deleted_at ON plans(deleted_at);

-- Indexes for plan_versions
CREATE INDEX idx_plan_versions_plan_id ON plan_versions(plan_id);
CREATE INDEX idx_plan_versions_version ON plan_versions(version);
CREATE INDEX idx_plan_versions_created_by ON plan_versions(created_by);
CREATE INDEX idx_plan_versions_created_at ON plan_versions(created_at);

-- Unique constraint for plan versions
CREATE UNIQUE INDEX idx_plan_versions_unique ON plan_versions(plan_id, version);

-- GIN indexes for JSONB fields for better search performance
CREATE INDEX idx_plans_steps_gin ON plans USING GIN (steps);
CREATE INDEX idx_plans_context_gin ON plans USING GIN (context);
CREATE INDEX idx_plan_versions_steps_gin ON plan_versions USING GIN (steps);
CREATE INDEX idx_plan_versions_context_gin ON plan_versions USING GIN (context);

-- Add check constraints for valid status values
ALTER TABLE plans 
ADD CONSTRAINT valid_plan_status CHECK (
    status IN ('DRAFT', 'REVIEWING', 'APPROVED', 'REJECTED', 'EXECUTING', 'COMPLETED', 'CANCELLED')
);

ALTER TABLE plan_versions 
ADD CONSTRAINT valid_plan_version_status CHECK (
    status IN ('DRAFT', 'REVIEWING', 'APPROVED', 'REJECTED', 'EXECUTING', 'COMPLETED', 'CANCELLED')
);

-- Add check constraints for version numbers
ALTER TABLE plans
ADD CONSTRAINT valid_plan_version CHECK (version > 0);

ALTER TABLE plan_versions
ADD CONSTRAINT valid_plan_version_number CHECK (version > 0);

-- Create trigger for auto-updating updated_at column
CREATE TRIGGER update_plans_updated_at 
    BEFORE UPDATE ON plans 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE plans IS 'AI implementation plans for tasks';
COMMENT ON COLUMN plans.task_id IS 'Reference to the task this plan belongs to';
COMMENT ON COLUMN plans.title IS 'Title of the implementation plan';
COMMENT ON COLUMN plans.description IS 'Detailed description of the plan';
COMMENT ON COLUMN plans.status IS 'Plan status: DRAFT, REVIEWING, APPROVED, REJECTED, EXECUTING, COMPLETED, CANCELLED';
COMMENT ON COLUMN plans.version IS 'Current version number of the plan';
COMMENT ON COLUMN plans.steps IS 'Implementation steps as JSONB array';
COMMENT ON COLUMN plans.context IS 'Plan context and metadata as JSONB object';
COMMENT ON COLUMN plans.created_by IS 'User or system that created the plan';
COMMENT ON COLUMN plans.approved_by IS 'User who approved the plan';
COMMENT ON COLUMN plans.approved_at IS 'Timestamp when plan was approved';
COMMENT ON COLUMN plans.rejected_by IS 'User who rejected the plan';
COMMENT ON COLUMN plans.rejected_at IS 'Timestamp when plan was rejected';

COMMENT ON TABLE plan_versions IS 'Version history for implementation plans';
COMMENT ON COLUMN plan_versions.plan_id IS 'Reference to the parent plan';
COMMENT ON COLUMN plan_versions.version IS 'Version number of this plan version';
COMMENT ON COLUMN plan_versions.title IS 'Title of the plan at this version';
COMMENT ON COLUMN plan_versions.description IS 'Description of the plan at this version';
COMMENT ON COLUMN plan_versions.status IS 'Status of the plan at this version';
COMMENT ON COLUMN plan_versions.steps IS 'Implementation steps at this version as JSONB array';
COMMENT ON COLUMN plan_versions.context IS 'Plan context at this version as JSONB object';
COMMENT ON COLUMN plan_versions.created_by IS 'User who created this version';
COMMENT ON COLUMN plan_versions.change_log IS 'Description of changes made in this version';

-- Create a function to automatically create plan versions on significant updates
CREATE OR REPLACE FUNCTION create_plan_version_on_update()
RETURNS TRIGGER AS $$
BEGIN
    -- Only create version if title, description, steps, or context changed
    IF OLD.title != NEW.title OR 
       OLD.description != NEW.description OR 
       OLD.steps != NEW.steps OR 
       OLD.context != NEW.context THEN
        
        INSERT INTO plan_versions (
            plan_id, 
            version, 
            title, 
            description, 
            status, 
            steps, 
            context, 
            created_by,
            change_log
        ) VALUES (
            OLD.id,
            OLD.version,
            OLD.title,
            OLD.description,
            OLD.status,
            OLD.steps,
            OLD.context,
            NEW.created_by,
            'Automatic version created on plan update'
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Note: The trigger for auto-versioning is commented out as it's handled in the application layer
-- This gives more control over when versions are created
-- CREATE TRIGGER create_plan_version_trigger
--     AFTER UPDATE ON plans
--     FOR EACH ROW EXECUTE FUNCTION create_plan_version_on_update();

-- Example JSONB structure for steps field:
-- [
--   {
--     "id": "step-1",
--     "description": "Analyze requirements",
--     "action": "analysis",
--     "parameters": {"type": "requirements"},
--     "order": 1,
--     "completed": false,
--     "completed_at": null
--   },
--   {
--     "id": "step-2", 
--     "description": "Design solution",
--     "action": "design",
--     "parameters": {"type": "technical"},
--     "order": 2,
--     "completed": true,
--     "completed_at": "2024-01-15T10:30:00Z"
--   }
-- ]

-- Example JSONB structure for context field:
-- {
--   "task_title": "Implement user authentication",
--   "task_description": "Add JWT-based authentication system",
--   "task_priority": "HIGH",
--   "estimated_hours": 8.0,
--   "requirements": ["security", "scalability"],
--   "constraints": ["existing_db_schema", "api_compatibility"]
-- }