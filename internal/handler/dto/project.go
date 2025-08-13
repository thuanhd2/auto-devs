package dto

import (
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/google/uuid"
)

// Project request DTOs
type ProjectCreateRequest struct {
	Name                string `json:"name" binding:"required,min=1,max=255" example:"My Project"`
	Description         string `json:"description" binding:"max=1000" example:"Project description"`
	WorktreeBasePath    string `json:"worktree_base_path" binding:"required,max=500" example:"/tmp/projects/repo"`
	InitWorkspaceScript string `json:"init_workspace_script" example:"npm install && npm run build"`
}

type ProjectUpdateRequest struct {
	Name                *string `json:"name,omitempty" binding:"omitempty,min=1,max=255" example:"Updated Project Name"`
	Description         *string `json:"description,omitempty" binding:"omitempty,max=1000" example:"Updated description"`
	RepositoryURL       *string `json:"repository_url,omitempty" binding:"omitempty,url,max=500" example:"https://github.com/user/repo.git"`
	WorktreeBasePath    *string `json:"worktree_base_path,omitempty" binding:"omitempty,max=500" example:"/tmp/projects/repo"`
	InitWorkspaceScript *string `json:"init_workspace_script,omitempty" example:"npm install && npm run build"`
}

// Project response DTOs
type ProjectResponse struct {
	ID                  uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name                string    `json:"name" example:"My Project"`
	Description         string    `json:"description" example:"Project description"`
	RepositoryURL       string    `json:"repository_url,omitempty" example:"https://github.com/user/repo.git"`
	WorktreeBasePath    string    `json:"worktree_base_path,omitempty" example:"/tmp/projects/repo"`
	InitWorkspaceScript string    `json:"init_workspace_script,omitempty" example:"npm install && npm run build"`
	CreatedAt           time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt           time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
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
	TotalTasks        int            `json:"total_tasks"`
	TasksByStatus     map[string]int `json:"tasks_by_status"`
	CompletionPercent float64        `json:"completion_percent"`
	LastActivityAt    *time.Time     `json:"last_activity_at"`
	RecentActivity    int            `json:"recent_activity"`
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

type UpdateRepositoryURLRequest struct {
	RepositoryURL string `json:"repository_url" binding:"required,url,max=500" example:"https://github.com/user/repo.git"`
}

type GitStatusResponse struct {
	GitEnabled       bool                      `json:"git_enabled"`
	WorktreeExists   bool                      `json:"worktree_exists"`
	RepositoryValid  bool                      `json:"repository_valid"`
	CurrentBranch    string                    `json:"current_branch,omitempty"`
	RemoteURL        string                    `json:"remote_url,omitempty"`
	OnMainBranch     bool                      `json:"on_main_branch"`
	WorkingDirStatus *WorkingDirStatusResponse `json:"working_dir_status,omitempty"`
	Status           string                    `json:"status"`
}

type WorkingDirStatusResponse struct {
	IsClean            bool `json:"is_clean"`
	HasStagedChanges   bool `json:"has_staged_changes"`
	HasUnstagedChanges bool `json:"has_unstaged_changes"`
	HasUntrackedFiles  bool `json:"has_untracked_files"`
}

// Helper functions to convert between entity and DTO
func (p *ProjectResponse) FromEntity(project *entity.Project) {
	p.ID = project.ID
	p.Name = project.Name
	p.Description = project.Description
	p.RepositoryURL = project.RepositoryURL
	p.WorktreeBasePath = project.WorktreeBasePath
	p.InitWorkspaceScript = project.InitWorkspaceScript
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
	// Convert map[entity.TaskStatus]int to map[string]int
	tasksByStatus := make(map[string]int)
	for status, count := range stats.TaskCounts {
		tasksByStatus[string(status)] = count
	}

	// Calculate recent activity (tasks updated in last 7 days)
	recentActivity := 0
	if stats.LastActivityAt != nil {
		// For now, we'll set a default value. In a real implementation,
		// you might want to calculate this from the database
		recentActivity = stats.TotalTasks
	}

	return ProjectStatisticsResponse{
		TotalTasks:        stats.TotalTasks,
		TasksByStatus:     tasksByStatus,
		CompletionPercent: stats.CompletionPercent,
		LastActivityAt:    stats.LastActivityAt,
		RecentActivity:    recentActivity,
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

func GitStatusResponseFromUsecase(status *usecase.GitStatus) GitStatusResponse {
	response := GitStatusResponse{
		GitEnabled:      status.GitEnabled,
		WorktreeExists:  status.WorktreeExists,
		RepositoryValid: status.RepositoryValid,
		CurrentBranch:   status.CurrentBranch,
		RemoteURL:       status.RemoteURL,
		OnMainBranch:    status.OnMainBranch,
		Status:          status.Status,
	}

	if status.WorkingDirStatus != nil {
		response.WorkingDirStatus = &WorkingDirStatusResponse{
			IsClean:            status.WorkingDirStatus.IsClean,
			HasStagedChanges:   status.WorkingDirStatus.HasStagedChanges,
			HasUnstagedChanges: status.WorkingDirStatus.HasUnstagedChanges,
			HasUntrackedFiles:  status.WorkingDirStatus.HasUntrackedFiles,
		}
	}

	return response
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
