-- Rollback log level constraint to original values
ALTER TABLE execution_logs DROP CONSTRAINT IF EXISTS valid_log_level;

ALTER TABLE execution_logs
ADD CONSTRAINT valid_log_level CHECK (
    log_level IN (
        'INFO',
        'WARNING',
        'ERROR',
        'DEBUG'
    )
);