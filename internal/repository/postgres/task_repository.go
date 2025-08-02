package postgres

import (
	"context"
	"fmt"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type taskRepository struct {
	db *database.GormDB
}

// NewTaskRepository creates a new PostgreSQL task repository
func NewTaskRepository(db *database.GormDB) repository.TaskRepository {
	return &taskRepository{db: db}
}

// Create creates a new task
func (r *taskRepository) Create(ctx context.Context, task *entity.Task) error {
	// Generate UUID if not provided
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}

	// Set default status if not provided
	if task.Status == "" {
		task.Status = entity.TaskStatusTodo
	}

	result := r.db.WithContext(ctx).Create(task)
	if result.Error != nil {
		return fmt.Errorf("failed to create task: %w", result.Error)
	}

	return nil
}

// GetByID retrieves a task by ID
func (r *taskRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Task, error) {
	var task entity.Task

	result := r.db.WithContext(ctx).First(&task, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get task: %w", result.Error)
	}

	return &task, nil
}

// GetByProjectID retrieves all tasks for a specific project
func (r *taskRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error) {
	var tasks []entity.Task

	result := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get tasks by project: %w", result.Error)
	}

	// Convert to slice of pointers
	taskPtrs := make([]*entity.Task, len(tasks))
	for i := range tasks {
		taskPtrs[i] = &tasks[i]
	}

	return taskPtrs, nil
}

// Update updates an existing task
func (r *taskRepository) Update(ctx context.Context, task *entity.Task) error {
	// First check if task exists
	var existingTask entity.Task
	result := r.db.WithContext(ctx).First(&existingTask, "id = ?", task.ID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return fmt.Errorf("task not found with id %s", task.ID)
		}
		return fmt.Errorf("failed to check task existence: %w", result.Error)
	}

	// Update the task
	result = r.db.WithContext(ctx).Save(task)
	if result.Error != nil {
		return fmt.Errorf("failed to update task: %w", result.Error)
	}

	return nil
}

// Delete deletes a task by ID (soft delete)
func (r *taskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Task{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("task not found with id %s", id)
	}

	return nil
}

// UpdateStatus updates the status of a task
func (r *taskRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TaskStatus) error {
	result := r.db.WithContext(ctx).Model(&entity.Task{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update task status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("task not found with id %s", id)
	}

	return nil
}

// GetByStatus retrieves all tasks with a specific status
func (r *taskRepository) GetByStatus(ctx context.Context, status entity.TaskStatus) ([]*entity.Task, error) {
	var tasks []entity.Task

	result := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC").Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get tasks by status: %w", result.Error)
	}

	// Convert to slice of pointers
	taskPtrs := make([]*entity.Task, len(tasks))
	for i := range tasks {
		taskPtrs[i] = &tasks[i]
	}

	return taskPtrs, nil
}
