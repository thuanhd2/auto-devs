package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type worktreeRepository struct {
	db *gorm.DB
}

func NewWorktreeRepository(db *database.GormDB) *worktreeRepository {
	return &worktreeRepository{db: db.DB}
}

// Create creates a new worktree record
func (r *worktreeRepository) Create(ctx context.Context, worktree *entity.Worktree) error {
	err := r.db.WithContext(ctx).Create(worktree).Error
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}
	return nil
}

// GetByID retrieves a worktree by ID
func (r *worktreeRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Worktree, error) {
	var worktree entity.Worktree
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&worktree).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("worktree not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}
	return &worktree, nil
}

// GetByTaskID retrieves a worktree by task ID
func (r *worktreeRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.Worktree, error) {
	var worktree entity.Worktree
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).First(&worktree).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("worktree not found for task %s", taskID)
		}
		return nil, fmt.Errorf("failed to get worktree by task: %w", err)
	}
	return &worktree, nil
}

// GetByProjectID retrieves all worktrees for a project
func (r *worktreeRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Worktree, error) {
	var worktrees []*entity.Worktree
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&worktrees).Error
	if err != nil {
		return nil, err
	}
	return worktrees, nil
}

// Update updates a worktree record
func (r *worktreeRepository) Update(ctx context.Context, worktree *entity.Worktree) error {
	// First check if the record exists
	var existing entity.Worktree
	err := r.db.WithContext(ctx).Where("id = ?", worktree.ID).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("worktree not found with id %s", worktree.ID)
		}
		return fmt.Errorf("failed to check worktree existence: %w", err)
	}
	
	// Update the record
	return r.db.WithContext(ctx).Save(worktree).Error
}

// Delete soft deletes a worktree record
func (r *worktreeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// First check if the record exists
	var existing entity.Worktree
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("worktree not found with id %s", id)
		}
		return fmt.Errorf("failed to check worktree existence: %w", err)
	}
	
	// Delete the record
	return r.db.WithContext(ctx).Delete(&entity.Worktree{}, id).Error
}

// UpdateStatus updates the status of a worktree
func (r *worktreeRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.WorktreeStatus) error {
	// First check if the record exists
	var existing entity.Worktree
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("worktree not found with id %s", id)
		}
		return fmt.Errorf("failed to check worktree existence: %w", err)
	}
	
	// Update the status
	return r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("id = ?", id).Update("status", status).Error
}

// GetByStatus retrieves worktrees by status
func (r *worktreeRepository) GetByStatus(ctx context.Context, status entity.WorktreeStatus) ([]*entity.Worktree, error) {
	var worktrees []*entity.Worktree
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&worktrees).Error
	if err != nil {
		return nil, err
	}
	return worktrees, nil
}

// GetByStatuses retrieves worktrees by multiple statuses
func (r *worktreeRepository) GetByStatuses(ctx context.Context, statuses []entity.WorktreeStatus) ([]*entity.Worktree, error) {
	var worktrees []*entity.Worktree
	err := r.db.WithContext(ctx).Where("status IN ?", statuses).Find(&worktrees).Error
	if err != nil {
		return nil, err
	}
	return worktrees, nil
}

// BulkUpdateStatus updates status for multiple worktrees
func (r *worktreeRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.WorktreeStatus) error {
	return r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("id IN ?", ids).Update("status", status).Error
}

// GetWorktreesWithFilters retrieves worktrees with advanced filtering
func (r *worktreeRepository) GetWorktreesWithFilters(ctx context.Context, filters entity.WorktreeFilters) ([]*entity.Worktree, error) {
	query := r.db.WithContext(ctx).Model(&entity.Worktree{})

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
		if filters.OrderDir != nil && *filters.OrderDir == "desc" {
			orderDir = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", *filters.OrderBy, orderDir))
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

	var worktrees []*entity.Worktree
	err := query.Find(&worktrees).Error
	if err != nil {
		return nil, err
	}

	return worktrees, nil
}

// GetByBranchName retrieves worktrees by branch name
func (r *worktreeRepository) GetByBranchName(ctx context.Context, branchName string) ([]*entity.Worktree, error) {
	var worktrees []*entity.Worktree
	err := r.db.WithContext(ctx).Where("branch_name = ?", branchName).Find(&worktrees).Error
	if err != nil {
		return nil, err
	}
	return worktrees, nil
}

// GetByWorktreePath retrieves a worktree by worktree path
func (r *worktreeRepository) GetByWorktreePath(ctx context.Context, worktreePath string) (*entity.Worktree, error) {
	var worktree entity.Worktree
	err := r.db.WithContext(ctx).Where("worktree_path = ?", worktreePath).First(&worktree).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("worktree not found with path %s", worktreePath)
		}
		return nil, fmt.Errorf("failed to get worktree by path: %w", err)
	}
	return &worktree, nil
}

// BulkDelete soft deletes multiple worktrees
func (r *worktreeRepository) BulkDelete(ctx context.Context, worktreeIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Worktree{}, worktreeIDs).Error
}

// BulkDeleteByProjectID soft deletes all worktrees for a project
func (r *worktreeRepository) BulkDeleteByProjectID(ctx context.Context, projectID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("project_id = ?", projectID).Delete(&entity.Worktree{}).Error
}

// BulkDeleteByTaskIDs soft deletes worktrees for multiple tasks
func (r *worktreeRepository) BulkDeleteByTaskIDs(ctx context.Context, taskIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Where("task_id IN ?", taskIDs).Delete(&entity.Worktree{}).Error
}

// GetWorktreeStatistics gets worktree statistics for a project
func (r *worktreeRepository) GetWorktreeStatistics(ctx context.Context, projectID uuid.UUID) (*entity.WorktreeStatistics, error) {
	var stats entity.WorktreeStatistics
	stats.ProjectID = projectID
	stats.GeneratedAt = time.Now()

	// Get total worktrees
	var totalCount int64
	err := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("project_id = ?", projectID).Count(&totalCount).Error
	if err != nil {
		return nil, err
	}
	stats.TotalWorktrees = int(totalCount)
	if err != nil {
		return nil, err
	}

	// Get worktrees by status
	var statusCounts []struct {
		Status entity.WorktreeStatus `json:"status"`
		Count  int                   `json:"count"`
	}

	err = r.db.WithContext(ctx).Model(&entity.Worktree{}).
		Select("status, count(*) as count").
		Where("project_id = ?", projectID).
		Group("status").
		Find(&statusCounts).Error
	if err != nil {
		return nil, err
	}

	stats.WorktreesByStatus = make(map[entity.WorktreeStatus]int)
	for _, sc := range statusCounts {
		stats.WorktreesByStatus[sc.Status] = sc.Count
	}

	// Get specific counts
	stats.ActiveWorktrees = stats.WorktreesByStatus[entity.WorktreeStatusActive]
	stats.CompletedWorktrees = stats.WorktreesByStatus[entity.WorktreeStatusCompleted]
	stats.ErrorWorktrees = stats.WorktreesByStatus[entity.WorktreeStatusError]

	// Calculate average creation time (simplified - in a real implementation you'd track actual creation time)
	stats.AverageCreationTime = 30.0 // Placeholder value in seconds

	return &stats, nil
}

// GetActiveWorktreesCount gets the count of active worktrees for a project
func (r *worktreeRepository) GetActiveWorktreesCount(ctx context.Context, projectID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Worktree{}).
		Where("project_id = ? AND status = ?", projectID, entity.WorktreeStatusActive).
		Count(&count).Error
	return int(count), err
}

// GetWorktreesByStatusCount gets worktrees count by status for a project
func (r *worktreeRepository) GetWorktreesByStatusCount(ctx context.Context, projectID uuid.UUID) (map[entity.WorktreeStatus]int, error) {
	var statusCounts []struct {
		Status entity.WorktreeStatus `json:"status"`
		Count  int                   `json:"count"`
	}

	err := r.db.WithContext(ctx).Model(&entity.Worktree{}).
		Select("status, count(*) as count").
		Where("project_id = ?", projectID).
		Group("status").
		Find(&statusCounts).Error
	if err != nil {
		return nil, err
	}

	result := make(map[entity.WorktreeStatus]int)
	for _, sc := range statusCounts {
		result[sc.Status] = sc.Count
	}

	return result, nil
}

// CheckDuplicateWorktreePath checks if a worktree path already exists
func (r *worktreeRepository) CheckDuplicateWorktreePath(ctx context.Context, worktreePath string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("worktree_path = ?", worktreePath)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// CheckDuplicateBranchName checks if a branch name already exists in a project
func (r *worktreeRepository) CheckDuplicateBranchName(ctx context.Context, projectID uuid.UUID, branchName string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&entity.Worktree{}).
		Where("project_id = ? AND branch_name = ?", projectID, branchName)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// ValidateWorktreeExists checks if a worktree exists
func (r *worktreeRepository) ValidateWorktreeExists(ctx context.Context, worktreeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Worktree{}).Where("id = ?", worktreeID).Count(&count).Error
	return count > 0, err
}

// ValidateTaskExists checks if a task exists
func (r *worktreeRepository) ValidateTaskExists(ctx context.Context, taskID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Task{}).Where("id = ?", taskID).Count(&count).Error
	return count > 0, err
}

// ValidateProjectExists checks if a project exists
func (r *worktreeRepository) ValidateProjectExists(ctx context.Context, projectID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Project{}).Where("id = ?", projectID).Count(&count).Error
	return count > 0, err
}

// GetOrphanedWorktrees gets worktrees that don't have corresponding tasks
func (r *worktreeRepository) GetOrphanedWorktrees(ctx context.Context) ([]*entity.Worktree, error) {
	var worktrees []*entity.Worktree
	err := r.db.WithContext(ctx).
		Joins("LEFT JOIN tasks ON worktrees.task_id = tasks.id").
		Where("tasks.id IS NULL").
		Find(&worktrees).Error
	if err != nil {
		return nil, err
	}
	return worktrees, nil
}

// CleanupCompletedWorktrees cleans up completed worktrees older than specified days
func (r *worktreeRepository) CleanupCompletedWorktrees(ctx context.Context, olderThanDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)
	return r.db.WithContext(ctx).
		Where("status = ? AND updated_at < ?", entity.WorktreeStatusCompleted, cutoffDate).
		Delete(&entity.Worktree{}).Error
}

// CleanupErrorWorktrees cleans up error worktrees older than specified days
func (r *worktreeRepository) CleanupErrorWorktrees(ctx context.Context, olderThanDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)
	return r.db.WithContext(ctx).
		Where("status = ? AND updated_at < ?", entity.WorktreeStatusError, cutoffDate).
		Delete(&entity.Worktree{}).Error
}
