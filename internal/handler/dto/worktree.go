package dto

import (
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/google/uuid"
)

// CreateWorktreeRequest represents a request to create a worktree for a task
type CreateWorktreeRequest struct {
	TaskID     uuid.UUID `json:"task_id" binding:"required"`
	ProjectID  uuid.UUID `json:"project_id" binding:"required"`
	TaskTitle  string    `json:"task_title" binding:"required"`
	Repository string    `json:"repository,omitempty"` // Optional repository URL to clone
}

// CleanupWorktreeRequest represents a request to cleanup a worktree for a task
type CleanupWorktreeRequest struct {
	TaskID     uuid.UUID `json:"task_id" binding:"required"`
	ProjectID  uuid.UUID `json:"project_id" binding:"required"`
	BranchName string    `json:"branch_name,omitempty"` // Optional branch name to delete
	Force      bool      `json:"force"`                 // Force cleanup even if worktree is active
}

// UpdateWorktreeStatusRequest represents a request to update worktree status
type UpdateWorktreeStatusRequest struct {
	Status entity.WorktreeStatus `json:"status" binding:"required"`
}

// WorktreeResponse represents a worktree response
type WorktreeResponse struct {
	Worktree *entity.Worktree `json:"worktree"`
	Message  string           `json:"message,omitempty"`
}

// WorktreesResponse represents a list of worktrees response
type WorktreesResponse struct {
	Worktrees []*entity.Worktree `json:"worktrees"`
	Count     int                `json:"count"`
}

// WorktreeValidationResponse represents a worktree validation response
type WorktreeValidationResponse struct {
	ValidationResult *usecase.WorktreeValidationResult `json:"validation_result"`
}

// WorktreeHealthResponse represents a worktree health response
type WorktreeHealthResponse struct {
	Health *usecase.WorktreeHealthInfo `json:"health"`
}

// BranchInfoResponse represents a branch info response
type BranchInfoResponse struct {
	BranchInfo *usecase.BranchInfo `json:"branch_info"`
}

// WorktreeStatisticsResponse represents a worktree statistics response
type WorktreeStatisticsResponse struct {
	Statistics *entity.WorktreeStatistics `json:"statistics"`
}

// WorktreeCountResponse represents a worktree count response
type WorktreeCountResponse struct {
	Count int `json:"count"`
}

// WorktreeFiltersRequest represents worktree filters for API requests
type WorktreeFiltersRequest struct {
	ProjectID     *uuid.UUID              `json:"project_id,omitempty"`
	TaskID        *uuid.UUID              `json:"task_id,omitempty"`
	Statuses      []entity.WorktreeStatus `json:"statuses,omitempty"`
	BranchName    *string                 `json:"branch_name,omitempty"`
	CreatedAfter  *time.Time              `json:"created_after,omitempty"`
	CreatedBefore *time.Time              `json:"created_before,omitempty"`
	Limit         *int                    `json:"limit,omitempty"`
	Offset        *int                    `json:"offset,omitempty"`
	OrderBy       *string                 `json:"order_by,omitempty"`
	OrderDir      *string                 `json:"order_dir,omitempty"`
}

// WorktreeFiltersResponse represents a filtered worktrees response
type WorktreeFiltersResponse struct {
	Worktrees []*entity.Worktree `json:"worktrees"`
	Count     int                `json:"count"`
	Limit     *int               `json:"limit,omitempty"`
	Offset    *int               `json:"offset,omitempty"`
	Total     int                `json:"total"`
}

// WorktreeSummary represents a summary of worktree information
type WorktreeSummary struct {
	ID          uuid.UUID             `json:"id"`
	TaskID      uuid.UUID             `json:"task_id"`
	ProjectID   uuid.UUID             `json:"project_id"`
	BranchName  string                `json:"branch_name"`
	Status      entity.WorktreeStatus `json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	TaskTitle   string                `json:"task_title,omitempty"`
	ProjectName string                `json:"project_name,omitempty"`
	IsHealthy   bool                  `json:"is_healthy,omitempty"`
	HealthScore int                   `json:"health_score,omitempty"`
}

// WorktreeSummaryResponse represents a worktree summary response
type WorktreeSummaryResponse struct {
	Worktrees []*WorktreeSummary `json:"worktrees"`
	Count     int                `json:"count"`
}

// WorktreeBulkOperationRequest represents a bulk operation request for worktrees
type WorktreeBulkOperationRequest struct {
	WorktreeIDs []uuid.UUID            `json:"worktree_ids" binding:"required,min=1"`
	Operation   string                 `json:"operation" binding:"required"` // "delete", "update_status", "cleanup"
	Status      *entity.WorktreeStatus `json:"status,omitempty"`             // For update_status operation
	Force       *bool                  `json:"force,omitempty"`              // For cleanup operation
}

// WorktreeBulkOperationResponse represents a bulk operation response
type WorktreeBulkOperationResponse struct {
	Operation    string      `json:"operation"`
	SuccessCount int         `json:"success_count"`
	FailedCount  int         `json:"failed_count"`
	FailedIDs    []uuid.UUID `json:"failed_ids,omitempty"`
	Message      string      `json:"message"`
}

// WorktreeMetrics represents worktree metrics for monitoring
type WorktreeMetrics struct {
	TotalWorktrees     int            `json:"total_worktrees"`
	ActiveWorktrees    int            `json:"active_worktrees"`
	CompletedWorktrees int            `json:"completed_worktrees"`
	ErrorWorktrees     int            `json:"error_worktrees"`
	CreatingWorktrees  int            `json:"creating_worktrees"`
	CleaningWorktrees  int            `json:"cleaning_worktrees"`
	WorktreesByStatus  map[string]int `json:"worktrees_by_status"`
	AverageHealthScore float64        `json:"average_health_score"`
	UnhealthyWorktrees int            `json:"unhealthy_worktrees"`
	OrphanedWorktrees  int            `json:"orphaned_worktrees"`
	GeneratedAt        time.Time      `json:"generated_at"`
}

// WorktreeMetricsResponse represents a worktree metrics response
type WorktreeMetricsResponse struct {
	Metrics *WorktreeMetrics `json:"metrics"`
}

// WorktreeError represents a worktree error
type WorktreeError struct {
	WorktreeID uuid.UUID `json:"worktree_id"`
	TaskID     uuid.UUID `json:"task_id"`
	Error      string    `json:"error"`
	Timestamp  time.Time `json:"timestamp"`
	Severity   string    `json:"severity"` // "low", "medium", "high", "critical"
}

// WorktreeErrorsResponse represents a worktree errors response
type WorktreeErrorsResponse struct {
	Errors []*WorktreeError `json:"errors"`
	Count  int              `json:"count"`
}

// WorktreeRecoveryRequest represents a worktree recovery request
type WorktreeRecoveryRequest struct {
	WorktreeID uuid.UUID `json:"worktree_id" binding:"required"`
	Force      bool      `json:"force"` // Force recovery even if worktree is not in error state
}

// WorktreeRecoveryResponse represents a worktree recovery response
type WorktreeRecoveryResponse struct {
	WorktreeID    uuid.UUID `json:"worktree_id"`
	Success       bool      `json:"success"`
	Message       string    `json:"message"`
	RecoverySteps []string  `json:"recovery_steps,omitempty"`
}

// WorktreeConfiguration represents worktree configuration
type WorktreeConfiguration struct {
	BaseDirectory   string        `json:"base_directory"`
	MaxPathLength   int           `json:"max_path_length"`
	MinDiskSpace    int64         `json:"min_disk_space"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	EnableLogging   bool          `json:"enable_logging"`
	LogLevel        string        `json:"log_level"`
}

// WorktreeConfigurationResponse represents a worktree configuration response
type WorktreeConfigurationResponse struct {
	Configuration *WorktreeConfiguration `json:"configuration"`
}

// WorktreeConfigurationUpdateRequest represents a worktree configuration update request
type WorktreeConfigurationUpdateRequest struct {
	BaseDirectory   *string        `json:"base_directory,omitempty"`
	MaxPathLength   *int           `json:"max_path_length,omitempty"`
	MinDiskSpace    *int64         `json:"min_disk_space,omitempty"`
	CleanupInterval *time.Duration `json:"cleanup_interval,omitempty"`
	EnableLogging   *bool          `json:"enable_logging,omitempty"`
	LogLevel        *string        `json:"log_level,omitempty"`
}
