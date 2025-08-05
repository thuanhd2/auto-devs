package repository

import (
	"context"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// ExecutionRepository defines the interface for execution data persistence
type ExecutionRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, execution *entity.Execution) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Execution, error)
	GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]*entity.Execution, error)
	Update(ctx context.Context, execution *entity.Execution) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Status management
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ExecutionStatus) error
	UpdateProgress(ctx context.Context, id uuid.UUID, progress float64) error
	UpdateError(ctx context.Context, id uuid.UUID, error string) error
	MarkCompleted(ctx context.Context, id uuid.UUID, completedAt time.Time, result *entity.ExecutionResult) error
	MarkFailed(ctx context.Context, id uuid.UUID, completedAt time.Time, error string) error

	// Filtering and search
	GetByStatus(ctx context.Context, status entity.ExecutionStatus) ([]*entity.Execution, error)
	GetByStatuses(ctx context.Context, statuses []entity.ExecutionStatus) ([]*entity.Execution, error)
	GetActive(ctx context.Context) ([]*entity.Execution, error)
	GetCompleted(ctx context.Context, limit int) ([]*entity.Execution, error)
	GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.Execution, error)

	// Advanced queries
	GetWithProcesses(ctx context.Context, id uuid.UUID) (*entity.Execution, error)
	GetWithLogs(ctx context.Context, id uuid.UUID, logLimit int) (*entity.Execution, error)
	GetExecutionStats(ctx context.Context, taskID *uuid.UUID) (*ExecutionStats, error)
	GetRecentExecutions(ctx context.Context, limit int) ([]*entity.Execution, error)

	// Bulk operations
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.ExecutionStatus) error
	BulkDelete(ctx context.Context, ids []uuid.UUID) error
	CleanupOldExecutions(ctx context.Context, olderThan time.Time) (int64, error)

	// Validation
	ValidateExecutionExists(ctx context.Context, id uuid.UUID) (bool, error)
	ValidateTaskExists(ctx context.Context, taskID uuid.UUID) (bool, error)
}

// ExecutionStats represents execution statistics
type ExecutionStats struct {
	TotalExecutions     int64                            `json:"total_executions"`
	CompletedExecutions int64                            `json:"completed_executions"`
	FailedExecutions    int64                            `json:"failed_executions"`
	AverageProgress     float64                          `json:"average_progress"`
	AverageDuration     time.Duration                    `json:"average_duration"`
	StatusDistribution  map[entity.ExecutionStatus]int64 `json:"status_distribution"`
	RecentActivity      []*entity.Execution              `json:"recent_activity"`
}

// ExecutionFilters represents filtering options for executions
type ExecutionFilters struct {
	TaskID        *uuid.UUID
	Statuses      []entity.ExecutionStatus
	StartedAfter  *time.Time
	StartedBefore *time.Time
	MinProgress   *float64
	MaxProgress   *float64
	WithErrors    *bool
	Limit         *int
	Offset        *int
	OrderBy       *string // "started_at", "completed_at", "progress", "status"
	OrderDir      *string // "asc", "desc"
}