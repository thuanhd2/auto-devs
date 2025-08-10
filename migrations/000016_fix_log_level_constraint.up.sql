-- Fix log level constraint to allow both WARN and WARNING
ALTER TABLE execution_logs DROP CONSTRAINT IF EXISTS valid_log_level;

ALTER TABLE execution_logs
ADD CONSTRAINT valid_log_level CHECK (
    log_level IN (
        'INFO',
        'WARNING',
        'WARN',
        'ERROR',
        'DEBUG'
    )
);

-- Add comment for documentation
COMMENT ON CONSTRAINT valid_log_level ON execution_logs IS 'Log level: INFO, WARNING, WARN, ERROR, DEBUG';