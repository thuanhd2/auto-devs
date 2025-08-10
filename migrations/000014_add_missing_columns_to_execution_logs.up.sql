-- Add missing columns to execution_logs table
-- Migration: 000014_add_missing_columns_to_execution_logs

-- Add metadata column for additional JSON metadata
ALTER TABLE execution_logs ADD COLUMN metadata JSONB;

-- Add created_at column for tracking when the log was created
ALTER TABLE execution_logs
ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Create index for better performance when querying by created_at
CREATE INDEX IF NOT EXISTS idx_execution_logs_created_at ON execution_logs (created_at);

-- Create index for better performance when querying by metadata (GIN index for JSONB)
CREATE INDEX IF NOT EXISTS idx_execution_logs_metadata ON execution_logs USING GIN (metadata);

-- Add comments for documentation
COMMENT ON COLUMN execution_logs.metadata IS 'Additional metadata as JSON for flexible log information storage';

COMMENT ON COLUMN execution_logs.created_at IS 'Timestamp when the log entry was created';