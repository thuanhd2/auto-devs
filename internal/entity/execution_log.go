package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LogLevel represents the level of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
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
	// ProcessID   *uuid.UUID `json:"process_id,omitempty" gorm:"type:uuid;index"`
	Level     LogLevel  `json:"log_level" gorm:"column:log_level; type:varchar(10);not null;index"`
	Message   string    `json:"message" gorm:"type:text;not null"`
	Timestamp time.Time `json:"timestamp" gorm:"not null;index"`
	Source    string    `json:"source" gorm:"type:varchar(50)"`       // stdout, stderr, system, etc.
	Metadata  JSONB     `json:"metadata,omitempty" gorm:"type:jsonb"` // Additional metadata as JSON
    // Structured fields parsed by backend
    LogType       string `json:"log_type" gorm:"type:varchar(20);index"`
    ToolName      string `json:"tool_name,omitempty" gorm:"type:varchar(100);index"`
    ToolUseID     string `json:"tool_use_id,omitempty" gorm:"type:varchar(100);index"`
    ParsedContent JSONB  `json:"parsed_content,omitempty" gorm:"type:jsonb"`
    IsError       *bool  `json:"is_error,omitempty" gorm:"type:boolean"`
    DurationMs    *int   `json:"duration_ms,omitempty" gorm:"type:int"`
    NumTurns      *int   `json:"num_turns,omitempty" gorm:"type:int"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Line      int       `json:"line" gorm:"type:int"`

	// Relationships
	Execution *Execution `json:"execution,omitempty" gorm:"foreignKey:ExecutionID;references:ID"`
	// Process   *Process   `json:"process,omitempty" gorm:"foreignKey:ProcessID;references:ID"`
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
