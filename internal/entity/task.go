package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskStatus string

const (
	TaskStatusTODO           TaskStatus = "TODO"
	TaskStatusPLANNING      TaskStatus = "PLANNING"
	TaskStatusPLANREVIEWING TaskStatus = "PLAN_REVIEWING"
	TaskStatusIMPLEMENTING  TaskStatus = "IMPLEMENTING"
	TaskStatusCODEREVIEWING TaskStatus = "CODE_REVIEWING"
	TaskStatusDONE          TaskStatus = "DONE"
	TaskStatusCANCELLED     TaskStatus = "CANCELLED"
)

// TaskStatusTransitions defines valid transitions between task statuses
var TaskStatusTransitions = map[TaskStatus][]TaskStatus{
	TaskStatusTODO: {
		TaskStatusPLANNING,
		TaskStatusCANCELLED,
	},
	TaskStatusPLANNING: {
		TaskStatusPLANREVIEWING,
		TaskStatusTODO,
		TaskStatusCANCELLED,
	},
	TaskStatusPLANREVIEWING: {
		TaskStatusIMPLEMENTING,
		TaskStatusPLANNING,
		TaskStatusCANCELLED,
	},
	TaskStatusIMPLEMENTING: {
		TaskStatusCODEREVIEWING,
		TaskStatusPLANREVIEWING,
		TaskStatusCANCELLED,
	},
	TaskStatusCODEREVIEWING: {
		TaskStatusDONE,
		TaskStatusIMPLEMENTING,
		TaskStatusCANCELLED,
	},
	TaskStatusDONE: {
		TaskStatusTODO, // Allow reopening tasks
	},
	TaskStatusCANCELLED: {
		TaskStatusTODO, // Allow reactivating cancelled tasks
	},
}

// IsValid checks if the task status is valid
func (ts TaskStatus) IsValid() bool {
	switch ts {
	case TaskStatusTODO, TaskStatusPLANNING, TaskStatusPLANREVIEWING,
		 TaskStatusIMPLEMENTING, TaskStatusCODEREVIEWING, TaskStatusDONE, TaskStatusCANCELLED:
		return true
	default:
		return false
	}
}

// String returns the string representation of TaskStatus
func (ts TaskStatus) String() string {
	return string(ts)
}

// CanTransitionTo checks if the current status can transition to the target status
func (ts TaskStatus) CanTransitionTo(target TaskStatus) bool {
	allowedTransitions, exists := TaskStatusTransitions[ts]
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

// GetAllStatuses returns all valid task statuses
func GetAllTaskStatuses() []TaskStatus {
	return []TaskStatus{
		TaskStatusTODO,
		TaskStatusPLANNING,
		TaskStatusPLANREVIEWING,
		TaskStatusIMPLEMENTING,
		TaskStatusCODEREVIEWING,
		TaskStatusDONE,
		TaskStatusCANCELLED,
	}
}

// GetStatusDisplayName returns a user-friendly display name for the status
func (ts TaskStatus) GetDisplayName() string {
	switch ts {
	case TaskStatusTODO:
		return "To Do"
	case TaskStatusPLANNING:
		return "Planning"
	case TaskStatusPLANREVIEWING:
		return "Plan Review"
	case TaskStatusIMPLEMENTING:
		return "Implementing"
	case TaskStatusCODEREVIEWING:
		return "Code Review"
	case TaskStatusDONE:
		return "Done"
	case TaskStatusCANCELLED:
		return "Cancelled"
	default:
		return string(ts)
	}
}

type Task struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID   uuid.UUID      `json:"project_id" gorm:"type:uuid;not null" validate:"required"`
	Title       string         `json:"title" gorm:"size:255;not null" validate:"required,min=1,max=255"`
	Description string         `json:"description" gorm:"size:1000" validate:"max=1000"`
	Status      TaskStatus     `json:"status" gorm:"size:50;not null;default:'TODO'" validate:"required,oneof=TODO PLANNING PLAN_REVIEWING IMPLEMENTING CODE_REVIEWING DONE CANCELLED"`
	BranchName  *string        `json:"branch_name,omitempty" gorm:"size:255"`
	PullRequest *string        `json:"pull_request,omitempty" gorm:"size:255"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// TaskStatusHistory tracks changes to task status over time
type TaskStatusHistory struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID     uuid.UUID      `json:"task_id" gorm:"type:uuid;not null" validate:"required"`
	FromStatus *TaskStatus    `json:"from_status,omitempty" gorm:"size:50"` // null for initial status
	ToStatus   TaskStatus     `json:"to_status" gorm:"size:50;not null" validate:"required"`
	ChangedBy  *string        `json:"changed_by,omitempty" gorm:"size:255"` // user ID or system identifier
	Reason     *string        `json:"reason,omitempty" gorm:"size:500"`     // optional reason for status change
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Task Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

// TaskStatusValidationError represents an error when attempting invalid status transitions
type TaskStatusValidationError struct {
	CurrentStatus TaskStatus
	TargetStatus  TaskStatus
	Message       string
}

func (e *TaskStatusValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("invalid status transition from %s to %s", e.CurrentStatus, e.TargetStatus)
}

// ValidateStatusTransition validates if a status transition is allowed
func ValidateStatusTransition(from, to TaskStatus) error {
	if !from.IsValid() {
		return &TaskStatusValidationError{
			CurrentStatus: from,
			TargetStatus:  to,
			Message:       fmt.Sprintf("invalid current status: %s", from),
		}
	}
	
	if !to.IsValid() {
		return &TaskStatusValidationError{
			CurrentStatus: from,
			TargetStatus:  to,
			Message:       fmt.Sprintf("invalid target status: %s", to),
		}
	}
	
	if !from.CanTransitionTo(to) {
		return &TaskStatusValidationError{
			CurrentStatus: from,
			TargetStatus:  to,
		}
	}
	
	return nil
}

// TaskStatusStats represents statistics for task statuses
type TaskStatusStats struct {
	Status TaskStatus `json:"status"`
	Count  int        `json:"count"`
}

// TaskStatusAnalytics represents comprehensive status analytics
type TaskStatusAnalytics struct {
	ProjectID              uuid.UUID         `json:"project_id"`
	StatusDistribution     []TaskStatusStats `json:"status_distribution"`
	AverageTimeInStatus    map[TaskStatus]float64 `json:"average_time_in_status"` // in hours
	TransitionCount        map[string]int    `json:"transition_count"`           // from_status->to_status counts
	TotalTasks             int               `json:"total_tasks"`
	CompletedTasks         int               `json:"completed_tasks"`
	CompletionRate         float64           `json:"completion_rate"`
	GeneratedAt            time.Time         `json:"generated_at"`
}
