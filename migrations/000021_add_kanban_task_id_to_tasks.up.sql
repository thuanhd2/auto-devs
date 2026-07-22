ALTER TABLE tasks ADD COLUMN kanban_task_id VARCHAR(64);
CREATE INDEX idx_tasks_kanban_task_id ON tasks(kanban_task_id) WHERE kanban_task_id IS NOT NULL;
