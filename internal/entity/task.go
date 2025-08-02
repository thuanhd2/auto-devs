package entity

import (
	"time"

	"github.com/google/uuid"
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
	ID          uuid.UUID  `json:"id" db:"id"`
	ProjectID   uuid.UUID  `json:"project_id" db:"project_id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Status      TaskStatus `json:"status" db:"status"`
	BranchName  *string    `json:"branch_name,omitempty" db:"branch_name"`
	PullRequest *string    `json:"pull_request,omitempty" db:"pull_request"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}