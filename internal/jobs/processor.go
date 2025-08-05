package jobs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Processor handles background job processing
type Processor struct {
	taskUsecase     usecase.TaskUsecase
	projectUsecase  usecase.ProjectUsecase
	worktreeUsecase usecase.WorktreeUsecase
	logger          *slog.Logger
}

// NewProcessor creates a new job processor
func NewProcessor(
	taskUsecase usecase.TaskUsecase,
	projectUsecase usecase.ProjectUsecase,
	worktreeUsecase usecase.WorktreeUsecase,
) *Processor {
	return &Processor{
		taskUsecase:     taskUsecase,
		projectUsecase:  projectUsecase,
		worktreeUsecase: worktreeUsecase,
		logger:          slog.Default().With("component", "job-processor"),
	}
}

// ProcessTaskPlanning processes task planning jobs
func (p *Processor) ProcessTaskPlanning(ctx context.Context, task *asynq.Task) error {
	p.logger.Info("Processing task planning job!!!!!!")

	payload, err := ParseTaskPlanningPayload(task)
	if err != nil {
		return fmt.Errorf("failed to parse task planning payload: %w", err)
	}

	p.logger.Info("Processing task planning job",
		"task_id", payload.TaskID,
		"branch_name", payload.BranchName,
		"project_id", payload.ProjectID)

	// Step 1: Update task status to PLANNING
	err = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusPLANNING)
	if err != nil {
		p.logger.Error("Failed to update task status to PLANNING",
			"task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to update task status to PLANNING: %w", err)
	}

	p.logger.Info("Updated task status to PLANNING!!!!!!")
	// Step 2: Get project details
	project, err := p.projectUsecase.GetByID(ctx, payload.ProjectID)
	if err != nil {
		// Revert task status on failure
		_ = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusTODO)
		p.logger.Error("Failed to get project",
			"project_id", payload.ProjectID, "error", err)
		return fmt.Errorf("failed to get project: %w", err)
	}

	p.logger.Info("Got project details!!!!!!")

	// Step 3: Create git worktree
	projectTask, err := p.taskUsecase.GetByID(ctx, payload.TaskID)
	if err != nil {
		p.logger.Error("Failed to get task", "task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to get task: %w", err)
	}

	worktree, err := p.createWorktree(ctx, project, projectTask)
	if err != nil {
		// Update task status back to TODO on failure
		_ = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusTODO)
		p.logger.Error("Failed to create worktree",
			"task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	p.logger.Info("Created worktree!!!!!!")

	// Step 4: Update task with worktree path and branch name
	err = p.updateTaskWithGitInfo(ctx, payload.TaskID, worktree.BranchName, worktree.WorktreePath)
	if err != nil {
		// Cleanup worktree on failure
		_ = p.cleanupWorktree(ctx, worktree.WorktreePath)
		_ = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusTODO)
		p.logger.Error("Failed to update task with git info",
			"task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to update task with git info: %w", err)
	}

	p.logger.Info("Updated task with git info!!!!!!")

	// Step 5: Run AI executor for planning (placeholder for now)
	planContent, err := p.runPlanningExecutor(ctx, payload.TaskID, worktree.WorktreePath)
	if err != nil {
		// Don't cleanup worktree, but update status back
		_ = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusPLANNING)
		p.logger.Error("Failed to run planning executor",
			"task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to run planning executor: %w", err)
	}

	p.logger.Info("Ran planning executor!!!!!!")

	// Step 6: Save plan and update status to PLAN_REVIEWING
	err = p.savePlanAndUpdateStatus(ctx, payload.TaskID, planContent)
	if err != nil {
		p.logger.Error("Failed to save plan",
			"task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to save plan: %w", err)
	}

	p.logger.Info("Task planning completed successfully", "task_id", payload.TaskID)
	return nil
}

// updateTaskStatus updates the task status
func (p *Processor) updateTaskStatus(ctx context.Context, taskID uuid.UUID, status entity.TaskStatus) error {
	p.logger.Info("Updating task status", "task_id", taskID, "status", status)
	_, err := p.taskUsecase.UpdateStatus(ctx, taskID, status)
	p.logger.Info("Updated task status", "task_id", taskID, "status", status)
	return err
}

// createWorktree creates a git worktree for the task
func (p *Processor) createWorktree(ctx context.Context, project *entity.Project, task *entity.Task) (*entity.Worktree, error) {
	if project.WorktreeBasePath == "" {
		return nil, fmt.Errorf("project has no worktree base path configured")
	}

	p.logger.Info("Creating worktree",
		"project_id", project.ID,
		"task_id", task.ID,
		"branch_name", task.BranchName)

	// Create worktree from the specified branch
	worktree, err := p.worktreeUsecase.CreateWorktreeForTask(ctx, usecase.CreateWorktreeRequest{
		TaskID:    task.ID,
		ProjectID: project.ID,
		TaskTitle: task.Title,
	})
	if err != nil {
		p.logger.Error("Failed to create worktree",
			"project_id", project.ID,
			"task_id", task.ID,
			"error", err)
		return nil, fmt.Errorf("failed to create git worktree: %w", err)
	}

	p.logger.Info("Worktree created successfully",
		"project_id", project.ID,
		"task_id", task.ID,
		"worktree_path", worktree.WorktreePath)

	return worktree, nil
}

// updateTaskWithGitInfo updates the task with git information
func (p *Processor) updateTaskWithGitInfo(ctx context.Context, taskID uuid.UUID, branchName, worktreePath string) error {
	updateReq := usecase.UpdateTaskRequest{
		BranchName:   &branchName,
		WorktreePath: &worktreePath,
	}

	_, err := p.taskUsecase.Update(ctx, taskID, updateReq)
	return err
}

// cleanupWorktree removes the worktree directory
func (p *Processor) cleanupWorktree(ctx context.Context, worktreePath string) error {
	if worktreePath == "" {
		p.logger.Warn("Empty worktree path, skipping cleanup")
		return nil
	}

	p.logger.Info("Cleaning up worktree", "path", worktreePath)

	// This would use git worktree remove command
	// For now, just log the cleanup attempt
	// TODO: Implement actual worktree cleanup using git commands
	p.logger.Warn("Worktree cleanup not implemented yet", "path", worktreePath)
	return nil
}

// runPlanningExecutor runs the AI executor for planning
func (p *Processor) runPlanningExecutor(ctx context.Context, taskID uuid.UUID, worktreePath string) (string, error) {
	// Placeholder for AI executor integration
	// This would spawn Claude Code CLI or other AI tools to generate a plan
	p.logger.Info("Running planning executor", "task_id", taskID, "worktree_path", worktreePath)

	// For now, return a mock plan
	return fmt.Sprintf("Planning generated for task %s at %s", taskID, worktreePath), nil
}

// savePlanAndUpdateStatus saves the generated plan and updates task status
func (p *Processor) savePlanAndUpdateStatus(ctx context.Context, taskID uuid.UUID, planContent string) error {
	// Update task with plan content (assuming there's a plan field)
	// For now, we'll just update the status to PLAN_REVIEWING
	_, err := p.taskUsecase.UpdateStatus(ctx, taskID, entity.TaskStatusPLANREVIEWING)
	return err
}
