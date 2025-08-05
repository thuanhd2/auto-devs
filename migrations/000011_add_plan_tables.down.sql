-- Drop the trigger and function
DROP TRIGGER IF EXISTS create_plan_version_trigger ON plans;
DROP FUNCTION IF EXISTS create_plan_version_on_update();

-- Drop indexes
DROP INDEX IF EXISTS idx_plans_task_id;
DROP INDEX IF EXISTS idx_plans_status;
DROP INDEX IF EXISTS idx_plans_created_by;
DROP INDEX IF EXISTS idx_plans_approved_by;
DROP INDEX IF EXISTS idx_plans_rejected_by;
DROP INDEX IF EXISTS idx_plans_created_at;
DROP INDEX IF EXISTS idx_plans_updated_at;
DROP INDEX IF EXISTS idx_plans_deleted_at;

DROP INDEX IF EXISTS idx_plan_versions_plan_id;
DROP INDEX IF EXISTS idx_plan_versions_version;
DROP INDEX IF EXISTS idx_plan_versions_created_by;
DROP INDEX IF EXISTS idx_plan_versions_created_at;
DROP INDEX IF EXISTS idx_plan_versions_unique;

-- Drop GIN indexes for JSONB fields
DROP INDEX IF EXISTS idx_plans_steps_gin;
DROP INDEX IF EXISTS idx_plans_context_gin;
DROP INDEX IF EXISTS idx_plan_versions_steps_gin;
DROP INDEX IF EXISTS idx_plan_versions_context_gin;

-- Drop triggers
DROP TRIGGER IF EXISTS update_plans_updated_at ON plans;

-- Drop tables (plan_versions first due to foreign key constraint)
DROP TABLE IF EXISTS plan_versions;
DROP TABLE IF EXISTS plans;