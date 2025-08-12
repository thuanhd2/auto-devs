package entity

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskStatus string

const (
	TaskStatusTODO          TaskStatus = "TODO"
	TaskStatusPLANNING      TaskStatus = "PLANNING"
	TaskStatusPLANREVIEWING TaskStatus = "PLAN_REVIEWING"
	TaskStatusIMPLEMENTING  TaskStatus = "IMPLEMENTING"
	TaskStatusCODEREVIEWING TaskStatus = "CODE_REVIEWING"
	TaskStatusDONE          TaskStatus = "DONE"
	TaskStatusCANCELLED     TaskStatus = "CANCELLED"
)

type TaskGitStatus string

const (
	TaskGitStatusNone      TaskGitStatus = "none"
	TaskGitStatusCreating  TaskGitStatus = "creating"
	TaskGitStatusActive    TaskGitStatus = "active"
	TaskGitStatusCompleted TaskGitStatus = "completed"
	TaskGitStatusCleaning  TaskGitStatus = "cleaning"
	TaskGitStatusError     TaskGitStatus = "error"
)

// TaskPriority represents the priority level of a task
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "LOW"
	TaskPriorityMedium TaskPriority = "MEDIUM"
	TaskPriorityHigh   TaskPriority = "HIGH"
	TaskPriorityUrgent TaskPriority = "URGENT"
)

// IsValid checks if the task priority is valid
func (tp TaskPriority) IsValid() bool {
	switch tp {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh, TaskPriorityUrgent:
		return true
	default:
		return false
	}
}

// String returns the string representation of TaskPriority
func (tp TaskPriority) String() string {
	return string(tp)
}

// GetDisplayName returns a user-friendly display name for the priority
func (tp TaskPriority) GetDisplayName() string {
	switch tp {
	case TaskPriorityLow:
		return "Low"
	case TaskPriorityMedium:
		return "Medium"
	case TaskPriorityHigh:
		return "High"
	case TaskPriorityUrgent:
		return "Urgent"
	default:
		return string(tp)
	}
}

// GetAllTaskPriorities returns all valid task priorities
func GetAllTaskPriorities() []TaskPriority {
	return []TaskPriority{
		TaskPriorityLow,
		TaskPriorityMedium,
		TaskPriorityHigh,
		TaskPriorityUrgent,
	}
}

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

// TaskGitStatusTransitions defines valid transitions between task Git statuses
var TaskGitStatusTransitions = map[TaskGitStatus][]TaskGitStatus{
	TaskGitStatusNone: {
		TaskGitStatusCreating,
	},
	TaskGitStatusCreating: {
		TaskGitStatusActive,
		TaskGitStatusError,
	},
	TaskGitStatusActive: {
		TaskGitStatusCompleted,
		TaskGitStatusCleaning,
		TaskGitStatusError,
	},
	TaskGitStatusCompleted: {
		TaskGitStatusCleaning,
	},
	TaskGitStatusCleaning: {
		TaskGitStatusNone,
		TaskGitStatusError, // If cleanup fails
	},
	TaskGitStatusError: {
		TaskGitStatusCreating, // Allow retry
		TaskGitStatusCleaning, // Allow cleanup after error
	},
}

// IsValid checks if the task Git status is valid
func (tgs TaskGitStatus) IsValid() bool {
	switch tgs {
	case TaskGitStatusNone, TaskGitStatusCreating, TaskGitStatusActive,
		TaskGitStatusCompleted, TaskGitStatusCleaning, TaskGitStatusError:
		return true
	default:
		return false
	}
}

// String returns the string representation of TaskGitStatus
func (tgs TaskGitStatus) String() string {
	return string(tgs)
}

// CanTransitionTo checks if the current Git status can transition to the target status
func (tgs TaskGitStatus) CanTransitionTo(target TaskGitStatus) bool {
	allowedTransitions, exists := TaskGitStatusTransitions[tgs]
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

// GetDisplayName returns a user-friendly display name for the Git status
func (tgs TaskGitStatus) GetDisplayName() string {
	switch tgs {
	case TaskGitStatusNone:
		return "None"
	case TaskGitStatusCreating:
		return "Creating"
	case TaskGitStatusActive:
		return "Active"
	case TaskGitStatusCompleted:
		return "Completed"
	case TaskGitStatusCleaning:
		return "Cleaning"
	case TaskGitStatusError:
		return "Error"
	default:
		return string(tgs)
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
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID      uuid.UUID      `json:"project_id" gorm:"type:uuid;not null" validate:"required"`
	Title          string         `json:"title" gorm:"size:255;not null" validate:"required,min=1,max=255"`
	Description    string         `json:"description" gorm:"size:1000" validate:"max=1000"`
	Status         TaskStatus     `json:"status" gorm:"size:50;not null;default:'TODO'" validate:"required,oneof=TODO PLANNING PLAN_REVIEWING IMPLEMENTING CODE_REVIEWING DONE CANCELLED"`
	Priority       TaskPriority   `json:"priority" gorm:"size:20;default:'MEDIUM'" validate:"oneof=LOW MEDIUM HIGH URGENT"`
	BranchName     *string        `json:"branch_name,omitempty" gorm:"size:255"`
	PullRequest    *string        `json:"pull_request,omitempty" gorm:"size:255"`
	WorktreePath   *string        `json:"worktree_path,omitempty" gorm:"type:text"`
	GitStatus      TaskGitStatus  `json:"git_status" gorm:"size:50;default:'none'"`
	EstimatedHours *float64       `json:"estimated_hours,omitempty" gorm:"type:decimal(5,2)" validate:"min=0,max=999.99"`
	ActualHours    *float64       `json:"actual_hours,omitempty" gorm:"type:decimal(5,2)" validate:"min=0,max=999.99"`
	Tags           []string       `json:"tags,omitempty" gorm:"-"` // Will be stored as JSON in database
	TagsJSON       string         `json:"-" gorm:"column:tags;type:jsonb"`
	ParentTaskID   *uuid.UUID     `json:"parent_task_id,omitempty" gorm:"type:uuid"`
	IsArchived     bool           `json:"is_archived" gorm:"default:false"`
	IsTemplate     bool           `json:"is_template" gorm:"default:false"`
	TemplateID     *uuid.UUID     `json:"template_id,omitempty" gorm:"type:uuid"`
	AssignedTo     *string        `json:"assigned_to,omitempty" gorm:"size:255"` // User ID for future assignment
	DueDate        *time.Time     `json:"due_date,omitempty"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	BaseBranchName *string        `json:"base_branch_name,omitempty" gorm:"size:255"`

	// Relationships
	Project    *Project       `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	ParentTask *Task          `json:"parent_task,omitempty" gorm:"foreignKey:ParentTaskID"`
	Subtasks   []Task         `json:"subtasks,omitempty" gorm:"foreignKey:ParentTaskID"`
	AuditLogs  []TaskAuditLog `json:"audit_logs,omitempty" gorm:"foreignKey:TaskID"`
	Plans      []Plan         `json:"plan,omitempty" gorm:"foreignKey:TaskID"`
}

// TaskAuditLog tracks all modifications to tasks
type TaskAuditLog struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID    uuid.UUID      `json:"task_id" gorm:"type:uuid;not null" validate:"required"`
	Action    string         `json:"action" gorm:"size:100;not null" validate:"required"` // "created", "updated", "deleted", "status_changed", etc.
	FieldName *string        `json:"field_name,omitempty" gorm:"size:100"`                // Which field was changed
	OldValue  *string        `json:"old_value,omitempty" gorm:"size:1000"`                // Previous value
	NewValue  *string        `json:"new_value,omitempty" gorm:"size:1000"`                // New value
	ChangedBy *string        `json:"changed_by,omitempty" gorm:"size:255"`                // User ID who made the change
	IPAddress *string        `json:"ip_address,omitempty" gorm:"size:45"`                 // IP address of the change
	UserAgent *string        `json:"user_agent,omitempty" gorm:"size:500"`                // User agent string
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Task Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

// TaskTemplate represents reusable task templates
type TaskTemplate struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID      uuid.UUID      `json:"project_id" gorm:"type:uuid;not null" validate:"required"`
	Name           string         `json:"name" gorm:"size:255;not null" validate:"required,min=1,max=255"`
	Description    string         `json:"description" gorm:"size:1000" validate:"max=1000"`
	Title          string         `json:"title" gorm:"size:255;not null" validate:"required,min=1,max=255"`
	Priority       TaskPriority   `json:"priority" gorm:"size:20;default:'MEDIUM'" validate:"oneof=LOW MEDIUM HIGH URGENT"`
	EstimatedHours *float64       `json:"estimated_hours,omitempty" gorm:"type:decimal(5,2)" validate:"min=0,max=999.99"`
	Tags           []string       `json:"tags,omitempty" gorm:"-"` // Will be stored as JSON in database
	TagsJSON       string         `json:"-" gorm:"column:tags;type:jsonb"`
	IsGlobal       bool           `json:"is_global" gorm:"default:false"` // Available across all projects
	CreatedBy      *string        `json:"created_by,omitempty" gorm:"size:255"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

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

// TaskGitStatusValidationError represents an error when attempting invalid Git status transitions
type TaskGitStatusValidationError struct {
	CurrentStatus TaskGitStatus
	TargetStatus  TaskGitStatus
	Message       string
}

func (e *TaskGitStatusValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("invalid Git status transition from %s to %s", e.CurrentStatus, e.TargetStatus)
}

// ValidateGitStatusTransition validates if a Git status transition is allowed
func ValidateGitStatusTransition(from, to TaskGitStatus) error {
	if !from.IsValid() {
		return &TaskGitStatusValidationError{
			CurrentStatus: from,
			TargetStatus:  to,
			Message:       fmt.Sprintf("invalid current Git status: %s", from),
		}
	}

	if !to.IsValid() {
		return &TaskGitStatusValidationError{
			CurrentStatus: from,
			TargetStatus:  to,
			Message:       fmt.Sprintf("invalid target Git status: %s", to),
		}
	}

	if !from.CanTransitionTo(to) {
		return &TaskGitStatusValidationError{
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
	ProjectID           uuid.UUID              `json:"project_id"`
	StatusDistribution  []TaskStatusStats      `json:"status_distribution"`
	AverageTimeInStatus map[TaskStatus]float64 `json:"average_time_in_status"` // in hours
	TransitionCount     map[string]int         `json:"transition_count"`       // from_status->to_status counts
	TotalTasks          int                    `json:"total_tasks"`
	CompletedTasks      int                    `json:"completed_tasks"`
	CompletionRate      float64                `json:"completion_rate"`
	GeneratedAt         time.Time              `json:"generated_at"`
}

// TaskSearchResult represents a search result with relevance score
type TaskSearchResult struct {
	Task    *Task   `json:"task"`
	Score   float64 `json:"score"`   // Relevance score for search results
	Matched string  `json:"matched"` // Which field matched the search
}

// TaskBulkOperation represents a bulk operation on multiple tasks
type TaskBulkOperation struct {
	TaskIDs []uuid.UUID `json:"task_ids" validate:"required,min=1"`
	Action  string      `json:"action" validate:"required"` // "delete", "update_status", "archive", "unarchive"
	Data    interface{} `json:"data,omitempty"`             // Additional data for the operation
}

// TaskExportFormat represents the format for task export
type TaskExportFormat string

const (
	TaskExportFormatCSV  TaskExportFormat = "csv"
	TaskExportFormatJSON TaskExportFormat = "json"
	TaskExportFormatXML  TaskExportFormat = "xml"
)

// TaskFilters represents comprehensive filtering options for tasks
type TaskFilters struct {
	ProjectID      *uuid.UUID
	Statuses       []TaskStatus
	Priorities     []TaskPriority
	Tags           []string
	ParentTaskID   *uuid.UUID
	AssignedTo     *string
	CreatedAfter   *time.Time
	CreatedBefore  *time.Time
	UpdatedAfter   *time.Time
	UpdatedBefore  *time.Time
	DueDateAfter   *time.Time
	DueDateBefore  *time.Time
	SearchTerm     *string
	IsArchived     *bool
	IsTemplate     *bool
	HasSubtasks    *bool
	EstimatedHours *float64
	ActualHours    *float64
	Limit          *int
	Offset         *int
	OrderBy        *string // "created_at", "updated_at", "title", "status", "priority", "due_date"
	OrderDir       *string // "asc", "desc"
}

// TaskDependency represents dependencies between tasks
type TaskDependency struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID          uuid.UUID `json:"task_id" gorm:"type:uuid;not null"`
	DependsOnTaskID uuid.UUID `json:"depends_on_task_id" gorm:"type:uuid;not null"`
	DependencyType  string    `json:"dependency_type" gorm:"size:50;default:'blocks'"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Task          *Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
	DependsOnTask *Task `json:"depends_on_task,omitempty" gorm:"foreignKey:DependsOnTaskID"`
}

// TaskComment represents comments on tasks
type TaskComment struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID    uuid.UUID      `json:"task_id" gorm:"type:uuid;not null"`
	Comment   string         `json:"comment" gorm:"not null"`
	CreatedBy string         `json:"created_by" gorm:"size:255;not null"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Task *Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

// TaskAttachment represents file attachments for tasks
type TaskAttachment struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID     uuid.UUID      `json:"task_id" gorm:"type:uuid;not null"`
	Filename   string         `json:"filename" gorm:"size:255;not null"`
	FilePath   string         `json:"file_path" gorm:"size:500;not null"`
	FileSize   int64          `json:"file_size" gorm:"not null"`
	MimeType   string         `json:"mime_type" gorm:"size:100"`
	UploadedBy string         `json:"uploaded_by" gorm:"size:255;not null"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Task *Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

// TaskStatistics represents comprehensive task statistics for a project
type TaskStatistics struct {
	ProjectID             uuid.UUID            `json:"project_id"`
	TotalTasks            int                  `json:"total_tasks"`
	CompletedTasks        int                  `json:"completed_tasks"`
	InProgressTasks       int                  `json:"in_progress_tasks"`
	ArchivedTasks         int                  `json:"archived_tasks"`
	TasksByPriority       map[TaskPriority]int `json:"tasks_by_priority"`
	TasksByStatus         map[TaskStatus]int   `json:"tasks_by_status"`
	AverageCompletionTime float64              `json:"average_completion_time"` // in hours
	TotalEstimatedHours   float64              `json:"total_estimated_hours"`
	TotalActualHours      float64              `json:"total_actual_hours"`
	OverdueTasks          int                  `json:"overdue_tasks"`
	GeneratedAt           time.Time            `json:"generated_at"`
}

// BeforeCreate GORM hook to convert Tags to TagsJSON before saving
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if len(t.Tags) > 0 {
		tagsJSON, err := json.Marshal(t.Tags)
		if err != nil {
			return err
		}
		t.TagsJSON = string(tagsJSON)
	} else {
		t.TagsJSON = "[]"
	}
	return nil
}

// BeforeUpdate GORM hook to convert Tags to TagsJSON before updating
func (t *Task) BeforeUpdate(tx *gorm.DB) error {
	if len(t.Tags) > 0 {
		tagsJSON, err := json.Marshal(t.Tags)
		if err != nil {
			return err
		}
		t.TagsJSON = string(tagsJSON)
	} else {
		t.TagsJSON = "[]"
	}
	return nil
}

// AfterFind GORM hook to convert TagsJSON to Tags after loading
func (t *Task) AfterFind(tx *gorm.DB) error {
	if t.TagsJSON != "" {
		return json.Unmarshal([]byte(t.TagsJSON), &t.Tags)
	}
	return nil
}

// BeforeCreate GORM hook for TaskTemplate
func (tt *TaskTemplate) BeforeCreate(tx *gorm.DB) error {
	if len(tt.Tags) > 0 {
		tagsJSON, err := json.Marshal(tt.Tags)
		if err != nil {
			return err
		}
		tt.TagsJSON = string(tagsJSON)
	} else {
		tt.TagsJSON = "[]"
	}
	return nil
}

// BeforeUpdate GORM hook for TaskTemplate
func (tt *TaskTemplate) BeforeUpdate(tx *gorm.DB) error {
	if len(tt.Tags) > 0 {
		tagsJSON, err := json.Marshal(tt.Tags)
		if err != nil {
			return err
		}
		tt.TagsJSON = string(tagsJSON)
	} else {
		tt.TagsJSON = "[]"
	}
	return nil
}

// AfterFind GORM hook for TaskTemplate
func (tt *TaskTemplate) AfterFind(tx *gorm.DB) error {
	if tt.TagsJSON != "" {
		return json.Unmarshal([]byte(tt.TagsJSON), &tt.Tags)
	}
	return nil
}
