package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProcessStatus represents the current status of a process
type ProcessStatus string

const (
	ProcessStatusStarting ProcessStatus = "starting"
	ProcessStatusRunning  ProcessStatus = "running"
	ProcessStatusStopped  ProcessStatus = "stopped"
	ProcessStatusKilled   ProcessStatus = "killed"
	ProcessStatusError    ProcessStatus = "error"
)

// IsValid checks if the process status is valid
func (ps ProcessStatus) IsValid() bool {
	switch ps {
	case ProcessStatusStarting, ProcessStatusRunning, ProcessStatusStopped,
		ProcessStatusKilled, ProcessStatusError:
		return true
	default:
		return false
	}
}

// Process represents an AI execution process
type Process struct {
	ID          uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ExecutionID uuid.UUID     `json:"execution_id" gorm:"type:uuid;not null;index"`
	Command     string        `json:"command" gorm:"type:text;not null"`
	WorkDir     string        `json:"work_dir" gorm:"type:varchar(512)"`
	PID         int           `json:"pid" gorm:"index"`
	Status      ProcessStatus `json:"status" gorm:"type:varchar(20);not null;index"`
	StartTime   time.Time     `json:"start_time" gorm:"not null"`
	EndTime     *time.Time    `json:"end_time,omitempty"`
	ExitCode    *int          `json:"exit_code,omitempty"`
	Error       string        `json:"error,omitempty" gorm:"type:text"`
	CPUUsage    float64       `json:"cpu_usage" gorm:"default:0.0"`
	MemoryUsage uint64        `json:"memory_usage" gorm:"default:0"`
	CreatedAt   time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Execution *Execution `json:"execution,omitempty" gorm:"foreignKey:ExecutionID;references:ID"`
}

// TableName returns the table name for GORM
func (Process) TableName() string {
	return "processes"
}

// BeforeCreate sets default values before creating
func (p *Process) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// IsRunning checks if the process is currently running
func (p *Process) IsRunning() bool {
	return p.Status == ProcessStatusRunning
}

// IsCompleted checks if the process is in a completed state
func (p *Process) IsCompleted() bool {
	return p.Status == ProcessStatusStopped ||
		p.Status == ProcessStatusKilled ||
		p.Status == ProcessStatusError
}

// GetDuration returns the duration the process has been running
func (p *Process) GetDuration() time.Duration {
	if p.EndTime != nil {
		return p.EndTime.Sub(p.StartTime)
	}
	return time.Since(p.StartTime)
}