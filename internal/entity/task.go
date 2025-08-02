package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskStatus string

const (
	TaskStatusTodo         TaskStatus = "TODO"
	TaskStatusPlanning     TaskStatus = "PLANNING"
	TaskStatusPlanReview   TaskStatus = "PLAN_REVIEWING"
	TaskStatusImplementing TaskStatus = "IMPLEMENTING"
	TaskStatusCodeReview   TaskStatus = "CODE_REVIEWING"
	TaskStatusDone         TaskStatus = "DONE"
	TaskStatusCancelled    TaskStatus = "CANCELLED"
)

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
