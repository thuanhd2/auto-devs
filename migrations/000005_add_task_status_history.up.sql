-- Create task_status_history table
CREATE TABLE task_status_history (
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
CREATE INDEX idx_task_status_history_task_id ON task_status_history (task_id);
CREATE INDEX idx_task_status_history_to_status ON task_status_history (to_status);
CREATE INDEX idx_task_status_history_created_at ON task_status_history (created_at);
CREATE INDEX idx_task_status_history_deleted_at ON task_status_history (deleted_at);

-- Function to automatically create status history when task status changes
CREATE OR REPLACE FUNCTION track_task_status_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Only track if status actually changed
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO task_status_history (task_id, from_status, to_status, changed_by, created_at)
        VALUES (NEW.id, OLD.status, NEW.status, 'system', NOW());
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically track status changes
CREATE TRIGGER track_task_status_change_trigger
    AFTER UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION track_task_status_change();

-- Insert initial status history for existing tasks
INSERT INTO task_status_history (task_id, from_status, to_status, changed_by, created_at)
SELECT id, NULL, status, 'system', created_at
FROM tasks
WHERE deleted_at IS NULL;