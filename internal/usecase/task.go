package usecase

import (
	"context"
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
	Delete(ctx context.Context, id uuid.UUID) error
	GetByStatus(ctx context.Context, status entity.TaskStatus) ([]*entity.Task, error)
	GetWithProject(ctx context.Context, id uuid.UUID) (*entity.Task, error)
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

type taskUsecase struct {
	taskRepo repository.TaskRepository
}

func NewTaskUsecase(taskRepo repository.TaskRepository) TaskUsecase {
	return &taskUsecase{
		taskRepo: taskRepo,
	}
}

func (u *taskUsecase) Create(ctx context.Context, req CreateTaskRequest) (*entity.Task, error) {
	task := &entity.Task{
		ID:          uuid.New(),
		ProjectID:   req.ProjectID,
		Title:       req.Title,
		Description: req.Description,
		Status:      entity.TaskStatusTodo,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.taskRepo.Create(ctx, task); err != nil {
		return nil, err
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