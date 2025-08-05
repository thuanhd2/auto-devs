package repository

import (
	"context"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// PlanRepository defines the interface for plan data persistence
type PlanRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, plan *entity.Plan) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Plan, error)
	GetByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.Plan, error)
	Update(ctx context.Context, plan *entity.Plan) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Status-based queries
	ListByStatus(ctx context.Context, status entity.PlanStatus) ([]*entity.Plan, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.PlanStatus) error

	// Advanced queries
	ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Plan, error)
	ListByTaskIDs(ctx context.Context, taskIDs []uuid.UUID) ([]*entity.Plan, error)
	GetLatestByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.Plan, error)

	// Content management
	UpdateContent(ctx context.Context, id uuid.UUID, content string) error
	SearchByContent(ctx context.Context, query string, projectID *uuid.UUID) ([]*entity.Plan, error)

	// Versioning support
	CreateVersion(ctx context.Context, planID uuid.UUID, content string, createdBy string) (*entity.PlanVersion, error)
	GetVersions(ctx context.Context, planID uuid.UUID) ([]*entity.PlanVersion, error)
	GetVersion(ctx context.Context, planID uuid.UUID, version int) (*entity.PlanVersion, error)
	RestoreVersion(ctx context.Context, planID uuid.UUID, version int) error
	CompareVersions(ctx context.Context, planID uuid.UUID, fromVersion, toVersion int) (*entity.PlanVersionComparison, error)

	// Bulk operations
	BulkUpdateStatus(ctx context.Context, planIDs []uuid.UUID, status entity.PlanStatus) error
	BulkDelete(ctx context.Context, planIDs []uuid.UUID) error

	// Statistics and analytics
	GetPlanStatistics(ctx context.Context, projectID uuid.UUID) (*entity.PlanStatistics, error)
	GetStatusDistribution(ctx context.Context, projectID *uuid.UUID) (map[entity.PlanStatus]int, error)

	// Validation helpers
	ValidatePlanExists(ctx context.Context, planID uuid.UUID) (bool, error)
	CheckDuplicatePlanForTask(ctx context.Context, taskID uuid.UUID, excludeID *uuid.UUID) (bool, error)
}