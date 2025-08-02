package postgres

import (
	"context"
	"fmt"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectRepository struct {
	db *database.GormDB
}

// NewProjectRepository creates a new PostgreSQL project repository
func NewProjectRepository(db *database.GormDB) repository.ProjectRepository {
	return &projectRepository{db: db}
}

// Create creates a new project
func (r *projectRepository) Create(ctx context.Context, project *entity.Project) error {
	// Generate UUID if not provided
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}

	result := r.db.WithContext(ctx).Create(project)
	if result.Error != nil {
		return fmt.Errorf("failed to create project: %w", result.Error)
	}

	return nil
}

// GetByID retrieves a project by ID
func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	var project entity.Project

	result := r.db.WithContext(ctx).First(&project, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get project: %w", result.Error)
	}

	return &project, nil
}

// GetAll retrieves all projects
func (r *projectRepository) GetAll(ctx context.Context) ([]*entity.Project, error) {
	var projects []entity.Project

	result := r.db.WithContext(ctx).Order("created_at DESC").Find(&projects)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get projects: %w", result.Error)
	}

	// Convert to slice of pointers
	projectPtrs := make([]*entity.Project, len(projects))
	for i := range projects {
		projectPtrs[i] = &projects[i]
	}

	return projectPtrs, nil
}

// Update updates an existing project
func (r *projectRepository) Update(ctx context.Context, project *entity.Project) error {
	result := r.db.WithContext(ctx).Save(project)
	if result.Error != nil {
		return fmt.Errorf("failed to update project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("project not found with id %s", project.ID)
	}

	return nil
}

// Delete deletes a project by ID (soft delete)
func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Project{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("project not found with id %s", id)
	}

	return nil
}

// GetWithTaskCount retrieves a project with its task count
func (r *projectRepository) GetWithTaskCount(ctx context.Context, id uuid.UUID) (*repository.ProjectWithTaskCount, error) {
	var project entity.Project
	var taskCount int64

	// Get project
	result := r.db.WithContext(ctx).First(&project, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get project: %w", result.Error)
	}

	// Get task count
	result = r.db.WithContext(ctx).Model(&entity.Task{}).Where("project_id = ?", id).Count(&taskCount)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get task count: %w", result.Error)
	}

	return &repository.ProjectWithTaskCount{
		Project:   &project,
		TaskCount: int(taskCount),
	}, nil
}
