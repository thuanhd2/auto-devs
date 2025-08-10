package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/google/uuid"
)

// ExecutionUsecase defines the interface for execution business logic
type ExecutionUsecase interface {
	// Basic operations
	Create(ctx context.Context, req CreateExecutionRequest) (*entity.Execution, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Execution, error)
	GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]*entity.Execution, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateExecutionRequest) (*entity.Execution, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Status management
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ExecutionStatus) (*entity.Execution, error)
	UpdateProgress(ctx context.Context, id uuid.UUID, progress float64) (*entity.Execution, error)
	MarkCompleted(ctx context.Context, id uuid.UUID, result *entity.ExecutionResult) (*entity.Execution, error)
	MarkFailed(ctx context.Context, id uuid.UUID, errorMsg string) (*entity.Execution, error)

	// Advanced queries
	GetWithLogs(ctx context.Context, id uuid.UUID, logLimit int) (*entity.Execution, error)
	GetWithProcesses(ctx context.Context, id uuid.UUID) (*entity.Execution, error)
	GetByStatusFiltered(ctx context.Context, req GetExecutionsFilterRequest) ([]*entity.Execution, int64, error)
	GetExecutionStats(ctx context.Context, taskID *uuid.UUID) (*repository.ExecutionStats, error)
	GetRecentExecutions(ctx context.Context, limit int) ([]*entity.Execution, error)

	// Log operations
	GetExecutionLogs(ctx context.Context, executionID uuid.UUID, req GetExecutionLogsRequest) ([]*entity.ExecutionLog, int64, error)
	AddExecutionLog(ctx context.Context, req AddExecutionLogRequest) (*entity.ExecutionLog, error)
	BatchAddLogs(ctx context.Context, logs []AddExecutionLogRequest) error
	GetLogStats(ctx context.Context, executionID uuid.UUID) (*repository.LogStats, error)

	// Validation
	ValidateExecutionExists(ctx context.Context, id uuid.UUID) error
	ValidateTaskExists(ctx context.Context, taskID uuid.UUID) error
}

// Request DTOs for usecase
type CreateExecutionRequest struct {
	TaskID uuid.UUID `json:"task_id"`
}

type UpdateExecutionRequest struct {
	Status   *entity.ExecutionStatus `json:"status,omitempty"`
	Progress *float64                `json:"progress,omitempty"`
	Error    *string                 `json:"error,omitempty"`
	Result   *entity.ExecutionResult `json:"result,omitempty"`
}

type GetExecutionsFilterRequest struct {
	TaskID        *uuid.UUID
	Statuses      []entity.ExecutionStatus
	StartedAfter  *time.Time
	StartedBefore *time.Time
	WithErrors    *bool
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
}

type GetExecutionLogsRequest struct {
	Levels     []entity.LogLevel
	Sources    []string
	SearchTerm *string
	TimeAfter  *time.Time
	TimeBefore *time.Time
	Limit      int
	Offset     int
	OrderBy    string
	OrderDir   string
}

type AddExecutionLogRequest struct {
	ExecutionID uuid.UUID       `json:"execution_id"`
	ProcessID   *uuid.UUID      `json:"process_id,omitempty"`
	Level       entity.LogLevel `json:"level"`
	Message     string          `json:"message"`
	Source      string          `json:"source"`
	Metadata    string          `json:"metadata,omitempty"`
	Timestamp   *time.Time      `json:"timestamp,omitempty"`
}

// ExecutionUsecaseImpl implements ExecutionUsecase
type ExecutionUsecaseImpl struct {
	executionRepo    repository.ExecutionRepository
	executionLogRepo repository.ExecutionLogRepository
	taskRepo         repository.TaskRepository
}

// NewExecutionUsecase creates a new execution usecase
func NewExecutionUsecase(
	executionRepo repository.ExecutionRepository,
	executionLogRepo repository.ExecutionLogRepository,
	taskRepo repository.TaskRepository,
) ExecutionUsecase {
	return &ExecutionUsecaseImpl{
		executionRepo:    executionRepo,
		executionLogRepo: executionLogRepo,
		taskRepo:         taskRepo,
	}
}

// Create creates a new execution
func (u *ExecutionUsecaseImpl) Create(ctx context.Context, req CreateExecutionRequest) (*entity.Execution, error) {
	// Validate that the task exists
	if err := u.ValidateTaskExists(ctx, req.TaskID); err != nil {
		return nil, err
	}

	execution := &entity.Execution{
		TaskID:    req.TaskID,
		Status:    entity.ExecutionStatusPending,
		StartedAt: time.Now(),
		Progress:  0.0,
	}

	if err := u.executionRepo.Create(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	return execution, nil
}

// GetByID retrieves an execution by ID
func (u *ExecutionUsecaseImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Execution, error) {
	execution, err := u.executionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}
	return execution, nil
}

// GetByTaskID retrieves all executions for a task
func (u *ExecutionUsecaseImpl) GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]*entity.Execution, error) {
	executions, err := u.executionRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get executions for task: %w", err)
	}
	return executions, nil
}

// Update updates an execution
func (u *ExecutionUsecaseImpl) Update(ctx context.Context, id uuid.UUID, req UpdateExecutionRequest) (*entity.Execution, error) {
	execution, err := u.executionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("execution not found: %w", err)
	}

	if req.Status != nil {
		execution.Status = *req.Status
	}
	if req.Progress != nil {
		execution.Progress = *req.Progress
	}
	if req.Error != nil {
		execution.ErrorMessage = *req.Error
	}

	if err := u.executionRepo.Update(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to update execution: %w", err)
	}

	return execution, nil
}

// Delete deletes an execution
func (u *ExecutionUsecaseImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := u.ValidateExecutionExists(ctx, id); err != nil {
		return err
	}

	if err := u.executionRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete execution: %w", err)
	}

	return nil
}

// UpdateStatus updates the execution status
func (u *ExecutionUsecaseImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ExecutionStatus) (*entity.Execution, error) {
	if err := u.executionRepo.UpdateStatus(ctx, id, status); err != nil {
		return nil, fmt.Errorf("failed to update execution status: %w", err)
	}

	return u.GetByID(ctx, id)
}

// UpdateProgress updates the execution progress
func (u *ExecutionUsecaseImpl) UpdateProgress(ctx context.Context, id uuid.UUID, progress float64) (*entity.Execution, error) {
	if progress < 0.0 || progress > 1.0 {
		return nil, fmt.Errorf("progress must be between 0.0 and 1.0")
	}

	if err := u.executionRepo.UpdateProgress(ctx, id, progress); err != nil {
		return nil, fmt.Errorf("failed to update execution progress: %w", err)
	}

	return u.GetByID(ctx, id)
}

// MarkCompleted marks an execution as completed
func (u *ExecutionUsecaseImpl) MarkCompleted(ctx context.Context, id uuid.UUID, result *entity.ExecutionResult) (*entity.Execution, error) {
	completedAt := time.Now()
	if err := u.executionRepo.MarkCompleted(ctx, id, completedAt, result); err != nil {
		return nil, fmt.Errorf("failed to mark execution as completed: %w", err)
	}

	return u.GetByID(ctx, id)
}

// MarkFailed marks an execution as failed
func (u *ExecutionUsecaseImpl) MarkFailed(ctx context.Context, id uuid.UUID, errorMsg string) (*entity.Execution, error) {
	completedAt := time.Now()
	if err := u.executionRepo.MarkFailed(ctx, id, completedAt, errorMsg); err != nil {
		return nil, fmt.Errorf("failed to mark execution as failed: %w", err)
	}

	return u.GetByID(ctx, id)
}

// GetWithLogs retrieves an execution with its logs
func (u *ExecutionUsecaseImpl) GetWithLogs(ctx context.Context, id uuid.UUID, logLimit int) (*entity.Execution, error) {
	execution, err := u.executionRepo.GetWithLogs(ctx, id, logLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution with logs: %w", err)
	}
	return execution, nil
}

// GetWithProcesses retrieves an execution with its processes
func (u *ExecutionUsecaseImpl) GetWithProcesses(ctx context.Context, id uuid.UUID) (*entity.Execution, error) {
	execution, err := u.executionRepo.GetWithProcesses(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution with processes: %w", err)
	}
	return execution, nil
}

// GetByStatusFiltered retrieves executions with filtering
func (u *ExecutionUsecaseImpl) GetByStatusFiltered(ctx context.Context, req GetExecutionsFilterRequest) ([]*entity.Execution, int64, error) {
	// Convert request to repository filters
	filters := repository.ExecutionFilters{
		TaskID:        req.TaskID,
		Statuses:      req.Statuses,
		StartedAfter:  req.StartedAfter,
		StartedBefore: req.StartedBefore,
		WithErrors:    req.WithErrors,
		Limit:         &req.Limit,
		Offset:        &req.Offset,
		OrderBy:       &req.OrderBy,
		OrderDir:      &req.OrderDir,
	}

	log.Println("filters", filters)

	// For now, return simple implementation
	// In a real implementation, you'd extend the repository interface to support filters
	if req.TaskID != nil {
		executions, err := u.executionRepo.GetByTaskID(ctx, *req.TaskID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get filtered executions: %w", err)
		}
		return executions, int64(len(executions)), nil
	}

	// If no TaskID filter, return recent executions
	executions, err := u.executionRepo.GetRecentExecutions(ctx, req.Limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get recent executions: %w", err)
	}
	return executions, int64(len(executions)), nil
}

// GetExecutionStats retrieves execution statistics
func (u *ExecutionUsecaseImpl) GetExecutionStats(ctx context.Context, taskID *uuid.UUID) (*repository.ExecutionStats, error) {
	stats, err := u.executionRepo.GetExecutionStats(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution stats: %w", err)
	}
	return stats, nil
}

// GetRecentExecutions retrieves recent executions
func (u *ExecutionUsecaseImpl) GetRecentExecutions(ctx context.Context, limit int) ([]*entity.Execution, error) {
	executions, err := u.executionRepo.GetRecentExecutions(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent executions: %w", err)
	}
	return executions, nil
}

// GetExecutionLogs retrieves execution logs with filtering
func (u *ExecutionUsecaseImpl) GetExecutionLogs(ctx context.Context, executionID uuid.UUID, req GetExecutionLogsRequest) ([]*entity.ExecutionLog, int64, error) {
	if err := u.ValidateExecutionExists(ctx, executionID); err != nil {
		return nil, 0, err
	}

	// For simple implementation, return all logs for the execution
	logs, err := u.executionLogRepo.GetByExecutionID(ctx, executionID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get execution logs: %w", err)
	}

	return logs, int64(len(logs)), nil
}

// AddExecutionLog adds a new execution log
func (u *ExecutionUsecaseImpl) AddExecutionLog(ctx context.Context, req AddExecutionLogRequest) (*entity.ExecutionLog, error) {
	if err := u.ValidateExecutionExists(ctx, req.ExecutionID); err != nil {
		return nil, err
	}

	timestamp := time.Now()
	if req.Timestamp != nil {
		timestamp = *req.Timestamp
	}

	metadata := entity.JSONB{}
	if req.Metadata != "" {
		err := json.Unmarshal([]byte(req.Metadata), &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	log := &entity.ExecutionLog{
		ExecutionID: req.ExecutionID,
		// ProcessID:   req.ProcessID,
		Level:     req.Level,
		Message:   req.Message,
		Source:    req.Source,
		Metadata:  metadata,
		Timestamp: timestamp,
	}

	if err := u.executionLogRepo.Create(ctx, log); err != nil {
		return nil, fmt.Errorf("failed to add execution log: %w", err)
	}

	return log, nil
}

// BatchAddLogs adds multiple logs in a batch
func (u *ExecutionUsecaseImpl) BatchAddLogs(ctx context.Context, logReqs []AddExecutionLogRequest) error {
	logs := make([]*entity.ExecutionLog, len(logReqs))
	for i, req := range logReqs {
		timestamp := time.Now()
		if req.Timestamp != nil {
			timestamp = *req.Timestamp
		}

		metadata := entity.JSONB{}
		if req.Metadata != "" {
			err := json.Unmarshal([]byte(req.Metadata), &metadata)
			if err != nil {
				return fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		logs[i] = &entity.ExecutionLog{
			ExecutionID: req.ExecutionID,
			// ProcessID:   req.ProcessID,
			Level:     req.Level,
			Message:   req.Message,
			Source:    req.Source,
			Metadata:  metadata,
			Timestamp: timestamp,
		}
	}

	if err := u.executionLogRepo.BatchCreate(ctx, logs); err != nil {
		return fmt.Errorf("failed to batch add execution logs: %w", err)
	}

	return nil
}

// GetLogStats retrieves log statistics
func (u *ExecutionUsecaseImpl) GetLogStats(ctx context.Context, executionID uuid.UUID) (*repository.LogStats, error) {
	if err := u.ValidateExecutionExists(ctx, executionID); err != nil {
		return nil, err
	}

	stats, err := u.executionLogRepo.GetLogStats(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get log stats: %w", err)
	}
	return stats, nil
}

// ValidateExecutionExists validates that an execution exists
func (u *ExecutionUsecaseImpl) ValidateExecutionExists(ctx context.Context, id uuid.UUID) error {
	exists, err := u.executionRepo.ValidateExecutionExists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to validate execution existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("execution not found")
	}
	return nil
}

// ValidateTaskExists validates that a task exists
func (u *ExecutionUsecaseImpl) ValidateTaskExists(ctx context.Context, taskID uuid.UUID) error {
	exists, err := u.executionRepo.ValidateTaskExists(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to validate task existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("task not found")
	}
	return nil
}
