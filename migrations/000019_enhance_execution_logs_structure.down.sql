-- Rollback migration: 000019_enhance_execution_logs_structure

-- Drop indexes
DROP INDEX IF EXISTS idx_execution_logs_parsed_content;
DROP INDEX IF EXISTS idx_execution_logs_tool_use_id;
DROP INDEX IF EXISTS idx_execution_logs_tool_name;
DROP INDEX IF EXISTS idx_execution_logs_log_type;

-- Drop columns
ALTER TABLE execution_logs
    DROP COLUMN IF EXISTS num_turns,
    DROP COLUMN IF EXISTS duration_ms,
    DROP COLUMN IF EXISTS is_error,
    DROP COLUMN IF EXISTS parsed_content,
    DROP COLUMN IF EXISTS tool_use_id,
    DROP COLUMN IF EXISTS tool_name,
    DROP COLUMN IF EXISTS log_type;
