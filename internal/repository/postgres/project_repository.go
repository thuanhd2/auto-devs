package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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



// Update updates an existing project
func (r *projectRepository) Update(ctx context.Context, project *entity.Project) error {
	// First check if project exists
	var existingProject entity.Project
	result := r.db.WithContext(ctx).First(&existingProject, "id = ?", project.ID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return fmt.Errorf("project not found with id %s", project.ID)
		}
		return fmt.Errorf("failed to check project existence: %w", result.Error)
	}

	// Update the project
	result = r.db.WithContext(ctx).Save(project)
	if result.Error != nil {
		return fmt.Errorf("failed to update project: %w", result.Error)
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



// GetAllWithParams retrieves projects with search, filtering, sorting and pagination
func (r *projectRepository) GetAllWithParams(ctx context.Context, params repository.GetProjectsParams) ([]*entity.Project, int, error) {
	var projects []entity.Project
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Project{})

	// Apply archived filter
	if params.Archived != nil {
		if *params.Archived {
			query = query.Unscoped().Where("deleted_at IS NOT NULL")
		} else {
			query = query.Where("deleted_at IS NULL")
		}
	}

	// Apply search filter
	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// Apply sorting
	orderClause := "created_at" // default
	switch params.SortBy {
	case "name":
		orderClause = "name"
	case "created_at":
		orderClause = "created_at"
	case "task_count":
		// For task count sorting, we need to join with tasks table
		query = query.Select("projects.*, COUNT(tasks.id) as task_count").
			Joins("LEFT JOIN tasks ON projects.id = tasks.project_id AND tasks.deleted_at IS NULL").
			Group("projects.id")
		orderClause = "task_count"
	}

	if params.SortOrder == "asc" {
		orderClause += " ASC"
	} else {
		orderClause += " DESC"
	}
	query = query.Order(orderClause)

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	query = query.Offset(offset).Limit(params.PageSize)

	// Execute query
	result := query.Find(&projects)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to get projects: %w", result.Error)
	}

	// Convert to slice of pointers
	projectPtrs := make([]*entity.Project, len(projects))
	for i := range projects {
		projectPtrs[i] = &projects[i]
	}

	return projectPtrs, int(total), nil
}

// GetTaskStatistics retrieves task count statistics for a project
func (r *projectRepository) GetTaskStatistics(ctx context.Context, projectID uuid.UUID) (map[entity.TaskStatus]int, error) {
	var results []struct {
		Status entity.TaskStatus `json:"status"`
		Count  int               `json:"count"`
	}

	result := r.db.WithContext(ctx).
		Model(&entity.Task{}).
		Select("status, COUNT(*) as count").
		Where("project_id = ?", projectID).
		Group("status").
		Find(&results)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get task statistics: %w", result.Error)
	}

	taskCounts := make(map[entity.TaskStatus]int)
	for _, r := range results {
		taskCounts[r.Status] = r.Count
	}

	return taskCounts, nil
}

// GetLastActivityAt retrieves the last activity timestamp for a project
func (r *projectRepository) GetLastActivityAt(ctx context.Context, projectID uuid.UUID) (*time.Time, error) {
	var lastActivity sql.NullTime

	// Get the most recent task update time for this project
	result := r.db.WithContext(ctx).
		Model(&entity.Task{}).
		Select("MAX(updated_at)").
		Where("project_id = ?", projectID).
		Scan(&lastActivity)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get last activity: %w", result.Error)
	}

	// If no tasks exist, use project's updated_at
	if !lastActivity.Valid {
		var project entity.Project
		result := r.db.WithContext(ctx).Select("updated_at").First(&project, "id = ?", projectID)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to get project updated_at: %w", result.Error)
		}
		return &project.UpdatedAt, nil
	}

	return &lastActivity.Time, nil
}

// Archive soft deletes a project (sets deleted_at)
func (r *projectRepository) Archive(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Project{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to archive project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("project not found with id %s", id)
	}

	return nil
}

// Restore undeletes a project (clears deleted_at)
func (r *projectRepository) Restore(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Unscoped().Model(&entity.Project{}).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil)

	if result.Error != nil {
		return fmt.Errorf("failed to restore project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("archived project not found with id %s", id)
	}

	return nil
}

// CheckNameExists checks if a project name already exists
func (r *projectRepository) CheckNameExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.Project{}).Where("name = ?", name)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	result := query.Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check name existence: %w", result.Error)
	}

	return count > 0, nil
}

// GetSettings retrieves project settings
func (r *projectRepository) GetSettings(ctx context.Context, projectID uuid.UUID) (*entity.ProjectSettings, error) {
	var settings entity.ProjectSettings

	result := r.db.WithContext(ctx).First(&settings, "project_id = ?", projectID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("settings not found")
		}
		return nil, fmt.Errorf("failed to get settings: %w", result.Error)
	}

	return &settings, nil
}

// CreateSettings creates new project settings
func (r *projectRepository) CreateSettings(ctx context.Context, settings *entity.ProjectSettings) error {
	if settings.ID == uuid.Nil {
		settings.ID = uuid.New()
	}

	result := r.db.WithContext(ctx).Create(settings)
	if result.Error != nil {
		return fmt.Errorf("failed to create settings: %w", result.Error)
	}

	return nil
}

// UpdateSettings updates existing project settings
func (r *projectRepository) UpdateSettings(ctx context.Context, settings *entity.ProjectSettings) error {
	result := r.db.WithContext(ctx).Save(settings)
	if result.Error != nil {
		return fmt.Errorf("failed to update settings: %w", result.Error)
	}

	return nil
}
