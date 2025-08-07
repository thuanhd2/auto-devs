package jobs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/internal/service/ai"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/auto-devs/auto-devs/internal/websocket"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Processor handles background job processing
type Processor struct {
	taskUsecase      usecase.TaskUsecase
	projectUsecase   usecase.ProjectUsecase
	worktreeUsecase  usecase.WorktreeUsecase
	planningService  *ai.PlanningService
	executionService *ai.ExecutionService
	planRepo         repository.PlanRepository
	wsService        *websocket.Service
	redisBroker      *RedisBrokerClient // Redis broker client for cross-process messaging
	logger           *slog.Logger
}

// NewProcessor creates a new job processor
func NewProcessor(
	taskUsecase usecase.TaskUsecase,
	projectUsecase usecase.ProjectUsecase,
	worktreeUsecase usecase.WorktreeUsecase,
	planningService *ai.PlanningService,
	executionService *ai.ExecutionService,
	planRepo repository.PlanRepository,
	wsService *websocket.Service,
) *Processor {
	return &Processor{
		taskUsecase:      taskUsecase,
		projectUsecase:   projectUsecase,
		worktreeUsecase:  worktreeUsecase,
		planningService:  planningService,
		executionService: executionService,
		planRepo:         planRepo,
		wsService:        wsService,
		logger:           slog.Default().With("component", "job-processor"),
	}
}

// NewProcessorWithRedisBroker creates a new job processor with Redis broker
func NewProcessorWithRedisBroker(
	taskUsecase usecase.TaskUsecase,
	projectUsecase usecase.ProjectUsecase,
	worktreeUsecase usecase.WorktreeUsecase,
	planningService *ai.PlanningService,
	executionService *ai.ExecutionService,
	planRepo repository.PlanRepository,
	wsService *websocket.Service,
	redisBroker *RedisBrokerClient,
) *Processor {
	return &Processor{
		taskUsecase:      taskUsecase,
		projectUsecase:   projectUsecase,
		worktreeUsecase:  worktreeUsecase,
		planningService:  planningService,
		executionService: executionService,
		planRepo:         planRepo,
		wsService:        wsService,
		redisBroker:      redisBroker,
		logger:           slog.Default().With("component", "job-processor"),
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

	// Step 1: Check current task status and update to PLANNING if needed
	currentTask, err := p.taskUsecase.GetByID(ctx, payload.TaskID)
	if err != nil {
		p.logger.Error("Failed to get task for status check",
			"task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Only update status to PLANNING if it's not already PLANNING
	// This handles cases where the status was already updated by the handler
	if currentTask.Status != entity.TaskStatusPLANNING {
		err = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusPLANNING)
		if err != nil {
			p.logger.Error("Failed to update task status to PLANNING",
				"task_id", payload.TaskID, "error", err)
			return fmt.Errorf("failed to update task status to PLANNING: %w", err)
		}
		p.logger.Info("Updated task status to PLANNING!!!!!!")
	} else {
		p.logger.Info("Task status is already PLANNING, skipping status update!!!!!!")
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

	// Step 5: Run AI executor for planning
	planContent, err := p.runPlanningExecutor(ctx, payload.TaskID, worktree.WorktreePath)
	if err != nil {
		// On planning failure, revert task status but keep worktree for manual planning
		revertErr := p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusTODO)
		if revertErr != nil {
			p.logger.Error("Failed to revert task status after planning failure",
				"task_id", payload.TaskID, "revert_error", revertErr)
		}
		p.logger.Error("Failed to run planning executor",
			"task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to run planning executor: %w", err)
	}

	p.logger.Info("Ran planning executor!!!!!!")

	// Step 6: Save plan and update status to PLAN_REVIEWING
	err = p.savePlanAndUpdateStatus(ctx, payload.TaskID, planContent)
	if err != nil {
		// If plan saving fails, revert task status to PLANNING so it can be retried
		revertErr := p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusPLANNING)
		if revertErr != nil {
			p.logger.Error("Failed to revert task status after plan save failure",
				"task_id", payload.TaskID, "revert_error", revertErr)
		}
		p.logger.Error("Failed to save plan",
			"task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to save plan: %w", err)
	}

	p.logger.Info("Task planning completed successfully", "task_id", payload.TaskID)
	return nil
}

func (p *Processor) ProcessTaskImplementation(ctx context.Context, task *asynq.Task) error {
	p.logger.Info("Processing task implementation job!!!!!!")

	payload, err := ParseTaskImplementationPayload(task)
	if err != nil {
		return fmt.Errorf("failed to parse task implementation payload: %w", err)
	}

	p.logger.Info("Processing task implementation job",
		"task_id", payload.TaskID,
		"project_id", payload.ProjectID)

	// Step 1: Check current task status and update to IMPLEMENTING if needed
	currentTask, err := p.taskUsecase.GetByID(ctx, payload.TaskID)
	if err != nil {
		p.logger.Error("Failed to get task for status check",
			"task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Only update status to IMPLEMENTING if it's not already IMPLEMENTING
	// This handles cases where the status was already updated by the handler
	if currentTask.Status != entity.TaskStatusIMPLEMENTING {
		err = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusIMPLEMENTING)
		if err != nil {
			p.logger.Error("Failed to update task status to IMPLEMENTING",
				"task_id", payload.TaskID, "error", err)
			return fmt.Errorf("failed to update task status to IMPLEMENTING: %w", err)
		}
		p.logger.Info("Updated task status to IMPLEMENTING")
	} else {
		p.logger.Info("Task status is already IMPLEMENTING, skipping status update")
	}

	p.logger.Info("Updated task status to IMPLEMENTING")

	// Step 2: Get the task and check if it has git worktree
	projectTask, err := p.taskUsecase.GetByID(ctx, payload.TaskID)
	if err != nil {
		// Revert task status on failure
		_ = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusPLANREVIEWING)
		p.logger.Error("Failed to get task", "task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check if task has worktree path
	if projectTask.WorktreePath == nil || *projectTask.WorktreePath == "" {
		// Revert task status on failure
		_ = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusPLANREVIEWING)
		p.logger.Error("Task does not have worktree path", "task_id", payload.TaskID)
		return fmt.Errorf("task does not have worktree path set")
	}

	p.logger.Info("Task has valid worktree path", "task_id", payload.TaskID, "worktree_path", *projectTask.WorktreePath)

	// Step 3: Get the approved plan for the task
	plan, err := p.planRepo.GetByTaskID(ctx, payload.TaskID)
	if err != nil {
		// Revert task status on failure
		_ = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusPLANREVIEWING)
		p.logger.Error("Failed to get plan for task", "task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to get plan for task: %w", err)
	}

	// TODO: Need approve plan first
	// Step 4: Validate plan status - ensure it's APPROVED
	if plan.Status != entity.PlanStatusAPPROVED && plan.Status != entity.PlanStatusREVIEWING {
		// Revert task status on failure
		_ = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusPLANREVIEWING)
		p.logger.Error("Plan is not approved", "task_id", payload.TaskID, "plan_status", plan.Status)
		return fmt.Errorf("plan is not approved, current status: %s", plan.Status)
	}

	p.logger.Info("Plan is approved and ready for implementation", "task_id", payload.TaskID, "plan_id", plan.ID)

	// Step 5: Convert entity.Plan to ai.Plan format for the execution service
	aiPlan := p.convertEntityPlanToAIPlan(plan)

	// Step 6: Start AI execution using executionService.StartExecution()
	execution, err := p.executionService.StartExecution(payload.TaskID.String(), *aiPlan)
	if err != nil {
		// Revert task status on failure
		_ = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusPLANREVIEWING)
		p.logger.Error("Failed to start AI execution", "task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to start AI execution: %w", err)
	}

	p.logger.Info("AI execution started successfully",
		"task_id", payload.TaskID,
		"execution_id", execution.ID,
		"execution_status", execution.Status)

	// Step 7: Monitor execution progress (in a real implementation, this would be done via callbacks)
	// For now, we'll just log that monitoring should be handled by the execution service callbacks
	p.logger.Info("Implementation execution started, progress will be monitored via execution service callbacks",
		"task_id", payload.TaskID,
		"execution_id", execution.ID)

	// Note: The execution service will handle updating task status to CODE_REVIEWING on completion
	// or back to IMPLEMENTING on failure through its callback mechanism

	return nil
}

// convertEntityPlanToAIPlan converts database Plan entity to AI service Plan format
func (p *Processor) convertEntityPlanToAIPlan(plan *entity.Plan) *ai.Plan {
	// For this implementation, we'll create a structured plan from the markdown content
	// In a more sophisticated version, the content could be parsed to extract steps
	return &ai.Plan{
		ID:          plan.ID.String(),
		TaskID:      plan.TaskID.String(),
		Description: "Implementation plan",
		Steps: []ai.PlanStep{
			{
				ID:          "1",
				Description: "Execute implementation based on plan content",
				Action:      "implement",
				Parameters: map[string]string{
					"content":       plan.Content,
					"worktree_path": "", // Will be set by execution service
				},
				Order: 1,
			},
		},
		Context: map[string]string{
			"plan_content": plan.Content,
			"plan_status":  string(plan.Status),
			"created_at":   plan.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
		CreatedAt: plan.CreatedAt,
	}
}

// updateTaskStatus updates the task status and broadcasts WebSocket notification
func (p *Processor) updateTaskStatus(ctx context.Context, taskID uuid.UUID, status entity.TaskStatus) error {
	p.logger.Info("Updating task status", "task_id", taskID, "status", status)

	// Get the current task to track the old status
	currentTask, err := p.taskUsecase.GetByID(ctx, taskID)
	if err != nil {
		p.logger.Error("Failed to get current task", "task_id", taskID, "error", err)
		return err
	}

	oldStatus := currentTask.Status

	// Update the task status
	task, err := p.taskUsecase.UpdateStatus(ctx, taskID, status)
	if err != nil {
		p.logger.Error("Failed to update task status", "task_id", taskID, "status", status, "error", err)
		return err
	}

	p.logger.Info("Updated task status", "task_id", taskID, "status", status)

	// Send WebSocket notifications if status actually changed
	if oldStatus != status {
		// Create changes map for task update notification
		changes := map[string]interface{}{
			"status": map[string]interface{}{
				"old": oldStatus,
				"new": status,
			},
		}

		// Convert task to response format for WebSocket
		taskResponse := map[string]interface{}{
			"id":         task.ID.String(),
			"project_id": task.ProjectID.String(),
			"title":      task.Title,
			"status":     string(task.Status),
			"updated_at": task.UpdatedAt,
		}

		// Try Redis broker first, then fallback to WebSocket service
		var notificationErr error

		if p.redisBroker != nil {
			// Use Redis broker for cross-process messaging
			if err := p.redisBroker.PublishTaskUpdated(task.ID, task.ProjectID, changes, taskResponse); err != nil {
				p.logger.Warn("Failed to publish via Redis broker, falling back to WebSocket service",
					"task_id", taskID, "error", err)
				notificationErr = err
			} else {
				p.logger.Debug("Published task update via Redis broker", "task_id", taskID)
			}

			// Send status changed notification via Redis broker
			if err := p.redisBroker.PublishStatusChanged(task.ID, task.ProjectID, "task",
				string(oldStatus), string(status)); err != nil {
				p.logger.Warn("Failed to publish status change via Redis broker",
					"task_id", taskID, "error", err)
			}
		}

		// Fallback to WebSocket service if Redis broker failed or not available
		if p.redisBroker == nil || notificationErr != nil {
			// Send task updated notification via service
			if err := p.wsService.NotifyTaskUpdated(task.ID, task.ProjectID, changes, taskResponse); err != nil {
				p.logger.Error("Failed to send WebSocket task update notification",
					"task_id", taskID, "error", err)
			}

			// Send status changed notification via service
			if err := p.wsService.NotifyStatusChanged(task.ID, task.ProjectID, "task",
				string(oldStatus), string(status)); err != nil {
				p.logger.Error("Failed to send WebSocket status change notification",
					"task_id", taskID, "error", err)
			}
		}

		p.logger.Info("Sent WebSocket notifications for status change",
			"task_id", taskID, "old_status", oldStatus, "new_status", status)
	}

	return nil
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
	p.logger.Info("Running planning executor", "task_id", taskID, "worktree_path", worktreePath)

	// Fetch the task from database
	task, err := p.taskUsecase.GetByID(ctx, taskID)
	if err != nil {
		p.logger.Error("Failed to get task for planning", "task_id", taskID, "error", err)
		return "", fmt.Errorf("failed to get task: %w", err)
	}

	p.logger.Info("Retrieved task for planning", "task_id", taskID, "title", task.Title)

	// Use PlanningService to generate a structured plan
	aiPlan, err := p.planningService.GeneratePlan(*task)
	if err != nil {
		p.logger.Error("Failed to generate AI plan", "task_id", taskID, "error", err)
		return "", fmt.Errorf("failed to generate AI plan: %w", err)
	}

	p.logger.Info("Generated AI plan", "task_id", taskID, "plan_id", aiPlan.ID)

	// Convert the plan to markdown using the service method
	planMarkdown := p.planningService.GetPlanAsMarkdown(aiPlan)

	p.logger.Info("Converted plan to markdown", "task_id", taskID, "content_length", len(planMarkdown))

	return planMarkdown, nil
}

// savePlanAndUpdateStatus saves the generated plan and updates task status
func (p *Processor) savePlanAndUpdateStatus(ctx context.Context, taskID uuid.UUID, planContent string) error {
	p.logger.Info("Saving plan and updating task status", "task_id", taskID)

	// Create a new Plan entity
	plan := &entity.Plan{
		TaskID:  taskID,
		Status:  entity.PlanStatusDRAFT,
		Content: planContent,
	}

	// Save the plan to the database
	err := p.planRepo.Create(ctx, plan)
	if err != nil {
		p.logger.Error("Failed to create plan", "task_id", taskID, "error", err)
		return fmt.Errorf("failed to create plan: %w", err)
	}

	p.logger.Info("Plan created successfully", "task_id", taskID, "plan_id", plan.ID)

	// Update the plan status to REVIEWING since the plan is ready for review
	err = p.planRepo.UpdateStatus(ctx, plan.ID, entity.PlanStatusREVIEWING)
	if err != nil {
		p.logger.Error("Failed to update plan status", "plan_id", plan.ID, "error", err)
		return fmt.Errorf("failed to update plan status: %w", err)
	}

	p.logger.Info("Plan status updated to REVIEWING", "plan_id", plan.ID)

	// Update task status to PLAN_REVIEWING with WebSocket broadcast
	err = p.updateTaskStatus(ctx, taskID, entity.TaskStatusPLANREVIEWING)
	if err != nil {
		p.logger.Error("Failed to update task status", "task_id", taskID, "error", err)
		return fmt.Errorf("failed to update task status: %w", err)
	}

	p.logger.Info("Task status updated to PLAN_REVIEWING", "task_id", taskID)
	return nil
}
