-- Enhance execution_logs to store structured parsed data
-- Migration: 000019_enhance_execution_logs_structure

ALTER TABLE execution_logs
    ADD COLUMN IF NOT EXISTS log_type VARCHAR(20),
    ADD COLUMN IF NOT EXISTS tool_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS tool_use_id VARCHAR(100),
    ADD COLUMN IF NOT EXISTS parsed_content JSONB,
    ADD COLUMN IF NOT EXISTS is_error BOOLEAN,
    ADD COLUMN IF NOT EXISTS duration_ms INTEGER,
    ADD COLUMN IF NOT EXISTS num_turns INTEGER;

-- Indexes for queryability
CREATE INDEX IF NOT EXISTS idx_execution_logs_log_type ON execution_logs (log_type);
CREATE INDEX IF NOT EXISTS idx_execution_logs_tool_name ON execution_logs (tool_name);
CREATE INDEX IF NOT EXISTS idx_execution_logs_tool_use_id ON execution_logs (tool_use_id);
CREATE INDEX IF NOT EXISTS idx_execution_logs_parsed_content ON execution_logs USING GIN (parsed_content);

-- Comments for documentation
COMMENT ON COLUMN execution_logs.log_type IS 'Semantic type of log line: user, assistant, tool_result, result, system, etc.';
COMMENT ON COLUMN execution_logs.tool_name IS 'Name of the tool used by assistant when applicable';
COMMENT ON COLUMN execution_logs.tool_use_id IS 'Correlation ID for tool use/result pairing';
COMMENT ON COLUMN execution_logs.parsed_content IS 'Structured JSON parsed from raw message';
COMMENT ON COLUMN execution_logs.is_error IS 'Flag indicating this log represents an error state';
COMMENT ON COLUMN execution_logs.duration_ms IS 'Duration in milliseconds for result/summary logs';
COMMENT ON COLUMN execution_logs.num_turns IS 'Number of conversation turns if available';
