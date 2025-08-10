package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExecutionStatus represents the current status of an execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "PENDING"
	ExecutionStatusRunning   ExecutionStatus = "RUNNING"
	ExecutionStatusPaused    ExecutionStatus = "PAUSED"
	ExecutionStatusCompleted ExecutionStatus = "COMPLETED"
	ExecutionStatusFailed    ExecutionStatus = "FAILED"
	ExecutionStatusCancelled ExecutionStatus = "CANCELLED"
)

// IsValid checks if the execution status is valid
func (es ExecutionStatus) IsValid() bool {
	switch es {
	case ExecutionStatusPending, ExecutionStatusRunning, ExecutionStatusPaused,
		ExecutionStatusCompleted, ExecutionStatusFailed, ExecutionStatusCancelled:
		return true
	default:
		return false
	}
}

// Execution represents an AI execution instance
type Execution struct {
	ID           uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID       uuid.UUID       `json:"task_id" gorm:"type:uuid;not null;index"`
	Status       ExecutionStatus `json:"status" gorm:"type:varchar(20);not null;index"`
	StartedAt    time.Time       `json:"started_at" gorm:"not null"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty" gorm:"type:text"`
	Progress     float64         `json:"progress" gorm:"default:0.0;check:progress >= 0 AND progress <= 1"`
	Result       *string         `json:"result,omitempty" gorm:"type:jsonb"` // JSON serialized ExecutionResult
	CreatedAt    time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Task      *Task          `json:"task,omitempty" gorm:"foreignKey:TaskID;references:ID"`
	Processes []Process      `json:"processes,omitempty" gorm:"foreignKey:ExecutionID;references:ID"`
	Logs      []ExecutionLog `json:"logs,omitempty" gorm:"foreignKey:ExecutionID;references:ID"`
}

// ExecutionResult represents the result of an execution
type ExecutionResult struct {
	Output   string                 `json:"output"`
	Files    []string               `json:"files"`
	Metrics  map[string]interface{} `json:"metrics"`
	Duration time.Duration          `json:"duration"`
}

// TableName returns the table name for GORM
func (Execution) TableName() string {
	return "executions"
}

// BeforeCreate sets default values before creating
func (e *Execution) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.Progress < 0 {
		e.Progress = 0.0
	}
	return nil
}

// IsCompleted checks if the execution is in a completed state
func (e *Execution) IsCompleted() bool {
	return e.Status == ExecutionStatusCompleted ||
		e.Status == ExecutionStatusFailed ||
		e.Status == ExecutionStatusCancelled
}

// IsActive checks if the execution is currently active
func (e *Execution) IsActive() bool {
	return e.Status == ExecutionStatusRunning ||
		e.Status == ExecutionStatusPaused
}

// GetDuration returns the duration of the execution
func (e *Execution) GetDuration() time.Duration {
	if e.CompletedAt != nil {
		return e.CompletedAt.Sub(e.StartedAt)
	}
	return time.Since(e.StartedAt)
}
