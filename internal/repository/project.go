package repository

import (
	"context"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *entity.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error)
	GetAll(ctx context.Context) ([]*entity.Project, error)
	Update(ctx context.Context, project *entity.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetWithTaskCount(ctx context.Context, id uuid.UUID) (*ProjectWithTaskCount, error)
}

type ProjectWithTaskCount struct {
	*entity.Project
	TaskCount int `json:"task_count"`
}