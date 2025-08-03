package usecase

import (
	"context"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/google/uuid"
)

type ProjectUsecase interface {
	Create(ctx context.Context, req CreateProjectRequest) (*entity.Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error)
	GetAll(ctx context.Context) ([]*entity.Project, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateProjectRequest) (*entity.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetWithTasks(ctx context.Context, id uuid.UUID) (*entity.Project, error)
}

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	RepoURL     string `json:"repo_url" binding:"required"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RepoURL     string `json:"repo_url"`
}

type projectUsecase struct {
	projectRepo repository.ProjectRepository
}

func NewProjectUsecase(projectRepo repository.ProjectRepository) ProjectUsecase {
	return &projectUsecase{
		projectRepo: projectRepo,
	}
}

func (u *projectUsecase) Create(ctx context.Context, req CreateProjectRequest) (*entity.Project, error) {
	project := &entity.Project{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		RepoURL:     req.RepoURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (u *projectUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	return u.projectRepo.GetByID(ctx, id)
}

func (u *projectUsecase) GetAll(ctx context.Context) ([]*entity.Project, error) {
	return u.projectRepo.GetAll(ctx)
}

func (u *projectUsecase) Update(ctx context.Context, id uuid.UUID, req UpdateProjectRequest) (*entity.Project, error) {
	project, err := u.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.RepoURL != "" {
		project.RepoURL = req.RepoURL
	}
	project.UpdatedAt = time.Now()

	if err := u.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (u *projectUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.projectRepo.Delete(ctx, id)
}

func (u *projectUsecase) GetWithTasks(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	project, err := u.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// The repository should handle loading tasks via GORM preloading
	// For now, we'll return the project as-is since the relationship is defined
	return project, nil
}