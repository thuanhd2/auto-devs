package repository

import (
	"context"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

type TaskRepository interface {
	Create(ctx context.Context, task *entity.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Task, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error)
	Update(ctx context.Context, task *entity.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TaskStatus) error
	GetByStatus(ctx context.Context, status entity.TaskStatus) ([]*entity.Task, error)
}