package repository

import (
	"context"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

type TaskRepository interface {
	Create(ctx context.Context, task *entity.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Task, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error)
	Update(ctx context.Context, task *entity.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TaskStatus) error
	UpdateStatusWithHistory(ctx context.Context, id uuid.UUID, status entity.TaskStatus, changedBy *string, reason *string) error
	GetByStatus(ctx context.Context, status entity.TaskStatus) ([]*entity.Task, error)
	GetByStatuses(ctx context.Context, statuses []entity.TaskStatus) ([]*entity.Task, error)
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.TaskStatus, changedBy *string) error
	GetStatusHistory(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskStatusHistory, error)
	GetStatusAnalytics(ctx context.Context, projectID uuid.UUID) (*entity.TaskStatusAnalytics, error)
	GetTasksWithFilters(ctx context.Context, filters TaskFilters) ([]*entity.Task, error)
}

// TaskFilters represents filtering options for tasks
type TaskFilters struct {
	ProjectID    *uuid.UUID
	Statuses     []entity.TaskStatus
	CreatedAfter *time.Time
	CreatedBefore *time.Time
	SearchTerm   *string
	Limit        *int
	Offset       *int
	OrderBy      *string // "created_at", "updated_at", "title", "status"
	OrderDir     *string // "asc", "desc"
}