-- Migration: Enhance Task Management with Priority, Tags, Relationships, and Audit Trail
-- This migration adds comprehensive task management features

-- Add new columns to tasks table
ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS priority VARCHAR(20) DEFAULT 'MEDIUM' CHECK (
    priority IN (
        'LOW',
        'MEDIUM',
        'HIGH',
        'URGENT'
    )
);

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS estimated_hours DECIMAL(5, 2) CHECK (
    estimated_hours >= 0
    AND estimated_hours <= 999.99
);

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS actual_hours DECIMAL(5, 2) CHECK (
    actual_hours >= 0
    AND actual_hours <= 999.99
);

ALTER TABLE tasks ADD COLUMN IF NOT EXISTS tags JSONB;

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS parent_task_id UUID REFERENCES tasks (id) ON DELETE SET NULL;

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS is_archived BOOLEAN DEFAULT FALSE;

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS is_template BOOLEAN DEFAULT FALSE;

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS template_id UUID REFERENCES tasks (id) ON DELETE SET NULL;

ALTER TABLE tasks ADD COLUMN IF NOT EXISTS assigned_to VARCHAR(255);

ALTER TABLE tasks ADD COLUMN IF NOT EXISTS due_date TIMESTAMP;

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks (priority);

CREATE INDEX IF NOT EXISTS idx_tasks_parent_task_id ON tasks (parent_task_id);

CREATE INDEX IF NOT EXISTS idx_tasks_is_archived ON tasks (is_archived);

CREATE INDEX IF NOT EXISTS idx_tasks_is_template ON tasks (is_template);

CREATE INDEX IF NOT EXISTS idx_tasks_assigned_to ON tasks (assigned_to);

CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks (due_date);

CREATE INDEX IF NOT EXISTS idx_tasks_estimated_hours ON tasks (estimated_hours);

CREATE INDEX IF NOT EXISTS idx_tasks_actual_hours ON tasks (actual_hours);

-- Create GIN index for tags JSONB column for efficient searching
CREATE INDEX IF NOT EXISTS idx_tasks_tags_gin ON tasks USING GIN (tags);

-- Create task_audit_logs table for tracking all task modifications
CREATE TABLE IF NOT EXISTS task_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    task_id UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    field_name VARCHAR(100),
    old_value TEXT,
    new_value TEXT,
    changed_by VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create indexes for task_audit_logs
CREATE INDEX IF NOT EXISTS idx_task_audit_logs_task_id ON task_audit_logs (task_id);

CREATE INDEX IF NOT EXISTS idx_task_audit_logs_action ON task_audit_logs (action);

CREATE INDEX IF NOT EXISTS idx_task_audit_logs_changed_by ON task_audit_logs (changed_by);

CREATE INDEX IF NOT EXISTS idx_task_audit_logs_created_at ON task_audit_logs (created_at);

-- Create task_templates table for reusable task templates
CREATE TABLE IF NOT EXISTS task_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    title VARCHAR(255) NOT NULL,
    priority VARCHAR(20) DEFAULT 'MEDIUM' CHECK (
        priority IN (
            'LOW',
            'MEDIUM',
            'HIGH',
            'URGENT'
        )
    ),
    estimated_hours DECIMAL(5, 2) CHECK (
        estimated_hours >= 0
        AND estimated_hours <= 999.99
    ),
    tags JSONB,
    is_global BOOLEAN DEFAULT FALSE,
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create indexes for task_templates
CREATE INDEX IF NOT EXISTS idx_task_templates_project_id ON task_templates (project_id);

CREATE INDEX IF NOT EXISTS idx_task_templates_is_global ON task_templates (is_global);

CREATE INDEX IF NOT EXISTS idx_task_templates_created_by ON task_templates (created_by);

CREATE INDEX IF NOT EXISTS idx_task_templates_tags_gin ON task_templates USING GIN (tags);

-- Create task_dependencies table for tracking task dependencies
CREATE TABLE IF NOT EXISTS task_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    task_id UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    depends_on_task_id UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    dependency_type VARCHAR(50) DEFAULT 'blocks' CHECK (
        dependency_type IN (
            'blocks',
            'requires',
            'related'
        )
    ),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (task_id, depends_on_task_id)
);

-- Create indexes for task_dependencies
CREATE INDEX IF NOT EXISTS idx_task_dependencies_task_id ON task_dependencies (task_id);

CREATE INDEX IF NOT EXISTS idx_task_dependencies_depends_on ON task_dependencies (depends_on_task_id);

CREATE INDEX IF NOT EXISTS idx_task_dependencies_type ON task_dependencies (dependency_type);

-- Create task_comments table for task comments
CREATE TABLE IF NOT EXISTS task_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    task_id UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    comment TEXT NOT NULL,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create indexes for task_comments
CREATE INDEX IF NOT EXISTS idx_task_comments_task_id ON task_comments (task_id);

CREATE INDEX IF NOT EXISTS idx_task_comments_created_by ON task_comments (created_by);

CREATE INDEX IF NOT EXISTS idx_task_comments_created_at ON task_comments (created_at);

-- Create task_attachments table for task file attachments
CREATE TABLE IF NOT EXISTS task_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    task_id UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100),
    uploaded_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create indexes for task_attachments
CREATE INDEX IF NOT EXISTS idx_task_attachments_task_id ON task_attachments (task_id);

CREATE INDEX IF NOT EXISTS idx_task_attachments_uploaded_by ON task_attachments (uploaded_by);

CREATE INDEX IF NOT EXISTS idx_task_attachments_created_at ON task_attachments (created_at);

-- Add constraints to prevent circular dependencies
ALTER TABLE task_dependencies
ADD CONSTRAINT check_no_self_dependency CHECK (task_id != depends_on_task_id);

-- Create function to prevent circular dependencies
CREATE OR REPLACE FUNCTION check_circular_dependencies()
RETURNS TRIGGER AS $$
BEGIN
    -- Check for circular dependencies using recursive CTE
    IF EXISTS (
        WITH RECURSIVE dependency_chain AS (
            SELECT task_id, depends_on_task_id, 1 as depth
            FROM task_dependencies
            WHERE task_id = NEW.task_id
            
            UNION ALL
            
            SELECT td.task_id, td.depends_on_task_id, dc.depth + 1
            FROM task_dependencies td
            JOIN dependency_chain dc ON td.task_id = dc.depends_on_task_id
            WHERE dc.depth < 10 -- Prevent infinite recursion
        )
        SELECT 1 FROM dependency_chain
        WHERE depends_on_task_id = NEW.task_id
    ) THEN
        RAISE EXCEPTION 'Circular dependency detected';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to check for circular dependencies
CREATE TRIGGER check_circular_dependencies_trigger
    BEFORE INSERT OR UPDATE ON task_dependencies
    FOR EACH ROW
    EXECUTE FUNCTION check_circular_dependencies();

-- Create function to automatically create audit log entries
CREATE OR REPLACE FUNCTION create_task_audit_log()
RETURNS TRIGGER AS $$
BEGIN
    -- Log task creation
    IF TG_OP = 'INSERT' THEN
        INSERT INTO task_audit_logs (task_id, action, new_value, changed_by)
        VALUES (NEW.id, 'created', 'Task created', COALESCE(NEW.assigned_to, 'system'));
        RETURN NEW;
    END IF;
    
    -- Log task updates
    IF TG_OP = 'UPDATE' THEN
        -- Log title changes
        IF OLD.title IS DISTINCT FROM NEW.title THEN
            INSERT INTO task_audit_logs (task_id, action, field_name, old_value, new_value, changed_by)
            VALUES (NEW.id, 'updated', 'title', OLD.title, NEW.title, COALESCE(NEW.assigned_to, 'system'));
        END IF;
        
        -- Log description changes
        IF OLD.description IS DISTINCT FROM NEW.description THEN
            INSERT INTO task_audit_logs (task_id, action, field_name, old_value, new_value, changed_by)
            VALUES (NEW.id, 'updated', 'description', OLD.description, NEW.description, COALESCE(NEW.assigned_to, 'system'));
        END IF;
        
        -- Log status changes
        IF OLD.status IS DISTINCT FROM NEW.status THEN
            INSERT INTO task_audit_logs (task_id, action, field_name, old_value, new_value, changed_by)
            VALUES (NEW.id, 'status_changed', 'status', OLD.status::text, NEW.status::text, COALESCE(NEW.assigned_to, 'system'));
        END IF;
        
        -- Log priority changes
        IF OLD.priority IS DISTINCT FROM NEW.priority THEN
            INSERT INTO task_audit_logs (task_id, action, field_name, old_value, new_value, changed_by)
            VALUES (NEW.id, 'updated', 'priority', OLD.priority::text, NEW.priority::text, COALESCE(NEW.assigned_to, 'system'));
        END IF;
        
        -- Log other field changes
        IF OLD.estimated_hours IS DISTINCT FROM NEW.estimated_hours THEN
            INSERT INTO task_audit_logs (task_id, action, field_name, old_value, new_value, changed_by)
            VALUES (NEW.id, 'updated', 'estimated_hours', COALESCE(OLD.estimated_hours::text, ''), COALESCE(NEW.estimated_hours::text, ''), COALESCE(NEW.assigned_to, 'system'));
        END IF;
        
        IF OLD.actual_hours IS DISTINCT FROM NEW.actual_hours THEN
            INSERT INTO task_audit_logs (task_id, action, field_name, old_value, new_value, changed_by)
            VALUES (NEW.id, 'updated', 'actual_hours', COALESCE(OLD.actual_hours::text, ''), COALESCE(NEW.actual_hours::text, ''), COALESCE(NEW.assigned_to, 'system'));
        END IF;
        
        IF OLD.due_date IS DISTINCT FROM NEW.due_date THEN
            INSERT INTO task_audit_logs (task_id, action, field_name, old_value, new_value, changed_by)
            VALUES (NEW.id, 'updated', 'due_date', COALESCE(OLD.due_date::text, ''), COALESCE(NEW.due_date::text, ''), COALESCE(NEW.assigned_to, 'system'));
        END IF;
        
        IF OLD.is_archived IS DISTINCT FROM NEW.is_archived THEN
            INSERT INTO task_audit_logs (task_id, action, field_name, old_value, new_value, changed_by)
            VALUES (NEW.id, 'archived_status_changed', 'is_archived', OLD.is_archived::text, NEW.is_archived::text, COALESCE(NEW.assigned_to, 'system'));
        END IF;
        
        RETURN NEW;
    END IF;
    
    -- Log task deletion
    IF TG_OP = 'DELETE' THEN
        INSERT INTO task_audit_logs (task_id, action, old_value, changed_by)
        VALUES (OLD.id, 'deleted', 'Task deleted', COALESCE(OLD.assigned_to, 'system'));
        RETURN OLD;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for automatic audit logging
CREATE TRIGGER task_audit_log_trigger
    AFTER INSERT OR UPDATE OR DELETE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION create_task_audit_log();

-- Create function to update task statistics
CREATE OR REPLACE FUNCTION update_task_statistics()
RETURNS TRIGGER AS $$
BEGIN
    -- Update project task count when tasks are created/deleted
    IF TG_OP = 'INSERT' THEN
        UPDATE projects 
        SET updated_at = NOW()
        WHERE id = NEW.project_id;
        RETURN NEW;
    END IF;
    
    IF TG_OP = 'DELETE' THEN
        UPDATE projects 
        SET updated_at = NOW()
        WHERE id = OLD.project_id;
        RETURN OLD;
    END IF;
    
    -- Update project when task status changes
    IF TG_OP = 'UPDATE' AND OLD.status IS DISTINCT FROM NEW.status THEN
        UPDATE projects 
        SET updated_at = NOW()
        WHERE id = NEW.project_id;
        RETURN NEW;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for updating project statistics
CREATE TRIGGER update_project_statistics_trigger
    AFTER INSERT OR UPDATE OR DELETE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_task_statistics();