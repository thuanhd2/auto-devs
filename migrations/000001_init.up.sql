-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create projects table
CREATE TABLE dax_projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    repo_url VARCHAR(500) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create tasks table
CREATE TABLE dax_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES dax_projects(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'TODO',
    branch_name VARCHAR(255),
    pull_request VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT valid_status CHECK (status IN ('TODO', 'PLANNING', 'PLAN_REVIEWING', 'IMPLEMENTING', 'CODE_REVIEWING', 'DONE', 'CANCELLED'))
);

-- Create indexes for better performance
CREATE INDEX idx_tasks_project_id ON dax_tasks(project_id);
CREATE INDEX idx_tasks_status ON dax_tasks(status);
CREATE INDEX idx_tasks_created_at ON dax_tasks(created_at);
CREATE INDEX idx_projects_created_at ON dax_projects(created_at);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for auto-updating updated_at columns
CREATE TRIGGER update_projects_updated_at 
    BEFORE UPDATE ON dax_projects 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tasks_updated_at 
    BEFORE UPDATE ON dax_tasks 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();