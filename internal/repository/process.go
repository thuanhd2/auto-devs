package repository

import (
	"context"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// ProcessRepository defines the interface for process data persistence
type ProcessRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, process *entity.Process) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Process, error)
	GetByExecutionID(ctx context.Context, executionID uuid.UUID) ([]*entity.Process, error)
	Update(ctx context.Context, process *entity.Process) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Process management
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ProcessStatus) error
	UpdatePID(ctx context.Context, id uuid.UUID, pid int) error
	UpdateResourceUsage(ctx context.Context, id uuid.UUID, cpuUsage float64, memoryUsage uint64) error
	MarkCompleted(ctx context.Context, id uuid.UUID, endTime time.Time, exitCode *int) error
	MarkFailed(ctx context.Context, id uuid.UUID, endTime time.Time, error string) error

	// Filtering and search
	GetByStatus(ctx context.Context, status entity.ProcessStatus) ([]*entity.Process, error)
	GetByStatuses(ctx context.Context, statuses []entity.ProcessStatus) ([]*entity.Process, error)
	GetRunning(ctx context.Context) ([]*entity.Process, error)
	GetCompleted(ctx context.Context, limit int) ([]*entity.Process, error)
	GetByPID(ctx context.Context, pid int) (*entity.Process, error)
	GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.Process, error)

	// Advanced queries
	GetProcessStats(ctx context.Context, executionID *uuid.UUID) (*ProcessStats, error)
	GetLongRunningProcesses(ctx context.Context, threshold time.Duration) ([]*entity.Process, error)
	GetHighResourceProcesses(ctx context.Context, cpuThreshold float64, memoryThreshold uint64) ([]*entity.Process, error)
	GetRecentProcesses(ctx context.Context, limit int) ([]*entity.Process, error)

	// Process monitoring
	GetActiveProcessesByExecution(ctx context.Context, executionID uuid.UUID) ([]*entity.Process, error)
	CountActiveProcesses(ctx context.Context) (int64, error)
	GetResourceUsageSummary(ctx context.Context, executionID *uuid.UUID) (*ResourceUsageSummary, error)

	// Bulk operations
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.ProcessStatus) error
	BulkDelete(ctx context.Context, ids []uuid.UUID) error
	CleanupOldProcesses(ctx context.Context, olderThan time.Time) (int64, error)
	TerminateProcessesByExecution(ctx context.Context, executionID uuid.UUID) error

	// Validation
	ValidateProcessExists(ctx context.Context, id uuid.UUID) (bool, error)
	ValidateExecutionExists(ctx context.Context, executionID uuid.UUID) (bool, error)
}

// ProcessStats represents process statistics
type ProcessStats struct {
	TotalProcesses        int64                           `json:"total_processes"`
	RunningProcesses      int64                           `json:"running_processes"`
	CompletedProcesses    int64                           `json:"completed_processes"`
	FailedProcesses       int64                           `json:"failed_processes"`
	AverageDuration       time.Duration                   `json:"average_duration"`
	AverageCPUUsage       float64                         `json:"average_cpu_usage"`
	AverageMemoryUsage    uint64                          `json:"average_memory_usage"`
	StatusDistribution    map[entity.ProcessStatus]int64  `json:"status_distribution"`
	RecentActivity        []*entity.Process               `json:"recent_activity"`
}

// ResourceUsageSummary represents resource usage summary
type ResourceUsageSummary struct {
	TotalCPUUsage    float64 `json:"total_cpu_usage"`
	AverageCPUUsage  float64 `json:"average_cpu_usage"`
	PeakCPUUsage     float64 `json:"peak_cpu_usage"`
	TotalMemoryUsage uint64  `json:"total_memory_usage"`
	AverageMemoryUsage uint64 `json:"average_memory_usage"`
	PeakMemoryUsage  uint64  `json:"peak_memory_usage"`
	ProcessCount     int64   `json:"process_count"`
}

// ProcessFilters represents filtering options for processes
type ProcessFilters struct {
	ExecutionID   *uuid.UUID
	Statuses      []entity.ProcessStatus
	StartedAfter  *time.Time
	StartedBefore *time.Time
	MinCPUUsage   *float64
	MaxCPUUsage   *float64
	MinMemoryUsage *uint64
	MaxMemoryUsage *uint64
	WithErrors    *bool
	Limit         *int
	Offset        *int
	OrderBy       *string // "start_time", "end_time", "cpu_usage", "memory_usage", "status"
	OrderDir      *string // "asc", "desc"
}