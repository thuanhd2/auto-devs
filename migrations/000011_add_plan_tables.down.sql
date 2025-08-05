-- Drop triggers first
DROP TRIGGER IF EXISTS create_initial_plan_version ON plans;
DROP TRIGGER IF EXISTS update_plans_updated_at ON plans;

-- Drop functions
DROP FUNCTION IF EXISTS create_initial_plan_version();
DROP FUNCTION IF EXISTS update_plans_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_plans_content_fts;
DROP INDEX IF EXISTS idx_plan_versions_deleted_at;
DROP INDEX IF EXISTS idx_plan_versions_created_at;
DROP INDEX IF EXISTS idx_plan_versions_version;
DROP INDEX IF EXISTS idx_plan_versions_plan_id;
DROP INDEX IF EXISTS idx_plans_deleted_at;
DROP INDEX IF EXISTS idx_plans_created_at;
DROP INDEX IF EXISTS idx_plans_status;
DROP INDEX IF EXISTS idx_plans_task_id;

-- Drop tables (order matters due to foreign key constraints)
DROP TABLE IF EXISTS plan_versions;
DROP TABLE IF EXISTS plans;