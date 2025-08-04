-- Create executions table for AI execution tracking
CREATE TABLE executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    process_id INTEGER,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    exit_code INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Create processes table for process management tracking  
CREATE TABLE processes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    process_id INTEGER NOT NULL,
    command TEXT NOT NULL,
    working_directory TEXT NOT NULL,
    environment_vars JSONB,
    resource_usage JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'RUNNING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Create execution_logs table for execution logging
CREATE TABLE execution_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    log_level VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    source TEXT NOT NULL
);

-- Create indexes for better performance
CREATE INDEX idx_executions_task_id ON executions(task_id);
CREATE INDEX idx_executions_status ON executions(status);
CREATE INDEX idx_executions_process_id ON executions(process_id);
CREATE INDEX idx_executions_started_at ON executions(started_at);

CREATE INDEX idx_processes_execution_id ON processes(execution_id);
CREATE INDEX idx_processes_process_id ON processes(process_id);
CREATE INDEX idx_processes_status ON processes(status);

CREATE INDEX idx_execution_logs_execution_id ON execution_logs(execution_id);
CREATE INDEX idx_execution_logs_log_level ON execution_logs(log_level);
CREATE INDEX idx_execution_logs_timestamp ON execution_logs(timestamp);
CREATE INDEX idx_execution_logs_source ON execution_logs(source);

-- Add check constraints for valid status values
ALTER TABLE executions 
ADD CONSTRAINT valid_execution_status CHECK (
    status IN ('PENDING', 'RUNNING', 'COMPLETED', 'FAILED', 'CANCELLED')
);

ALTER TABLE processes
ADD CONSTRAINT valid_process_status CHECK (
    status IN ('RUNNING', 'TERMINATED', 'KILLED')
);

ALTER TABLE execution_logs
ADD CONSTRAINT valid_log_level CHECK (
    log_level IN ('INFO', 'WARNING', 'ERROR', 'DEBUG')
);

-- Create triggers for auto-updating updated_at columns
CREATE TRIGGER update_executions_updated_at 
    BEFORE UPDATE ON executions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_processes_updated_at 
    BEFORE UPDATE ON processes 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE executions IS 'AI execution tracking for task automation';
COMMENT ON COLUMN executions.task_id IS 'Reference to the task being executed';
COMMENT ON COLUMN executions.status IS 'Execution status: PENDING, RUNNING, COMPLETED, FAILED, CANCELLED';
COMMENT ON COLUMN executions.process_id IS 'Operating system process ID';
COMMENT ON COLUMN executions.started_at IS 'Timestamp when execution started';
COMMENT ON COLUMN executions.completed_at IS 'Timestamp when execution completed';
COMMENT ON COLUMN executions.error_message IS 'Error message if execution failed';
COMMENT ON COLUMN executions.exit_code IS 'Process exit code';

COMMENT ON TABLE processes IS 'Process management tracking for AI CLI tools';
COMMENT ON COLUMN processes.execution_id IS 'Reference to the parent execution';
COMMENT ON COLUMN processes.process_id IS 'Operating system process ID';
COMMENT ON COLUMN processes.command IS 'Full CLI command executed';
COMMENT ON COLUMN processes.working_directory IS 'Working directory for process execution';
COMMENT ON COLUMN processes.environment_vars IS 'Environment variables as JSON';
COMMENT ON COLUMN processes.resource_usage IS 'Resource usage metrics as JSON (cpu_percent, memory_mb, etc.)';
COMMENT ON COLUMN processes.status IS 'Process status: RUNNING, TERMINATED, KILLED';

COMMENT ON TABLE execution_logs IS 'Execution logs for AI processes';
COMMENT ON COLUMN execution_logs.execution_id IS 'Reference to the parent execution';
COMMENT ON COLUMN execution_logs.log_level IS 'Log level: INFO, WARNING, ERROR, DEBUG';
COMMENT ON COLUMN execution_logs.message IS 'Log message content';
COMMENT ON COLUMN execution_logs.timestamp IS 'Log entry timestamp';
COMMENT ON COLUMN execution_logs.source IS 'Log source: stdout, stderr, system';