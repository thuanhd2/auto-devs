package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/google/uuid"
)

type TaskUsecase interface {
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
}

type CreateTaskRequest struct {
	ProjectID   uuid.UUID `json:"project_id" binding:"required"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	BranchName  string `json:"branch_name"`
	PullRequest string `json:"pull_request"`
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
	ProjectID     *uuid.UUID          `json:"project_id,omitempty"`
	Statuses      []entity.TaskStatus `json:"statuses,omitempty"`
	CreatedAfter  *time.Time          `json:"created_after,omitempty"`
	CreatedBefore *time.Time          `json:"created_before,omitempty"`
	SearchTerm    *string             `json:"search_term,omitempty"`
	Limit         *int                `json:"limit,omitempty"`
	Offset        *int                `json:"offset,omitempty"`
	OrderBy       *string             `json:"order_by,omitempty"`
	OrderDir      *string             `json:"order_dir,omitempty"`
}

type taskUsecase struct {
	taskRepo            repository.TaskRepository
	projectRepo         repository.ProjectRepository
	notificationUsecase NotificationUsecase
}

func NewTaskUsecase(taskRepo repository.TaskRepository, projectRepo repository.ProjectRepository, notificationUsecase NotificationUsecase) TaskUsecase {
	return &taskUsecase{
		taskRepo:            taskRepo,
		projectRepo:         projectRepo,
		notificationUsecase: notificationUsecase,
	}
}

func (u *taskUsecase) Create(ctx context.Context, req CreateTaskRequest) (*entity.Task, error) {
	task := &entity.Task{
		ID:          uuid.New(),
		ProjectID:   req.ProjectID,
		Title:       req.Title,
		Description: req.Description,
		Status:      entity.TaskStatusTODO,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.BranchName != "" {
		task.BranchName = &req.BranchName
	}
	if req.PullRequest != "" {
		task.PullRequest = &req.PullRequest
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

	// Convert to repository filters
	filters := repository.TaskFilters{
		ProjectID:     req.ProjectID,
		Statuses:      req.Statuses,
		CreatedAfter:  req.CreatedAfter,
		CreatedBefore: req.CreatedBefore,
		SearchTerm:    req.SearchTerm,
		Limit:         req.Limit,
		Offset:        req.Offset,
		OrderBy:       req.OrderBy,
		OrderDir:      req.OrderDir,
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
