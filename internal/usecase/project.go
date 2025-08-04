package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
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
}

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	RepoURL     string `json:"repo_url" binding:"required"`

	// Git-related fields
	RepositoryURL    string `json:"repository_url"`
	MainBranch       string `json:"main_branch"`
	WorktreeBasePath string `json:"worktree_base_path"`
	GitAuthMethod    string `json:"git_auth_method"`
	GitEnabled       bool   `json:"git_enabled"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RepoURL     string `json:"repo_url"`

	// Git-related fields
	RepositoryURL    string `json:"repository_url"`
	MainBranch       string `json:"main_branch"`
	WorktreeBasePath string `json:"worktree_base_path"`
	GitAuthMethod    string `json:"git_auth_method"`
	GitEnabled       bool   `json:"git_enabled"`
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
}

func NewProjectUsecase(projectRepo repository.ProjectRepository, auditUsecase AuditUsecase) ProjectUsecase {
	return &projectUsecase{
		projectRepo:  projectRepo,
		auditUsecase: auditUsecase,
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
	if err := validateRepoURL(req.RepoURL); err != nil {
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
		ID:          uuid.New(),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		RepoURL:     strings.TrimSpace(req.RepoURL),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),

		// Git-related fields
		RepositoryURL:    strings.TrimSpace(req.RepositoryURL),
		MainBranch:       strings.TrimSpace(req.MainBranch),
		WorktreeBasePath: strings.TrimSpace(req.WorktreeBasePath),
		GitAuthMethod:    strings.TrimSpace(req.GitAuthMethod),
		GitEnabled:       req.GitEnabled,
	}

	if err := u.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Log the create operation
	if u.auditUsecase != nil {
		_ = u.auditUsecase.LogProjectOperation(ctx, entity.AuditActionCreate, project.ID, nil, project, fmt.Sprintf("Created project '%s'", project.Name))
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
	if req.RepoURL != "" {
		if err := validateRepoURL(req.RepoURL); err != nil {
			return nil, err
		}
		oldProject.RepoURL = strings.TrimSpace(req.RepoURL)
	}

	// Update Git-related fields
	if req.RepositoryURL != "" {
		oldProject.RepositoryURL = strings.TrimSpace(req.RepositoryURL)
	}
	if req.MainBranch != "" {
		oldProject.MainBranch = strings.TrimSpace(req.MainBranch)
	}
	if req.WorktreeBasePath != "" {
		oldProject.WorktreeBasePath = strings.TrimSpace(req.WorktreeBasePath)
	}
	if req.GitAuthMethod != "" {
		oldProject.GitAuthMethod = strings.TrimSpace(req.GitAuthMethod)
	}
	// GitEnabled is a boolean, so we need to check if it's explicitly set
	// For now, we'll always update it if provided
	oldProject.GitEnabled = req.GitEnabled

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
