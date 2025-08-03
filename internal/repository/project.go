package repository

import (
	"context"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *entity.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error)
	GetAll(ctx context.Context) ([]*entity.Project, error)
	GetAllWithParams(ctx context.Context, params GetProjectsParams) ([]*entity.Project, int, error)
	Update(ctx context.Context, project *entity.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetWithTaskCount(ctx context.Context, id uuid.UUID) (*ProjectWithTaskCount, error)
	GetTaskStatistics(ctx context.Context, projectID uuid.UUID) (map[entity.TaskStatus]int, error)
	GetLastActivityAt(ctx context.Context, projectID uuid.UUID) (*time.Time, error)
	Archive(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
	CheckNameExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error)
	GetSettings(ctx context.Context, projectID uuid.UUID) (*entity.ProjectSettings, error)
	CreateSettings(ctx context.Context, settings *entity.ProjectSettings) error
	UpdateSettings(ctx context.Context, settings *entity.ProjectSettings) error
}

type ProjectWithTaskCount struct {
	*entity.Project
	TaskCount int `json:"task_count"`
}

type GetProjectsParams struct {
	Search    string
	SortBy    string // name, created_at, task_count
	SortOrder string // asc, desc
	Page      int
	PageSize  int
	Archived  *bool
}