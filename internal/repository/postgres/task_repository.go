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
func (r *taskRepository) GetTasksWithFilters(ctx context.Context, filters entity.TaskFilters) ([]*entity.Task, error) {
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

// SearchTasks performs full-text search on tasks
func (r *taskRepository) SearchTasks(ctx context.Context, query string, projectID *uuid.UUID) ([]*entity.TaskSearchResult, error) {
	searchQuery := r.db.WithContext(ctx).Model(&entity.Task{}).
		Select("*, ts_rank(to_tsvector('english', title || ' ' || COALESCE(description, '')), plainto_tsquery('english', ?)) as rank", query).
		Where("to_tsvector('english', title || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', ?)", query)

	if projectID != nil {
		searchQuery = searchQuery.Where("project_id = ?", *projectID)
	}

	searchQuery = searchQuery.Order("rank DESC")

	var tasks []entity.Task
	if err := searchQuery.Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}

	results := make([]*entity.TaskSearchResult, len(tasks))
	for i, task := range tasks {
		results[i] = &entity.TaskSearchResult{
			Task:    &task,
			Score:   0.8, // Placeholder score
			Matched: "title",
		}
	}

	return results, nil
}

// GetTasksByPriority retrieves tasks by priority level
func (r *taskRepository) GetTasksByPriority(ctx context.Context, priority entity.TaskPriority) ([]*entity.Task, error) {
	var tasks []entity.Task

	result := r.db.WithContext(ctx).Where("priority = ?", priority).Order("created_at DESC").Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get tasks by priority: %w", result.Error)
	}

	taskPtrs := make([]*entity.Task, len(tasks))
	for i := range tasks {
		taskPtrs[i] = &tasks[i]
	}

	return taskPtrs, nil
}

// GetTasksByTags retrieves tasks that have any of the specified tags
func (r *taskRepository) GetTasksByTags(ctx context.Context, tags []string) ([]*entity.Task, error) {
	var tasks []entity.Task

	// Using JSONB containment operator
	tagConditions := make([]string, len(tags))
	args := make([]interface{}, len(tags))
	for i, tag := range tags {
		tagConditions[i] = "tags @> ?"
		args[i] = fmt.Sprintf(`["%s"]`, tag)
	}

	query := r.db.WithContext(ctx).Where(strings.Join(tagConditions, " OR "), args...)
	result := query.Order("created_at DESC").Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get tasks by tags: %w", result.Error)
	}

	taskPtrs := make([]*entity.Task, len(tasks))
	for i := range tasks {
		taskPtrs[i] = &tasks[i]
	}

	return taskPtrs, nil
}

// GetArchivedTasks retrieves archived tasks
func (r *taskRepository) GetArchivedTasks(ctx context.Context, projectID *uuid.UUID) ([]*entity.Task, error) {
	query := r.db.WithContext(ctx).Where("is_archived = ?", true)

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	var tasks []entity.Task
	result := query.Order("created_at DESC").Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get archived tasks: %w", result.Error)
	}

	taskPtrs := make([]*entity.Task, len(tasks))
	for i := range tasks {
		taskPtrs[i] = &tasks[i]
	}

	return taskPtrs, nil
}

// GetTasksWithSubtasks retrieves tasks with their subtasks
func (r *taskRepository) GetTasksWithSubtasks(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error) {
	var tasks []entity.Task

	result := r.db.WithContext(ctx).Preload("Subtasks").Where("project_id = ?", projectID).Order("created_at DESC").Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get tasks with subtasks: %w", result.Error)
	}

	taskPtrs := make([]*entity.Task, len(tasks))
	for i := range tasks {
		taskPtrs[i] = &tasks[i]
	}

	return taskPtrs, nil
}

// GetSubtasks retrieves all subtasks of a parent task
func (r *taskRepository) GetSubtasks(ctx context.Context, parentTaskID uuid.UUID) ([]*entity.Task, error) {
	var tasks []entity.Task

	result := r.db.WithContext(ctx).Where("parent_task_id = ?", parentTaskID).Order("created_at ASC").Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get subtasks: %w", result.Error)
	}

	taskPtrs := make([]*entity.Task, len(tasks))
	for i := range tasks {
		taskPtrs[i] = &tasks[i]
	}

	return taskPtrs, nil
}

// GetParentTask retrieves the parent task of a subtask
func (r *taskRepository) GetParentTask(ctx context.Context, taskID uuid.UUID) (*entity.Task, error) {
	var task entity.Task

	result := r.db.WithContext(ctx).Joins("JOIN tasks subtask ON subtask.parent_task_id = tasks.id").
		Where("subtask.id = ?", taskID).First(&task)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("parent task not found for task %s", taskID)
		}
		return nil, fmt.Errorf("failed to get parent task: %w", result.Error)
	}

	return &task, nil
}

// UpdateParentTask updates the parent task relationship
func (r *taskRepository) UpdateParentTask(ctx context.Context, taskID uuid.UUID, parentTaskID *uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&entity.Task{}).Where("id = ?", taskID).Update("parent_task_id", parentTaskID)
	if result.Error != nil {
		return fmt.Errorf("failed to update parent task: %w", result.Error)
	}

	return nil
}

// BulkDelete deletes multiple tasks
func (r *taskRepository) BulkDelete(ctx context.Context, taskIDs []uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id IN ?", taskIDs).Delete(&entity.Task{})
	if result.Error != nil {
		return fmt.Errorf("failed to bulk delete tasks: %w", result.Error)
	}

	return nil
}

// BulkArchive archives multiple tasks
func (r *taskRepository) BulkArchive(ctx context.Context, taskIDs []uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&entity.Task{}).Where("id IN ?", taskIDs).Update("is_archived", true)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk archive tasks: %w", result.Error)
	}

	return nil
}

// BulkUnarchive unarchives multiple tasks
func (r *taskRepository) BulkUnarchive(ctx context.Context, taskIDs []uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&entity.Task{}).Where("id IN ?", taskIDs).Update("is_archived", false)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk unarchive tasks: %w", result.Error)
	}

	return nil
}

// BulkUpdatePriority updates priority for multiple tasks
func (r *taskRepository) BulkUpdatePriority(ctx context.Context, taskIDs []uuid.UUID, priority entity.TaskPriority) error {
	result := r.db.WithContext(ctx).Model(&entity.Task{}).Where("id IN ?", taskIDs).Update("priority", priority)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk update priority: %w", result.Error)
	}

	return nil
}

// BulkAssign assigns multiple tasks to a user
func (r *taskRepository) BulkAssign(ctx context.Context, taskIDs []uuid.UUID, assignedTo string) error {
	result := r.db.WithContext(ctx).Model(&entity.Task{}).Where("id IN ?", taskIDs).Update("assigned_to", assignedTo)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk assign tasks: %w", result.Error)
	}

	return nil
}

// CreateTemplate creates a new task template
func (r *taskRepository) CreateTemplate(ctx context.Context, template *entity.TaskTemplate) error {
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}

	result := r.db.WithContext(ctx).Create(template)
	if result.Error != nil {
		return fmt.Errorf("failed to create template: %w", result.Error)
	}

	return nil
}

// GetTemplates retrieves task templates
func (r *taskRepository) GetTemplates(ctx context.Context, projectID uuid.UUID, includeGlobal bool) ([]*entity.TaskTemplate, error) {
	query := r.db.WithContext(ctx).Model(&entity.TaskTemplate{})

	if includeGlobal {
		query = query.Where("project_id = ? OR is_global = ?", projectID, true)
	} else {
		query = query.Where("project_id = ?", projectID)
	}

	var templates []entity.TaskTemplate
	result := query.Order("created_at DESC").Find(&templates)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get templates: %w", result.Error)
	}

	templatePtrs := make([]*entity.TaskTemplate, len(templates))
	for i := range templates {
		templatePtrs[i] = &templates[i]
	}

	return templatePtrs, nil
}

// GetTemplateByID retrieves a specific template
func (r *taskRepository) GetTemplateByID(ctx context.Context, id uuid.UUID) (*entity.TaskTemplate, error) {
	var template entity.TaskTemplate

	result := r.db.WithContext(ctx).First(&template, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("template not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get template: %w", result.Error)
	}

	return &template, nil
}

// UpdateTemplate updates a task template
func (r *taskRepository) UpdateTemplate(ctx context.Context, template *entity.TaskTemplate) error {
	result := r.db.WithContext(ctx).Save(template)
	if result.Error != nil {
		return fmt.Errorf("failed to update template: %w", result.Error)
	}

	return nil
}

// DeleteTemplate deletes a task template
func (r *taskRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.TaskTemplate{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete template: %w", result.Error)
	}

	return nil
}

// CreateTaskFromTemplate creates a new task from a template
func (r *taskRepository) CreateTaskFromTemplate(ctx context.Context, templateID uuid.UUID, projectID uuid.UUID, createdBy string) (*entity.Task, error) {
	template, err := r.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	task := &entity.Task{
		ID:             uuid.New(),
		ProjectID:      projectID,
		Title:          template.Title,
		Description:    template.Description,
		Status:         entity.TaskStatusTODO,
		Priority:       template.Priority,
		EstimatedHours: template.EstimatedHours,
		Tags:           template.Tags,
		TemplateID:     &templateID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := r.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

// GetAuditLogs retrieves audit logs for a task
func (r *taskRepository) GetAuditLogs(ctx context.Context, taskID uuid.UUID, limit *int) ([]*entity.TaskAuditLog, error) {
	query := r.db.WithContext(ctx).Where("task_id = ?", taskID).Order("created_at DESC")

	if limit != nil {
		query = query.Limit(*limit)
	}

	var logs []entity.TaskAuditLog
	result := query.Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", result.Error)
	}

	logPtrs := make([]*entity.TaskAuditLog, len(logs))
	for i := range logs {
		logPtrs[i] = &logs[i]
	}

	return logPtrs, nil
}

// CreateAuditLog creates a new audit log entry
func (r *taskRepository) CreateAuditLog(ctx context.Context, auditLog *entity.TaskAuditLog) error {
	if auditLog.ID == uuid.Nil {
		auditLog.ID = uuid.New()
	}

	result := r.db.WithContext(ctx).Create(auditLog)
	if result.Error != nil {
		return fmt.Errorf("failed to create audit log: %w", result.Error)
	}

	return nil
}

// GetTaskStatistics retrieves comprehensive task statistics
func (r *taskRepository) GetTaskStatistics(ctx context.Context, projectID uuid.UUID) (*entity.TaskStatistics, error) {
	stats := &entity.TaskStatistics{
		ProjectID:       projectID,
		TasksByPriority: make(map[entity.TaskPriority]int),
		TasksByStatus:   make(map[entity.TaskStatus]int),
		GeneratedAt:     time.Now(),
	}

	// Get total tasks
	var totalTasks int64
	if err := r.db.WithContext(ctx).Model(&entity.Task{}).Where("project_id = ?", projectID).Count(&totalTasks).Error; err != nil {
		return nil, fmt.Errorf("failed to count total tasks: %w", err)
	}
	stats.TotalTasks = int(totalTasks)

	// Get completed tasks
	var completedTasks int64
	if err := r.db.WithContext(ctx).Model(&entity.Task{}).Where("project_id = ? AND status = ?", projectID, entity.TaskStatusDONE).Count(&completedTasks).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed tasks: %w", err)
	}
	stats.CompletedTasks = int(completedTasks)

	// Get in progress tasks
	var inProgressTasks int64
	if err := r.db.WithContext(ctx).Model(&entity.Task{}).Where("project_id = ? AND status IN ?", projectID, []entity.TaskStatus{entity.TaskStatusIMPLEMENTING, entity.TaskStatusCODEREVIEWING}).Count(&inProgressTasks).Error; err != nil {
		return nil, fmt.Errorf("failed to count in progress tasks: %w", err)
	}
	stats.InProgressTasks = int(inProgressTasks)

	// Get archived tasks
	var archivedTasks int64
	if err := r.db.WithContext(ctx).Model(&entity.Task{}).Where("project_id = ? AND is_archived = ?", projectID, true).Count(&archivedTasks).Error; err != nil {
		return nil, fmt.Errorf("failed to count archived tasks: %w", err)
	}
	stats.ArchivedTasks = int(archivedTasks)

	// Calculate completion rate (stored in AverageCompletionTime field for now)
	if stats.TotalTasks > 0 {
		stats.AverageCompletionTime = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	return stats, nil
}

// AddDependency adds a dependency between tasks
func (r *taskRepository) AddDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID, dependencyType string) error {
	dependency := &entity.TaskDependency{
		ID:              uuid.New(),
		TaskID:          taskID,
		DependsOnTaskID: dependsOnTaskID,
		DependencyType:  dependencyType,
		CreatedAt:       time.Now(),
	}

	result := r.db.WithContext(ctx).Create(dependency)
	if result.Error != nil {
		return fmt.Errorf("failed to add dependency: %w", result.Error)
	}

	return nil
}

// RemoveDependency removes a dependency between tasks
func (r *taskRepository) RemoveDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("task_id = ? AND depends_on_task_id = ?", taskID, dependsOnTaskID).Delete(&entity.TaskDependency{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove dependency: %w", result.Error)
	}

	return nil
}

// GetDependencies retrieves dependencies for a task
func (r *taskRepository) GetDependencies(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskDependency, error) {
	var dependencies []entity.TaskDependency

	result := r.db.WithContext(ctx).Where("task_id = ?", taskID).Find(&dependencies)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", result.Error)
	}

	dependencyPtrs := make([]*entity.TaskDependency, len(dependencies))
	for i := range dependencies {
		dependencyPtrs[i] = &dependencies[i]
	}

	return dependencyPtrs, nil
}

// GetDependents retrieves tasks that depend on the given task
func (r *taskRepository) GetDependents(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskDependency, error) {
	var dependencies []entity.TaskDependency

	result := r.db.WithContext(ctx).Where("depends_on_task_id = ?", taskID).Find(&dependencies)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get dependents: %w", result.Error)
	}

	dependencyPtrs := make([]*entity.TaskDependency, len(dependencies))
	for i := range dependencies {
		dependencyPtrs[i] = &dependencies[i]
	}

	return dependencyPtrs, nil
}

// AddComment adds a comment to a task
func (r *taskRepository) AddComment(ctx context.Context, comment *entity.TaskComment) error {
	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}

	result := r.db.WithContext(ctx).Create(comment)
	if result.Error != nil {
		return fmt.Errorf("failed to add comment: %w", result.Error)
	}

	return nil
}

// GetComments retrieves comments for a task
func (r *taskRepository) GetComments(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskComment, error) {
	var comments []entity.TaskComment

	result := r.db.WithContext(ctx).Where("task_id = ?", taskID).Order("created_at ASC").Find(&comments)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get comments: %w", result.Error)
	}

	commentPtrs := make([]*entity.TaskComment, len(comments))
	for i := range comments {
		commentPtrs[i] = &comments[i]
	}

	return commentPtrs, nil
}

// UpdateComment updates a comment
func (r *taskRepository) UpdateComment(ctx context.Context, comment *entity.TaskComment) error {
	result := r.db.WithContext(ctx).Save(comment)
	if result.Error != nil {
		return fmt.Errorf("failed to update comment: %w", result.Error)
	}

	return nil
}

// DeleteComment deletes a comment
func (r *taskRepository) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.TaskComment{}, "id = ?", commentID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete comment: %w", result.Error)
	}

	return nil
}

// AddAttachment adds a file attachment to a task
func (r *taskRepository) AddAttachment(ctx context.Context, attachment *entity.TaskAttachment) error {
	if attachment.ID == uuid.Nil {
		attachment.ID = uuid.New()
	}

	result := r.db.WithContext(ctx).Create(attachment)
	if result.Error != nil {
		return fmt.Errorf("failed to add attachment: %w", result.Error)
	}

	return nil
}

// GetAttachments retrieves attachments for a task
func (r *taskRepository) GetAttachments(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskAttachment, error) {
	var attachments []entity.TaskAttachment

	result := r.db.WithContext(ctx).Where("task_id = ?", taskID).Order("created_at ASC").Find(&attachments)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get attachments: %w", result.Error)
	}

	attachmentPtrs := make([]*entity.TaskAttachment, len(attachments))
	for i := range attachments {
		attachmentPtrs[i] = &attachments[i]
	}

	return attachmentPtrs, nil
}

// DeleteAttachment deletes a file attachment
func (r *taskRepository) DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.TaskAttachment{}, "id = ?", attachmentID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete attachment: %w", result.Error)
	}

	return nil
}

// ExportTasks exports tasks in the specified format
func (r *taskRepository) ExportTasks(ctx context.Context, filters entity.TaskFilters, format entity.TaskExportFormat) ([]byte, error) {
	// This is a placeholder implementation
	// In a real implementation, you would query the tasks and format them according to the specified format
	return []byte("exported tasks"), nil
}

// CheckDuplicateTitle checks if a task title already exists in a project
func (r *taskRepository) CheckDuplicateTitle(ctx context.Context, projectID uuid.UUID, title string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&entity.Task{}).Where("project_id = ? AND LOWER(title) = LOWER(?)", projectID, title)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check duplicate title: %w", err)
	}

	return count > 0, nil
}

// ValidateTaskExists checks if a task exists
func (r *taskRepository) ValidateTaskExists(ctx context.Context, taskID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Task{}).Where("id = ?", taskID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to validate task exists: %w", err)
	}

	return count > 0, nil
}

// ValidateProjectExists checks if a project exists
func (r *taskRepository) ValidateProjectExists(ctx context.Context, projectID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Project{}).Where("id = ?", projectID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to validate project exists: %w", err)
	}

	return count > 0, nil
}
