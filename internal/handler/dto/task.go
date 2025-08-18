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
	Description string    `json:"description" binding:"max=5000" example:"Add JWT-based authentication system"`
}

type TaskUpdateRequest struct {
	Title       *string            `json:"title,omitempty" binding:"omitempty,min=1,max=255" example:"Updated task title"`
	Description *string            `json:"description,omitempty" binding:"omitempty,max=5000" example:"Updated description"`
	Status      *entity.TaskStatus `json:"status,omitempty" binding:"omitempty,oneof=TODO PLANNING PLAN_REVIEWING IMPLEMENTING CODE_REVIEWING DONE CANCELLED" example:"TODO"`
	BranchName  *string            `json:"branch_name,omitempty" binding:"omitempty,max=255" example:"feature/user-auth"`
	PullRequest *string            `json:"pull_request,omitempty" binding:"omitempty,max=255" example:"https://github.com/user/repo/pull/123"`
}

type TaskStatusUpdateRequest struct {
	Status entity.TaskStatus `json:"status" binding:"required,oneof=TODO PLANNING PLAN_REVIEWING IMPLEMENTING CODE_REVIEWING DONE CANCELLED" example:"TODO"`
}

type TaskStatusUpdateWithHistoryRequest struct {
	Status    entity.TaskStatus `json:"status" binding:"required,oneof=TODO PLANNING PLAN_REVIEWING IMPLEMENTING CODE_REVIEWING DONE CANCELLED" example:"TODO"`
	ChangedBy *string           `json:"changed_by,omitempty" example:"user123"`
	Reason    *string           `json:"reason,omitempty" example:"Requirements changed"`
}

type BulkStatusUpdateRequest struct {
	TaskIDs   []uuid.UUID       `json:"task_ids" binding:"required" example:"[\"123e4567-e89b-12d3-a456-426614174000\"]"`
	Status    entity.TaskStatus `json:"status" binding:"required,oneof=TODO PLANNING PLAN_REVIEWING IMPLEMENTING CODE_REVIEWING DONE CANCELLED" example:"TODO"`
	ChangedBy *string           `json:"changed_by,omitempty" example:"user123"`
}

type TaskAdvancedFilterQuery struct {
	ProjectID     *string    `form:"project_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status        *string    `form:"status" example:"TODO"`
	Statuses      []string   `form:"statuses" example:"TODO,PLANNING"`
	CreatedAfter  *time.Time `form:"created_after" example:"2024-01-01T00:00:00Z"`
	CreatedBefore *time.Time `form:"created_before" example:"2024-12-31T23:59:59Z"`
	SearchTerm    *string    `form:"search" example:"authentication"`
	Limit         *int       `form:"limit" example:"10"`
	Offset        *int       `form:"offset" example:"0"`
	OrderBy       *string    `form:"order_by" example:"created_at"`
	OrderDir      *string    `form:"order_dir" example:"desc"`
}

// Task response DTOs
type TaskResponse struct {
	ID           uuid.UUID            `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ProjectID    uuid.UUID            `json:"project_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title        string               `json:"title" example:"Implement user authentication"`
	Description  string               `json:"description" example:"Add JWT-based authentication system"`
	Status       entity.TaskStatus    `json:"status" example:"TODO"`
	GitStatus    entity.TaskGitStatus `json:"git_status" example:"none"`
	BranchName   *string              `json:"branch_name,omitempty" example:"feature/user-auth"`
	PullRequest  *string              `json:"pull_request,omitempty" example:"https://github.com/user/repo/pull/123"`
	WorktreePath *string              `json:"worktree_path,omitempty" example:"/tmp/worktrees/task-123"`
	CreatedAt    time.Time            `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt    time.Time            `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

type TaskWithProjectResponse struct {
	TaskResponse
	Project ProjectResponse `json:"project"`
}

type TaskListResponse struct {
	Tasks []TaskResponse `json:"tasks"`
	Total int            `json:"total"`
}

type TaskPlansResponse struct {
	Plans []PlanResponse `json:"plans"`
}

type TasksByStatusResponse struct {
	Status entity.TaskStatus `json:"status"`
	Tasks  []TaskResponse    `json:"tasks"`
	Count  int               `json:"count"`
}

type TaskStatusHistoryResponse struct {
	ID         uuid.UUID          `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	TaskID     uuid.UUID          `json:"task_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	FromStatus *entity.TaskStatus `json:"from_status,omitempty" example:"TODO"`
	ToStatus   entity.TaskStatus  `json:"to_status" example:"PLANNING"`
	ChangedBy  *string            `json:"changed_by,omitempty" example:"user123"`
	Reason     *string            `json:"reason,omitempty" example:"Requirements changed"`
	CreatedAt  time.Time          `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

type TaskStatusAnalyticsResponse struct {
	ProjectID           uuid.UUID                     `json:"project_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	StatusDistribution  []TaskStatusStatsResponse     `json:"status_distribution"`
	AverageTimeInStatus map[entity.TaskStatus]float64 `json:"average_time_in_status"`
	TransitionCount     map[string]int                `json:"transition_count"`
	TotalTasks          int                           `json:"total_tasks" example:"50"`
	CompletedTasks      int                           `json:"completed_tasks" example:"20"`
	CompletionRate      float64                       `json:"completion_rate" example:"40.0"`
	GeneratedAt         time.Time                     `json:"generated_at" example:"2024-01-15T10:30:00Z"`
}

type TaskStatusStatsResponse struct {
	Status entity.TaskStatus `json:"status" example:"TODO"`
	Count  int               `json:"count" example:"10"`
}

type TaskStatusValidationResponse struct {
	Valid         bool              `json:"valid" example:"true"`
	CurrentStatus entity.TaskStatus `json:"current_status" example:"TODO"`
	TargetStatus  entity.TaskStatus `json:"target_status" example:"PLANNING"`
	Message       string            `json:"message,omitempty" example:"Transition is valid"`
}

type TaskGitStatusUpdateRequest struct {
	GitStatus entity.TaskGitStatus `json:"git_status" binding:"required,oneof=none creating active completed cleaning error" example:"active"`
}

type TaskGitStatusValidationResponse struct {
	Valid            bool                 `json:"valid" example:"true"`
	CurrentGitStatus entity.TaskGitStatus `json:"current_git_status" example:"none"`
	TargetGitStatus  entity.TaskGitStatus `json:"target_git_status" example:"creating"`
	Message          string               `json:"message,omitempty" example:"Git status transition is valid"`
}

// Helper functions to convert between entity and DTO
func (t *TaskResponse) FromEntity(task *entity.Task) {
	t.ID = task.ID
	t.ProjectID = task.ProjectID
	t.Title = task.Title
	t.Description = task.Description
	t.Status = task.Status
	t.GitStatus = task.GitStatus
	t.BranchName = task.BranchName
	t.PullRequest = task.PullRequest
	t.WorktreePath = task.WorktreePath
	t.CreatedAt = task.CreatedAt
	t.UpdatedAt = task.UpdatedAt
}

func (t *TaskWithProjectResponse) FromEntity(task *entity.Task) {
	t.TaskResponse.FromEntity(task)
	if task.Project != nil {
		t.Project.FromEntity(task.Project)
	}
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

func TaskStatusHistoryResponseFromEntity(history *entity.TaskStatusHistory) TaskStatusHistoryResponse {
	return TaskStatusHistoryResponse{
		ID:         history.ID,
		TaskID:     history.TaskID,
		FromStatus: history.FromStatus,
		ToStatus:   history.ToStatus,
		ChangedBy:  history.ChangedBy,
		Reason:     history.Reason,
		CreatedAt:  history.CreatedAt,
	}
}

func TaskStatusHistoryListFromEntities(histories []*entity.TaskStatusHistory) []TaskStatusHistoryResponse {
	responses := make([]TaskStatusHistoryResponse, len(histories))
	for i, history := range histories {
		responses[i] = TaskStatusHistoryResponseFromEntity(history)
	}
	return responses
}

func TaskStatusAnalyticsResponseFromEntity(analytics *entity.TaskStatusAnalytics) TaskStatusAnalyticsResponse {
	// Convert status distribution
	statusDist := make([]TaskStatusStatsResponse, len(analytics.StatusDistribution))
	for i, stat := range analytics.StatusDistribution {
		statusDist[i] = TaskStatusStatsResponse{
			Status: stat.Status,
			Count:  stat.Count,
		}
	}

	return TaskStatusAnalyticsResponse{
		ProjectID:           analytics.ProjectID,
		StatusDistribution:  statusDist,
		AverageTimeInStatus: analytics.AverageTimeInStatus,
		TransitionCount:     analytics.TransitionCount,
		TotalTasks:          analytics.TotalTasks,
		CompletedTasks:      analytics.CompletedTasks,
		CompletionRate:      analytics.CompletionRate,
		GeneratedAt:         analytics.GeneratedAt,
	}
}

// Start Planning DTOs
type StartPlanningRequest struct {
	BranchName string `json:"branch_name" binding:"required" example:"main"`
	AIType     string `json:"ai_type" binding:"required" example:"claude-code"`
}

type StartPlanningResponse struct {
	Message string `json:"message" example:"Planning started successfully"`
	JobID   string `json:"job_id" example:"task-123-planning-456"`
}

// Approve Plan DTOs
type ApprovePlanRequest struct {
	AIType string `json:"ai_type" binding:"required" example:"claude-code"`
}

// Git Branches DTOs
type GitBranchResponse struct {
	Name        string `json:"name" example:"main"`
	IsCurrent   bool   `json:"is_current" example:"true"`
	LastCommit  string `json:"last_commit,omitempty" example:"abc123def"`
	LastUpdated string `json:"last_updated,omitempty" example:"2024-01-15T10:30:00Z"`
}

type ListBranchesResponse struct {
	Branches []GitBranchResponse `json:"branches"`
	Total    int                 `json:"total"`
}

type PlanUpdateRequest struct {
	Content string `json:"content" binding:"required" example:"Implement user authentication"`
}
