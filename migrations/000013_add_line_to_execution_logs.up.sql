-- Add line field to execution_logs table
ALTER TABLE execution_logs ADD COLUMN line INTEGER;

-- Add updated_at field to execution_logs table
ALTER TABLE execution_logs
ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Create index for better performance when querying by execution_id and line
CREATE INDEX IF NOT EXISTS idx_execution_logs_execution_id_line ON execution_logs (execution_id, line);

-- Create trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_execution_logs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_execution_logs_updated_at
    BEFORE UPDATE ON execution_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_execution_logs_updated_at();

-- Add comment for documentation
COMMENT ON COLUMN execution_logs.line IS 'Line number in the execution output for ordering and deduplication';

COMMENT ON COLUMN execution_logs.updated_at IS 'Timestamp when the log was last updated';