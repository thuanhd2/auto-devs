package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorktreeStatus string

const (
	WorktreeStatusCreating  WorktreeStatus = "creating"
	WorktreeStatusActive    WorktreeStatus = "active"
	WorktreeStatusCompleted WorktreeStatus = "completed"
	WorktreeStatusCleaning  WorktreeStatus = "cleaning"
	WorktreeStatusError     WorktreeStatus = "error"
)

// IsValid checks if the worktree status is valid
func (ws WorktreeStatus) IsValid() bool {
	switch ws {
	case WorktreeStatusCreating, WorktreeStatusActive, WorktreeStatusCompleted, WorktreeStatusCleaning, WorktreeStatusError:
		return true
	default:
		return false
	}
}

// String returns the string representation of WorktreeStatus
func (ws WorktreeStatus) String() string {
	return string(ws)
}

// GetDisplayName returns a user-friendly display name for the status
func (ws WorktreeStatus) GetDisplayName() string {
	switch ws {
	case WorktreeStatusCreating:
		return "Creating"
	case WorktreeStatusActive:
		return "Active"
	case WorktreeStatusCompleted:
		return "Completed"
	case WorktreeStatusCleaning:
		return "Cleaning"
	case WorktreeStatusError:
		return "Error"
	default:
		return string(ws)
	}
}

// GetAllWorktreeStatuses returns all valid worktree statuses
func GetAllWorktreeStatuses() []WorktreeStatus {
	return []WorktreeStatus{
		WorktreeStatusCreating,
		WorktreeStatusActive,
		WorktreeStatusCompleted,
		WorktreeStatusCleaning,
		WorktreeStatusError,
	}
}

// Worktree represents a Git worktree for a task
type Worktree struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID       uuid.UUID      `json:"task_id" gorm:"type:uuid;not null" validate:"required"`
	ProjectID    uuid.UUID      `json:"project_id" gorm:"type:uuid;not null" validate:"required"`
	BranchName   string         `json:"branch_name" gorm:"size:255;not null" validate:"required"`
	WorktreePath string         `json:"worktree_path" gorm:"type:text;not null" validate:"required"`
	Status       WorktreeStatus `json:"status" gorm:"size:50;not null;default:'creating'" validate:"required"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Task    Task    `json:"task,omitempty" gorm:"foreignKey:TaskID"`
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// WorktreeStatusTransitions defines valid transitions between worktree statuses
var WorktreeStatusTransitions = map[WorktreeStatus][]WorktreeStatus{
	WorktreeStatusCreating: {
		WorktreeStatusActive,
		WorktreeStatusError,
	},
	WorktreeStatusActive: {
		WorktreeStatusCompleted,
		WorktreeStatusCleaning,
		WorktreeStatusError,
	},
	WorktreeStatusCompleted: {
		WorktreeStatusCleaning,
		WorktreeStatusActive, // Allow reactivation
	},
	WorktreeStatusCleaning: {
		WorktreeStatusError, // If cleanup fails
	},
	WorktreeStatusError: {
		WorktreeStatusCreating, // Allow retry
		WorktreeStatusCleaning, // Allow cleanup after error
	},
}

// CanTransitionTo checks if the current status can transition to the target status
func (ws WorktreeStatus) CanTransitionTo(target WorktreeStatus) bool {
	allowedTransitions, exists := WorktreeStatusTransitions[ws]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == target {
			return true
		}
	}
	return false
}

// WorktreeStatusValidationError represents an error when attempting invalid status transitions
type WorktreeStatusValidationError struct {
	CurrentStatus WorktreeStatus
	TargetStatus  WorktreeStatus
	Message       string
}

func (e *WorktreeStatusValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "invalid worktree status transition from " + string(e.CurrentStatus) + " to " + string(e.TargetStatus)
}

// ValidateWorktreeStatusTransition validates if a status transition is allowed
func ValidateWorktreeStatusTransition(from, to WorktreeStatus) error {
	if !from.IsValid() {
		return &WorktreeStatusValidationError{
			CurrentStatus: from,
			TargetStatus:  to,
			Message:       "invalid current status: " + string(from),
		}
	}

	if !to.IsValid() {
		return &WorktreeStatusValidationError{
			CurrentStatus: from,
			TargetStatus:  to,
			Message:       "invalid target status: " + string(to),
		}
	}

	if !from.CanTransitionTo(to) {
		return &WorktreeStatusValidationError{
			CurrentStatus: from,
			TargetStatus:  to,
		}
	}

	return nil
}

// WorktreeFilters represents filtering options for worktrees
type WorktreeFilters struct {
	ProjectID     *uuid.UUID
	TaskID        *uuid.UUID
	Statuses      []WorktreeStatus
	BranchName    *string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Limit         *int
	Offset        *int
	OrderBy       *string // "created_at", "updated_at", "branch_name", "status"
	OrderDir      *string // "asc", "desc"
}

// WorktreeStatistics represents worktree statistics for a project
type WorktreeStatistics struct {
	ProjectID           uuid.UUID              `json:"project_id"`
	TotalWorktrees      int                    `json:"total_worktrees"`
	ActiveWorktrees     int                    `json:"active_worktrees"`
	CompletedWorktrees  int                    `json:"completed_worktrees"`
	ErrorWorktrees      int                    `json:"error_worktrees"`
	WorktreesByStatus   map[WorktreeStatus]int `json:"worktrees_by_status"`
	AverageCreationTime float64                `json:"average_creation_time"` // in seconds
	GeneratedAt         time.Time              `json:"generated_at"`
}
