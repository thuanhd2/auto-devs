package git

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// ProjectGitServiceInterface defines the interface for project Git operations
type ProjectGitServiceInterface interface {
	UpdateProjectRepositoryURL(ctx context.Context, projectID uuid.UUID, worktreeBasePath string, updateRepoURL func(uuid.UUID, string) error) error
	SetupProjectGit(ctx context.Context, projectID uuid.UUID, worktreeBasePath string, updateRepoURL func(uuid.UUID, string) error) error
	GetGitStatus(ctx context.Context, worktreeBasePath string) (*RepositoryInfo, error)
}

// ProjectGitService handles Git operations for projects
type ProjectGitService struct {
	gitManager *GitManager
	logger     *slog.Logger
}

// NewProjectGitService creates a new ProjectGitService instance
func NewProjectGitService(gitManager *GitManager) *ProjectGitService {
	return &ProjectGitService{
		gitManager: gitManager,
		logger:     slog.Default().With("component", "project-git-service"),
	}
}

// UpdateProjectRepositoryURL automatically updates the repository URL for a project
// by reading the remote URL from the Git repository at the worktree base path
func (s *ProjectGitService) UpdateProjectRepositoryURL(ctx context.Context, projectID uuid.UUID, worktreeBasePath string, updateRepoURL func(uuid.UUID, string) error) error {
	if worktreeBasePath == "" {
		return fmt.Errorf("project has no worktree base path configured")
	}

	// Get remote URL from Git repository
	remoteURL, err := s.gitManager.commands.GetRemoteURL(ctx, worktreeBasePath, "origin")
	if err != nil {
		s.logger.Warn("Failed to get remote URL", "project_id", projectID, "error", err)
		return fmt.Errorf("failed to get remote URL from Git repository: %w", err)
	}

	if remoteURL == "" {
		s.logger.Warn("No remote URL found", "project_id", projectID)
		return fmt.Errorf("no remote URL found in Git repository")
	}

	// Update project with the repository URL
	err = updateRepoURL(projectID, remoteURL)
	if err != nil {
		return fmt.Errorf("failed to update project repository URL: %w", err)
	}

	s.logger.Info("Updated project repository URL",
		"project_id", projectID,
		"repository_url", remoteURL)

	return nil
}

// SetupProjectGit initializes Git for a project
func (s *ProjectGitService) SetupProjectGit(ctx context.Context, projectID uuid.UUID, worktreeBasePath string, updateRepoURL func(uuid.UUID, string) error) error {
	if worktreeBasePath == "" {
		return fmt.Errorf("project has no worktree base path configured")
	}

	// Validate the repository
	repoInfo, err := s.gitManager.ValidateRepository(ctx, worktreeBasePath)
	if err != nil {
		s.logger.Warn("Repository validation failed", "project_id", projectID, "error", err)
		return fmt.Errorf("failed to validate repository: %w", err)
	}

	// Repository is valid, try to get remote URL
	remoteURL, err := s.gitManager.commands.GetRemoteURL(ctx, worktreeBasePath, "origin")
	if err == nil && remoteURL != "" {
		// Update project with the repository URL
		err = updateRepoURL(projectID, remoteURL)
		if err != nil {
			s.logger.Warn("Failed to update repository URL", "project_id", projectID, "error", err)
		} else {
			s.logger.Info("Updated project repository URL",
				"project_id", projectID,
				"repository_url", remoteURL)
		}
	}

	s.logger.Info("Setup Git for project",
		"project_id", projectID,
		"worktree_path", worktreeBasePath,
		"repository_valid", true,
		"remote_url", repoInfo.RemoteURL)

	return nil
}

func (s *ProjectGitService) GetGitStatus(ctx context.Context, worktreeBasePath string) (*RepositoryInfo, error) {
	if worktreeBasePath == "" {
		return nil, fmt.Errorf("project has no worktree base path configured")
	}

	repoInfo, err := s.gitManager.ValidateRepository(ctx, worktreeBasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to validate repository: %w", err)
	}

	return repoInfo, nil
}
