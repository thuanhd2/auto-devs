package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/google/uuid"
)

// JobClientInterface defines the interface for job client operations
type JobClientInterface interface {
	EnqueueTaskPlanning(payload *TaskPlanningPayload, delay time.Duration) (string, error)
	EnqueueTaskImplementation(payload *TaskImplementationPayload, delay time.Duration) (string, error)
}

// TaskPlanningPayload represents the payload for task planning jobs
type TaskPlanningPayload struct {
	TaskID     uuid.UUID `json:"task_id"`
	BranchName string    `json:"branch_name"`
	ProjectID  uuid.UUID `json:"project_id"`
}

// TaskImplementationPayload represents the payload for task implementation jobs
type TaskImplementationPayload struct {
	TaskID    uuid.UUID `json:"task_id"`
	ProjectID uuid.UUID `json:"project_id"`
}

type TaskUsecase interface {
	// Basic CRUD operations
	Create(ctx context.Context, req CreateTaskRequest) (*entity.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Task, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateTaskRequest) (*entity.Task, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TaskStatus) (*entity.Task, error)
	UpdateStatusWithHistory(ctx context.Context, req UpdateStatusRequest) (*entity.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByStatus(ctx context.Context, status entity.TaskStatus) ([]*entity.Task, error)
	GetByStatuses(ctx context.Context, statuses []entity.TaskStatus) ([]*entity.Task, error)
	GetWithProject(ctx context.Context, id uuid.UUID) (*entity.Task, error)
	BulkUpdateStatus(ctx context.Context, req BulkUpdateStatusRequest) error
	GetStatusHistory(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskStatusHistory, error)
	GetStatusAnalytics(ctx context.Context, projectID uuid.UUID) (*entity.TaskStatusAnalytics, error)
	GetTasksWithFilters(ctx context.Context, req GetTasksFilterRequest) ([]*entity.Task, error)
	ValidateStatusTransition(ctx context.Context, taskID uuid.UUID, newStatus entity.TaskStatus) error

	// Advanced filtering and search
	SearchTasks(ctx context.Context, query string, projectID *uuid.UUID) ([]*entity.TaskSearchResult, error)
	GetTasksByPriority(ctx context.Context, priority entity.TaskPriority) ([]*entity.Task, error)
	GetTasksByTags(ctx context.Context, tags []string) ([]*entity.Task, error)
	GetArchivedTasks(ctx context.Context, projectID *uuid.UUID) ([]*entity.Task, error)
	GetTasksWithSubtasks(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error)

	// Parent-child relationships
	GetSubtasks(ctx context.Context, parentTaskID uuid.UUID) ([]*entity.Task, error)
	GetParentTask(ctx context.Context, taskID uuid.UUID) (*entity.Task, error)
	UpdateParentTask(ctx context.Context, taskID uuid.UUID, parentTaskID *uuid.UUID) error
	CreateSubtask(ctx context.Context, parentTaskID uuid.UUID, req CreateTaskRequest) (*entity.Task, error)

	// Bulk operations
	BulkDelete(ctx context.Context, taskIDs []uuid.UUID) error
	BulkArchive(ctx context.Context, taskIDs []uuid.UUID) error
	BulkUnarchive(ctx context.Context, taskIDs []uuid.UUID) error
	BulkUpdatePriority(ctx context.Context, taskIDs []uuid.UUID, priority entity.TaskPriority) error
	BulkAssign(ctx context.Context, taskIDs []uuid.UUID, assignedTo string) error

	// Templates
	CreateTemplate(ctx context.Context, req CreateTemplateRequest) (*entity.TaskTemplate, error)
	GetTemplates(ctx context.Context, projectID uuid.UUID, includeGlobal bool) ([]*entity.TaskTemplate, error)
	GetTemplateByID(ctx context.Context, id uuid.UUID) (*entity.TaskTemplate, error)
	UpdateTemplate(ctx context.Context, id uuid.UUID, req UpdateTemplateRequest) (*entity.TaskTemplate, error)
	DeleteTemplate(ctx context.Context, id uuid.UUID) error
	CreateTaskFromTemplate(ctx context.Context, templateID uuid.UUID, projectID uuid.UUID, createdBy string) (*entity.Task, error)

	// Audit trail
	GetAuditLogs(ctx context.Context, taskID uuid.UUID, limit *int) ([]*entity.TaskAuditLog, error)

	// Statistics and analytics
	GetTaskStatistics(ctx context.Context, projectID uuid.UUID) (*entity.TaskStatistics, error)

	// Dependencies
	AddDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID, dependencyType string) error
	RemoveDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) error
	GetDependencies(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskDependency, error)
	GetDependents(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskDependency, error)

	// Comments
	AddComment(ctx context.Context, req AddCommentRequest) (*entity.TaskComment, error)
	GetComments(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskComment, error)
	UpdateComment(ctx context.Context, commentID uuid.UUID, req UpdateCommentRequest) (*entity.TaskComment, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID) error

	// Export functionality
	ExportTasks(ctx context.Context, filters entity.TaskFilters, format entity.TaskExportFormat) ([]byte, error)

	// Validation
	CheckDuplicateTitle(ctx context.Context, projectID uuid.UUID, title string, excludeID *uuid.UUID) (bool, error)

	// Git status management
	UpdateGitStatus(ctx context.Context, taskID uuid.UUID, gitStatus entity.TaskGitStatus) (*entity.Task, error)
	ValidateGitStatusTransition(ctx context.Context, taskID uuid.UUID, newGitStatus entity.TaskGitStatus) error

	// Planning workflow
	StartPlanning(ctx context.Context, taskID uuid.UUID, branchName string) (string, error) // returns job ID
	ApprovePlan(ctx context.Context, taskID uuid.UUID) (string, error) // returns job ID
	ListGitBranches(ctx context.Context, projectID uuid.UUID) ([]GitBranch, error)
}

type CreateTaskRequest struct {
	ProjectID      uuid.UUID           `json:"project_id" binding:"required"`
	Title          string              `json:"title" binding:"required"`
	Description    string              `json:"description"`
	Priority       entity.TaskPriority `json:"priority"`
	EstimatedHours *float64            `json:"estimated_hours"`
	Tags           []string            `json:"tags"`
	ParentTaskID   *uuid.UUID          `json:"parent_task_id"`
	AssignedTo     *string             `json:"assigned_to"`
	DueDate        *time.Time          `json:"due_date"`
	BranchName     *string             `json:"branch_name"`
	PullRequest    *string             `json:"pull_request"`
}

type UpdateTaskRequest struct {
	Title          string               `json:"title"`
	Description    string               `json:"description"`
	Priority       *entity.TaskPriority `json:"priority"`
	EstimatedHours *float64             `json:"estimated_hours"`
	ActualHours    *float64             `json:"actual_hours"`
	Tags           []string             `json:"tags"`
	AssignedTo     *string              `json:"assigned_to"`
	DueDate        *time.Time           `json:"due_date"`
	BranchName     *string              `json:"branch_name"`
	PullRequest    *string              `json:"pull_request"`
	WorktreePath   *string              `json:"worktree_path"`
}

type UpdateStatusRequest struct {
	TaskID    uuid.UUID         `json:"task_id" binding:"required"`
	Status    entity.TaskStatus `json:"status" binding:"required"`
	ChangedBy *string           `json:"changed_by,omitempty"`
	Reason    *string           `json:"reason,omitempty"`
}

type BulkUpdateStatusRequest struct {
	TaskIDs   []uuid.UUID       `json:"task_ids" binding:"required"`
	Status    entity.TaskStatus `json:"status" binding:"required"`
	ChangedBy *string           `json:"changed_by,omitempty"`
}

type GetTasksFilterRequest struct {
	ProjectID      *uuid.UUID
	Statuses       []entity.TaskStatus
	Priorities     []entity.TaskPriority
	Tags           []string
	ParentTaskID   *uuid.UUID
	AssignedTo     *string
	CreatedAfter   *time.Time
	CreatedBefore  *time.Time
	UpdatedAfter   *time.Time
	UpdatedBefore  *time.Time
	DueDateAfter   *time.Time
	DueDateBefore  *time.Time
	SearchTerm     *string
	IsArchived     *bool
	IsTemplate     *bool
	HasSubtasks    *bool
	EstimatedHours *float64
	ActualHours    *float64
	Limit          *int
	Offset         *int
	OrderBy        *string
	OrderDir       *string
}

type CreateTemplateRequest struct {
	ProjectID      uuid.UUID           `json:"project_id" binding:"required"`
	Name           string              `json:"name" binding:"required"`
	Description    string              `json:"description"`
	Title          string              `json:"title" binding:"required"`
	Priority       entity.TaskPriority `json:"priority"`
	EstimatedHours *float64            `json:"estimated_hours"`
	Tags           []string            `json:"tags"`
	IsGlobal       bool                `json:"is_global"`
	CreatedBy      string              `json:"created_by"`
}

type UpdateTemplateRequest struct {
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	Title          string               `json:"title"`
	Priority       *entity.TaskPriority `json:"priority"`
	EstimatedHours *float64             `json:"estimated_hours"`
	Tags           []string             `json:"tags"`
	IsGlobal       *bool                `json:"is_global"`
}

type AddCommentRequest struct {
	TaskID    uuid.UUID `json:"task_id" binding:"required"`
	Comment   string    `json:"comment" binding:"required"`
	CreatedBy string    `json:"created_by" binding:"required"`
}

type UpdateCommentRequest struct {
	Comment string `json:"comment" binding:"required"`
}

type taskUsecase struct {
	taskRepo            repository.TaskRepository
	projectRepo         repository.ProjectRepository
	notificationUsecase NotificationUsecase
	worktreeUsecase     WorktreeUsecase
	jobClient           JobClientInterface
}

func NewTaskUsecase(taskRepo repository.TaskRepository, projectRepo repository.ProjectRepository, notificationUsecase NotificationUsecase, worktreeUsecase WorktreeUsecase, jobClient JobClientInterface) TaskUsecase {
	return &taskUsecase{
		taskRepo:            taskRepo,
		projectRepo:         projectRepo,
		notificationUsecase: notificationUsecase,
		worktreeUsecase:     worktreeUsecase,
		jobClient:           jobClient,
	}
}

func (u *taskUsecase) Create(ctx context.Context, req CreateTaskRequest) (*entity.Task, error) {
	// Validate project exists
	if exists, err := u.taskRepo.ValidateProjectExists(ctx, req.ProjectID); err != nil {
		return nil, fmt.Errorf("failed to validate project: %w", err)
	} else if !exists {
		return nil, fmt.Errorf("project not found")
	}

	// Check for duplicate title
	if isDuplicate, err := u.taskRepo.CheckDuplicateTitle(ctx, req.ProjectID, req.Title, nil); err != nil {
		return nil, fmt.Errorf("failed to check duplicate title: %w", err)
	} else if isDuplicate {
		return nil, fmt.Errorf("task with title '%s' already exists in this project", req.Title)
	}

	// Validate parent task if provided
	if req.ParentTaskID != nil {
		if exists, err := u.taskRepo.ValidateTaskExists(ctx, *req.ParentTaskID); err != nil {
			return nil, fmt.Errorf("failed to validate parent task: %w", err)
		} else if !exists {
			return nil, fmt.Errorf("parent task not found")
		}
	}

	// Set default priority if not provided
	if req.Priority == "" {
		req.Priority = entity.TaskPriorityMedium
	}

	task := &entity.Task{
		ID:             uuid.New(),
		ProjectID:      req.ProjectID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         entity.TaskStatusTODO,
		Priority:       req.Priority,
		EstimatedHours: req.EstimatedHours,
		Tags:           req.Tags,
		ParentTaskID:   req.ParentTaskID,
		AssignedTo:     req.AssignedTo,
		DueDate:        req.DueDate,
		BranchName:     req.BranchName,
		PullRequest:    req.PullRequest,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := u.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	// Send task created notification
	if u.notificationUsecase != nil {
		project, err := u.projectRepo.GetByID(ctx, task.ProjectID)
		if err == nil {
			// Don't fail task creation if notification fails
			_ = u.notificationUsecase.SendTaskCreatedNotification(ctx, task, project)
		}
	}

	return task, nil
}

func (u *taskUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Task, error) {
	return u.taskRepo.GetByID(ctx, id)
}

func (u *taskUsecase) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error) {
	return u.taskRepo.GetByProjectID(ctx, projectID)
}

func (u *taskUsecase) Update(ctx context.Context, id uuid.UUID, req UpdateTaskRequest) (*entity.Task, error) {
	task, err := u.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check for duplicate title if title is being changed
	if req.Title != "" && req.Title != task.Title {
		if isDuplicate, err := u.taskRepo.CheckDuplicateTitle(ctx, task.ProjectID, req.Title, &id); err != nil {
			return nil, fmt.Errorf("failed to check duplicate title: %w", err)
		} else if isDuplicate {
			return nil, fmt.Errorf("task with title '%s' already exists in this project", req.Title)
		}
		task.Title = req.Title
	}

	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.EstimatedHours != nil {
		task.EstimatedHours = req.EstimatedHours
	}
	if req.ActualHours != nil {
		task.ActualHours = req.ActualHours
	}
	if req.Tags != nil {
		task.Tags = req.Tags
	}
	if req.AssignedTo != nil {
		task.AssignedTo = req.AssignedTo
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}
	if req.BranchName != nil {
		task.BranchName = req.BranchName
	}
	if req.PullRequest != nil {
		task.PullRequest = req.PullRequest
	}

	task.UpdatedAt = time.Now()
	if req.WorktreePath != nil {
		task.WorktreePath = req.WorktreePath
	}

	task.UpdatedAt = time.Now()

	if err := u.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (u *taskUsecase) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TaskStatus) (*entity.Task, error) {
	if err := u.taskRepo.UpdateStatus(ctx, id, status); err != nil {
		return nil, err
	}

	return u.taskRepo.GetByID(ctx, id)
}

func (u *taskUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.taskRepo.Delete(ctx, id)
}

func (u *taskUsecase) GetWithProject(ctx context.Context, id uuid.UUID) (*entity.Task, error) {
	task, err := u.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// The repository should handle loading project via GORM preloading
	// For now, we'll return the task as-is since the relationship is defined
	return task, nil
}

func (u *taskUsecase) GetByStatus(ctx context.Context, status entity.TaskStatus) ([]*entity.Task, error) {
	return u.taskRepo.GetByStatus(ctx, status)
}

// UpdateStatusWithHistory updates task status with validation and history tracking
func (u *taskUsecase) UpdateStatusWithHistory(ctx context.Context, req UpdateStatusRequest) (*entity.Task, error) {
	// Validate the status transition first
	if err := u.ValidateStatusTransition(ctx, req.TaskID, req.Status); err != nil {
		return nil, err
	}

	// Update status with history
	if err := u.taskRepo.UpdateStatusWithHistory(ctx, req.TaskID, req.Status, req.ChangedBy, req.Reason); err != nil {
		return nil, err
	}

	// Get updated task
	updatedTask, err := u.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		return nil, err
	}

	// Handle worktree operations based on status change
	if u.worktreeUsecase != nil {
		if err := u.handleWorktreeOperations(ctx, updatedTask, req.Status); err != nil {
			// Log error but don't fail the status update
			// TODO: Add proper logging here
		}
	}

	// Send status change notification
	if u.notificationUsecase != nil {
		project, err := u.projectRepo.GetByID(ctx, updatedTask.ProjectID)
		if err == nil {
			notificationData := entity.TaskStatusChangeNotificationData{
				TaskID:      req.TaskID,
				TaskTitle:   updatedTask.Title,
				FromStatus:  &updatedTask.Status,
				ToStatus:    req.Status,
				ChangedBy:   req.ChangedBy,
				Reason:      req.Reason,
				ProjectID:   updatedTask.ProjectID,
				ProjectName: project.Name,
			}
			// Don't fail status update if notification fails
			_ = u.notificationUsecase.SendTaskStatusChangeNotification(ctx, notificationData)
		}
	}

	return updatedTask, nil
}

// GetByStatuses retrieves tasks with multiple statuses
func (u *taskUsecase) GetByStatuses(ctx context.Context, statuses []entity.TaskStatus) ([]*entity.Task, error) {
	// Validate all statuses
	for _, status := range statuses {
		if !status.IsValid() {
			return nil, fmt.Errorf("invalid status: %s", status)
		}
	}

	return u.taskRepo.GetByStatuses(ctx, statuses)
}

// BulkUpdateStatus updates multiple tasks to the same status
func (u *taskUsecase) BulkUpdateStatus(ctx context.Context, req BulkUpdateStatusRequest) error {
	if len(req.TaskIDs) == 0 {
		return fmt.Errorf("no task IDs provided")
	}

	// Validate target status
	if !req.Status.IsValid() {
		return fmt.Errorf("invalid target status: %s", req.Status)
	}

	// This will validate transitions for each task individually in the repository
	return u.taskRepo.BulkUpdateStatus(ctx, req.TaskIDs, req.Status, req.ChangedBy)
}

// GetStatusHistory retrieves status change history for a task
func (u *taskUsecase) GetStatusHistory(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskStatusHistory, error) {
	// Verify task exists
	if _, err := u.taskRepo.GetByID(ctx, taskID); err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	return u.taskRepo.GetStatusHistory(ctx, taskID)
}

// GetStatusAnalytics generates comprehensive status analytics for a project
func (u *taskUsecase) GetStatusAnalytics(ctx context.Context, projectID uuid.UUID) (*entity.TaskStatusAnalytics, error) {
	return u.taskRepo.GetStatusAnalytics(ctx, projectID)
}

// GetTasksWithFilters retrieves tasks with various filtering options
func (u *taskUsecase) GetTasksWithFilters(ctx context.Context, req GetTasksFilterRequest) ([]*entity.Task, error) {
	// Validate statuses if provided
	for _, status := range req.Statuses {
		if !status.IsValid() {
			return nil, fmt.Errorf("invalid status filter: %s", status)
		}
	}

	// Validate priorities if provided
	for _, priority := range req.Priorities {
		if !priority.IsValid() {
			return nil, fmt.Errorf("invalid priority filter: %s", priority)
		}
	}

	// Convert to entity filters
	filters := entity.TaskFilters{
		ProjectID:      req.ProjectID,
		Statuses:       req.Statuses,
		Priorities:     req.Priorities,
		Tags:           req.Tags,
		ParentTaskID:   req.ParentTaskID,
		AssignedTo:     req.AssignedTo,
		CreatedAfter:   req.CreatedAfter,
		CreatedBefore:  req.CreatedBefore,
		UpdatedAfter:   req.UpdatedAfter,
		UpdatedBefore:  req.UpdatedBefore,
		DueDateAfter:   req.DueDateAfter,
		DueDateBefore:  req.DueDateBefore,
		SearchTerm:     req.SearchTerm,
		IsArchived:     req.IsArchived,
		IsTemplate:     req.IsTemplate,
		HasSubtasks:    req.HasSubtasks,
		EstimatedHours: req.EstimatedHours,
		ActualHours:    req.ActualHours,
		Limit:          req.Limit,
		Offset:         req.Offset,
		OrderBy:        req.OrderBy,
		OrderDir:       req.OrderDir,
	}

	return u.taskRepo.GetTasksWithFilters(ctx, filters)
}

// ValidateStatusTransition validates if a status transition is allowed for a specific task
func (u *taskUsecase) ValidateStatusTransition(ctx context.Context, taskID uuid.UUID, newStatus entity.TaskStatus) error {
	// Get current task
	task, err := u.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Validate transition using entity logic
	return entity.ValidateStatusTransition(task.Status, newStatus)
}

// SearchTasks performs full-text search on tasks
func (u *taskUsecase) SearchTasks(ctx context.Context, query string, projectID *uuid.UUID) ([]*entity.TaskSearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	return u.taskRepo.SearchTasks(ctx, query, projectID)
}

// GetTasksByPriority retrieves tasks by priority level
func (u *taskUsecase) GetTasksByPriority(ctx context.Context, priority entity.TaskPriority) ([]*entity.Task, error) {
	if !priority.IsValid() {
		return nil, fmt.Errorf("invalid priority: %s", priority)
	}

	return u.taskRepo.GetTasksByPriority(ctx, priority)
}

// GetTasksByTags retrieves tasks that have any of the specified tags
func (u *taskUsecase) GetTasksByTags(ctx context.Context, tags []string) ([]*entity.Task, error) {
	if len(tags) == 0 {
		return nil, fmt.Errorf("at least one tag must be provided")
	}

	return u.taskRepo.GetTasksByTags(ctx, tags)
}

// GetArchivedTasks retrieves archived tasks
func (u *taskUsecase) GetArchivedTasks(ctx context.Context, projectID *uuid.UUID) ([]*entity.Task, error) {
	return u.taskRepo.GetArchivedTasks(ctx, projectID)
}

// GetTasksWithSubtasks retrieves tasks with their subtasks
func (u *taskUsecase) GetTasksWithSubtasks(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error) {
	return u.taskRepo.GetTasksWithSubtasks(ctx, projectID)
}

// GetSubtasks retrieves all subtasks of a parent task
func (u *taskUsecase) GetSubtasks(ctx context.Context, parentTaskID uuid.UUID) ([]*entity.Task, error) {
	return u.taskRepo.GetSubtasks(ctx, parentTaskID)
}

// GetParentTask retrieves the parent task of a subtask
func (u *taskUsecase) GetParentTask(ctx context.Context, taskID uuid.UUID) (*entity.Task, error) {
	return u.taskRepo.GetParentTask(ctx, taskID)
}

// UpdateParentTask updates the parent task relationship
func (u *taskUsecase) UpdateParentTask(ctx context.Context, taskID uuid.UUID, parentTaskID *uuid.UUID) error {
	// Validate task exists
	if exists, err := u.taskRepo.ValidateTaskExists(ctx, taskID); err != nil {
		return fmt.Errorf("failed to validate task: %w", err)
	} else if !exists {
		return fmt.Errorf("task not found")
	}

	// Validate parent task if provided
	if parentTaskID != nil {
		if exists, err := u.taskRepo.ValidateTaskExists(ctx, *parentTaskID); err != nil {
			return fmt.Errorf("failed to validate parent task: %w", err)
		} else if !exists {
			return fmt.Errorf("parent task not found")
		}
	}

	return u.taskRepo.UpdateParentTask(ctx, taskID, parentTaskID)
}

// CreateSubtask creates a new subtask
func (u *taskUsecase) CreateSubtask(ctx context.Context, parentTaskID uuid.UUID, req CreateTaskRequest) (*entity.Task, error) {
	// Validate parent task exists
	if exists, err := u.taskRepo.ValidateTaskExists(ctx, parentTaskID); err != nil {
		return nil, fmt.Errorf("failed to validate parent task: %w", err)
	} else if !exists {
		return nil, fmt.Errorf("parent task not found")
	}

	// Set parent task ID
	req.ParentTaskID = &parentTaskID

	return u.Create(ctx, req)
}

// BulkDelete deletes multiple tasks
func (u *taskUsecase) BulkDelete(ctx context.Context, taskIDs []uuid.UUID) error {
	if len(taskIDs) == 0 {
		return fmt.Errorf("no task IDs provided")
	}

	return u.taskRepo.BulkDelete(ctx, taskIDs)
}

// BulkArchive archives multiple tasks
func (u *taskUsecase) BulkArchive(ctx context.Context, taskIDs []uuid.UUID) error {
	if len(taskIDs) == 0 {
		return fmt.Errorf("no task IDs provided")
	}

	return u.taskRepo.BulkArchive(ctx, taskIDs)
}

// BulkUnarchive unarchives multiple tasks
func (u *taskUsecase) BulkUnarchive(ctx context.Context, taskIDs []uuid.UUID) error {
	if len(taskIDs) == 0 {
		return fmt.Errorf("no task IDs provided")
	}

	return u.taskRepo.BulkUnarchive(ctx, taskIDs)
}

// BulkUpdatePriority updates priority for multiple tasks
func (u *taskUsecase) BulkUpdatePriority(ctx context.Context, taskIDs []uuid.UUID, priority entity.TaskPriority) error {
	if len(taskIDs) == 0 {
		return fmt.Errorf("no task IDs provided")
	}

	if !priority.IsValid() {
		return fmt.Errorf("invalid priority: %s", priority)
	}

	return u.taskRepo.BulkUpdatePriority(ctx, taskIDs, priority)
}

// BulkAssign assigns multiple tasks to a user
func (u *taskUsecase) BulkAssign(ctx context.Context, taskIDs []uuid.UUID, assignedTo string) error {
	if len(taskIDs) == 0 {
		return fmt.Errorf("no task IDs provided")
	}

	if strings.TrimSpace(assignedTo) == "" {
		return fmt.Errorf("assigned_to cannot be empty")
	}

	return u.taskRepo.BulkAssign(ctx, taskIDs, assignedTo)
}

// CreateTemplate creates a new task template
func (u *taskUsecase) CreateTemplate(ctx context.Context, req CreateTemplateRequest) (*entity.TaskTemplate, error) {
	// Validate project exists
	if exists, err := u.taskRepo.ValidateProjectExists(ctx, req.ProjectID); err != nil {
		return nil, fmt.Errorf("failed to validate project: %w", err)
	} else if !exists {
		return nil, fmt.Errorf("project not found")
	}

	template := &entity.TaskTemplate{
		ID:             uuid.New(),
		ProjectID:      req.ProjectID,
		Name:           req.Name,
		Description:    req.Description,
		Title:          req.Title,
		Priority:       req.Priority,
		EstimatedHours: req.EstimatedHours,
		Tags:           req.Tags,
		IsGlobal:       req.IsGlobal,
		CreatedBy:      &req.CreatedBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := u.taskRepo.CreateTemplate(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// GetTemplates retrieves task templates
func (u *taskUsecase) GetTemplates(ctx context.Context, projectID uuid.UUID, includeGlobal bool) ([]*entity.TaskTemplate, error) {
	return u.taskRepo.GetTemplates(ctx, projectID, includeGlobal)
}

// GetTemplateByID retrieves a specific template
func (u *taskUsecase) GetTemplateByID(ctx context.Context, id uuid.UUID) (*entity.TaskTemplate, error) {
	return u.taskRepo.GetTemplateByID(ctx, id)
}

// UpdateTemplate updates a task template
func (u *taskUsecase) UpdateTemplate(ctx context.Context, id uuid.UUID, req UpdateTemplateRequest) (*entity.TaskTemplate, error) {
	template, err := u.taskRepo.GetTemplateByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.Title != "" {
		template.Title = req.Title
	}
	if req.Priority != nil {
		template.Priority = *req.Priority
	}
	if req.EstimatedHours != nil {
		template.EstimatedHours = req.EstimatedHours
	}
	if req.Tags != nil {
		template.Tags = req.Tags
	}
	if req.IsGlobal != nil {
		template.IsGlobal = *req.IsGlobal
	}

	template.UpdatedAt = time.Now()

	if err := u.taskRepo.UpdateTemplate(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// DeleteTemplate deletes a task template
func (u *taskUsecase) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return u.taskRepo.DeleteTemplate(ctx, id)
}

// CreateTaskFromTemplate creates a new task from a template
func (u *taskUsecase) CreateTaskFromTemplate(ctx context.Context, templateID uuid.UUID, projectID uuid.UUID, createdBy string) (*entity.Task, error) {
	return u.taskRepo.CreateTaskFromTemplate(ctx, templateID, projectID, createdBy)
}

// GetAuditLogs retrieves audit logs for a task
func (u *taskUsecase) GetAuditLogs(ctx context.Context, taskID uuid.UUID, limit *int) ([]*entity.TaskAuditLog, error) {
	// Verify task exists
	if _, err := u.taskRepo.GetByID(ctx, taskID); err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	return u.taskRepo.GetAuditLogs(ctx, taskID, limit)
}

// GetTaskStatistics retrieves comprehensive task statistics
func (u *taskUsecase) GetTaskStatistics(ctx context.Context, projectID uuid.UUID) (*entity.TaskStatistics, error) {
	return u.taskRepo.GetTaskStatistics(ctx, projectID)
}

// AddDependency adds a dependency between tasks
func (u *taskUsecase) AddDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID, dependencyType string) error {
	// Validate both tasks exist
	if exists, err := u.taskRepo.ValidateTaskExists(ctx, taskID); err != nil {
		return fmt.Errorf("failed to validate task: %w", err)
	} else if !exists {
		return fmt.Errorf("task not found")
	}

	if exists, err := u.taskRepo.ValidateTaskExists(ctx, dependsOnTaskID); err != nil {
		return fmt.Errorf("failed to validate dependency task: %w", err)
	} else if !exists {
		return fmt.Errorf("dependency task not found")
	}

	// Validate dependency type
	validTypes := []string{"blocks", "requires", "related"}
	isValid := false
	for _, validType := range validTypes {
		if dependencyType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid dependency type: %s", dependencyType)
	}

	return u.taskRepo.AddDependency(ctx, taskID, dependsOnTaskID, dependencyType)
}

// RemoveDependency removes a dependency between tasks
func (u *taskUsecase) RemoveDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) error {
	return u.taskRepo.RemoveDependency(ctx, taskID, dependsOnTaskID)
}

// GetDependencies retrieves dependencies for a task
func (u *taskUsecase) GetDependencies(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskDependency, error) {
	return u.taskRepo.GetDependencies(ctx, taskID)
}

// GetDependents retrieves tasks that depend on the given task
func (u *taskUsecase) GetDependents(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskDependency, error) {
	return u.taskRepo.GetDependents(ctx, taskID)
}

// AddComment adds a comment to a task
func (u *taskUsecase) AddComment(ctx context.Context, req AddCommentRequest) (*entity.TaskComment, error) {
	// Validate task exists
	if exists, err := u.taskRepo.ValidateTaskExists(ctx, req.TaskID); err != nil {
		return nil, fmt.Errorf("failed to validate task: %w", err)
	} else if !exists {
		return nil, fmt.Errorf("task not found")
	}

	comment := &entity.TaskComment{
		ID:        uuid.New(),
		TaskID:    req.TaskID,
		Comment:   req.Comment,
		CreatedBy: req.CreatedBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := u.taskRepo.AddComment(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

// GetComments retrieves comments for a task
func (u *taskUsecase) GetComments(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskComment, error) {
	return u.taskRepo.GetComments(ctx, taskID)
}

// UpdateComment updates a comment
func (u *taskUsecase) UpdateComment(ctx context.Context, commentID uuid.UUID, req UpdateCommentRequest) (*entity.TaskComment, error) {
	// Get existing comment
	comments, err := u.taskRepo.GetComments(ctx, uuid.Nil) // We need to get the comment by ID, but the interface doesn't support it yet
	if err != nil {
		return nil, err
	}

	// Find the comment (this is a temporary workaround)
	var comment *entity.TaskComment
	for _, c := range comments {
		if c.ID == commentID {
			comment = c
			break
		}
	}

	if comment == nil {
		return nil, fmt.Errorf("comment not found")
	}

	comment.Comment = req.Comment
	comment.UpdatedAt = time.Now()

	if err := u.taskRepo.UpdateComment(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

// DeleteComment deletes a comment
func (u *taskUsecase) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	return u.taskRepo.DeleteComment(ctx, commentID)
}

// ExportTasks exports tasks in the specified format
func (u *taskUsecase) ExportTasks(ctx context.Context, filters entity.TaskFilters, format entity.TaskExportFormat) ([]byte, error) {
	return u.taskRepo.ExportTasks(ctx, filters, format)
}

// CheckDuplicateTitle checks if a task title already exists in a project
func (u *taskUsecase) CheckDuplicateTitle(ctx context.Context, projectID uuid.UUID, title string, excludeID *uuid.UUID) (bool, error) {
	return u.taskRepo.CheckDuplicateTitle(ctx, projectID, title, excludeID)
}

// handleWorktreeOperations handles worktree operations based on task status changes
func (u *taskUsecase) handleWorktreeOperations(ctx context.Context, task *entity.Task, newStatus entity.TaskStatus) error {
	switch newStatus {
	case entity.TaskStatusIMPLEMENTING:
		// Create worktree when task moves to IMPLEMENTING
		return u.createWorktreeForTask(ctx, task)
	case entity.TaskStatusDONE:
		// Complete worktree when task is done
		return u.completeWorktreeForTask(ctx, task)
	case entity.TaskStatusCANCELLED:
		// Cleanup worktree when task is cancelled
		return u.cleanupWorktreeForTask(ctx, task)
	default:
		// No worktree operations needed for other statuses
		return nil
	}
}

// createWorktreeForTask creates a worktree for a task
func (u *taskUsecase) createWorktreeForTask(ctx context.Context, task *entity.Task) error {
	// Check if worktree already exists
	existingWorktree, err := u.worktreeUsecase.GetWorktreeByTaskID(ctx, task.ID)
	if err == nil && existingWorktree != nil {
		// Worktree already exists, update Git status to active if needed
		if task.GitStatus != entity.TaskGitStatusActive {
			return u.updateTaskGitStatus(ctx, task, entity.TaskGitStatusActive)
		}
		return nil
	}

	// Update Git status to creating
	if err := u.updateTaskGitStatus(ctx, task, entity.TaskGitStatusCreating); err != nil {
		return fmt.Errorf("failed to update Git status: %w", err)
	}

	// Create worktree
	_, err = u.worktreeUsecase.CreateWorktreeForTask(ctx, CreateWorktreeRequest{
		TaskID:    task.ID,
		ProjectID: task.ProjectID,
		TaskTitle: task.Title,
	})
	if err != nil {
		// Update Git status to error if creation fails
		u.updateTaskGitStatus(ctx, task, entity.TaskGitStatusError)
		return err
	}

	// Git status will be updated to active by the worktree usecase
	return nil
}

// completeWorktreeForTask marks a worktree as completed for a task
func (u *taskUsecase) completeWorktreeForTask(ctx context.Context, task *entity.Task) error {
	// Check if worktree exists
	existingWorktree, err := u.worktreeUsecase.GetWorktreeByTaskID(ctx, task.ID)
	if err != nil || existingWorktree == nil {
		// No worktree to complete
		return nil
	}

	// Update Git status to completed
	if err := u.updateTaskGitStatus(ctx, task, entity.TaskGitStatusCompleted); err != nil {
		return fmt.Errorf("failed to update Git status: %w", err)
	}

	// Update worktree status to completed
	return u.worktreeUsecase.UpdateWorktreeStatus(ctx, existingWorktree.ID, entity.WorktreeStatusCompleted)
}

// cleanupWorktreeForTask cleans up worktree for a task
func (u *taskUsecase) cleanupWorktreeForTask(ctx context.Context, task *entity.Task) error {
	// Check if worktree exists
	existingWorktree, err := u.worktreeUsecase.GetWorktreeByTaskID(ctx, task.ID)
	if err != nil || existingWorktree == nil {
		// No worktree to cleanup
		return nil
	}

	// Update Git status to cleaning
	if err := u.updateTaskGitStatus(ctx, task, entity.TaskGitStatusCleaning); err != nil {
		return fmt.Errorf("failed to update Git status: %w", err)
	}

	// Cleanup worktree
	err = u.worktreeUsecase.CleanupWorktreeForTask(ctx, CleanupWorktreeRequest{
		TaskID:    task.ID,
		ProjectID: task.ProjectID,
		Force:     true, // Force cleanup for completed/cancelled tasks
	})
	if err != nil {
		// Update Git status to error if cleanup fails
		u.updateTaskGitStatus(ctx, task, entity.TaskGitStatusError)
		return err
	}

	// Git status will be updated to none by the worktree usecase
	return nil
}

// updateTaskGitStatus updates the Git status of a task with validation
func (u *taskUsecase) updateTaskGitStatus(ctx context.Context, task *entity.Task, newGitStatus entity.TaskGitStatus) error {
	// Validate Git status transition
	if err := entity.ValidateGitStatusTransition(task.GitStatus, newGitStatus); err != nil {
		return fmt.Errorf("invalid Git status transition: %w", err)
	}

	// Update task Git status
	task.GitStatus = newGitStatus
	return u.taskRepo.Update(ctx, task)
}

// UpdateGitStatus updates the Git status of a task
func (u *taskUsecase) UpdateGitStatus(ctx context.Context, taskID uuid.UUID, gitStatus entity.TaskGitStatus) (*entity.Task, error) {
	// Get current task
	task, err := u.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Update Git status with validation
	if err := u.updateTaskGitStatus(ctx, task, gitStatus); err != nil {
		return nil, err
	}

	// Return updated task
	return u.taskRepo.GetByID(ctx, taskID)
}

// ValidateGitStatusTransition validates if a Git status transition is allowed for a specific task
func (u *taskUsecase) ValidateGitStatusTransition(ctx context.Context, taskID uuid.UUID, newGitStatus entity.TaskGitStatus) error {
	// Get current task
	task, err := u.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Validate transition using entity logic
	return entity.ValidateGitStatusTransition(task.GitStatus, newGitStatus)
}

// StartPlanning starts the planning process for a task
func (u *taskUsecase) StartPlanning(ctx context.Context, taskID uuid.UUID, branchName string) (string, error) {
	// Get task to validate it exists and is in TODO status
	task, err := u.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("failed to get task: %w", err)
	}

	if task.Status != entity.TaskStatusTODO {
		return "", fmt.Errorf("task must be in TODO status to start planning, current status: %s", task.Status)
	}

	// Update task with branch name
	_, err = u.Update(ctx, taskID, UpdateTaskRequest{
		BranchName: &branchName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to update task with branch name: %w", err)
	}

	// Enqueue the planning job using asynq client
	payload := &TaskPlanningPayload{
		TaskID:     taskID,
		BranchName: branchName,
		ProjectID:  task.ProjectID,
	}

	jobID, err := u.jobClient.EnqueueTaskPlanning(payload, 0)
	if err != nil {
		return "", fmt.Errorf("failed to enqueue planning job: %w", err)
	}

	return jobID, nil
}

// ApprovePlan approves the plan for a task and starts implementation
func (u *taskUsecase) ApprovePlan(ctx context.Context, taskID uuid.UUID) (string, error) {
	// Get task to validate it exists and is in PLAN_REVIEWING status
	task, err := u.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("failed to get task: %w", err)
	}

	if task.Status != entity.TaskStatusPLANREVIEWING {
		return "", fmt.Errorf("task must be in PLAN_REVIEWING status to approve plan, current status: %s", task.Status)
	}

	// Note: Status update to IMPLEMENTING is now handled by the WebSocket handler
	// to provide immediate UI feedback with WebSocket notifications
	
	// Enqueue the implementation job using asynq client
	payload := &TaskImplementationPayload{
		TaskID:    taskID,
		ProjectID: task.ProjectID,
	}

	jobID, err := u.jobClient.EnqueueTaskImplementation(payload, 0)
	if err != nil {
		return "", fmt.Errorf("failed to enqueue implementation job: %w", err)
	}

	return jobID, nil
}

// ListGitBranches lists all Git branches for a project (delegated to project usecase)
func (u *taskUsecase) ListGitBranches(ctx context.Context, projectID uuid.UUID) ([]GitBranch, error) {
	// This is a bit awkward - we'd need project usecase here
	// For now, return empty list as this will be handled by project usecase
	return []GitBranch{}, fmt.Errorf("method should be called on project usecase instead")
}
