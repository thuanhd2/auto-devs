package git

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// GitManager provides high-level Git operations and management
type GitManager struct {
	commands   *GitCommands
	validator  *GitValidator
	logger     *slog.Logger
	config     *ManagerConfig
}

// ManagerConfig contains configuration for the GitManager
type ManagerConfig struct {
	DefaultTimeout   time.Duration
	MaxRetries      int
	WorkingDir      string
	EnableLogging   bool
	LogLevel        slog.Level
}

// NewGitManager creates a new GitManager instance
func NewGitManager(config *ManagerConfig) (*GitManager, error) {
	// Set default configuration
	if config == nil {
		config = &ManagerConfig{
			DefaultTimeout: 30 * time.Second,
			MaxRetries:     3,
			EnableLogging:  true,
			LogLevel:       slog.LevelInfo,
		}
	}
	
	// Initialize command executor
	executor, err := NewDefaultCommandExecutor()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Git command executor: %w", err)
	}
	
	// Initialize commands and validator
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	
	// Setup logger
	var logger *slog.Logger
	if config.EnableLogging {
		logger = slog.Default().With("component", "git-manager")
	} else {
		logger = slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	
	return &GitManager{
		commands:  commands,
		validator: validator,
		logger:    logger,
		config:    config,
	}, nil
}

// Initialize performs initial setup and validation
func (m *GitManager) Initialize(ctx context.Context) error {
	m.logger.Info("Initializing Git manager")
	
	// Validate Git installation
	if err := m.validator.ValidateGitInstallation(ctx); err != nil {
		m.logger.Error("Git installation validation failed", "error", err)
		return fmt.Errorf("Git installation validation failed: %w", err)
	}
	
	version, err := m.commands.Version(ctx)
	if err == nil {
		m.logger.Info("Git installation validated", "version", version)
	}
	
	return nil
}

// ValidateRepository validates a repository and returns detailed information
func (m *GitManager) ValidateRepository(ctx context.Context, repoPath string) (*RepositoryInfo, error) {
	if repoPath == "" && m.config.WorkingDir != "" {
		repoPath = m.config.WorkingDir
	}
	
	m.logger.Debug("Validating repository", "path", repoPath)
	
	info, err := m.validator.ValidateRepository(ctx, repoPath)
	if err != nil {
		m.logger.Error("Repository validation failed", "path", repoPath, "error", err)
		return nil, err
	}
	
	m.logger.Info("Repository validated successfully", 
		"path", repoPath, 
		"branch", info.CurrentBranch,
		"clean", info.WorkingDirStatus.IsClean)
	
	return info, nil
}

// CloneRepository clones a repository with validation and error handling
func (m *GitManager) CloneRepository(ctx context.Context, request *CloneRequest) (*RepositoryInfo, error) {
	m.logger.Info("Starting repository clone", 
		"url", request.URL, 
		"destination", request.Destination)
	
	// Validate repository URL
	if err := m.validator.ValidateRepositoryURL(ctx, request.URL); err != nil {
		m.logger.Error("Invalid repository URL", "url", request.URL, "error", err)
		return nil, fmt.Errorf("repository URL validation failed: %w", err)
	}
	
	// Ensure destination directory
	if request.Destination == "" {
		return nil, fmt.Errorf("destination directory is required")
	}
	
	// Perform clone operation with retry logic
	err := m.executeWithRetry(ctx, func() error {
		return m.commands.Clone(ctx, request.URL, request.Destination, request.Options)
	})
	
	if err != nil {
		m.logger.Error("Repository clone failed", "url", request.URL, "error", err)
		return nil, fmt.Errorf("clone operation failed: %w", err)
	}
	
	// Validate the cloned repository
	info, err := m.validator.ValidateRepository(ctx, request.Destination)
	if err != nil {
		m.logger.Error("Cloned repository validation failed", "path", request.Destination, "error", err)
		return nil, fmt.Errorf("cloned repository validation failed: %w", err)
	}
	
	m.logger.Info("Repository cloned successfully", 
		"url", request.URL, 
		"destination", request.Destination,
		"branch", info.CurrentBranch)
	
	return info, nil
}

// CreateBranch creates a new branch with validation
func (m *GitManager) CreateBranch(ctx context.Context, request *CreateBranchRequest) error {
	workingDir := m.getWorkingDir(request.WorkingDir)
	
	m.logger.Info("Creating branch", 
		"name", request.BranchName, 
		"working_dir", workingDir)
	
	// Validate branch name
	if err := m.validator.ValidateBranchName(request.BranchName); err != nil {
		m.logger.Error("Invalid branch name", "name", request.BranchName, "error", err)
		return fmt.Errorf("branch name validation failed: %w", err)
	}
	
	// Check if branch already exists
	exists, err := m.validator.CheckBranchExists(ctx, workingDir, request.BranchName)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}
	
	if exists && !request.Force {
		return fmt.Errorf("%w: branch '%s' already exists", ErrBranchAlreadyExists, request.BranchName)
	}
	
	// Create the branch
	err = m.executeWithRetry(ctx, func() error {
		return m.commands.CreateBranch(ctx, workingDir, request.BranchName, request.StartPoint)
	})
	
	if err != nil {
		m.logger.Error("Branch creation failed", "name", request.BranchName, "error", err)
		return fmt.Errorf("branch creation failed: %w", err)
	}
	
	m.logger.Info("Branch created successfully", "name", request.BranchName)
	return nil
}

// SwitchBranch switches to a different branch
func (m *GitManager) SwitchBranch(ctx context.Context, request *SwitchBranchRequest) error {
	workingDir := m.getWorkingDir(request.WorkingDir)
	
	m.logger.Info("Switching branch", 
		"name", request.BranchName, 
		"working_dir", workingDir)
	
	// Validate branch name
	if err := m.validator.ValidateBranchName(request.BranchName); err != nil {
		return fmt.Errorf("branch name validation failed: %w", err)
	}
	
	// Check working directory status if required
	if !request.Force {
		status, err := m.validator.ValidateWorkingDirectory(ctx, workingDir)
		if err != nil {
			return fmt.Errorf("failed to validate working directory: %w", err)
		}
		
		if !status.IsClean {
			return fmt.Errorf("%w: working directory has uncommitted changes", ErrWorkingDirDirty)
		}
	}
	
	// Switch to the branch
	err := m.executeWithRetry(ctx, func() error {
		return m.commands.Checkout(ctx, workingDir, request.BranchName, request.CreateIfNotExists)
	})
	
	if err != nil {
		m.logger.Error("Branch switch failed", "name", request.BranchName, "error", err)
		return fmt.Errorf("branch switch failed: %w", err)
	}
	
	m.logger.Info("Branch switched successfully", "name", request.BranchName)
	return nil
}

// GetRepositoryStatus returns current repository status and information
func (m *GitManager) GetRepositoryStatus(ctx context.Context, workingDir string) (*RepositoryStatus, error) {
	workingDir = m.getWorkingDir(workingDir)
	
	m.logger.Debug("Getting repository status", "working_dir", workingDir)
	
	// Validate repository
	info, err := m.validator.ValidateRepository(ctx, workingDir)
	if err != nil {
		return nil, fmt.Errorf("repository validation failed: %w", err)
	}
	
	// Get additional status information
	branches, err := m.commands.ListBranches(ctx, workingDir, &ListBranchesOptions{})
	if err != nil {
		m.logger.Warn("Failed to list branches", "error", err)
		branches = []string{} // Continue with empty branch list
	}
	
	status := &RepositoryStatus{
		Repository: *info,
		Branches:   branches,
		IsValid:    true,
	}
	
	return status, nil
}

// GetBranches returns a list of branches in the repository
func (m *GitManager) GetBranches(ctx context.Context, request *ListBranchesRequest) ([]string, error) {
	workingDir := m.getWorkingDir(request.WorkingDir)
	
	m.logger.Debug("Listing branches", "working_dir", workingDir)
	
	branches, err := m.commands.ListBranches(ctx, workingDir, request.Options)
	if err != nil {
		m.logger.Error("Failed to list branches", "error", err)
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	
	return branches, nil
}

// ValidateGitConfig validates Git configuration
func (m *GitManager) ValidateGitConfig(ctx context.Context, workingDir string) (*GitConfig, error) {
	workingDir = m.getWorkingDir(workingDir)
	
	m.logger.Debug("Validating Git configuration", "working_dir", workingDir)
	
	config, err := m.validator.ValidateGitConfig(ctx, workingDir)
	if err != nil {
		m.logger.Error("Git configuration validation failed", "error", err)
		return nil, fmt.Errorf("Git configuration validation failed: %w", err)
	}
	
	m.logger.Info("Git configuration validated", 
		"user_name", config.UserName,
		"user_email", config.UserEmail)
	
	return config, nil
}

// Helper methods

// executeWithRetry executes a function with retry logic
func (m *GitManager) executeWithRetry(ctx context.Context, operation func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= m.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retrying
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
			
			m.logger.Debug("Retrying operation", "attempt", attempt+1, "max_retries", m.config.MaxRetries)
		}
		
		err := operation()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Don't retry certain types of errors
		if !m.shouldRetry(err) {
			break
		}
	}
	
	return fmt.Errorf("operation failed after %d retries: %w", m.config.MaxRetries, lastErr)
}

// shouldRetry determines if an error should trigger a retry
func (m *GitManager) shouldRetry(err error) bool {
	// Don't retry validation errors or authentication errors
	if IsAuthenticationError(err) || IsBranchError(err) || IsRepositoryError(err) {
		return false
	}
	
	// Don't retry if it's a Git command timeout or cancellation
	if err == ErrCommandTimeout || err == ErrCommandCancelled {
		return false
	}
	
	// Retry network-related errors and temporary failures
	return true
}

// getWorkingDir returns the working directory, using config default if not provided
func (m *GitManager) getWorkingDir(workingDir string) string {
	if workingDir != "" {
		return workingDir
	}
	if m.config.WorkingDir != "" {
		return m.config.WorkingDir
	}
	return "."
}

// Request types for Git operations

// CloneRequest represents a repository clone request
type CloneRequest struct {
	URL         string
	Destination string
	Options     *CloneOptions
}

// CreateBranchRequest represents a branch creation request
type CreateBranchRequest struct {
	BranchName string
	StartPoint string
	WorkingDir string
	Force      bool
}

// SwitchBranchRequest represents a branch switch request
type SwitchBranchRequest struct {
	BranchName        string
	WorkingDir        string
	CreateIfNotExists bool
	Force             bool
}

// ListBranchesRequest represents a list branches request
type ListBranchesRequest struct {
	WorkingDir string
	Options    *ListBranchesOptions
}

// Response types for Git operations

// RepositoryStatus represents the complete status of a repository
type RepositoryStatus struct {
	Repository RepositoryInfo
	Branches   []string
	IsValid    bool
}