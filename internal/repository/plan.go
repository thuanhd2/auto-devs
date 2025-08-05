package repository

import (
	"context"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// PlanRepository defines the interface for plan data access
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
	
	// Advanced filtering
	GetPlansWithFilters(ctx context.Context, filters entity.PlanFilters) ([]*entity.Plan, error)
	
	// Versioning support
	CreateVersion(ctx context.Context, planID uuid.UUID, changeLog string, createdBy *string) (*entity.PlanVersion, error)
	GetVersions(ctx context.Context, planID uuid.UUID) ([]*entity.PlanVersion, error)
	GetVersionByNumber(ctx context.Context, planID uuid.UUID, version int) (*entity.PlanVersion, error)
	RollbackToVersion(ctx context.Context, planID uuid.UUID, version int, createdBy *string) error
	CompareVersions(ctx context.Context, planID uuid.UUID, fromVersion, toVersion int) (*PlanVersionComparison, error)
	
	// Plan step management
	UpdateStepStatus(ctx context.Context, planID uuid.UUID, stepID string, completed bool) error
	GetPlanProgress(ctx context.Context, planID uuid.UUID) (*PlanProgress, error)
	
	// Bulk operations
	BulkUpdateStatus(ctx context.Context, planIDs []uuid.UUID, status entity.PlanStatus) error
	BulkDelete(ctx context.Context, planIDs []uuid.UUID) error
	
	// Analytics and reporting
	GetPlanStatistics(ctx context.Context, taskID *uuid.UUID) (*PlanStatistics, error)
	GetStatusDistribution(ctx context.Context) (map[entity.PlanStatus]int, error)
	
	// Validation
	ValidatePlanExists(ctx context.Context, planID uuid.UUID) (bool, error)
	CheckDuplicateTitle(ctx context.Context, taskID uuid.UUID, title string, excludeID *uuid.UUID) (bool, error)
}

// PlanVersionComparison represents the differences between two plan versions
type PlanVersionComparison struct {
	PlanID      uuid.UUID                 `json:"plan_id"`
	FromVersion int                       `json:"from_version"`
	ToVersion   int                       `json:"to_version"`
	Changes     []PlanVersionChange       `json:"changes"`
	Summary     PlanVersionChangeSummary  `json:"summary"`
}

// PlanVersionChange represents a single change between versions
type PlanVersionChange struct {
	Field     string      `json:"field"`      // "title", "description", "steps", "context", "status"
	Type      string      `json:"type"`       // "added", "removed", "modified"
	OldValue  interface{} `json:"old_value,omitempty"`
	NewValue  interface{} `json:"new_value,omitempty"`
	StepIndex *int        `json:"step_index,omitempty"` // For step-specific changes
}

// PlanVersionChangeSummary provides a high-level summary of changes
type PlanVersionChangeSummary struct {
	TotalChanges    int `json:"total_changes"`
	StepsAdded      int `json:"steps_added"`
	StepsRemoved    int `json:"steps_removed"`
	StepsModified   int `json:"steps_modified"`
	MetaDataChanges int `json:"metadata_changes"`
}

// PlanProgress represents the current progress of a plan
type PlanProgress struct {
	PlanID              uuid.UUID `json:"plan_id"`
	TotalSteps          int       `json:"total_steps"`
	CompletedSteps      int       `json:"completed_steps"`
	CompletionPercentage float64   `json:"completion_percentage"`
	CurrentStep         *string   `json:"current_step,omitempty"`
	EstimatedCompletion *string   `json:"estimated_completion,omitempty"`
}

// PlanStatistics represents comprehensive statistics for plans
type PlanStatistics struct {
	TaskID                *uuid.UUID                   `json:"task_id,omitempty"`
	TotalPlans            int                          `json:"total_plans"`
	StatusDistribution    map[entity.PlanStatus]int    `json:"status_distribution"`
	AverageStepsPerPlan   float64                      `json:"average_steps_per_plan"`
	AverageCompletionTime *float64                     `json:"average_completion_time,omitempty"` // in hours
	VersionStatistics     PlanVersionStatistics        `json:"version_statistics"`
	MostActiveCreators    []PlanCreatorStats           `json:"most_active_creators"`
}

// PlanVersionStatistics represents version-related statistics
type PlanVersionStatistics struct {
	TotalVersions         int     `json:"total_versions"`
	AverageVersionsPerPlan float64 `json:"average_versions_per_plan"`
	MostVersionedPlan     *struct {
		PlanID   uuid.UUID `json:"plan_id"`
		Title    string    `json:"title"`
		Versions int       `json:"versions"`
	} `json:"most_versioned_plan,omitempty"`
}

// PlanCreatorStats represents statistics for plan creators
type PlanCreatorStats struct {
	CreatedBy   string `json:"created_by"`
	PlansCount  int    `json:"plans_count"`
	VersionsCount int  `json:"versions_count"`
}