package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/internal/service/git"
	"github.com/google/uuid"
)

type ProjectUsecase interface {
	Create(ctx context.Context, req CreateProjectRequest) (*entity.Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error)
	GetAll(ctx context.Context, params GetProjectsParams) (*GetProjectsResult, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateProjectRequest) (*entity.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetWithTasks(ctx context.Context, id uuid.UUID) (*entity.Project, error)
	GetStatistics(ctx context.Context, id uuid.UUID) (*ProjectStatistics, error)
	Archive(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
	CheckNameExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error)
	GetSettings(ctx context.Context, projectID uuid.UUID) (*entity.ProjectSettings, error)
	UpdateSettings(ctx context.Context, projectID uuid.UUID, settings *entity.ProjectSettings) (*entity.ProjectSettings, error)
	UpdateRepositoryURL(ctx context.Context, projectID uuid.UUID, repositoryURL string) error
	ReinitGitRepository(ctx context.Context, projectID uuid.UUID) error
	GetGitStatus(ctx context.Context, projectID uuid.UUID) (*GitStatus, error)
	ListBranches(ctx context.Context, projectID uuid.UUID) ([]GitBranch, error)
}

type CreateProjectRequest struct {
	Name                string `json:"name" binding:"required"`
	Description         string `json:"description"`
	WorktreeBasePath    string `json:"worktree_base_path" binding:"required"`
	InitWorkspaceScript string `json:"init_workspace_script"`
}

type UpdateProjectRequest struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	RepositoryURL       string `json:"repository_url"`
	WorktreeBasePath    string `json:"worktree_base_path"`
	InitWorkspaceScript string `json:"init_workspace_script"`
}

type GetProjectsParams struct {
	Search    string
	SortBy    string // name, created_at, task_count
	SortOrder string // asc, desc
	Page      int
	PageSize  int
	Archived  *bool
}

type GetProjectsResult struct {
	Projects []*entity.Project `json:"projects"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

type ProjectStatistics struct {
	TaskCounts        map[entity.TaskStatus]int `json:"task_counts"`
	TotalTasks        int                       `json:"total_tasks"`
	CompletionPercent float64                   `json:"completion_percent"`
	LastActivityAt    *time.Time                `json:"last_activity_at"`
}

type GitStatus struct {
	GitEnabled       bool              `json:"git_enabled"`
	WorktreeExists   bool              `json:"worktree_exists"`
	RepositoryValid  bool              `json:"repository_valid"`
	CurrentBranch    string            `json:"current_branch,omitempty"`
	RemoteURL        string            `json:"remote_url,omitempty"`
	OnMainBranch     bool              `json:"on_main_branch"`
	WorkingDirStatus *WorkingDirStatus `json:"working_dir_status,omitempty"`
	Status           string            `json:"status"`
}

type WorkingDirStatus struct {
	IsClean            bool `json:"is_clean"`
	HasStagedChanges   bool `json:"has_staged_changes"`
	HasUnstagedChanges bool `json:"has_unstaged_changes"`
	HasUntrackedFiles  bool `json:"has_untracked_files"`
}

// Validation errors

// GitBranch represents a Git branch with metadata
type GitBranch struct {
	Name        string `json:"name"`
	IsCurrent   bool   `json:"is_current"`
	LastCommit  string `json:"last_commit,omitempty"`
	LastUpdated string `json:"last_updated,omitempty"`
}

// Validation errors
var (
	ErrProjectNameRequired = errors.New("project name is required")
	ErrProjectNameTooShort = errors.New("project name must be at least 3 characters")
	ErrProjectNameTooLong  = errors.New("project name must not exceed 255 characters")
	ErrProjectNameExists   = errors.New("project name already exists")
	ErrDescriptionTooLong  = errors.New("description must not exceed 1000 characters")
	ErrRepoURLRequired     = errors.New("repository URL is required")
	ErrRepoURLInvalid      = errors.New("repository URL is invalid")
	ErrRepoURLTooLong      = errors.New("repository URL must not exceed 500 characters")
)

// validateProjectName validates project name according to business rules
func validateProjectName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrProjectNameRequired
	}
	if len(name) < 3 {
		return ErrProjectNameTooShort
	}
	if len(name) > 255 {
		return ErrProjectNameTooLong
	}
	return nil
}

// validateDescription validates project description
func validateDescription(description string) error {
	if len(description) > 1000 {
		return ErrDescriptionTooLong
	}
	return nil
}

// validateRepoURL validates repository URL format
func validateRepoURL(repoURL string) error {
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return ErrRepoURLRequired
	}
	if len(repoURL) > 500 {
		return ErrRepoURLTooLong
	}

	// Parse URL to validate format
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return ErrRepoURLInvalid
	}

	// Check if it's a valid HTTP/HTTPS URL
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrRepoURLInvalid
	}

	// Additional validation for common Git hosting patterns
	validGitPattern := regexp.MustCompile(`^https?://(github\.com|gitlab\.com|bitbucket\.org)/[\w\-.]+/\w+/?$`)
	if !validGitPattern.MatchString(repoURL) {
		// Allow other valid URLs but warn about common patterns
		if !strings.Contains(parsedURL.Host, ".") {
			return ErrRepoURLInvalid
		}
	}

	return nil
}

type projectUsecase struct {
	projectRepo  repository.ProjectRepository
	auditUsecase AuditUsecase
	gitService   git.ProjectGitServiceInterface
}

func NewProjectUsecase(projectRepo repository.ProjectRepository, auditUsecase AuditUsecase, gitService git.ProjectGitServiceInterface) ProjectUsecase {
	return &projectUsecase{
		projectRepo:  projectRepo,
		auditUsecase: auditUsecase,
		gitService:   gitService,
	}
}

func (u *projectUsecase) Create(ctx context.Context, req CreateProjectRequest) (*entity.Project, error) {
	// Validate input
	if err := validateProjectName(req.Name); err != nil {
		return nil, err
	}
	if err := validateDescription(req.Description); err != nil {
		return nil, err
	}

	// Check for duplicate name
	exists, err := u.CheckNameExists(ctx, req.Name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
	}
	if exists {
		return nil, ErrProjectNameExists
	}

	project := &entity.Project{
		ID:                  uuid.New(),
		Name:                strings.TrimSpace(req.Name),
		Description:         strings.TrimSpace(req.Description),
		RepositoryURL:       "", // Will be populated by git service later
		WorktreeBasePath:    strings.TrimSpace(req.WorktreeBasePath),
		InitWorkspaceScript: strings.TrimSpace(req.InitWorkspaceScript),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := u.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Log the create operation
	if u.auditUsecase != nil {
		_ = u.auditUsecase.LogProjectOperation(ctx, entity.AuditActionCreate, project.ID, nil, project, fmt.Sprintf("Created project '%s'", project.Name))
	}

	// Try to automatically update repository URL from Git
	// Use background context for async operation
	bgCtx := context.Background()
	err = u.gitService.UpdateProjectRepositoryURL(bgCtx, project.ID, project.WorktreeBasePath, func(projectID uuid.UUID, repoURL string) error {
		return u.UpdateRepositoryURL(bgCtx, projectID, repoURL)
	})
	if err != nil {
		// Log error but don't fail the project creation
		fmt.Printf("Failed to auto-update repository URL for project %s: %v\n", project.ID, err)
	}

	return project, nil
}

func (u *projectUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	return u.projectRepo.GetByID(ctx, id)
}

func (u *projectUsecase) GetAll(ctx context.Context, params GetProjectsParams) (*GetProjectsResult, error) {
	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	projects, total, err := u.projectRepo.GetAllWithParams(ctx, repository.GetProjectsParams{
		Search:    params.Search,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
		Page:      params.Page,
		PageSize:  params.PageSize,
		Archived:  params.Archived,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return &GetProjectsResult{
		Projects: projects,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

func (u *projectUsecase) Update(ctx context.Context, id uuid.UUID, req UpdateProjectRequest) (*entity.Project, error) {
	oldProject, err := u.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create a copy for audit logging
	originalProject := *oldProject

	// Validate and update fields if provided
	if req.Name != "" {
		if err := validateProjectName(req.Name); err != nil {
			return nil, err
		}
		// Check for duplicate name (excluding current project)
		exists, err := u.CheckNameExists(ctx, req.Name, &id)
		if err != nil {
			return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
		}
		if exists {
			return nil, ErrProjectNameExists
		}
		oldProject.Name = strings.TrimSpace(req.Name)
	}
	if req.Description != "" {
		if err := validateDescription(req.Description); err != nil {
			return nil, err
		}
		oldProject.Description = strings.TrimSpace(req.Description)
	}
	if req.RepositoryURL != "" {
		if err := validateRepoURL(req.RepositoryURL); err != nil {
			return nil, err
		}
		oldProject.RepositoryURL = strings.TrimSpace(req.RepositoryURL)
	}
	if req.WorktreeBasePath != "" {
		oldProject.WorktreeBasePath = strings.TrimSpace(req.WorktreeBasePath)
	}
	if req.InitWorkspaceScript != "" {
		oldProject.InitWorkspaceScript = strings.TrimSpace(req.InitWorkspaceScript)
	}

	oldProject.UpdatedAt = time.Now()

	if err := u.projectRepo.Update(ctx, oldProject); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Log the update operation
	if u.auditUsecase != nil {
		_ = u.auditUsecase.LogProjectOperation(ctx, entity.AuditActionUpdate, oldProject.ID, &originalProject, oldProject, fmt.Sprintf("Updated project '%s'", oldProject.Name))
	}

	return oldProject, nil
}

func (u *projectUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	// Get project for audit logging
	project, err := u.projectRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	err = u.projectRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Log the delete operation
	if u.auditUsecase != nil {
		_ = u.auditUsecase.LogProjectOperation(ctx, entity.AuditActionDelete, id, project, nil, fmt.Sprintf("Deleted project '%s'", project.Name))
	}

	return nil
}

func (u *projectUsecase) GetWithTasks(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	project, err := u.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// The repository should handle loading tasks via GORM preloading
	// For now, we'll return the project as-is since the relationship is defined
	return project, nil
}

func (u *projectUsecase) GetStatistics(ctx context.Context, id uuid.UUID) (*ProjectStatistics, error) {
	// Get project to ensure it exists
	_, err := u.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get task statistics
	taskCounts, err := u.projectRepo.GetTaskStatistics(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task statistics: %w", err)
	}

	// Calculate total tasks and completion percentage
	totalTasks := 0
	doneTasks := 0
	for status, count := range taskCounts {
		totalTasks += count
		if status == entity.TaskStatusDONE {
			doneTasks = count
		}
	}

	var completionPercent float64
	if totalTasks > 0 {
		completionPercent = float64(doneTasks) / float64(totalTasks) * 100
	}

	// Get last activity timestamp
	lastActivityAt, err := u.projectRepo.GetLastActivityAt(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get last activity: %w", err)
	}

	return &ProjectStatistics{
		TaskCounts:        taskCounts,
		TotalTasks:        totalTasks,
		CompletionPercent: completionPercent,
		LastActivityAt:    lastActivityAt,
	}, nil
}

func (u *projectUsecase) Archive(ctx context.Context, id uuid.UUID) error {
	project, err := u.projectRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	err = u.projectRepo.Archive(ctx, project.ID)
	if err != nil {
		return err
	}

	// Log the archive operation
	if u.auditUsecase != nil {
		_ = u.auditUsecase.LogProjectOperation(ctx, entity.AuditActionArchive, project.ID, project, nil, fmt.Sprintf("Archived project '%s'", project.Name))
	}

	return nil
}

func (u *projectUsecase) Restore(ctx context.Context, id uuid.UUID) error {
	err := u.projectRepo.Restore(ctx, id)
	if err != nil {
		return err
	}

	// Get restored project for audit logging
	project, err := u.projectRepo.GetByID(ctx, id)
	if err == nil && u.auditUsecase != nil {
		_ = u.auditUsecase.LogProjectOperation(ctx, entity.AuditActionRestore, id, nil, project, fmt.Sprintf("Restored project '%s'", project.Name))
	}

	return nil
}

func (u *projectUsecase) CheckNameExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	return u.projectRepo.CheckNameExists(ctx, strings.TrimSpace(name), excludeID)
}

func (u *projectUsecase) GetSettings(ctx context.Context, projectID uuid.UUID) (*entity.ProjectSettings, error) {
	// Verify project exists
	_, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	settings, err := u.projectRepo.GetSettings(ctx, projectID)
	if err != nil {
		// If settings don't exist, create default settings
		if err.Error() == "settings not found" {
			defaultSettings := &entity.ProjectSettings{
				ProjectID:            projectID,
				NotificationsEnabled: true,
				EmailNotifications:   false,
				GitBranch:            "main",
				GitAutoSync:          false,
				TaskPrefix:           "",
			}

			err = u.projectRepo.CreateSettings(ctx, defaultSettings)
			if err != nil {
				return nil, fmt.Errorf("failed to create default settings: %w", err)
			}

			return defaultSettings, nil
		}
		return nil, err
	}

	return settings, nil
}

func (u *projectUsecase) UpdateSettings(ctx context.Context, projectID uuid.UUID, settings *entity.ProjectSettings) (*entity.ProjectSettings, error) {
	// Verify project exists
	_, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	settings.ProjectID = projectID
	settings.UpdatedAt = time.Now()

	err = u.projectRepo.UpdateSettings(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to update settings: %w", err)
	}

	return settings, nil
}

func (u *projectUsecase) UpdateRepositoryURL(ctx context.Context, projectID uuid.UUID, repositoryURL string) error {
	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Update repository URL
	project.RepositoryURL = repositoryURL
	project.UpdatedAt = time.Now()

	if err := u.projectRepo.Update(ctx, project); err != nil {
		return fmt.Errorf("failed to update project repository URL: %w", err)
	}

	// Log the update operation
	if u.auditUsecase != nil {
		_ = u.auditUsecase.LogProjectOperation(ctx, entity.AuditActionUpdate, project.ID, nil, project, fmt.Sprintf("Updated repository URL to '%s'", repositoryURL))
	}

	return nil
}

func (u *projectUsecase) ReinitGitRepository(ctx context.Context, projectID uuid.UUID) error {
	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	repoInfo, err := u.gitService.GetGitStatus(ctx, project.WorktreeBasePath)
	if err != nil {
		return fmt.Errorf("failed to reinitialize git repository: %w", err)
	}

	if repoInfo.RemoteURL != project.RepositoryURL {
		u.UpdateRepositoryURL(ctx, projectID, repoInfo.RemoteURL)
	}

	return nil
}

func (u *projectUsecase) GetGitStatus(ctx context.Context, projectID uuid.UUID) (*GitStatus, error) {
	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	status := &GitStatus{
		GitEnabled:     project.RepositoryURL != "" || project.WorktreeBasePath != "",
		WorktreeExists: project.WorktreeBasePath != "",
		Status:         "Git status not implemented",
	}

	// TODO: Implement actual Git status checking using git service
	// For now, return basic status
	if project.RepositoryURL != "" {
		status.RepositoryValid = true
		status.RemoteURL = project.RepositoryURL
		status.OnMainBranch = true // Default assumption
	}

	return status, nil
}

// ListBranches lists all Git branches for a project
func (u *projectUsecase) ListBranches(ctx context.Context, projectID uuid.UUID) ([]GitBranch, error) {
	// Get project to ensure it exists and has Git configuration
	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if project.WorktreeBasePath == "" {
		return nil, fmt.Errorf("project has no worktree base path configured")
	}

	// TODO: Use git service to list actual branches
	// // For now, return mock branches
	// branches := []GitBranch{
	// 	{
	// 		Name:        "main",
	// 		IsCurrent:   true,
	// 		LastCommit:  "abc123def",
	// 		LastUpdated: "2024-01-15T10:30:00Z",
	// 	},
	// 	{
	// 		Name:        "develop",
	// 		IsCurrent:   false,
	// 		LastCommit:  "def456ghi",
	// 		LastUpdated: "2024-01-14T15:20:00Z",
	// 	},
	// 	{
	// 		Name:        "feature/user-auth",
	// 		IsCurrent:   false,
	// 		LastCommit:  "ghi789jkl",
	// 		LastUpdated: "2024-01-13T09:15:00Z",
	// 	},
	// }

	branches, err := u.gitService.ListBranches(ctx, project.WorktreeBasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	gitBranches := make([]GitBranch, len(branches))
	for i, branch := range branches {
		isCurrent := false
		if strings.HasPrefix(branch, "* ") {
			branch = strings.TrimPrefix(branch, "* ")
			isCurrent = true
		}
		gitBranches[i] = GitBranch{
			Name:        branch,
			IsCurrent:   isCurrent,
			LastCommit:  "",
			LastUpdated: "",
		}
	}

	// sort current branch to the top
	sort.Slice(gitBranches, func(i, j int) bool {
		return gitBranches[i].IsCurrent
	})

	return gitBranches, nil
}
