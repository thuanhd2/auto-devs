package dto

import (
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/google/uuid"
)

// Project request DTOs
type ProjectCreateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255" example:"My Project"`
	Description string `json:"description" binding:"max=1000" example:"Project description"`
	RepoURL     string `json:"repo_url" binding:"required,url,max=500" example:"https://github.com/user/repo"`
}

type ProjectUpdateRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=255" example:"Updated Project Name"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000" example:"Updated description"`
	RepoURL     *string `json:"repo_url,omitempty" binding:"omitempty,url,max=500" example:"https://github.com/user/updated-repo"`
}

// Project response DTOs
type ProjectResponse struct {
	ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string    `json:"name" example:"My Project"`
	Description string    `json:"description" example:"Project description"`
	RepoURL     string    `json:"repo_url" example:"https://github.com/user/repo"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

type ProjectWithTasksResponse struct {
	ProjectResponse
	Tasks []TaskResponse `json:"tasks"`
}

type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

type ProjectStatisticsResponse struct {
	TaskCounts        map[entity.TaskStatus]int `json:"task_counts"`
	TotalTasks        int                       `json:"total_tasks"`
	CompletionPercent float64                   `json:"completion_percent"`
	LastActivityAt    *time.Time                `json:"last_activity_at"`
}

type ProjectSettingsResponse struct {
	ID                   uuid.UUID `json:"id"`
	ProjectID            uuid.UUID `json:"project_id"`
	AutoArchiveDays      *int      `json:"auto_archive_days,omitempty"`
	NotificationsEnabled bool      `json:"notifications_enabled"`
	EmailNotifications   bool      `json:"email_notifications"`
	SlackWebhookURL      string    `json:"slack_webhook_url,omitempty"`
	GitBranch            string    `json:"git_branch"`
	GitAutoSync          bool      `json:"git_auto_sync"`
	TaskPrefix           string    `json:"task_prefix"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ProjectSettingsUpdateRequest struct {
	AutoArchiveDays      *int    `json:"auto_archive_days,omitempty"`
	NotificationsEnabled *bool   `json:"notifications_enabled,omitempty"`
	EmailNotifications   *bool   `json:"email_notifications,omitempty"`
	SlackWebhookURL      *string `json:"slack_webhook_url,omitempty"`
	GitBranch            *string `json:"git_branch,omitempty"`
	GitAutoSync          *bool   `json:"git_auto_sync,omitempty"`
	TaskPrefix           *string `json:"task_prefix,omitempty"`
}

// Helper functions to convert between entity and DTO
func (p *ProjectResponse) FromEntity(project *entity.Project) {
	p.ID = project.ID
	p.Name = project.Name
	p.Description = project.Description
	p.RepoURL = project.RepoURL
	p.CreatedAt = project.CreatedAt
	p.UpdatedAt = project.UpdatedAt
}

func (p *ProjectWithTasksResponse) FromEntity(project *entity.Project) {
	p.ProjectResponse.FromEntity(project)
	p.Tasks = make([]TaskResponse, len(project.Tasks))
	for i, task := range project.Tasks {
		p.Tasks[i].FromEntity(&task)
	}
}

func ProjectResponseFromEntity(project *entity.Project) ProjectResponse {
	var resp ProjectResponse
	resp.FromEntity(project)
	return resp
}

func ProjectListResponseFromEntities(projects []*entity.Project) ProjectListResponse {
	responses := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = ProjectResponseFromEntity(project)
	}
	return ProjectListResponse{
		Projects: responses,
		Total:    len(projects),
		Page:     1,
		PageSize: len(projects),
	}
}

func ProjectListResponseFromResult(result *usecase.GetProjectsResult) ProjectListResponse {
	responses := make([]ProjectResponse, len(result.Projects))
	for i, project := range result.Projects {
		responses[i] = ProjectResponseFromEntity(project)
	}
	return ProjectListResponse{
		Projects: responses,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}
}

func ProjectStatisticsResponseFromUsecase(stats *usecase.ProjectStatistics) ProjectStatisticsResponse {
	return ProjectStatisticsResponse{
		TaskCounts:        stats.TaskCounts,
		TotalTasks:        stats.TotalTasks,
		CompletionPercent: stats.CompletionPercent,
		LastActivityAt:    stats.LastActivityAt,
	}
}

func ProjectSettingsResponseFromEntity(settings *entity.ProjectSettings) ProjectSettingsResponse {
	return ProjectSettingsResponse{
		ID:                   settings.ID,
		ProjectID:            settings.ProjectID,
		AutoArchiveDays:      settings.AutoArchiveDays,
		NotificationsEnabled: settings.NotificationsEnabled,
		EmailNotifications:   settings.EmailNotifications,
		SlackWebhookURL:      settings.SlackWebhookURL,
		GitBranch:            settings.GitBranch,
		GitAutoSync:          settings.GitAutoSync,
		TaskPrefix:           settings.TaskPrefix,
		CreatedAt:            settings.CreatedAt,
		UpdatedAt:            settings.UpdatedAt,
	}
}

func (req *ProjectSettingsUpdateRequest) ToEntity() *entity.ProjectSettings {
	settings := &entity.ProjectSettings{}

	if req.AutoArchiveDays != nil {
		settings.AutoArchiveDays = req.AutoArchiveDays
	}
	if req.NotificationsEnabled != nil {
		settings.NotificationsEnabled = *req.NotificationsEnabled
	}
	if req.EmailNotifications != nil {
		settings.EmailNotifications = *req.EmailNotifications
	}
	if req.SlackWebhookURL != nil {
		settings.SlackWebhookURL = *req.SlackWebhookURL
	}
	if req.GitBranch != nil {
		settings.GitBranch = *req.GitBranch
	}
	if req.GitAutoSync != nil {
		settings.GitAutoSync = *req.GitAutoSync
	}
	if req.TaskPrefix != nil {
		settings.TaskPrefix = *req.TaskPrefix
	}

	return settings
}
