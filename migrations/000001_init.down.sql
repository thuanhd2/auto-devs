-- Drop triggers
DROP TRIGGER IF EXISTS update_tasks_updated_at ON dax_tasks;
DROP TRIGGER IF EXISTS update_projects_updated_at ON dax_projects;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_projects_created_at;
DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_project_id;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS dax_tasks;
DROP TABLE IF EXISTS dax_projects;

-- Drop UUID extension (only if no other tables are using it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";