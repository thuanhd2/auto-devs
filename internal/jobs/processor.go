package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	aiexecutors "github.com/auto-devs/auto-devs/internal/ai-executors"
	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/internal/service/ai"
	"github.com/auto-devs/auto-devs/internal/service/git"
	"github.com/auto-devs/auto-devs/internal/service/github"
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
	executionRepo    repository.ExecutionRepository
	executionLogRepo repository.ExecutionLogRepository
	wsService        *websocket.Service
	redisBroker      *RedisBrokerClient // Redis broker client for cross-process messaging
	gitManager       *git.GitManager
	prCreator        *github.PRCreator
	prRepo           repository.PullRequestRepository
	githubService    github.GitHubServiceInterface
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
	executionRepo repository.ExecutionRepository,
	executionLogRepo repository.ExecutionLogRepository,
	wsService *websocket.Service,
	gitManager *git.GitManager,
	prCreator *github.PRCreator,
	prRepo repository.PullRequestRepository,
	githubService github.GitHubServiceInterface,
) *Processor {
	return &Processor{
		taskUsecase:      taskUsecase,
		projectUsecase:   projectUsecase,
		worktreeUsecase:  worktreeUsecase,
		planningService:  planningService,
		executionService: executionService,
		planRepo:         planRepo,
		executionRepo:    executionRepo,
		executionLogRepo: executionLogRepo,
		wsService:        wsService,
		gitManager:       gitManager,
		prCreator:        prCreator,
		prRepo:           prRepo,
		githubService:    githubService,
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
	executionRepo repository.ExecutionRepository,
	executionLogRepo repository.ExecutionLogRepository,
	wsService *websocket.Service,
	redisBroker *RedisBrokerClient,
	gitManager *git.GitManager,
	prCreator *github.PRCreator,
	prRepo repository.PullRequestRepository,
	githubService github.GitHubServiceInterface,
) *Processor {
	return &Processor{
		taskUsecase:      taskUsecase,
		projectUsecase:   projectUsecase,
		worktreeUsecase:  worktreeUsecase,
		planningService:  planningService,
		executionService: executionService,
		planRepo:         planRepo,
		executionRepo:    executionRepo,
		executionLogRepo: executionLogRepo,
		wsService:        wsService,
		redisBroker:      redisBroker,
		gitManager:       gitManager,
		prCreator:        prCreator,
		prRepo:           prRepo,
		githubService:    githubService,
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

	// DO not create worktree if it already exists
	if projectTask.WorktreePath == nil || *projectTask.WorktreePath == "" {
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
	}
	// Step 5: Run AI executor for planning
	// reload projectTask with new worktree path
	projectTask, err = p.taskUsecase.GetByID(ctx, payload.TaskID)
	if err != nil {
		p.logger.Error("Failed to get task", "task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to get task: %w", err)
	}

	aiExecutor, err := p.getAiExecutor(payload.AIType)
	if err != nil {
		p.logger.Error("Failed to get AI executor", "task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to get AI executor: %w", err)
	}

	execution, err := p.executionService.StartExecution(projectTask, aiExecutor, true)
	if err != nil {
		p.logger.Error("Failed to start AI execution", "task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to start AI execution: %w", err)
	}

	// map execution to entity.Execution
	dbExecution := &entity.Execution{
		TaskID:    payload.TaskID,
		Status:    entity.ExecutionStatus(execution.Status),
		StartedAt: execution.StartedAt,
		Progress:  execution.Progress,
		Result:    nil,
	}

	err = p.executionRepo.Create(ctx, dbExecution)
	if err != nil {
		p.logger.Error("Failed to save execution to database", "task_id", payload.TaskID, "execution_id", execution.ID, "error", err)
		return fmt.Errorf("failed to save execution to database: %w", err)
	}

	stdoutChannel := make(chan string)
	stderrChannel := make(chan string)
	execution.RegisterStdoutChannel(stdoutChannel)
	execution.RegisterStderrChannel(stderrChannel)

	p.executionService.RunExecution(execution)

	go func() {
		for {
			time.Sleep(1 * time.Second)
			select {
			case <-execution.GetContextDoneChannel():
				backgroundCtx := context.Background()
				completedAt := time.Now()

				if execution.Error != "" {
					p.logger.Error("AI Planning execution failed", "task_id", payload.TaskID, "execution_id", execution.ID, "error", execution.Error)
					_ = p.updateTaskStatus(backgroundCtx, payload.TaskID, entity.TaskStatusTODO)
					err := p.executionRepo.MarkFailed(backgroundCtx, dbExecution.ID, completedAt, execution.Error)
					if err != nil {
						p.logger.Error("Failed to mark execution as failed", "error", err, "execution_id", dbExecution.ID)
					}
				} else {
					p.logger.Info("AI Planning execution completed successfully", "task_id", payload.TaskID, "execution_id", execution.ID)
					_ = p.updateTaskStatus(backgroundCtx, payload.TaskID, entity.TaskStatusPLANREVIEWING)
					err := p.executionRepo.MarkCompleted(backgroundCtx, dbExecution.ID, completedAt, nil)
					if err != nil {
						p.logger.Error("Failed to mark execution as completed", "error", err, "execution_id", dbExecution.ID)
					}
					result := execution.Result
					p.logger.Info("AI Planning execution result", "task_id", payload.TaskID, "execution_id", execution.ID, "result", result)
					if result != nil {
						planContent, err := aiExecutor.ParseOutputToPlan(result.Output)
						if err != nil {
							p.logger.Error("Failed to parse output to plan", "error", err, "execution_id", dbExecution.ID)
						}
						err = p.savePlanAndUpdateStatus(backgroundCtx, payload.TaskID, planContent)
						if err != nil {
							p.logger.Error("Failed to save plan", "error", err, "execution_id", dbExecution.ID)
						}
					}
				}
				return
			case stdout := <-stdoutChannel:
				p.logger.Info("AI Planning execution stdout", "task_id", payload.TaskID, "execution_id", execution.ID, "stdout", stdout)
				// Save stdout to execution database
				logs := aiExecutor.ParseOutputToLogs(stdout)
				// assign execution id to each log
				for _, log := range logs {
					log.ExecutionID = dbExecution.ID
				}
				err := p.executionLogRepo.BatchInsertOrUpdate(context.Background(), logs)
				if err != nil {
					p.logger.Error("Failed to insert or update logs", "error", err, "execution_id", dbExecution.ID)
				}
			case stderr := <-stderrChannel:
				p.logger.Error("AI Planning execution stderr", "task_id", payload.TaskID, "execution_id", execution.ID, "stderr", stderr)
			}
		}
	}()

	p.logger.Info("AI Planning execution started successfully",
		"task_id", payload.TaskID,
		"execution_id", execution.ID,
		"execution_status", execution.Status)

	p.logger.Info("Task planning is running background!", "task_id", payload.TaskID)
	return nil
}

func (p *Processor) getAiExecutor(aiType string) (ai.AiCodingCli, error) {
	switch aiType {
	case "claude-code":
		aiExecutor := aiexecutors.NewClaudeCodeExecutor()
		return aiExecutor, nil
	case "fake-code":
		aiExecutor := aiexecutors.NewFakeCodeExecutor()
		return aiExecutor, nil
	case "cursor-agent":
		aiExecutor := aiexecutors.NewCursorAgentExecutor()
		return aiExecutor, nil
	default:
		return nil, fmt.Errorf("invalid execution type: %s", aiType)
	}
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

	// Step 5: inject plan to task
	projectTask.Plans = []entity.Plan{*plan}

	// Step 6: Start AI execution using executionService.StartExecution()
	aiExecutor, err := p.getAiExecutor(payload.AIType)
	if err != nil {
		p.logger.Error("Failed to get AI executor", "task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to get AI executor: %w", err)
	}
	execution, err := p.executionService.StartExecution(projectTask, aiExecutor, false)
	if err != nil {
		// Revert task status on failure
		_ = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusPLANREVIEWING)
		p.logger.Error("Failed to start AI execution", "task_id", payload.TaskID, "error", err)
		return fmt.Errorf("failed to start AI execution: %w", err)
	}

	// Map AI execution to entity.Execution and save to database
	dbExecution := &entity.Execution{
		TaskID:    payload.TaskID,
		Status:    entity.ExecutionStatus(execution.Status),
		StartedAt: execution.StartedAt,
		Progress:  execution.Progress,
		Result:    nil,
	}

	err = p.executionRepo.Create(ctx, dbExecution)
	if err != nil {
		// Revert task status on failure
		_ = p.updateTaskStatus(ctx, payload.TaskID, entity.TaskStatusPLANREVIEWING)
		p.logger.Error("Failed to save execution to database", "task_id", payload.TaskID, "execution_id", execution.ID, "error", err)
		return fmt.Errorf("failed to save execution to database: %w", err)
	}

	p.logger.Info("Execution saved to database",
		"task_id", payload.TaskID,
		"ai_execution_id", execution.ID,
		"db_execution_id", dbExecution.ID)

	stdoutChannel := make(chan string)
	stderrChannel := make(chan string)
	execution.RegisterStdoutChannel(stdoutChannel)
	execution.RegisterStderrChannel(stderrChannel)

	p.executionService.RunExecution(execution)

	go func() {
		for {
			time.Sleep(1 * time.Second)
			select {
			case <-execution.GetContextDoneChannel():
				completedAt := time.Now()

				// Check if execution completed successfully or failed
				if execution.Error != "" {
					p.logger.Error("AI execution failed", "task_id", payload.TaskID, "execution_id", execution.ID, "error", execution.Error)
					_ = p.updateTaskStatus(context.Background(), payload.TaskID, entity.TaskStatusPLANREVIEWING) // Keep in implementing for retry

					// Mark execution as failed
					err := p.executionRepo.MarkFailed(context.Background(), dbExecution.ID, completedAt, execution.Error)
					if err != nil {
						p.logger.Error("Failed to mark execution as failed", "error", err, "execution_id", dbExecution.ID)
					}

					// Create failure log entry
					// failureLog := &entity.ExecutionLog{
					// 	ExecutionID: dbExecution.ID,
					// 	Level:       entity.LogLevelError,
					// 	Message:     fmt.Sprintf("Execution failed: %s", execution.Error),
					// 	Timestamp:   completedAt,
					// 	Source:      "system",
					// }
					// if err := p.executionLogRepo.Create(context.Background(), failureLog); err != nil {
					// 	p.logger.Error("Failed to save failure log", "error", err, "execution_id", dbExecution.ID)
					// }
				} else {
					p.logger.Info("AI execution completed successfully", "task_id", payload.TaskID, "execution_id", execution.ID)

					// Update execution status to COMPLETED
					err := p.executionRepo.MarkCompleted(context.Background(), dbExecution.ID, completedAt, nil)
					if err != nil {
						p.logger.Error("Failed to mark execution as completed", "error", err, "execution_id", dbExecution.ID)
					}
					// Execute PR creation workflow
					p.executePRCreationWorkflow(context.Background(), projectTask, plan, dbExecution)

					_ = p.updateTaskStatus(context.Background(), payload.TaskID, entity.TaskStatusCODEREVIEWING)

					// // Create completion log entry
					// completionLog := &entity.ExecutionLog{
					// 	ExecutionID: dbExecution.ID,
					// 	Level:       entity.LogLevelInfo,
					// 	Message:     "Execution completed successfully",
					// 	Timestamp:   completedAt,
					// 	Source:      "system",
					// }
					// if err := p.executionLogRepo.Create(context.Background(), completionLog); err != nil {
					// 	p.logger.Error("Failed to save completion log", "error", err, "execution_id", dbExecution.ID)
					// }
				}
				return
			case stdout := <-stdoutChannel:
				p.logger.Info("AI execution stdout", "task_id", payload.TaskID, "execution_id", execution.ID, "stdout", stdout)
				// Save stdout to execution database
				// stdoutLog := &entity.ExecutionLog{
				// 	ExecutionID: dbExecution.ID,
				// 	Level:       entity.LogLevelInfo,
				// 	Message:     stdout,
				// 	Timestamp:   time.Now(),
				// 	Source:      "stdout",
				// }
				// if err := p.executionLogRepo.Create(context.Background(), stdoutLog); err != nil {
				// 	p.logger.Error("Failed to save stdout log", "error", err, "execution_id", dbExecution.ID)
				// }
				logs := aiExecutor.ParseOutputToLogs(stdout)
				// assign execution id to each log
				for _, log := range logs {
					log.ExecutionID = dbExecution.ID
				}
				err := p.executionLogRepo.BatchInsertOrUpdate(context.Background(), logs)
				if err != nil {
					p.logger.Error("Failed to insert or update logs", "error", err, "execution_id", dbExecution.ID)
				}
			case stderr := <-stderrChannel:
				p.logger.Error("AI execution stderr", "task_id", payload.TaskID, "execution_id", execution.ID, "stderr", stderr)
				// Save stderr to execution database
				// stderrLog := &entity.ExecutionLog{
				// 	ExecutionID: dbExecution.ID,
				// 	Level:       entity.LogLevelError,
				// 	Message:     stderr,
				// 	Timestamp:   time.Now(),
				// 	Source:      "stderr",
				// }
				// if err := p.executionLogRepo.Create(context.Background(), stderrLog); err != nil {
				// 	p.logger.Error("Failed to save stderr log", "error", err, "execution_id", dbExecution.ID)
				// }
			}
		}
	}()

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

// executePRCreationWorkflow handles the automated PR creation workflow after successful AI implementation
func (p *Processor) executePRCreationWorkflow(ctx context.Context, projectTask *entity.Task, plan *entity.Plan, dbExecution *entity.Execution) {
	p.logger.Info("Starting PR creation workflow", "task_id", projectTask.ID)

	// Step 1: Check if task has a worktree path
	if projectTask.WorktreePath == nil {
		p.logger.Error("Task has no worktree path, cannot commit and push changes", "task_id", projectTask.ID)
		return
	}

	// Step 2: Check if there are pending changes in the worktree
	hasPendingChanges, err := p.gitManager.HasPendingChanges(ctx, *projectTask.WorktreePath)
	if err != nil {
		p.logger.Error("Failed to check pending changes", "error", err, "task_id", projectTask.ID)
		// Continue without failing the entire workflow
	}

	// Step 3: Commit and push changes if any exist
	if hasPendingChanges {
		commitMessage := fmt.Sprintf("Implement task: %s\n\nTask ID: %s\nAI Implementation completed via Auto-Devs\n\n- %s",
			projectTask.Title,
			projectTask.ID.String(),
			projectTask.Description)

		err = p.gitManager.CommitAndPush(ctx, *projectTask.WorktreePath, commitMessage, "origin", *projectTask.BranchName)
		if err != nil {
			p.logger.Error("Failed to commit and push changes", "error", err, "task_id", projectTask.ID)
			// Don't fail the workflow, but log the error
			return
		} else {
			p.logger.Info("Successfully committed and pushed changes", "task_id", projectTask.ID, "branch", *projectTask.BranchName)
		}
	} else {
		p.logger.Info("No pending changes to commit", "task_id", projectTask.ID)
	}

	// Step 4: Create PR using the existing PRCreator service
	if p.prCreator != nil && projectTask.BranchName != nil {
		project, err := p.projectUsecase.GetByID(ctx, projectTask.ProjectID)
		if err != nil {
			p.logger.Error("Failed to get project", "error", err, "task_id", projectTask.ID)
			return
		}
		projectTask.Project = project
		pr, err := p.prCreator.CreatePRFromImplementation(ctx, *projectTask, *dbExecution, plan)
		if err != nil {
			p.logger.Error("Failed to create PR", "error", err, "task_id", projectTask.ID)
			// Don't fail the workflow, log and continue
			return
		}

		// Step 5: Save PR to database
		if err := p.prRepo.Create(ctx, pr); err != nil {
			p.logger.Error("Failed to save PR to database", "error", err, "pr_id", pr.ID, "task_id", projectTask.ID)
		} else {
			p.logger.Info("PR created and saved successfully",
				"pr_number", pr.GitHubPRNumber,
				"task_id", projectTask.ID,
				"pr_id", pr.ID)

			// Step 6: Send WebSocket notification about PR creation
			p.sendPRNotification(ctx, projectTask.ProjectID, pr, "pr_created")
		}
	} else {
		p.logger.Warn("PR creation skipped - missing required services or branch name",
			"task_id", projectTask.ID,
			"has_pr_creator", p.prCreator != nil,
			"has_branch_name", projectTask.BranchName != nil)
	}
}

// sendPRNotification sends WebSocket notification about PR events
func (p *Processor) sendPRNotification(ctx context.Context, projectID uuid.UUID, pr *entity.PullRequest, eventType string) {
	if p.wsService != nil {
		data := map[string]interface{}{
			"type": eventType,
			"pr":   pr,
		}
		if err := p.wsService.SendProjectMessage(projectID, websocket.MessageTypePRUpdate, data); err != nil {
			p.logger.Error("Failed to send PR WebSocket notification", "error", err, "project_id", projectID, "pr_id", pr.ID)
		} else {
			p.logger.Debug("PR WebSocket notification sent successfully", "event_type", eventType, "pr_id", pr.ID)
		}
	}
}

// ProcessWorktreeCleanup processes worktree cleanup jobs
func (p *Processor) ProcessWorktreeCleanup(ctx context.Context, task *asynq.Task) error {
	p.logger.Info("Processing worktree cleanup job")

	_, err := ParseWorktreeCleanupPayload(task)
	if err != nil {
		return fmt.Errorf("failed to parse worktree cleanup payload: %w", err)
	}

	// Calculate cutoff time (7 days ago)
	cutoffTime := time.Now().AddDate(0, 0, -7)
	p.logger.Info("Looking for tasks eligible for cleanup", "cutoff_time", cutoffTime)

	// Get all tasks eligible for cleanup
	eligibleTasks, err := p.taskUsecase.GetTasksEligibleForWorktreeCleanup(ctx, cutoffTime)
	if err != nil {
		p.logger.Error("Failed to get tasks eligible for cleanup", "error", err)
		return fmt.Errorf("failed to get tasks eligible for cleanup: %w", err)
	}

	p.logger.Info("Found tasks eligible for cleanup", "count", len(eligibleTasks))

	// Process each eligible task
	successCount := 0
	errorCount := 0

	for _, t := range eligibleTasks {
		if err := p.cleanupTaskWorktree(ctx, t); err != nil {
			p.logger.Error("Failed to cleanup worktree for task",
				"task_id", t.ID,
				"worktree_path", *t.WorktreePath,
				"error", err)
			errorCount++
		} else {
			successCount++
		}
	}

	p.logger.Info("Completed worktree cleanup job",
		"total_tasks", len(eligibleTasks),
		"successful_cleanups", successCount,
		"failed_cleanups", errorCount)

	return nil
}

// cleanupTaskWorktree performs cleanup for a single task's worktree
func (p *Processor) cleanupTaskWorktree(ctx context.Context, task *entity.Task) error {
	if task.WorktreePath == nil || *task.WorktreePath == "" {
		p.logger.Warn("Task has no worktree path to cleanup", "task_id", task.ID)
		return nil
	}

	worktreePath := *task.WorktreePath
	p.logger.Info("Cleaning up worktree for task",
		"task_id", task.ID,
		"worktree_path", worktreePath,
		"status", task.Status)

	// Get project to determine base working directory
	project, err := p.projectUsecase.GetByID(ctx, task.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Step 1: Remove git worktree
	deleteReq := &git.DeleteWorktreeRequest{
		WorkingDir:   project.WorktreeBasePath,
		WorktreePath: worktreePath,
	}

	if err := p.gitManager.DeleteWorktree(ctx, deleteReq); err != nil {
		p.logger.Warn("Failed to delete git worktree, continuing with cleanup",
			"task_id", task.ID,
			"error", err)
		// Don't fail the entire cleanup if git worktree removal fails
	} else {
		p.logger.Info("Successfully removed git worktree", "task_id", task.ID)
	}

	// Step 2: Delete branch if it exists
	if task.BranchName != nil && *task.BranchName != "" {
		branchName := *task.BranchName
		if err := p.gitManager.DeleteBranch(ctx, project.WorktreeBasePath, branchName, true); err != nil {
			p.logger.Warn("Failed to delete branch, continuing with cleanup",
				"task_id", task.ID,
				"branch_name", branchName,
				"error", err)
			// Don't fail the entire cleanup if branch deletion fails
		} else {
			p.logger.Info("Successfully deleted branch",
				"task_id", task.ID,
				"branch_name", branchName)
		}
	}

	// Step 3: Remove worktree folder from filesystem
	if err := p.removeWorktreeFolder(worktreePath); err != nil {
		p.logger.Warn("Failed to remove worktree folder",
			"task_id", task.ID,
			"worktree_path", worktreePath,
			"error", err)
		// Don't fail the entire cleanup if folder removal fails
	} else {
		p.logger.Info("Successfully removed worktree folder", "task_id", task.ID)
	}

	// Step 4: Update task to clear worktree path and set git status to none
	updateReq := usecase.UpdateTaskRequest{
		WorktreePath: new(string), // Set to empty string
	}

	if _, err := p.taskUsecase.Update(ctx, task.ID, updateReq); err != nil {
		return fmt.Errorf("failed to update task worktree path after cleanup: %w", err)
	}

	// Update git status separately
	if _, err := p.taskUsecase.UpdateGitStatus(ctx, task.ID, entity.TaskGitStatusNone); err != nil {
		return fmt.Errorf("failed to update git status after cleanup: %w", err)
	}

	p.logger.Info("Successfully completed worktree cleanup for task", "task_id", task.ID)
	return nil
}

// removeWorktreeFolder removes the worktree folder from the filesystem
func (p *Processor) removeWorktreeFolder(worktreePath string) error {
	// Implementation would use os.RemoveAll to delete the folder
	// For safety, we'll add some basic validation
	if worktreePath == "" {
		return fmt.Errorf("empty worktree path")
	}

	// Basic safety check to ensure we're not deleting system directories
	if strings.Contains(worktreePath, "..") ||
		worktreePath == "/" ||
		strings.HasPrefix(worktreePath, "/bin") ||
		strings.HasPrefix(worktreePath, "/usr") ||
		strings.HasPrefix(worktreePath, "/etc") ||
		strings.HasPrefix(worktreePath, "/sys") ||
		strings.HasPrefix(worktreePath, "/proc") {
		return fmt.Errorf("unsafe worktree path: %s", worktreePath)
	}

	// Remove the folder
	if err := os.RemoveAll(worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree folder: %w", err)
	}

	p.logger.Info("Successfully removed worktree folder", "path", worktreePath)
	return nil
}

// ProcessPRStatusSync processes PR status sync jobs
func (p *Processor) ProcessPRStatusSync(ctx context.Context, task *asynq.Task) error {
	p.logger.Info("Processing PR status sync job")

	_, err := ParsePRStatusSyncPayload(task)
	if err != nil {
		return fmt.Errorf("failed to parse PR status sync payload: %w", err)
	}

	// Get all open PRs from database
	openPRs, err := p.prRepo.GetOpenPRs(ctx)
	if err != nil {
		p.logger.Error("Failed to get open PRs", "error", err)
		return fmt.Errorf("failed to get open PRs: %w", err)
	}

	p.logger.Info("Found open PRs to check", "count", len(openPRs))

	// Process each open PR
	for _, pr := range openPRs {
		if err := p.processSinglePR(ctx, pr); err != nil {
			p.logger.Error("Failed to process PR",
				"pr_id", pr.ID,
				"github_pr_number", pr.GitHubPRNumber,
				"repository", pr.Repository,
				"error", err)
			// Continue processing other PRs even if one fails
		}
	}

	p.logger.Info("Completed PR status sync job")
	return nil
}

// processSinglePR checks and updates the status of a single PR
func (p *Processor) processSinglePR(ctx context.Context, pr *entity.PullRequest) error {
	p.logger.Debug("Checking PR status",
		"pr_id", pr.ID,
		"github_pr_number", pr.GitHubPRNumber,
		"repository", pr.Repository,
		"current_status", pr.Status)

	// Get current PR status from GitHub
	updatedPR, err := p.githubService.GetPullRequest(ctx, pr.Repository, pr.GitHubPRNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR from GitHub: %w", err)
	}

	// Check if PR status has changed
	if pr.Status != updatedPR.Status {
		p.logger.Info("PR status changed",
			"pr_id", pr.ID,
			"github_pr_number", pr.GitHubPRNumber,
			"old_status", pr.Status,
			"new_status", updatedPR.Status)

		// Update PR status in database
		pr.Status = updatedPR.Status
		pr.MergedAt = updatedPR.MergedAt
		pr.ClosedAt = updatedPR.ClosedAt
		pr.MergeCommitSHA = updatedPR.MergeCommitSHA
		pr.MergedBy = updatedPR.MergedBy

		if err := p.prRepo.Update(ctx, pr); err != nil {
			return fmt.Errorf("failed to update PR status in database: %w", err)
		}

		// If PR was merged, automatically mark associated task as DONE
		if updatedPR.Status == entity.PullRequestStatusMerged {
			if err := p.autoCompleteTask(ctx, pr.TaskID); err != nil {
				p.logger.Error("Failed to auto-complete task",
					"task_id", pr.TaskID,
					"pr_id", pr.ID,
					"error", err)
				// Don't return error here as PR update was successful
			} else {
				p.logger.Info("Auto-completed task due to PR merge",
					"task_id", pr.TaskID,
					"pr_id", pr.ID,
					"github_pr_number", pr.GitHubPRNumber)
			}
		}

		// Send WebSocket notification about PR status change
		p.sendPRStatusChangeNotification(ctx, pr, string(pr.Status), string(updatedPR.Status))
	}

	return nil
}

// autoCompleteTask automatically marks a task as DONE when its PR is merged
func (p *Processor) autoCompleteTask(ctx context.Context, taskID uuid.UUID) error {
	p.logger.Info("Auto-completing task", "task_id", taskID)

	// Get current task to check if it's not already DONE
	currentTask, err := p.taskUsecase.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Only update if task is not already DONE
	if currentTask.Status != entity.TaskStatusDONE {
		// Update task status to DONE
		err = p.updateTaskStatus(ctx, taskID, entity.TaskStatusDONE)
		if err != nil {
			return fmt.Errorf("failed to update task status to DONE: %w", err)
		}

		p.logger.Info("Task auto-completed successfully", "task_id", taskID)
	} else {
		p.logger.Debug("Task is already DONE, skipping", "task_id", taskID)
	}

	return nil
}

// sendPRStatusChangeNotification sends WebSocket notification about PR status changes
func (p *Processor) sendPRStatusChangeNotification(ctx context.Context, pr *entity.PullRequest, oldStatus, newStatus string) {
	if p.wsService != nil {
		// Get the task to determine project ID
		task, err := p.taskUsecase.GetByID(ctx, pr.TaskID)
		if err != nil {
			p.logger.Error("Failed to get task for PR notification", "task_id", pr.TaskID, "error", err)
			return
		}

		data := map[string]interface{}{
			"type":       "pr_status_changed",
			"pr":         pr,
			"old_status": oldStatus,
			"new_status": newStatus,
		}

		if err := p.wsService.SendProjectMessage(task.ProjectID, websocket.MessageTypePRUpdate, data); err != nil {
			p.logger.Error("Failed to send PR status change WebSocket notification",
				"error", err,
				"project_id", task.ProjectID,
				"pr_id", pr.ID)
		} else {
			p.logger.Debug("PR status change WebSocket notification sent successfully",
				"pr_id", pr.ID,
				"old_status", oldStatus,
				"new_status", newStatus)
		}
	}
}
