package repository

import (
	"context"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// ExecutionLogRepository defines the interface for execution log persistence
type ExecutionLogRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, log *entity.ExecutionLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ExecutionLog, error)
	GetByExecutionID(ctx context.Context, executionID uuid.UUID) ([]*entity.ExecutionLog, error)
	GetByProcessID(ctx context.Context, processID uuid.UUID) ([]*entity.ExecutionLog, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Batch operations for performance
	BatchCreate(ctx context.Context, logs []*entity.ExecutionLog) error
	BatchCreateAsync(ctx context.Context, logs []*entity.ExecutionLog) error
	GetLogsBatch(ctx context.Context, executionID uuid.UUID, limit, offset int) ([]*entity.ExecutionLog, error)

	// Filtering and search
	GetByLevel(ctx context.Context, executionID uuid.UUID, level entity.LogLevel) ([]*entity.ExecutionLog, error)
	GetByLevels(ctx context.Context, executionID uuid.UUID, levels []entity.LogLevel) ([]*entity.ExecutionLog, error)
	GetBySource(ctx context.Context, executionID uuid.UUID, source string) ([]*entity.ExecutionLog, error)
	GetByDateRange(ctx context.Context, executionID uuid.UUID, startDate, endDate time.Time) ([]*entity.ExecutionLog, error)
	GetRecentLogs(ctx context.Context, executionID uuid.UUID, limit int) ([]*entity.ExecutionLog, error)

	// Advanced queries
	SearchLogs(ctx context.Context, executionID uuid.UUID, searchTerm string) ([]*entity.ExecutionLog, error)
	GetLogStats(ctx context.Context, executionID uuid.UUID) (*LogStats, error)
	GetErrorLogs(ctx context.Context, executionID uuid.UUID, limit int) ([]*entity.ExecutionLog, error)
	GetLogsByTimeWindow(ctx context.Context, executionID uuid.UUID, windowStart, windowEnd time.Time) ([]*entity.ExecutionLog, error)

	// Log management and cleanup
	RotateLogs(ctx context.Context, executionID uuid.UUID, maxLogs int) error
	CleanupOldLogs(ctx context.Context, olderThan time.Time) (int64, error)
	CleanupExecutionLogs(ctx context.Context, executionID uuid.UUID, keepRecent int) (int64, error)
	ArchiveLogs(ctx context.Context, executionID uuid.UUID, olderThan time.Time) (int64, error)

	// Bulk operations
	BulkDelete(ctx context.Context, ids []uuid.UUID) error
	BulkDeleteByExecution(ctx context.Context, executionID uuid.UUID) (int64, error)
	BulkDeleteByLevel(ctx context.Context, level entity.LogLevel, olderThan time.Time) (int64, error)

	// Validation
	ValidateLogExists(ctx context.Context, id uuid.UUID) (bool, error)
	ValidateExecutionExists(ctx context.Context, executionID uuid.UUID) (bool, error)
}

// LogStats represents log statistics
type LogStats struct {
	TotalLogs        int64                     `json:"total_logs"`
	LogsByLevel      map[entity.LogLevel]int64 `json:"logs_by_level"`
	LogsBySource     map[string]int64          `json:"logs_by_source"`
	ErrorCount       int64                     `json:"error_count"`
	WarningCount     int64                     `json:"warning_count"`
	FirstLogTime     *time.Time                `json:"first_log_time,omitempty"`
	LastLogTime      *time.Time                `json:"last_log_time,omitempty"`
	RecentErrorLogs  []*entity.ExecutionLog    `json:"recent_error_logs"`
	LogSizeBytes     int64                     `json:"log_size_bytes"`
}

// LogFilters represents filtering options for logs
type LogFilters struct {
	ExecutionID   *uuid.UUID
	ProcessID     *uuid.UUID
	Levels        []entity.LogLevel
	Sources       []string
	SearchTerm    *string
	TimeAfter     *time.Time
	TimeBefore    *time.Time
	Limit         *int
	Offset        *int
	OrderBy       *string // "timestamp", "level", "source"
	OrderDir      *string // "asc", "desc"
}

// LogBatchConfig represents configuration for batch operations
type LogBatchConfig struct {
	BatchSize     int           `json:"batch_size"`
	FlushInterval time.Duration `json:"flush_interval"`
	AsyncBuffer   int           `json:"async_buffer"`
	RetryAttempts int           `json:"retry_attempts"`
	RetryDelay    time.Duration `json:"retry_delay"`
}

// Default configuration values
var DefaultLogBatchConfig = LogBatchConfig{
	BatchSize:     1000,
	FlushInterval: time.Second * 5,
	AsyncBuffer:   10000,
	RetryAttempts: 3,
	RetryDelay:    time.Millisecond * 100,
}