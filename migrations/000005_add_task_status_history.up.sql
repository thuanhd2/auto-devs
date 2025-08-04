-- Create task_status_histories table
CREATE TABLE task_status_histories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    task_id UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    from_status VARCHAR(50), -- null for initial status
    to_status VARCHAR(50) NOT NULL,
    changed_by VARCHAR(255), -- user ID or system identifier  
    reason VARCHAR(500), -- optional reason for status change
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT valid_from_status CHECK (
        from_status IS NULL OR from_status IN (
            'TODO',
            'PLANNING',
            'PLAN_REVIEWING',
            'IMPLEMENTING',
            'CODE_REVIEWING',
            'DONE',
            'CANCELLED'
        )
    ),
    CONSTRAINT valid_to_status CHECK (
        to_status IN (
            'TODO',
            'PLANNING',
            'PLAN_REVIEWING',
            'IMPLEMENTING',
            'CODE_REVIEWING',
            'DONE',
            'CANCELLED'
        )
    )
);

-- Create indexes for better performance
CREATE INDEX idx_task_status_histories_task_id ON task_status_histories (task_id);
CREATE INDEX idx_task_status_histories_to_status ON task_status_histories (to_status);
CREATE INDEX idx_task_status_histories_created_at ON task_status_histories (created_at);
CREATE INDEX idx_task_status_histories_deleted_at ON task_status_histories (deleted_at);

-- Insert initial status history for existing tasks
INSERT INTO task_status_histories (task_id, from_status, to_status, changed_by, created_at)
SELECT id, NULL, status, 'system', created_at
FROM tasks
WHERE deleted_at IS NULL;