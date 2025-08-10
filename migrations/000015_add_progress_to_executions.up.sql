-- Add progress column to executions table
ALTER TABLE executions
ADD COLUMN progress DOUBLE PRECISION DEFAULT 0.0 CHECK (
    progress >= 0.0
    AND progress <= 1.0
);

-- Add result column to executions table
ALTER TABLE executions ADD COLUMN result JSONB;

-- Add comment for documentation
COMMENT ON COLUMN executions.progress IS 'Execution progress as decimal (0.0 to 1.0)';

COMMENT ON COLUMN executions.result IS 'Execution result as JSON data';

-- Create index for better performance on progress queries
CREATE INDEX idx_executions_progress ON executions (progress);

CREATE INDEX idx_executions_result ON executions USING GIN (result);