package repository

import (
	"context"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

type TaskRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, task *entity.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Task, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error)
	Update(ctx context.Context, task *entity.Task) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Status management
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TaskStatus) error
	UpdateStatusWithHistory(ctx context.Context, id uuid.UUID, status entity.TaskStatus, changedBy *string, reason *string) error
	GetByStatus(ctx context.Context, status entity.TaskStatus) ([]*entity.Task, error)
	GetByStatuses(ctx context.Context, statuses []entity.TaskStatus) ([]*entity.Task, error)
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.TaskStatus, changedBy *string) error

	// Advanced filtering and search
	GetTasksWithFilters(ctx context.Context, filters entity.TaskFilters) ([]*entity.Task, error)
	SearchTasks(ctx context.Context, query string, projectID *uuid.UUID) ([]*entity.TaskSearchResult, error)
	GetTasksByPriority(ctx context.Context, priority entity.TaskPriority) ([]*entity.Task, error)
	GetTasksByTags(ctx context.Context, tags []string) ([]*entity.Task, error)
	GetArchivedTasks(ctx context.Context, projectID *uuid.UUID) ([]*entity.Task, error)
	GetTasksWithSubtasks(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error)

	// Parent-child relationships
	GetSubtasks(ctx context.Context, parentTaskID uuid.UUID) ([]*entity.Task, error)
	GetParentTask(ctx context.Context, taskID uuid.UUID) (*entity.Task, error)
	UpdateParentTask(ctx context.Context, taskID uuid.UUID, parentTaskID *uuid.UUID) error

	// Bulk operations
	BulkDelete(ctx context.Context, taskIDs []uuid.UUID) error
	BulkArchive(ctx context.Context, taskIDs []uuid.UUID) error
	BulkUnarchive(ctx context.Context, taskIDs []uuid.UUID) error
	BulkUpdatePriority(ctx context.Context, taskIDs []uuid.UUID, priority entity.TaskPriority) error
	BulkAssign(ctx context.Context, taskIDs []uuid.UUID, assignedTo string) error

	// Templates
	CreateTemplate(ctx context.Context, template *entity.TaskTemplate) error
	GetTemplates(ctx context.Context, projectID uuid.UUID, includeGlobal bool) ([]*entity.TaskTemplate, error)
	GetTemplateByID(ctx context.Context, id uuid.UUID) (*entity.TaskTemplate, error)
	UpdateTemplate(ctx context.Context, template *entity.TaskTemplate) error
	DeleteTemplate(ctx context.Context, id uuid.UUID) error
	CreateTaskFromTemplate(ctx context.Context, templateID uuid.UUID, projectID uuid.UUID, createdBy string) (*entity.Task, error)

	// Audit trail
	GetAuditLogs(ctx context.Context, taskID uuid.UUID, limit *int) ([]*entity.TaskAuditLog, error)

	// Statistics and analytics
	GetStatusHistory(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskStatusHistory, error)
	GetStatusAnalytics(ctx context.Context, projectID uuid.UUID) (*entity.TaskStatusAnalytics, error)
	GetTaskStatistics(ctx context.Context, projectID uuid.UUID) (*entity.TaskStatistics, error)

	// Dependencies
	AddDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID, dependencyType string) error
	RemoveDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) error
	GetDependencies(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskDependency, error)
	GetDependents(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskDependency, error)

	// Comments
	AddComment(ctx context.Context, comment *entity.TaskComment) error
	GetComments(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskComment, error)
	UpdateComment(ctx context.Context, comment *entity.TaskComment) error
	DeleteComment(ctx context.Context, commentID uuid.UUID) error



	// Export functionality
	ExportTasks(ctx context.Context, filters entity.TaskFilters, format entity.TaskExportFormat) ([]byte, error)

	// Validation
	CheckDuplicateTitle(ctx context.Context, projectID uuid.UUID, title string, excludeID *uuid.UUID) (bool, error)
	ValidateTaskExists(ctx context.Context, taskID uuid.UUID) (bool, error)
	ValidateProjectExists(ctx context.Context, projectID uuid.UUID) (bool, error)

	// Worktree cleanup
	GetTasksEligibleForWorktreeCleanup(ctx context.Context, cutoffTime time.Time) ([]*entity.Task, error)
}

// TaskFilters represents filtering options for tasks (moved to entity package)
// This is kept for backward compatibility
type TaskFilters struct {
	ProjectID     *uuid.UUID
	Statuses      []entity.TaskStatus
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	SearchTerm    *string
	Limit         *int
	Offset        *int
	OrderBy       *string // "created_at", "updated_at", "title", "status"
	OrderDir      *string // "asc", "desc"
}
