package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

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
		task.Status = entity.TaskStatusTODO
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

// UpdateStatusWithHistory updates a task status and creates a history record
func (r *taskRepository) UpdateStatusWithHistory(ctx context.Context, id uuid.UUID, status entity.TaskStatus, changedBy *string, reason *string) error {
	// Get current task to validate transition
	currentTask, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get current task: %w", err)
	}

	// Validate status transition
	if err := entity.ValidateStatusTransition(currentTask.Status, status); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Start transaction
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update task status
		result := tx.Model(&entity.Task{}).Where("id = ?", id).Update("status", status)
		if result.Error != nil {
			return fmt.Errorf("failed to update task status: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("task not found with id %s", id)
		}

		// Create status history record
		history := &entity.TaskStatusHistory{
			TaskID:     id,
			FromStatus: &currentTask.Status,
			ToStatus:   status,
			ChangedBy:  changedBy,
			Reason:     reason,
		}

		if err := tx.Create(history).Error; err != nil {
			return fmt.Errorf("failed to create status history: %w", err)
		}

		return nil
	})
}

// GetByStatuses retrieves all tasks with specific statuses
func (r *taskRepository) GetByStatuses(ctx context.Context, statuses []entity.TaskStatus) ([]*entity.Task, error) {
	var tasks []entity.Task

	result := r.db.WithContext(ctx).Where("status IN ?", statuses).Order("created_at DESC").Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get tasks by statuses: %w", result.Error)
	}

	// Convert to slice of pointers
	taskPtrs := make([]*entity.Task, len(tasks))
	for i := range tasks {
		taskPtrs[i] = &tasks[i]
	}

	return taskPtrs, nil
}

// BulkUpdateStatus updates status for multiple tasks
func (r *taskRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.TaskStatus, changedBy *string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get current tasks to validate transitions
		var currentTasks []entity.Task
		if err := tx.Where("id IN ?", ids).Find(&currentTasks).Error; err != nil {
			return fmt.Errorf("failed to get current tasks: %w", err)
		}

		// Validate all transitions first
		for _, task := range currentTasks {
			if err := entity.ValidateStatusTransition(task.Status, status); err != nil {
				return fmt.Errorf("invalid status transition for task %s: %w", task.ID, err)
			}
		}

		// Update all tasks
		result := tx.Model(&entity.Task{}).Where("id IN ?", ids).Update("status", status)
		if result.Error != nil {
			return fmt.Errorf("failed to bulk update task status: %w", result.Error)
		}

		// Create history records for each task
		for _, task := range currentTasks {
			history := &entity.TaskStatusHistory{
				TaskID:     task.ID,
				FromStatus: &task.Status,
				ToStatus:   status,
				ChangedBy:  changedBy,
			}

			if err := tx.Create(history).Error; err != nil {
				return fmt.Errorf("failed to create status history for task %s: %w", task.ID, err)
			}
		}

		return nil
	})
}

// GetStatusHistory retrieves status history for a task
func (r *taskRepository) GetStatusHistory(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskStatusHistory, error) {
	var history []entity.TaskStatusHistory

	result := r.db.WithContext(ctx).Where("task_id = ?", taskID).Order("created_at ASC").Find(&history)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get status history: %w", result.Error)
	}

	// Convert to slice of pointers
	historyPtrs := make([]*entity.TaskStatusHistory, len(history))
	for i := range history {
		historyPtrs[i] = &history[i]
	}

	return historyPtrs, nil
}

// GetStatusAnalytics generates status analytics for a project
func (r *taskRepository) GetStatusAnalytics(ctx context.Context, projectID uuid.UUID) (*entity.TaskStatusAnalytics, error) {
	analytics := &entity.TaskStatusAnalytics{
		ProjectID:   projectID,
		GeneratedAt: time.Now(),
	}

	// Get status distribution
	var statusStats []entity.TaskStatusStats
	result := r.db.WithContext(ctx).
		Model(&entity.Task{}).
		Select("status, count(*) as count").
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Group("status").
		Find(&statusStats)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get status distribution: %w", result.Error)
	}

	analytics.StatusDistribution = statusStats

	// Calculate total and completed tasks
	analytics.TotalTasks = 0
	for _, stat := range statusStats {
		analytics.TotalTasks += stat.Count
		if stat.Status == entity.TaskStatusDONE {
			analytics.CompletedTasks = stat.Count
		}
	}

	// Calculate completion rate
	if analytics.TotalTasks > 0 {
		analytics.CompletionRate = float64(analytics.CompletedTasks) / float64(analytics.TotalTasks) * 100
	}

	// Get average time in status
	analytics.AverageTimeInStatus = make(map[entity.TaskStatus]float64)
	
	// Get transition counts
	analytics.TransitionCount = make(map[string]int)
	var transitions []struct {
		FromStatus *string
		ToStatus   string
		Count      int
	}

	transitionQuery := `
		SELECT 
			from_status,
			to_status,
			COUNT(*) as count
		FROM task_status_history 
		WHERE task_id IN (SELECT id FROM tasks WHERE project_id = ? AND deleted_at IS NULL)
		AND deleted_at IS NULL
		GROUP BY from_status, to_status
	`

	if err := r.db.WithContext(ctx).Raw(transitionQuery, projectID).Scan(&transitions).Error; err != nil {
		return nil, fmt.Errorf("failed to get transition counts: %w", err)
	}

	for _, t := range transitions {
		fromStatus := "INITIAL"
		if t.FromStatus != nil {
			fromStatus = *t.FromStatus
		}
		key := fmt.Sprintf("%s->%s", fromStatus, t.ToStatus)
		analytics.TransitionCount[key] = t.Count
	}

	return analytics, nil
}

// GetTasksWithFilters retrieves tasks with various filtering options
func (r *taskRepository) GetTasksWithFilters(ctx context.Context, filters repository.TaskFilters) ([]*entity.Task, error) {
	query := r.db.WithContext(ctx).Model(&entity.Task{})

	// Apply filters
	if filters.ProjectID != nil {
		query = query.Where("project_id = ?", *filters.ProjectID)
	}

	if len(filters.Statuses) > 0 {
		query = query.Where("status IN ?", filters.Statuses)
	}

	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
	}

	if filters.SearchTerm != nil && *filters.SearchTerm != "" {
		searchPattern := "%" + strings.ToLower(*filters.SearchTerm) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
	}

	// Apply ordering
	if filters.OrderBy != nil {
		orderDir := "ASC"
		if filters.OrderDir != nil && strings.ToUpper(*filters.OrderDir) == "DESC" {
			orderDir = "DESC"
		}
		
		validOrderBy := map[string]bool{
			"created_at": true,
			"updated_at": true,
			"title":      true,
			"status":     true,
		}
		
		if validOrderBy[*filters.OrderBy] {
			query = query.Order(fmt.Sprintf("%s %s", *filters.OrderBy, orderDir))
		}
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filters.Offset != nil {
		query = query.Offset(*filters.Offset)
	}

	if filters.Limit != nil {
		query = query.Limit(*filters.Limit)
	}

	var tasks []entity.Task
	result := query.Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get tasks with filters: %w", result.Error)
	}

	// Convert to slice of pointers
	taskPtrs := make([]*entity.Task, len(tasks))
	for i := range tasks {
		taskPtrs[i] = &tasks[i]
	}

	return taskPtrs, nil
}
