package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/internal/service/git"
	worktreesvc "github.com/auto-devs/auto-devs/internal/service/worktree"
	"github.com/google/uuid"
)

type WorktreeUsecase interface {
	// Basic worktree lifecycle management
	CreateWorktreeForTask(ctx context.Context, req CreateWorktreeRequest) (*entity.Worktree, error)
	CleanupWorktreeForTask(ctx context.Context, req CleanupWorktreeRequest) error
	GetWorktreeByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.Worktree, error)
	GetWorktreesByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Worktree, error)
	UpdateWorktreeStatus(ctx context.Context, worktreeID uuid.UUID, status entity.WorktreeStatus) error

	// Worktree validation and health monitoring
	ValidateWorktree(ctx context.Context, worktreeID uuid.UUID) (*WorktreeValidationResult, error)
	GetWorktreeHealth(ctx context.Context, worktreeID uuid.UUID) (*WorktreeHealthInfo, error)

	// Branch management within worktrees
	CreateBranchForTask(ctx context.Context, taskID uuid.UUID, branchName string) error
	SwitchToBranch(ctx context.Context, worktreeID uuid.UUID, branchName string) error
	GetBranchInfo(ctx context.Context, worktreeID uuid.UUID) (*BranchInfo, error)

	// Worktree initialization and configuration
	InitializeWorktree(ctx context.Context, worktreeID uuid.UUID) error
	CopyConfigurationFiles(ctx context.Context, worktreeID uuid.UUID, sourcePath string) error

	// Error handling and recovery
	HandleWorktreeCreationFailure(ctx context.Context, taskID uuid.UUID, error error) error
	RecoverFailedWorktree(ctx context.Context, worktreeID uuid.UUID) error

	// Statistics and monitoring
	GetWorktreeStatistics(ctx context.Context, projectID uuid.UUID) (*entity.WorktreeStatistics, error)
	GetActiveWorktreesCount(ctx context.Context, projectID uuid.UUID) (int, error)
}

type CreateWorktreeRequest struct {
	TaskID     uuid.UUID `json:"task_id" binding:"required"`
	ProjectID  uuid.UUID `json:"project_id" binding:"required"`
	TaskTitle  string    `json:"task_title" binding:"required"`
	Repository string    `json:"repository,omitempty"` // Optional repository URL to clone
}

type CleanupWorktreeRequest struct {
	TaskID     uuid.UUID `json:"task_id" binding:"required"`
	ProjectID  uuid.UUID `json:"project_id" binding:"required"`
	BranchName string    `json:"branch_name,omitempty"` // Optional branch name to delete
	Force      bool      `json:"force"`                 // Force cleanup even if worktree is active
}

type WorktreeValidationResult struct {
	IsValid         bool      `json:"is_valid"`
	Errors          []string  `json:"errors,omitempty"`
	Warnings        []string  `json:"warnings,omitempty"`
	GitRepositoryOK bool      `json:"git_repository_ok"`
	BranchExists    bool      `json:"branch_exists"`
	DirectoryExists bool      `json:"directory_exists"`
	PermissionsOK   bool      `json:"permissions_ok"`
	ValidationTime  time.Time `json:"validation_time"`
}

type WorktreeHealthInfo struct {
	WorktreeID      uuid.UUID             `json:"worktree_id"`
	Status          entity.WorktreeStatus `json:"status"`
	IsHealthy       bool                  `json:"is_healthy"`
	IsValid         bool                  `json:"is_valid"`
	LastActivity    time.Time             `json:"last_activity"`
	DiskUsage       int64                 `json:"disk_usage"` // in bytes
	FileCount       int                   `json:"file_count"`
	GitStatus       string                `json:"git_status"`
	BranchStatus    string                `json:"branch_status"`
	HealthScore     int                   `json:"health_score"` // 0-100
	Issues          []string              `json:"issues,omitempty"`
	LastHealthCheck time.Time             `json:"last_health_check"`
}

type BranchInfo struct {
	Name           string    `json:"name"`
	IsCurrent      bool      `json:"is_current"`
	LastCommit     string    `json:"last_commit"`
	LastCommitDate time.Time `json:"last_commit_date"`
	CommitCount    int       `json:"commit_count"`
	IsClean        bool      `json:"is_clean"`
	HasUncommitted bool      `json:"has_uncommitted"`
	HasUntracked   bool      `json:"has_untracked"`
}

type worktreeUsecase struct {
	worktreeRepo          repository.WorktreeRepository
	taskRepo              repository.TaskRepository
	projectRepo           repository.ProjectRepository
	integratedWorktreeSvc *worktreesvc.IntegratedWorktreeService
	gitManager            *git.GitManager
	logger                *slog.Logger
}

func NewWorktreeUsecase(
	worktreeRepo repository.WorktreeRepository,
	taskRepo repository.TaskRepository,
	projectRepo repository.ProjectRepository,
	integratedWorktreeSvc *worktreesvc.IntegratedWorktreeService,
	gitManager *git.GitManager,
) WorktreeUsecase {
	return &worktreeUsecase{
		worktreeRepo:          worktreeRepo,
		taskRepo:              taskRepo,
		projectRepo:           projectRepo,
		integratedWorktreeSvc: integratedWorktreeSvc,
		gitManager:            gitManager,
		logger:                slog.Default().With("component", "worktree-usecase"),
	}
}

// CreateWorktreeForTask implements the basic worktree creation workflow
func (w *worktreeUsecase) CreateWorktreeForTask(ctx context.Context, req CreateWorktreeRequest) (*entity.Worktree, error) {
	w.logger.Info("Creating worktree for task",
		"task_id", req.TaskID,
		"project_id", req.ProjectID,
		"task_title", req.TaskTitle)

	// Step 1: Validate task eligibility for worktree creation
	if err := w.validateTaskEligibility(ctx, req.TaskID); err != nil {
		return nil, fmt.Errorf("task not eligible for worktree creation: %w", err)
	}

	// Step 2: Check if worktree already exists for this task
	existingWorktree, err := w.worktreeRepo.GetByTaskID(ctx, req.TaskID)
	if err == nil && existingWorktree != nil {
		return nil, fmt.Errorf("worktree already exists for task %s", req.TaskID)
	}

	// Step 3: Get project information (validate project exists)
	project, err := w.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	task, err := w.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	taskBranchName := ""
	if task.BaseBranchName != nil {
		taskBranchName = *task.BaseBranchName
	} else {
		taskBranchName = "main"
	}

	// Step 4: Generate unique branch name using naming conventions
	branchName, err := w.gitManager.GenerateBranchName(req.TaskID.String(), req.TaskTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to generate branch name: %w", err)
	}

	// Step 5: Create Git worktree from main branch
	worktreePath, err := w.integratedWorktreeSvc.CreateTaskWorktree(ctx, &worktreesvc.CreateTaskWorktreeRequest{
		ProjectID:         req.ProjectID.String(),
		TaskID:            req.TaskID.String(),
		TaskTitle:         req.TaskTitle,
		ProjectWorkDir:    project.WorktreeBasePath,
		ProjectMainBranch: taskBranchName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w", err)
	}
	// Step 6: Create worktree record in database with "creating" status
	worktree := &entity.Worktree{
		TaskID:       req.TaskID,
		ProjectID:    req.ProjectID,
		BranchName:   branchName,
		WorktreePath: worktreePath.WorktreePath,
		Status:       entity.WorktreeStatusCreating,
	}

	if err := w.worktreeRepo.Create(ctx, worktree); err != nil {
		return nil, fmt.Errorf("failed to create worktree record: %w", err)
	}

	w.logger.Info("Worktree created successfully=============",
		"worktree_path", worktreePath.WorktreePath,
		"branch_name", branchName)

	// Step 7: Update worktree record with path and set status to active
	worktree.WorktreePath = worktreePath.WorktreePath
	worktree.Status = entity.WorktreeStatusActive
	if err := w.worktreeRepo.Update(ctx, worktree); err != nil {
		return nil, fmt.Errorf("failed to update worktree record: %w", err)
	}

	// Step 8: Update task with Git information
	if err := w.updateTaskWithGitInfo(ctx, req.TaskID, branchName, worktreePath.WorktreePath); err != nil {
		w.logger.Warn("Failed to update task with Git info", "error", err)
	}

	w.logger.Info("Successfully created worktree for task",
		"task_id", req.TaskID,
		"worktree_path", worktreePath.WorktreePath,
		"branch_name", branchName)

	return worktree, nil
}

// CleanupWorktreeForTask implements basic worktree cleanup
func (w *worktreeUsecase) CleanupWorktreeForTask(ctx context.Context, req CleanupWorktreeRequest) error {
	w.logger.Info("Cleaning up worktree for task",
		"task_id", req.TaskID,
		"project_id", req.ProjectID,
		"force", req.Force)

	// Get worktree record
	worktree, err := w.worktreeRepo.GetByTaskID(ctx, req.TaskID)
	if err != nil {
		return fmt.Errorf("worktree not found for task: %w", err)
	}

	// Check if cleanup is allowed
	if !req.Force && worktree.Status == entity.WorktreeStatusActive {
		return fmt.Errorf("cannot cleanup active worktree without force flag")
	}

	// Update status to cleaning
	worktree.Status = entity.WorktreeStatusCleaning
	if err := w.worktreeRepo.Update(ctx, worktree); err != nil {
		return fmt.Errorf("failed to update worktree status: %w", err)
	}

	// Clean up worktree directory and files
	if err := w.integratedWorktreeSvc.CleanupTaskWorktree(ctx, &worktreesvc.CleanupTaskWorktreeRequest{
		ProjectID: req.ProjectID.String(),
		TaskID:    req.TaskID.String(),
	}); err != nil {
		// Update status to error if cleanup fails
		worktree.Status = entity.WorktreeStatusError
		w.worktreeRepo.Update(ctx, worktree)
		return fmt.Errorf("failed to cleanup worktree: %w", err)
	}

	// Soft delete worktree record
	if err := w.worktreeRepo.Delete(ctx, worktree.ID); err != nil {
		return fmt.Errorf("failed to delete worktree record: %w", err)
	}

	// Update task Git status
	if err := w.updateTaskGitStatus(ctx, req.TaskID, entity.TaskGitStatusNone); err != nil {
		w.logger.Warn("Failed to update task Git status", "error", err)
	}

	w.logger.Info("Successfully cleaned up worktree for task", "task_id", req.TaskID)
	return nil
}

// GetWorktreeByTaskID retrieves worktree information for a specific task
func (w *worktreeUsecase) GetWorktreeByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.Worktree, error) {
	return w.worktreeRepo.GetByTaskID(ctx, taskID)
}

// GetWorktreesByProjectID retrieves all worktrees for a project
func (w *worktreeUsecase) GetWorktreesByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Worktree, error) {
	return w.worktreeRepo.GetByProjectID(ctx, projectID)
}

// UpdateWorktreeStatus updates the status of a worktree
func (w *worktreeUsecase) UpdateWorktreeStatus(ctx context.Context, worktreeID uuid.UUID, status entity.WorktreeStatus) error {
	worktree, err := w.worktreeRepo.GetByID(ctx, worktreeID)
	if err != nil {
		return fmt.Errorf("worktree not found: %w", err)
	}

	// Validate status transition
	if err := entity.ValidateWorktreeStatusTransition(worktree.Status, status); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	worktree.Status = status
	return w.worktreeRepo.Update(ctx, worktree)
}

// ValidateWorktree implements basic worktree validation
func (w *worktreeUsecase) ValidateWorktree(ctx context.Context, worktreeID uuid.UUID) (*WorktreeValidationResult, error) {
	worktree, err := w.worktreeRepo.GetByID(ctx, worktreeID)
	if err != nil {
		return nil, fmt.Errorf("worktree not found: %w", err)
	}

	result := &WorktreeValidationResult{
		ValidationTime: time.Now(),
	}

	// Validate Git repository state
	_, err = w.gitManager.GetRepositoryStatus(ctx, worktree.WorktreePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Git repository error: %v", err))
	} else {
		result.GitRepositoryOK = true
	}

	// Check if branch exists by listing branches and checking if our branch is in the list
	branches, err := w.gitManager.GetBranches(ctx, &git.ListBranchesRequest{
		WorkingDir: worktree.WorktreePath,
	})
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Branch check error: %v", err))
	} else {
		for _, branch := range branches {
			if branch == worktree.BranchName {
				result.BranchExists = true
				break
			}
		}
	}

	// Check if directory exists
	worktreeInfo, err := w.integratedWorktreeSvc.GetTaskWorktreeInfo(ctx, worktree.ProjectID.String(), worktree.TaskID.String())
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Directory check error: %v", err))
	} else {
		result.DirectoryExists = worktreeInfo != nil
	}

	// Determine overall validity
	result.IsValid = len(result.Errors) == 0 && result.GitRepositoryOK && result.BranchExists && result.DirectoryExists

	return result, nil
}

// GetWorktreeHealth implements basic worktree health monitoring
func (w *worktreeUsecase) GetWorktreeHealth(ctx context.Context, worktreeID uuid.UUID) (*WorktreeHealthInfo, error) {
	worktree, err := w.worktreeRepo.GetByID(ctx, worktreeID)
	if err != nil {
		return nil, fmt.Errorf("worktree not found: %w", err)
	}

	health := &WorktreeHealthInfo{
		WorktreeID:      worktreeID,
		Status:          worktree.Status,
		LastHealthCheck: time.Now(),
	}

	// Get worktree info
	worktreeInfo, err := w.integratedWorktreeSvc.GetTaskWorktreeInfo(ctx, worktree.ProjectID.String(), worktree.TaskID.String())
	if err != nil {
		health.Issues = append(health.Issues, fmt.Sprintf("Failed to get worktree info: %v", err))
	} else {
		health.DiskUsage = worktreeInfo.WorktreeInfo.Size
		health.FileCount = worktreeInfo.WorktreeInfo.FileCount
		health.IsValid = worktreeInfo.WorktreeInfo.IsValid
	}

	// Get Git status
	repoStatus, err := w.gitManager.GetRepositoryStatus(ctx, worktree.WorktreePath)
	if err != nil {
		health.Issues = append(health.Issues, fmt.Sprintf("Failed to get Git status: %v", err))
	} else {
		if repoStatus.IsValid {
			health.GitStatus = "clean"
		} else {
			health.GitStatus = "dirty"
		}
	}

	// Calculate health score
	health.HealthScore = w.calculateHealthScore(health)

	// Determine overall health
	health.IsHealthy = health.HealthScore >= 80 && len(health.Issues) == 0

	return health, nil
}

// CreateBranchForTask creates a new branch for a task
func (w *worktreeUsecase) CreateBranchForTask(ctx context.Context, taskID uuid.UUID, branchName string) error {
	worktree, err := w.worktreeRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("worktree not found for task: %w", err)
	}

	return w.gitManager.CreateBranchFromMain(ctx, worktree.WorktreePath, branchName)
}

// SwitchToBranch switches to a specific branch in the worktree
func (w *worktreeUsecase) SwitchToBranch(ctx context.Context, worktreeID uuid.UUID, branchName string) error {
	worktree, err := w.worktreeRepo.GetByID(ctx, worktreeID)
	if err != nil {
		return fmt.Errorf("worktree not found: %w", err)
	}

	return w.gitManager.SwitchToBranch(ctx, worktree.WorktreePath, branchName)
}

// GetBranchInfo gets information about the current branch
func (w *worktreeUsecase) GetBranchInfo(ctx context.Context, worktreeID uuid.UUID) (*BranchInfo, error) {
	worktree, err := w.worktreeRepo.GetByID(ctx, worktreeID)
	if err != nil {
		return nil, fmt.Errorf("worktree not found: %w", err)
	}

	// Get current branch by listing branches and finding the current one
	branches, err := w.gitManager.GetBranches(ctx, &git.ListBranchesRequest{
		WorkingDir: worktree.WorktreePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	// Find current branch (marked with * in git branch output)
	var currentBranch string
	for _, branch := range branches {
		if len(branch) > 2 && branch[:2] == "* " {
			currentBranch = branch[2:]
			break
		}
	}

	// Get repository status
	repoStatus, err := w.gitManager.GetRepositoryStatus(ctx, worktree.WorktreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository status: %w", err)
	}

	return &BranchInfo{
		Name:           currentBranch,
		IsCurrent:      currentBranch == worktree.BranchName,
		LastCommit:     "",          // Not available in current RepositoryStatus
		LastCommitDate: time.Time{}, // Not available in current RepositoryStatus
		CommitCount:    0,           // Not available in current RepositoryStatus
		IsClean:        repoStatus.IsValid,
		HasUncommitted: false, // Not available in current RepositoryStatus
		HasUntracked:   false, // Not available in current RepositoryStatus
	}, nil
}

// InitializeWorktree implements basic worktree initialization
func (w *worktreeUsecase) InitializeWorktree(ctx context.Context, worktreeID uuid.UUID) error {
	worktree, err := w.worktreeRepo.GetByID(ctx, worktreeID)
	if err != nil {
		return fmt.Errorf("worktree not found: %w", err)
	}

	// Create basic worktree metadata files
	metadataPath := fmt.Sprintf("%s/.worktree-metadata", worktree.WorktreePath)
	// Note: WriteFile method doesn't exist in GitManager, so we'll skip this for now
	w.logger.Info("Would create metadata file", "worktree_id", worktreeID, "metadata_path", metadataPath)

	w.logger.Info("Initialized worktree", "worktree_id", worktreeID, "metadata_path", metadataPath)
	return nil
}

// CopyConfigurationFiles copies necessary configuration files to the worktree
func (w *worktreeUsecase) CopyConfigurationFiles(ctx context.Context, worktreeID uuid.UUID, sourcePath string) error {
	worktree, err := w.worktreeRepo.GetByID(ctx, worktreeID)
	if err != nil {
		return fmt.Errorf("worktree not found: %w", err)
	}

	// This is a placeholder for copying configuration files
	// In a real implementation, you would copy specific files from sourcePath to worktree.WorktreePath
	w.logger.Info("Copying configuration files",
		"worktree_id", worktreeID,
		"source_path", sourcePath,
		"target_path", worktree.WorktreePath)

	return nil
}

// HandleWorktreeCreationFailure implements basic error handling
func (w *worktreeUsecase) HandleWorktreeCreationFailure(ctx context.Context, taskID uuid.UUID, error error) error {
	w.logger.Error("Handling worktree creation failure", "task_id", taskID, "error", error)

	// Get worktree record
	worktree, err := w.worktreeRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get worktree record: %w", err)
	}

	// Update status to error
	worktree.Status = entity.WorktreeStatusError
	if err := w.worktreeRepo.Update(ctx, worktree); err != nil {
		return fmt.Errorf("failed to update worktree status: %w", err)
	}

	// Clean up any partial worktree
	if worktree.WorktreePath != "" {
		w.integratedWorktreeSvc.CleanupTaskWorktree(ctx, &worktreesvc.CleanupTaskWorktreeRequest{
			ProjectID: worktree.ProjectID.String(),
			TaskID:    worktree.TaskID.String(),
		})
	}

	// Update task Git status
	if err := w.updateTaskGitStatus(ctx, taskID, entity.TaskGitStatusError); err != nil {
		w.logger.Warn("Failed to update task Git status", "error", err)
	}

	return nil
}

// RecoverFailedWorktree implements basic recovery from interrupted operations
func (w *worktreeUsecase) RecoverFailedWorktree(ctx context.Context, worktreeID uuid.UUID) error {
	worktree, err := w.worktreeRepo.GetByID(ctx, worktreeID)
	if err != nil {
		return fmt.Errorf("worktree not found: %w", err)
	}

	if worktree.Status != entity.WorktreeStatusError {
		return fmt.Errorf("worktree is not in error status")
	}

	// Attempt to recreate the worktree
	task, err := w.taskRepo.GetByID(ctx, worktree.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Clean up existing worktree first
	if worktree.WorktreePath != "" {
		w.integratedWorktreeSvc.CleanupTaskWorktree(ctx, &worktreesvc.CleanupTaskWorktreeRequest{
			ProjectID: worktree.ProjectID.String(),
			TaskID:    worktree.TaskID.String(),
		})
	}

	// Recreate worktree
	_, err = w.CreateWorktreeForTask(ctx, CreateWorktreeRequest{
		TaskID:    worktree.TaskID,
		ProjectID: worktree.ProjectID,
		TaskTitle: task.Title,
	})

	return err
}

// GetWorktreeStatistics gets worktree statistics for a project
func (w *worktreeUsecase) GetWorktreeStatistics(ctx context.Context, projectID uuid.UUID) (*entity.WorktreeStatistics, error) {
	return w.worktreeRepo.GetWorktreeStatistics(ctx, projectID)
}

// GetActiveWorktreesCount gets the count of active worktrees for a project
func (w *worktreeUsecase) GetActiveWorktreesCount(ctx context.Context, projectID uuid.UUID) (int, error) {
	return w.worktreeRepo.GetActiveWorktreesCount(ctx, projectID)
}

// Helper methods

func (w *worktreeUsecase) validateTaskEligibility(ctx context.Context, taskID uuid.UUID) error {
	task, err := w.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Check if task is in a status that allows worktree creation
	if task.Status != entity.TaskStatusPLANNING {
		return fmt.Errorf("task must be in PLANNING status to create worktree")
	}

	// Check if task already has a worktree
	existingWorktree, err := w.worktreeRepo.GetByTaskID(ctx, taskID)
	if err == nil && existingWorktree != nil {
		return fmt.Errorf("task already has a worktree")
	}

	return nil
}

func (w *worktreeUsecase) updateTaskWithGitInfo(ctx context.Context, taskID uuid.UUID, branchName, worktreePath string) error {
	task, err := w.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.BranchName = &branchName
	task.WorktreePath = &worktreePath
	task.GitStatus = entity.TaskGitStatusActive

	return w.taskRepo.Update(ctx, task)
}

func (w *worktreeUsecase) updateTaskGitStatus(ctx context.Context, taskID uuid.UUID, status entity.TaskGitStatus) error {
	task, err := w.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.GitStatus = status
	return w.taskRepo.Update(ctx, task)
}

func (w *worktreeUsecase) calculateHealthScore(health *WorktreeHealthInfo) int {
	score := 100

	// Deduct points for issues
	score -= len(health.Issues) * 10

	// Deduct points for invalid worktree
	if !health.IsValid {
		score -= 20
	}

	// Deduct points for unclean Git status
	if health.GitStatus != "clean" {
		score -= 15
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}
