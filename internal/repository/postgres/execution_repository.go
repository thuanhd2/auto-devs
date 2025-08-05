package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type executionRepository struct {
	db *database.GormDB
}

// NewExecutionRepository creates a new PostgreSQL execution repository
func NewExecutionRepository(db *database.GormDB) repository.ExecutionRepository {
	return &executionRepository{db: db}
}

// Create creates a new execution
func (r *executionRepository) Create(ctx context.Context, execution *entity.Execution) error {
	// Generate UUID if not provided
	if execution.ID == uuid.Nil {
		execution.ID = uuid.New()
	}

	// Set default status if not provided
	if execution.Status == "" {
		execution.Status = entity.ExecutionStatusPending
	}

	// Set started time if not provided
	if execution.StartedAt.IsZero() {
		execution.StartedAt = time.Now()
	}

	result := r.db.WithContext(ctx).Create(execution)
	if result.Error != nil {
		return fmt.Errorf("failed to create execution: %w", result.Error)
	}

	return nil
}

// GetByID retrieves an execution by ID
func (r *executionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Execution, error) {
	var execution entity.Execution

	result := r.db.WithContext(ctx).First(&execution, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("execution not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get execution: %w", result.Error)
	}

	return &execution, nil
}

// GetByTaskID retrieves all executions for a specific task
func (r *executionRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]*entity.Execution, error) {
	var executions []entity.Execution

	result := r.db.WithContext(ctx).Where("task_id = ?", taskID).Order("started_at DESC").Find(&executions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get executions by task: %w", result.Error)
	}

	// Convert to slice of pointers
	executionPtrs := make([]*entity.Execution, len(executions))
	for i := range executions {
		executionPtrs[i] = &executions[i]
	}

	return executionPtrs, nil
}

// Update updates an existing execution
func (r *executionRepository) Update(ctx context.Context, execution *entity.Execution) error {
	// First check if execution exists
	var existingExecution entity.Execution
	result := r.db.WithContext(ctx).First(&existingExecution, "id = ?", execution.ID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return fmt.Errorf("execution not found with id %s", execution.ID)
		}
		return fmt.Errorf("failed to check execution existence: %w", result.Error)
	}

	// Update the execution
	result = r.db.WithContext(ctx).Save(execution)
	if result.Error != nil {
		return fmt.Errorf("failed to update execution: %w", result.Error)
	}

	return nil
}

// Delete deletes an execution by ID (soft delete)
func (r *executionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Execution{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete execution: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("execution not found with id %s", id)
	}

	return nil
}

// UpdateStatus updates the status of an execution
func (r *executionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ExecutionStatus) error {
	result := r.db.WithContext(ctx).Model(&entity.Execution{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update execution status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("execution not found with id %s", id)
	}

	return nil
}

// UpdateProgress updates the progress of an execution
func (r *executionRepository) UpdateProgress(ctx context.Context, id uuid.UUID, progress float64) error {
	if progress < 0 || progress > 1 {
		return fmt.Errorf("progress must be between 0.0 and 1.0, got %f", progress)
	}

	result := r.db.WithContext(ctx).Model(&entity.Execution{}).Where("id = ?", id).Update("progress", progress)
	if result.Error != nil {
		return fmt.Errorf("failed to update execution progress: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("execution not found with id %s", id)
	}

	return nil
}

// UpdateError updates the error message of an execution
func (r *executionRepository) UpdateError(ctx context.Context, id uuid.UUID, error string) error {
	result := r.db.WithContext(ctx).Model(&entity.Execution{}).Where("id = ?", id).Update("error", error)
	if result.Error != nil {
		return fmt.Errorf("failed to update execution error: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("execution not found with id %s", id)
	}

	return nil
}

// MarkCompleted marks an execution as completed with result
func (r *executionRepository) MarkCompleted(ctx context.Context, id uuid.UUID, completedAt time.Time, result *entity.ExecutionResult) error {
	updates := map[string]interface{}{
		"status":       entity.ExecutionStatusCompleted,
		"completed_at": completedAt,
		"progress":     1.0,
	}

	if result != nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal execution result: %w", err)
		}
		updates["result"] = string(resultJSON)
	}

	dbResult := r.db.WithContext(ctx).Model(&entity.Execution{}).Where("id = ?", id).Updates(updates)
	if dbResult.Error != nil {
		return fmt.Errorf("failed to mark execution as completed: %w", dbResult.Error)
	}

	if dbResult.RowsAffected == 0 {
		return fmt.Errorf("execution not found with id %s", id)
	}

	return nil
}

// MarkFailed marks an execution as failed with error
func (r *executionRepository) MarkFailed(ctx context.Context, id uuid.UUID, completedAt time.Time, error string) error {
	updates := map[string]interface{}{
		"status":       entity.ExecutionStatusFailed,
		"completed_at": completedAt,
		"error":        error,
	}

	result := r.db.WithContext(ctx).Model(&entity.Execution{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to mark execution as failed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("execution not found with id %s", id)
	}

	return nil
}

// GetByStatus retrieves executions by status
func (r *executionRepository) GetByStatus(ctx context.Context, status entity.ExecutionStatus) ([]*entity.Execution, error) {
	var executions []entity.Execution

	result := r.db.WithContext(ctx).Where("status = ?", status).Order("started_at DESC").Find(&executions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get executions by status: %w", result.Error)
	}

	// Convert to slice of pointers
	executionPtrs := make([]*entity.Execution, len(executions))
	for i := range executions {
		executionPtrs[i] = &executions[i]
	}

	return executionPtrs, nil
}

// GetByStatuses retrieves executions by multiple statuses
func (r *executionRepository) GetByStatuses(ctx context.Context, statuses []entity.ExecutionStatus) ([]*entity.Execution, error) {
	var executions []entity.Execution

	result := r.db.WithContext(ctx).Where("status IN ?", statuses).Order("started_at DESC").Find(&executions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get executions by statuses: %w", result.Error)
	}

	// Convert to slice of pointers
	executionPtrs := make([]*entity.Execution, len(executions))
	for i := range executions {
		executionPtrs[i] = &executions[i]
	}

	return executionPtrs, nil
}

// GetActive retrieves all active executions
func (r *executionRepository) GetActive(ctx context.Context) ([]*entity.Execution, error) {
	activeStatuses := []entity.ExecutionStatus{
		entity.ExecutionStatusPending,
		entity.ExecutionStatusRunning,
		entity.ExecutionStatusPaused,
	}
	return r.GetByStatuses(ctx, activeStatuses)
}

// GetCompleted retrieves completed executions with limit
func (r *executionRepository) GetCompleted(ctx context.Context, limit int) ([]*entity.Execution, error) {
	var executions []entity.Execution

	query := r.db.WithContext(ctx).Where("status IN ?", []entity.ExecutionStatus{
		entity.ExecutionStatusCompleted,
		entity.ExecutionStatusFailed,
		entity.ExecutionStatusCancelled,
	}).Order("completed_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&executions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get completed executions: %w", result.Error)
	}

	// Convert to slice of pointers
	executionPtrs := make([]*entity.Execution, len(executions))
	for i := range executions {
		executionPtrs[i] = &executions[i]
	}

	return executionPtrs, nil
}

// GetByDateRange retrieves executions within a date range
func (r *executionRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.Execution, error) {
	var executions []entity.Execution

	result := r.db.WithContext(ctx).Where("started_at BETWEEN ? AND ?", startDate, endDate).Order("started_at DESC").Find(&executions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get executions by date range: %w", result.Error)
	}

	// Convert to slice of pointers
	executionPtrs := make([]*entity.Execution, len(executions))
	for i := range executions {
		executionPtrs[i] = &executions[i]
	}

	return executionPtrs, nil
}

// GetWithProcesses retrieves an execution with its processes
func (r *executionRepository) GetWithProcesses(ctx context.Context, id uuid.UUID) (*entity.Execution, error) {
	var execution entity.Execution

	result := r.db.WithContext(ctx).Preload("Processes").First(&execution, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("execution not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get execution with processes: %w", result.Error)
	}

	return &execution, nil
}

// GetWithLogs retrieves an execution with its logs
func (r *executionRepository) GetWithLogs(ctx context.Context, id uuid.UUID, logLimit int) (*entity.Execution, error) {
	var execution entity.Execution

	query := r.db.WithContext(ctx)
	if logLimit > 0 {
		query = query.Preload("Logs", func(db *gorm.DB) *gorm.DB {
			return db.Order("timestamp DESC").Limit(logLimit)
		})
	} else {
		query = query.Preload("Logs", func(db *gorm.DB) *gorm.DB {
			return db.Order("timestamp DESC")
		})
	}

	result := query.First(&execution, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("execution not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get execution with logs: %w", result.Error)
	}

	return &execution, nil
}

// GetExecutionStats retrieves execution statistics
func (r *executionRepository) GetExecutionStats(ctx context.Context, taskID *uuid.UUID) (*repository.ExecutionStats, error) {
	var stats repository.ExecutionStats

	query := r.db.WithContext(ctx).Model(&entity.Execution{})
	if taskID != nil {
		query = query.Where("task_id = ?", *taskID)
	}

	// Count total executions
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total executions: %w", err)
	}
	stats.TotalExecutions = totalCount

	// Count completed executions
	var completedCount int64
	if err := query.Where("status = ?", entity.ExecutionStatusCompleted).Count(&completedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed executions: %w", err)
	}
	stats.CompletedExecutions = completedCount

	// Count failed executions
	var failedCount int64
	if err := query.Where("status = ?", entity.ExecutionStatusFailed).Count(&failedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed executions: %w", err)
	}
	stats.FailedExecutions = failedCount

	// Calculate average progress
	var avgProgress float64
	if err := query.Select("AVG(progress)").Scan(&avgProgress).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate average progress: %w", err)
	}
	stats.AverageProgress = avgProgress

	// Status distribution
	statusDistribution := make(map[entity.ExecutionStatus]int64)
	var statusCounts []struct {
		Status entity.ExecutionStatus
		Count  int64
	}

	if err := query.Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get status distribution: %w", err)
	}

	for _, sc := range statusCounts {
		statusDistribution[sc.Status] = sc.Count
	}
	stats.StatusDistribution = statusDistribution

	// Recent activity (last 10 executions)
	var recentExecutions []entity.Execution
	if err := query.Order("started_at DESC").Limit(10).Find(&recentExecutions).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent executions: %w", err)
	}

	recentPtrs := make([]*entity.Execution, len(recentExecutions))
	for i := range recentExecutions {
		recentPtrs[i] = &recentExecutions[i]
	}
	stats.RecentActivity = recentPtrs

	return &stats, nil
}

// GetRecentExecutions retrieves recent executions with limit
func (r *executionRepository) GetRecentExecutions(ctx context.Context, limit int) ([]*entity.Execution, error) {
	var executions []entity.Execution

	query := r.db.WithContext(ctx).Order("started_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&executions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get recent executions: %w", result.Error)
	}

	// Convert to slice of pointers
	executionPtrs := make([]*entity.Execution, len(executions))
	for i := range executions {
		executionPtrs[i] = &executions[i]
	}

	return executionPtrs, nil
}

// BulkUpdateStatus updates status for multiple executions
func (r *executionRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.ExecutionStatus) error {
	result := r.db.WithContext(ctx).Model(&entity.Execution{}).Where("id IN ?", ids).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk update execution status: %w", result.Error)
	}

	return nil
}

// BulkDelete deletes multiple executions
func (r *executionRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Execution{}, "id IN ?", ids)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk delete executions: %w", result.Error)
	}

	return nil
}

// CleanupOldExecutions removes old executions
func (r *executionRepository) CleanupOldExecutions(ctx context.Context, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Unscoped().Delete(&entity.Execution{}, "started_at < ?", olderThan)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old executions: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// ValidateExecutionExists checks if an execution exists
func (r *executionRepository) ValidateExecutionExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&entity.Execution{}).Where("id = ?", id).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to validate execution existence: %w", result.Error)
	}

	return count > 0, nil
}

// ValidateTaskExists checks if a task exists
func (r *executionRepository) ValidateTaskExists(ctx context.Context, taskID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&entity.Task{}).Where("id = ?", taskID).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to validate task existence: %w", result.Error)
	}

	return count > 0, nil
}