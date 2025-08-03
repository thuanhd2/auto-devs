package dto

import (
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// Task request DTOs
type TaskCreateRequest struct {
	ProjectID   uuid.UUID `json:"project_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title       string    `json:"title" binding:"required,min=1,max=255" example:"Implement user authentication"`
	Description string    `json:"description" binding:"max=1000" example:"Add JWT-based authentication system"`
}

type TaskUpdateRequest struct {
	Title       *string `json:"title,omitempty" binding:"omitempty,min=1,max=255" example:"Updated task title"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000" example:"Updated description"`
	BranchName  *string `json:"branch_name,omitempty" binding:"omitempty,max=255" example:"feature/user-auth"`
	PullRequest *string `json:"pull_request,omitempty" binding:"omitempty,max=255" example:"https://github.com/user/repo/pull/123"`
}

type TaskStatusUpdateRequest struct {
	Status entity.TaskStatus `json:"status" binding:"required,oneof=TODO PLANNING PLAN_REVIEWING IMPLEMENTING CODE_REVIEWING DONE CANCELLED" example:"TODO"`
}

// Task response DTOs
type TaskResponse struct {
	ID          uuid.UUID         `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ProjectID   uuid.UUID         `json:"project_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title       string            `json:"title" example:"Implement user authentication"`
	Description string            `json:"description" example:"Add JWT-based authentication system"`
	Status      entity.TaskStatus `json:"status" example:"TODO"`
	BranchName  *string           `json:"branch_name,omitempty" example:"feature/user-auth"`
	PullRequest *string           `json:"pull_request,omitempty" example:"https://github.com/user/repo/pull/123"`
	CreatedAt   time.Time         `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt   time.Time         `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

type TaskWithProjectResponse struct {
	TaskResponse
	Project ProjectResponse `json:"project"`
}

type TaskListResponse struct {
	Tasks []TaskResponse `json:"tasks"`
	Total int            `json:"total"`
}

type TasksByStatusResponse struct {
	Status entity.TaskStatus `json:"status"`
	Tasks  []TaskResponse    `json:"tasks"`
	Count  int               `json:"count"`
}

// Helper functions to convert between entity and DTO
func (t *TaskResponse) FromEntity(task *entity.Task) {
	t.ID = task.ID
	t.ProjectID = task.ProjectID
	t.Title = task.Title
	t.Description = task.Description
	t.Status = task.Status
	t.BranchName = task.BranchName
	t.PullRequest = task.PullRequest
	t.CreatedAt = task.CreatedAt
	t.UpdatedAt = task.UpdatedAt
}

func (t *TaskWithProjectResponse) FromEntity(task *entity.Task) {
	t.TaskResponse.FromEntity(task)
	t.Project.FromEntity(&task.Project)
}

func TaskResponseFromEntity(task *entity.Task) TaskResponse {
	var resp TaskResponse
	resp.FromEntity(task)
	return resp
}

func TaskListResponseFromEntities(tasks []*entity.Task) TaskListResponse {
	responses := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = TaskResponseFromEntity(task)
	}
	return TaskListResponse{
		Tasks: responses,
		Total: len(tasks),
	}
}