package repository

import (
	"context"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

type WorktreeRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, worktree *entity.Worktree) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Worktree, error)
	GetByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.Worktree, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Worktree, error)
	Update(ctx context.Context, worktree *entity.Worktree) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Status management
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.WorktreeStatus) error
	GetByStatus(ctx context.Context, status entity.WorktreeStatus) ([]*entity.Worktree, error)
	GetByStatuses(ctx context.Context, statuses []entity.WorktreeStatus) ([]*entity.Worktree, error)
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.WorktreeStatus) error

	// Advanced filtering and search
	GetWorktreesWithFilters(ctx context.Context, filters entity.WorktreeFilters) ([]*entity.Worktree, error)
	GetByBranchName(ctx context.Context, branchName string) ([]*entity.Worktree, error)
	GetByWorktreePath(ctx context.Context, worktreePath string) (*entity.Worktree, error)

	// Bulk operations
	BulkDelete(ctx context.Context, worktreeIDs []uuid.UUID) error
	BulkDeleteByProjectID(ctx context.Context, projectID uuid.UUID) error
	BulkDeleteByTaskIDs(ctx context.Context, taskIDs []uuid.UUID) error

	// Statistics and analytics
	GetWorktreeStatistics(ctx context.Context, projectID uuid.UUID) (*entity.WorktreeStatistics, error)
	GetActiveWorktreesCount(ctx context.Context, projectID uuid.UUID) (int, error)
	GetWorktreesByStatusCount(ctx context.Context, projectID uuid.UUID) (map[entity.WorktreeStatus]int, error)

	// Validation
	CheckDuplicateWorktreePath(ctx context.Context, worktreePath string, excludeID *uuid.UUID) (bool, error)
	CheckDuplicateBranchName(ctx context.Context, projectID uuid.UUID, branchName string, excludeID *uuid.UUID) (bool, error)
	ValidateWorktreeExists(ctx context.Context, worktreeID uuid.UUID) (bool, error)
	ValidateTaskExists(ctx context.Context, taskID uuid.UUID) (bool, error)
	ValidateProjectExists(ctx context.Context, projectID uuid.UUID) (bool, error)

	// Cleanup operations
	GetOrphanedWorktrees(ctx context.Context) ([]*entity.Worktree, error)
	CleanupCompletedWorktrees(ctx context.Context, olderThanDays int) error
	CleanupErrorWorktrees(ctx context.Context, olderThanDays int) error
}
