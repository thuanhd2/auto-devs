package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectGitService handles Git operations for projects
type ProjectGitService struct {
	validator *GitValidator
	manager   *GitManager
	commands  *GitCommands
}

// NewProjectGitService creates a new ProjectGitService instance
func NewProjectGitService(validator *GitValidator, manager *GitManager, commands *GitCommands) *ProjectGitService {
	return &ProjectGitService{
		validator: validator,
		manager:   manager,
		commands:  commands,
	}
}

// GitProjectConfig represents Git configuration for a project
type GitProjectConfig struct {
	RepositoryURL    string `json:"repository_url" validate:"required"`
	MainBranch       string `json:"main_branch" validate:"required"`
	WorktreeBasePath string `json:"worktree_base_path" validate:"required"`
	GitAuthMethod    string `json:"git_auth_method" validate:"required,oneof=ssh https"`
	GitEnabled       bool   `json:"git_enabled"`
}

// ValidateGitProjectConfig validates Git configuration for a project
func (s *ProjectGitService) ValidateGitProjectConfig(ctx context.Context, config *GitProjectConfig) error {
	if !config.GitEnabled {
		return nil // Git is not enabled, no validation needed
	}

	// Validate required fields when Git is enabled
	if config.RepositoryURL == "" {
		return fmt.Errorf("repository URL is required when Git is enabled")
	}

	if config.MainBranch == "" {
		return fmt.Errorf("main branch is required when Git is enabled")
	}

	if config.WorktreeBasePath == "" {
		return fmt.Errorf("worktree base path is required when Git is enabled")
	}

	if config.GitAuthMethod == "" {
		return fmt.Errorf("Git authentication method is required when Git is enabled")
	}

	// Validate authentication method
	if config.GitAuthMethod != "ssh" && config.GitAuthMethod != "https" {
		return fmt.Errorf("invalid Git authentication method: %s (must be 'ssh' or 'https')", config.GitAuthMethod)
	}

	// Validate repository URL format
	if err := s.validator.ValidateRepositoryURL(ctx, config.RepositoryURL); err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	// Validate authentication method matches repository URL
	if err := s.validateAuthMethodMatchesURL(config.RepositoryURL, config.GitAuthMethod); err != nil {
		return err
	}

	// Validate worktree base path
	if err := s.validateWorktreeBasePath(config.WorktreeBasePath); err != nil {
		return fmt.Errorf("invalid worktree base path: %w", err)
	}

	return nil
}

// TestGitConnection tests if the Git repository is accessible with current configuration
func (s *ProjectGitService) TestGitConnection(ctx context.Context, config *GitProjectConfig) error {
	if !config.GitEnabled {
		return fmt.Errorf("Git is not enabled for this project")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Test repository accessibility
	if err := s.testRepositoryAccessibility(ctx, config, tempDir); err != nil {
		return fmt.Errorf("repository accessibility test failed: %w", err)
	}

	return nil
}

// SetupGitProject initializes Git integration for a project
func (s *ProjectGitService) SetupGitProject(ctx context.Context, config *GitProjectConfig) error {
	if !config.GitEnabled {
		return fmt.Errorf("Git is not enabled for this project")
	}

	// Validate configuration
	if err := s.ValidateGitProjectConfig(ctx, config); err != nil {
		return fmt.Errorf("invalid Git configuration: %w", err)
	}

	// Test connection
	if err := s.TestGitConnection(ctx, config); err != nil {
		return fmt.Errorf("Git connection test failed: %w", err)
	}

	// Create worktree directory if it doesn't exist
	if err := os.MkdirAll(config.WorktreeBasePath, 0o755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}

	// Clone repository if worktree is empty
	if err := s.setupRepository(ctx, config); err != nil {
		return fmt.Errorf("failed to setup repository: %w", err)
	}

	return nil
}

// GetGitProjectStatus returns the current status of Git integration for a project
func (s *ProjectGitService) GetGitProjectStatus(ctx context.Context, config *GitProjectConfig) (*GitProjectStatus, error) {
	if !config.GitEnabled {
		return &GitProjectStatus{
			GitEnabled: false,
			Status:     "Git integration is disabled",
		}, nil
	}

	status := &GitProjectStatus{
		GitEnabled: true,
		Config:     config,
	}

	// Check if worktree directory exists
	if _, err := os.Stat(config.WorktreeBasePath); err == nil {
		status.WorktreeExists = true
	} else {
		status.Status = "Worktree directory does not exist"
		return status, nil
	}

	// Validate repository
	repoInfo, err := s.validator.ValidateRepository(ctx, config.WorktreeBasePath)
	if err != nil {
		status.Status = fmt.Sprintf("Repository validation failed: %v", err)
		return status, nil
	}

	status.RepositoryValid = true
	status.CurrentBranch = repoInfo.CurrentBranch
	status.RemoteURL = repoInfo.RemoteURL
	status.WorkingDirStatus = repoInfo.WorkingDirStatus

	// Check if current branch matches main branch
	if repoInfo.CurrentBranch == config.MainBranch {
		status.OnMainBranch = true
	}

	status.Status = "Git integration is working properly"
	return status, nil
}

// GitProjectStatus represents the status of Git integration for a project
type GitProjectStatus struct {
	GitEnabled       bool              `json:"git_enabled"`
	Config           *GitProjectConfig `json:"config,omitempty"`
	WorktreeExists   bool              `json:"worktree_exists"`
	RepositoryValid  bool              `json:"repository_valid"`
	CurrentBranch    string            `json:"current_branch,omitempty"`
	RemoteURL        string            `json:"remote_url,omitempty"`
	OnMainBranch     bool              `json:"on_main_branch"`
	WorkingDirStatus WorkingDirStatus  `json:"working_dir_status,omitempty"`
	Status           string            `json:"status"`
}

// Helper methods

func (s *ProjectGitService) validateAuthMethodMatchesURL(repoURL, authMethod string) error {
	if strings.HasPrefix(repoURL, "https://") && authMethod != "https" {
		return fmt.Errorf("HTTPS repository URL requires 'https' authentication method")
	}

	if strings.HasPrefix(repoURL, "ssh://") || strings.Contains(repoURL, "@") && strings.Contains(repoURL, ":") {
		if authMethod != "ssh" {
			return fmt.Errorf("SSH repository URL requires 'ssh' authentication method")
		}
	}

	return nil
}

func (s *ProjectGitService) validateWorktreeBasePath(path string) error {
	if path == "" {
		return fmt.Errorf("worktree base path cannot be empty")
	}

	// Check if path is absolute
	if !filepath.IsAbs(path) {
		return fmt.Errorf("worktree base path must be an absolute path")
	}

	// Check if path contains invalid characters
	if strings.Contains(path, "..") {
		return fmt.Errorf("worktree base path cannot contain '..'")
	}

	return nil
}

func (s *ProjectGitService) testRepositoryAccessibility(ctx context.Context, config *GitProjectConfig, tempDir string) error {
	// Try to clone the repository to test accessibility
	cloneOptions := &CloneOptions{
		Branch:       config.MainBranch,
		Depth:        1, // Shallow clone for testing
		SingleBranch: true,
	}

	if err := s.commands.Clone(ctx, config.RepositoryURL, tempDir, cloneOptions); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

func (s *ProjectGitService) setupRepository(ctx context.Context, config *GitProjectConfig) error {
	// Check if worktree directory is empty
	if s.isDirectoryEmpty(config.WorktreeBasePath) {
		// Clone repository
		cloneOptions := &CloneOptions{
			Branch: config.MainBranch,
		}

		if err := s.commands.Clone(ctx, config.RepositoryURL, config.WorktreeBasePath, cloneOptions); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
	} else {
		// Check if it's already a Git repository
		isRepo, err := s.commands.IsRepository(ctx, config.WorktreeBasePath)
		if err != nil {
			return fmt.Errorf("failed to check if directory is a repository: %w", err)
		}

		if !isRepo {
			return fmt.Errorf("worktree directory is not empty and not a Git repository")
		}

		// Validate that it's the correct repository
		repoInfo, err := s.validator.ValidateRepository(ctx, config.WorktreeBasePath)
		if err != nil {
			return fmt.Errorf("failed to validate existing repository: %w", err)
		}

		if repoInfo.RemoteURL != config.RepositoryURL {
			return fmt.Errorf("existing repository has different remote URL: expected %s, got %s",
				config.RepositoryURL, repoInfo.RemoteURL)
		}
	}

	return nil
}

func (s *ProjectGitService) isDirectoryEmpty(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return true // Consider as empty if we can't read it
	}
	return len(entries) == 0
}
