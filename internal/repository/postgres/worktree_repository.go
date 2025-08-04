package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type worktreeRepository struct {
	db *database.GormDB
}

// NewWorktreeRepository creates a new PostgreSQL worktree repository
func NewWorktreeRepository(db *database.GormDB) repository.WorktreeRepository {
	return &worktreeRepository{db: db}
}

// Create creates a new worktree
func (r *worktreeRepository) Create(ctx context.Context, worktree *entity.Worktree) error {
	// Generate UUID if not provided
	if worktree.ID == uuid.Nil {
		worktree.ID = uuid.New()
	}

	// Set default status if not provided
	if worktree.Status == "" {
		worktree.Status = entity.WorktreeStatusCreating
	}

	result := r.db.WithContext(ctx).Create(worktree)
	if result.Error != nil {
		return fmt.Errorf("failed to create worktree: %w", result.Error)
	}

	return nil
}

// GetByID retrieves a worktree by ID
func (r *worktreeRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Worktree, error) {
	var worktree entity.Worktree

	result := r.db.WithContext(ctx).First(&worktree, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("worktree not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get worktree: %w", result.Error)
	}

	return &worktree, nil
}

// GetByTaskID retrieves a worktree by task ID
func (r *worktreeRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.Worktree, error) {
	var worktree entity.Worktree

	result := r.db.WithContext(ctx).Where("task_id = ?", taskID).First(&worktree)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("worktree not found for task %s", taskID)
		}
		return nil, fmt.Errorf("failed to get worktree by task ID: %w", result.Error)
	}

	return &worktree, nil
}

// GetByProjectID retrieves all worktrees for a specific project
func (r *worktreeRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Worktree, error) {
	var worktrees []entity.Worktree

	result := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").Find(&worktrees)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get worktrees by project: %w", result.Error)
	}

	// Convert to slice of pointers
	worktreePtrs := make([]*entity.Worktree, len(worktrees))
	for i := range worktrees {
		worktreePtrs[i] = &worktrees[i]
	}

	return worktreePtrs, nil
}

// Update updates an existing worktree
func (r *worktreeRepository) Update(ctx context.Context, worktree *entity.Worktree) error {
	// First check if worktree exists
	var existingWorktree entity.Worktree
	result := r.db.WithContext(ctx).First(&existingWorktree, "id = ?", worktree.ID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return fmt.Errorf("worktree not found with id %s", worktree.ID)
		}
		return fmt.Errorf("failed to check worktree existence: %w", result.Error)
	}

	// Update the worktree
	result = r.db.WithContext(ctx).Save(worktree)
	if result.Error != nil {
		return fmt.Errorf("failed to update worktree: %w", result.Error)
	}

	return nil
}

// Delete deletes a worktree by ID (soft delete)
func (r *worktreeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Worktree{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete worktree: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("worktree not found with id %s", id)
	}

	return nil
}

// UpdateStatus updates the status of a worktree
func (r *worktreeRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.WorktreeStatus) error {
	result := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update worktree status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("worktree not found with id %s", id)
	}

	return nil
}

// GetByStatus retrieves all worktrees with a specific status
func (r *worktreeRepository) GetByStatus(ctx context.Context, status entity.WorktreeStatus) ([]*entity.Worktree, error) {
	var worktrees []entity.Worktree

	result := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC").Find(&worktrees)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get worktrees by status: %w", result.Error)
	}

	// Convert to slice of pointers
	worktreePtrs := make([]*entity.Worktree, len(worktrees))
	for i := range worktrees {
		worktreePtrs[i] = &worktrees[i]
	}

	return worktreePtrs, nil
}

// GetByStatuses retrieves all worktrees with any of the specified statuses
func (r *worktreeRepository) GetByStatuses(ctx context.Context, statuses []entity.WorktreeStatus) ([]*entity.Worktree, error) {
	var worktrees []entity.Worktree

	result := r.db.WithContext(ctx).Where("status IN ?", statuses).Order("created_at DESC").Find(&worktrees)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get worktrees by statuses: %w", result.Error)
	}

	// Convert to slice of pointers
	worktreePtrs := make([]*entity.Worktree, len(worktrees))
	for i := range worktrees {
		worktreePtrs[i] = &worktrees[i]
	}

	return worktreePtrs, nil
}

// BulkUpdateStatus updates the status of multiple worktrees
func (r *worktreeRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.WorktreeStatus) error {
	result := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("id IN ?", ids).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk update worktree status: %w", result.Error)
	}

	return nil
}

// GetWorktreesWithFilters retrieves worktrees with advanced filtering
func (r *worktreeRepository) GetWorktreesWithFilters(ctx context.Context, filters entity.WorktreeFilters) ([]*entity.Worktree, error) {
	query := r.db.WithContext(ctx).Model(&entity.Worktree{})

	// Apply filters
	if filters.ProjectID != nil {
		query = query.Where("project_id = ?", *filters.ProjectID)
	}

	if filters.TaskID != nil {
		query = query.Where("task_id = ?", *filters.TaskID)
	}

	if len(filters.Statuses) > 0 {
		query = query.Where("status IN ?", filters.Statuses)
	}

	if filters.BranchName != nil {
		query = query.Where("branch_name ILIKE ?", "%"+*filters.BranchName+"%")
	}

	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
	}

	// Apply ordering
	if filters.OrderBy != nil {
		orderDir := "ASC"
		if filters.OrderDir != nil {
			orderDir = *filters.OrderDir
		}
		query = query.Order(*filters.OrderBy + " " + orderDir)
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filters.Limit != nil {
		query = query.Limit(*filters.Limit)
	}

	if filters.Offset != nil {
		query = query.Offset(*filters.Offset)
	}

	var worktrees []entity.Worktree
	result := query.Find(&worktrees)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get worktrees with filters: %w", result.Error)
	}

	// Convert to slice of pointers
	worktreePtrs := make([]*entity.Worktree, len(worktrees))
	for i := range worktrees {
		worktreePtrs[i] = &worktrees[i]
	}

	return worktreePtrs, nil
}

// GetByBranchName retrieves worktrees by branch name
func (r *worktreeRepository) GetByBranchName(ctx context.Context, branchName string) ([]*entity.Worktree, error) {
	var worktrees []entity.Worktree

	result := r.db.WithContext(ctx).Where("branch_name = ?", branchName).Order("created_at DESC").Find(&worktrees)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get worktrees by branch name: %w", result.Error)
	}

	// Convert to slice of pointers
	worktreePtrs := make([]*entity.Worktree, len(worktrees))
	for i := range worktrees {
		worktreePtrs[i] = &worktrees[i]
	}

	return worktreePtrs, nil
}

// GetByWorktreePath retrieves a worktree by worktree path
func (r *worktreeRepository) GetByWorktreePath(ctx context.Context, worktreePath string) (*entity.Worktree, error) {
	var worktree entity.Worktree

	result := r.db.WithContext(ctx).Where("worktree_path = ?", worktreePath).First(&worktree)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("worktree not found with path %s", worktreePath)
		}
		return nil, fmt.Errorf("failed to get worktree by path: %w", result.Error)
	}

	return &worktree, nil
}

// BulkDelete deletes multiple worktrees by IDs
func (r *worktreeRepository) BulkDelete(ctx context.Context, worktreeIDs []uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Worktree{}, "id IN ?", worktreeIDs)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk delete worktrees: %w", result.Error)
	}

	return nil
}

// BulkDeleteByProjectID deletes all worktrees for a project
func (r *worktreeRepository) BulkDeleteByProjectID(ctx context.Context, projectID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Worktree{}, "project_id = ?", projectID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete worktrees by project ID: %w", result.Error)
	}

	return nil
}

// BulkDeleteByTaskIDs deletes all worktrees for specified tasks
func (r *worktreeRepository) BulkDeleteByTaskIDs(ctx context.Context, taskIDs []uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Worktree{}, "task_id IN ?", taskIDs)
	if result.Error != nil {
		return fmt.Errorf("failed to delete worktrees by task IDs: %w", result.Error)
	}

	return nil
}

// GetWorktreeStatistics retrieves worktree statistics for a project
func (r *worktreeRepository) GetWorktreeStatistics(ctx context.Context, projectID uuid.UUID) (*entity.WorktreeStatistics, error) {
	var totalWorktrees int64
	var activeWorktrees int64
	var completedWorktrees int64
	var errorWorktrees int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("project_id = ?", projectID).Count(&totalWorktrees).Error; err != nil {
		return nil, fmt.Errorf("failed to count total worktrees: %w", err)
	}

	// Get active count
	if err := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("project_id = ? AND status = ?", projectID, entity.WorktreeStatusActive).Count(&activeWorktrees).Error; err != nil {
		return nil, fmt.Errorf("failed to count active worktrees: %w", err)
	}

	// Get completed count
	if err := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("project_id = ? AND status = ?", projectID, entity.WorktreeStatusCompleted).Count(&completedWorktrees).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed worktrees: %w", err)
	}

	// Get error count
	if err := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("project_id = ? AND status = ?", projectID, entity.WorktreeStatusError).Count(&errorWorktrees).Error; err != nil {
		return nil, fmt.Errorf("failed to count error worktrees: %w", err)
	}

	// Get worktrees by status count
	worktreesByStatus, err := r.GetWorktreesByStatusCount(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get worktrees by status count: %w", err)
	}

	// Calculate average creation time (simplified - could be enhanced with actual timing data)
	var avgCreationTime float64
	if totalWorktrees > 0 {
		// This is a placeholder - in a real implementation, you might track actual creation time
		avgCreationTime = 30.0 // 30 seconds as placeholder
	}

	stats := &entity.WorktreeStatistics{
		ProjectID:           projectID,
		TotalWorktrees:      int(totalWorktrees),
		ActiveWorktrees:     int(activeWorktrees),
		CompletedWorktrees:  int(completedWorktrees),
		ErrorWorktrees:      int(errorWorktrees),
		WorktreesByStatus:   worktreesByStatus,
		AverageCreationTime: avgCreationTime,
		GeneratedAt:         time.Now(),
	}

	return stats, nil
}

// GetActiveWorktreesCount retrieves the count of active worktrees for a project
func (r *worktreeRepository) GetActiveWorktreesCount(ctx context.Context, projectID uuid.UUID) (int, error) {
	var count int64

	result := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("project_id = ? AND status = ?", projectID, entity.WorktreeStatusActive).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count active worktrees: %w", result.Error)
	}

	return int(count), nil
}

// GetWorktreesByStatusCount retrieves the count of worktrees by status for a project
func (r *worktreeRepository) GetWorktreesByStatusCount(ctx context.Context, projectID uuid.UUID) (map[entity.WorktreeStatus]int, error) {
	var results []struct {
		Status entity.WorktreeStatus `json:"status"`
		Count  int64                 `json:"count"`
	}

	result := r.db.WithContext(ctx).Model(&entity.Worktree{}).
		Select("status, count(*) as count").
		Where("project_id = ?", projectID).
		Group("status").
		Find(&results)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get worktrees by status count: %w", result.Error)
	}

	statusCounts := make(map[entity.WorktreeStatus]int)
	for _, result := range results {
		statusCounts[result.Status] = int(result.Count)
	}

	return statusCounts, nil
}

// CheckDuplicateWorktreePath checks if a worktree path already exists
func (r *worktreeRepository) CheckDuplicateWorktreePath(ctx context.Context, worktreePath string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("worktree_path = ?", worktreePath)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	result := query.Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check duplicate worktree path: %w", result.Error)
	}

	return count > 0, nil
}

// CheckDuplicateBranchName checks if a branch name already exists in a project
func (r *worktreeRepository) CheckDuplicateBranchName(ctx context.Context, projectID uuid.UUID, branchName string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("project_id = ? AND branch_name = ?", projectID, branchName)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	result := query.Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check duplicate branch name: %w", result.Error)
	}

	return count > 0, nil
}

// ValidateWorktreeExists checks if a worktree exists
func (r *worktreeRepository) ValidateWorktreeExists(ctx context.Context, worktreeID uuid.UUID) (bool, error) {
	var count int64

	result := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("id = ?", worktreeID).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to validate worktree exists: %w", result.Error)
	}

	return count > 0, nil
}

// ValidateTaskExists checks if a task exists
func (r *worktreeRepository) ValidateTaskExists(ctx context.Context, taskID uuid.UUID) (bool, error) {
	var count int64

	result := r.db.WithContext(ctx).Model(&entity.Task{}).Where("id = ?", taskID).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to validate task exists: %w", result.Error)
	}

	return count > 0, nil
}

// ValidateProjectExists checks if a project exists
func (r *worktreeRepository) ValidateProjectExists(ctx context.Context, projectID uuid.UUID) (bool, error) {
	var count int64

	result := r.db.WithContext(ctx).Model(&entity.Project{}).Where("id = ?", projectID).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to validate project exists: %w", result.Error)
	}

	return count > 0, nil
}

// GetOrphanedWorktrees retrieves worktrees that reference non-existent tasks or projects
func (r *worktreeRepository) GetOrphanedWorktrees(ctx context.Context) ([]*entity.Worktree, error) {
	var worktrees []entity.Worktree

	// Find worktrees where task doesn't exist
	result := r.db.WithContext(ctx).
		Joins("LEFT JOIN tasks ON worktrees.task_id = tasks.id").
		Where("tasks.id IS NULL").
		Find(&worktrees)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get orphaned worktrees: %w", result.Error)
	}

	// Convert to slice of pointers
	worktreePtrs := make([]*entity.Worktree, len(worktrees))
	for i := range worktrees {
		worktreePtrs[i] = &worktrees[i]
	}

	return worktreePtrs, nil
}

// CleanupCompletedWorktrees removes completed worktrees older than specified days
func (r *worktreeRepository) CleanupCompletedWorktrees(ctx context.Context, olderThanDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

	result := r.db.WithContext(ctx).Delete(&entity.Worktree{}, "status = ? AND created_at < ?", entity.WorktreeStatusCompleted, cutoffDate)
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup completed worktrees: %w", result.Error)
	}

	return nil
}

// CleanupErrorWorktrees removes error worktrees older than specified days
func (r *worktreeRepository) CleanupErrorWorktrees(ctx context.Context, olderThanDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

	result := r.db.WithContext(ctx).Delete(&entity.Worktree{}, "status = ? AND created_at < ?", entity.WorktreeStatusError, cutoffDate)
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup error worktrees: %w", result.Error)
	}

	return nil
}
