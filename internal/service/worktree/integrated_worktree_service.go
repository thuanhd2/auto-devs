package worktree

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/auto-devs/auto-devs/config"
	"github.com/auto-devs/auto-devs/internal/service/git"
)

// IntegratedWorktreeService combines worktree and git operations
type IntegratedWorktreeService struct {
	worktreeManager *WorktreeManager
	gitManager      *git.GitManager
	logger          *slog.Logger
}

// IntegratedConfig contains configuration for the integrated service
type IntegratedConfig struct {
	Worktree *config.WorktreeConfig
	Git      *git.ManagerConfig
}

// NewIntegratedWorktreeService creates a new integrated worktree service
func NewIntegratedWorktreeService(config *IntegratedConfig) (*IntegratedWorktreeService, error) {
	// Initialize worktree manager
	worktreeManager, err := NewWorktreeManager(config.Worktree)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize worktree manager: %w", err)
	}

	// Initialize git manager
	gitManager, err := git.NewGitManager(config.Git)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git manager: %w", err)
	}

	// Initialize git manager
	if err := gitManager.Initialize(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize git manager: %w", err)
	}

	return &IntegratedWorktreeService{
		worktreeManager: worktreeManager,
		gitManager:      gitManager,
		logger:          slog.Default().With("component", "integrated-worktree-service"),
	}, nil
}

// CreateTaskWorktree creates a complete worktree setup for a task
func (iws *IntegratedWorktreeService) CreateTaskWorktree(ctx context.Context, request *CreateTaskWorktreeRequest) (*TaskWorktreeInfo, error) {
	iws.logger.Info("Creating task worktree",
		"project_id", request.ProjectID,
		"task_id", request.TaskID,
		"task_title", request.TaskTitle)

	// Generate worktree path
	worktreePath, err := iws.worktreeManager.GenerateWorktreePath(request.ProjectID, request.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate worktree path: %w", err)
	}

	// Create worktree directory
	_, err = iws.worktreeManager.CreateWorktree(ctx, request.ProjectID, request.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to create worktree directory: %w", err)
	}

	// Execute init workspace script if provided
	if request.InitWorkspaceScript != "" {
		if err := iws.executeInitScript(ctx, worktreePath, request.InitWorkspaceScript); err != nil {
			iws.logger.Warn("Failed to execute init workspace script", "error", err)
			// Continue with worktree creation even if script fails
		}
	}

	// Generate branch name
	branchName, err := iws.gitManager.GenerateBranchName(request.TaskID, request.TaskTitle)
	if err != nil {
		// Clean up worktree on error
		iws.worktreeManager.CleanupWorktree(ctx, worktreePath)
		return nil, fmt.Errorf("failed to generate branch name: %w", err)
	}

	// Create branch from main
	if err := iws.gitManager.CreateWorktree(ctx, &git.CreateWorktreeRequest{
		BaseWorkingDir:     request.ProjectWorkDir,
		BaseBranchName:     request.ProjectMainBranch,
		WorktreeWorkingDir: worktreePath,
		WorktreeBranchName: branchName,
	}); err != nil {
		// Clean up worktree on error
		iws.worktreeManager.CleanupWorktree(ctx, worktreePath)
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	// Get worktree info
	worktreeInfo, err := iws.worktreeManager.GetWorktreeInfo(worktreePath)
	if err != nil {
		iws.logger.Warn("Failed to get worktree info", "error", err)
	}

	// Get repository status
	repoStatus, err := iws.gitManager.GetRepositoryStatus(ctx, worktreePath)
	if err != nil {
		iws.logger.Warn("Failed to get repository status", "error", err)
	}

	info := &TaskWorktreeInfo{
		ProjectID:      request.ProjectID,
		TaskID:         request.TaskID,
		TaskTitle:      request.TaskTitle,
		WorktreePath:   worktreePath,
		BranchName:     branchName,
		CreatedAt:      time.Now(),
		WorktreeInfo:   worktreeInfo,
		RepositoryInfo: repoStatus,
	}

	iws.logger.Info("Task worktree created successfully",
		"worktree_path", worktreePath,
		"branch_name", branchName)

	return info, nil
}

// CleanupTaskWorktree cleans up a complete task worktree
func (iws *IntegratedWorktreeService) CleanupTaskWorktree(ctx context.Context, request *CleanupTaskWorktreeRequest) error {
	iws.logger.Info("Cleaning up task worktree",
		"project_id", request.ProjectID,
		"task_id", request.TaskID)

	// Generate worktree path
	worktreePath, err := iws.worktreeManager.GenerateWorktreePath(request.ProjectID, request.TaskID)
	if err != nil {
		return fmt.Errorf("failed to generate worktree path: %w", err)
	}

	// Check if worktree exists
	if !iws.worktreeManager.WorktreeExists(worktreePath) {
		iws.logger.Warn("Worktree does not exist, skipping cleanup", "path", worktreePath)
		return nil
	}

	// Delete branch from repository
	if err := iws.gitManager.DeleteWorktree(ctx, &git.DeleteWorktreeRequest{
		WorkingDir:   worktreePath,
		WorktreePath: worktreePath,
	}); err != nil {
		iws.logger.Warn("Failed to delete branch", "error", err)
		// Continue with cleanup even if branch deletion fails
	}

	// Clean up worktree directory
	if err := iws.worktreeManager.CleanupWorktree(ctx, worktreePath); err != nil {
		return fmt.Errorf("failed to cleanup worktree directory: %w", err)
	}

	iws.logger.Info("Task worktree cleaned up successfully",
		"worktree_path", worktreePath,
		"task_id", request.TaskID)

	return nil
}

// GetTaskWorktreeInfo gets complete information about a task worktree
func (iws *IntegratedWorktreeService) GetTaskWorktreeInfo(ctx context.Context, projectID, taskID string) (*TaskWorktreeInfo, error) {
	iws.logger.Debug("Getting task worktree info", "project_id", projectID, "task_id", taskID)

	// Generate worktree path
	worktreePath, err := iws.worktreeManager.GenerateWorktreePath(projectID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate worktree path: %w", err)
	}

	// Check if worktree exists
	if !iws.worktreeManager.WorktreeExists(worktreePath) {
		return nil, fmt.Errorf("worktree does not exist: %s", worktreePath)
	}

	// Get worktree info
	worktreeInfo, err := iws.worktreeManager.GetWorktreeInfo(worktreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree info: %w", err)
	}

	// Get repository status
	repoStatus, err := iws.gitManager.GetRepositoryStatus(ctx, worktreePath)
	if err != nil {
		iws.logger.Warn("Failed to get repository status", "error", err)
	}

	// Try to get current branch name
	branchName := ""
	if repoStatus != nil && len(repoStatus.Branches) > 0 {
		// Find the current branch (usually the first one or one that's not main/master)
		for _, branch := range repoStatus.Branches {
			if branch != "main" && branch != "master" {
				branchName = branch
				break
			}
		}
	}

	info := &TaskWorktreeInfo{
		ProjectID:      projectID,
		TaskID:         taskID,
		WorktreePath:   worktreePath,
		BranchName:     branchName,
		WorktreeInfo:   worktreeInfo,
		RepositoryInfo: repoStatus,
	}

	return info, nil
}

// ListProjectWorktrees lists all worktrees for a project with git information
func (iws *IntegratedWorktreeService) ListProjectWorktrees(ctx context.Context, projectID string) ([]*TaskWorktreeInfo, error) {
	iws.logger.Debug("Listing project worktrees", "project_id", projectID)

	// Get worktree paths
	worktreePaths, err := iws.worktreeManager.ListWorktrees(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	var worktreeInfos []*TaskWorktreeInfo

	for _, worktreePath := range worktreePaths {
		// Extract task ID from path
		taskID := iws.extractTaskIDFromPath(worktreePath)
		if taskID == "" {
			iws.logger.Warn("Could not extract task ID from path", "path", worktreePath)
			continue
		}

		// Get worktree info
		worktreeInfo, err := iws.worktreeManager.GetWorktreeInfo(worktreePath)
		if err != nil {
			iws.logger.Warn("Failed to get worktree info", "path", worktreePath, "error", err)
			continue
		}

		// Get repository status
		repoStatus, err := iws.gitManager.GetRepositoryStatus(ctx, worktreePath)
		if err != nil {
			iws.logger.Warn("Failed to get repository status", "path", worktreePath, "error", err)
		}

		// Try to get branch name
		branchName := ""
		if repoStatus != nil && len(repoStatus.Branches) > 0 {
			for _, branch := range repoStatus.Branches {
				if branch != "main" && branch != "master" {
					branchName = branch
					break
				}
			}
		}

		info := &TaskWorktreeInfo{
			ProjectID:      projectID,
			TaskID:         taskID,
			WorktreePath:   worktreePath,
			BranchName:     branchName,
			WorktreeInfo:   worktreeInfo,
			RepositoryInfo: repoStatus,
		}

		worktreeInfos = append(worktreeInfos, info)
	}

	return worktreeInfos, nil
}

// extractTaskIDFromPath extracts task ID from worktree path
func (iws *IntegratedWorktreeService) extractTaskIDFromPath(worktreePath string) string {
	// Path format: /tmp/test-integrated-worktrees/project-project-123/task-task-456
	// Extract the task ID from the last component that starts with "task-"

	// Split path by filepath separator
	parts := strings.Split(worktreePath, string(filepath.Separator))

	// Find the last component that starts with "task-"
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if strings.HasPrefix(part, "task-") {
			// Extract the task ID part after "task-"
			taskID := strings.TrimPrefix(part, "task-")
			if taskID != "" {
				return taskID
			}
		}
	}

	return ""
}

// executeInitScript executes the initialization script in the worktree directory
func (iws *IntegratedWorktreeService) executeInitScript(ctx context.Context, worktreePath string, script string) error {
	if script == "" {
		return nil
	}

	iws.logger.Info("Executing init workspace script", "path", worktreePath)

	// Create a context with timeout for script execution (5 minutes)
	scriptCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Execute script using bash
	cmd := exec.CommandContext(scriptCtx, "bash", "-c", script)
	cmd.Dir = worktreePath
	
	// Set environment variables
	cmd.Env = append(os.Environ(), 
		fmt.Sprintf("WORKTREE_PATH=%s", worktreePath),
		"TERM=xterm-256color",
	)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	
	// Log the output regardless of success or failure
	if len(output) > 0 {
		iws.logger.Info("Init script output", 
			"output", string(output),
			"path", worktreePath)
	}

	if err != nil {
		return fmt.Errorf("script execution failed: %w (output: %s)", err, string(output))
	}

	iws.logger.Info("Init workspace script executed successfully", "path", worktreePath)
	return nil
}

// initializeGitRepository initializes a Git repository in the specified directory
func (iws *IntegratedWorktreeService) initializeGitRepository(ctx context.Context, worktreePath string) error {
	iws.logger.Debug("Initializing Git repository", "path", worktreePath)

	// Initialize Git repository using git init
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = worktreePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize Git repository: %w", err)
	}

	// Create initial commit to establish main branch
	if err := iws.createInitialCommit(ctx, worktreePath); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	return nil
}

// createInitialCommit creates an initial commit in the repository
func (iws *IntegratedWorktreeService) createInitialCommit(ctx context.Context, worktreePath string) error {
	iws.logger.Debug("Creating initial commit", "path", worktreePath)

	// Create a README file
	readmeContent := "# Task Worktree\n\nThis is a worktree for task development.\n"
	readmePath := filepath.Join(worktreePath, "README.md")

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0o644); err != nil {
		return fmt.Errorf("failed to create README file: %w", err)
	}

	// Add file to Git
	cmd := exec.CommandContext(ctx, "git", "add", "README.md")
	cmd.Dir = worktreePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add README to Git: %w", err)
	}

	// Create commit
	cmd = exec.CommandContext(ctx, "git", "commit", "-m", "Initial commit")
	cmd.Dir = worktreePath
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test User", "GIT_AUTHOR_EMAIL=test@example.com")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	return nil
}

// Request and response types

// CreateTaskWorktreeRequest represents a request to create a task worktree
type CreateTaskWorktreeRequest struct {
	ProjectID           string `json:"project_id"`
	TaskID              string `json:"task_id"`
	TaskTitle           string `json:"task_title"`
	ProjectWorkDir      string `json:"project_work_dir"`
	ProjectMainBranch   string `json:"project_main_branch"`
	InitWorkspaceScript string `json:"init_workspace_script"`
}

// CleanupTaskWorktreeRequest represents a request to cleanup a task worktree
type CleanupTaskWorktreeRequest struct {
	ProjectID string `json:"project_id"`
	TaskID    string `json:"task_id"`
}

// TaskWorktreeInfo contains complete information about a task worktree
type TaskWorktreeInfo struct {
	ProjectID      string                `json:"project_id"`
	TaskID         string                `json:"task_id"`
	TaskTitle      string                `json:"task_title,omitempty"`
	WorktreePath   string                `json:"worktree_path"`
	BranchName     string                `json:"branch_name"`
	CreatedAt      time.Time             `json:"created_at"`
	WorktreeInfo   *WorktreeInfo         `json:"worktree_info,omitempty"`
	RepositoryInfo *git.RepositoryStatus `json:"repository_info,omitempty"`
}
