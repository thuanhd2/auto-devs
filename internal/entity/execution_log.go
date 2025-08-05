package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LogLevel represents the level of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// IsValid checks if the log level is valid
func (ll LogLevel) IsValid() bool {
	switch ll {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
		return true
	default:
		return false
	}
}

// ExecutionLog represents a log entry for an execution
type ExecutionLog struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ExecutionID uuid.UUID `json:"execution_id" gorm:"type:uuid;not null;index"`
	ProcessID   *uuid.UUID `json:"process_id,omitempty" gorm:"type:uuid;index"`
	Level       LogLevel  `json:"level" gorm:"type:varchar(10);not null;index"`
	Message     string    `json:"message" gorm:"type:text;not null"`
	Timestamp   time.Time `json:"timestamp" gorm:"not null;index"`
	Source      string    `json:"source" gorm:"type:varchar(50)"` // stdout, stderr, system, etc.
	Metadata    string    `json:"metadata,omitempty" gorm:"type:jsonb"` // Additional metadata as JSON
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Execution *Execution `json:"execution,omitempty" gorm:"foreignKey:ExecutionID;references:ID"`
	Process   *Process   `json:"process,omitempty" gorm:"foreignKey:ProcessID;references:ID"`
}

// TableName returns the table name for GORM
func (ExecutionLog) TableName() string {
	return "execution_logs"
}

// BeforeCreate sets default values before creating
func (el *ExecutionLog) BeforeCreate(tx *gorm.DB) error {
	if el.ID == uuid.Nil {
		el.ID = uuid.New()
	}
	if el.Timestamp.IsZero() {
		el.Timestamp = time.Now()
	}
	return nil
}