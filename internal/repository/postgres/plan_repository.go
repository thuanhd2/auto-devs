package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type planRepository struct {
	db *database.GormDB
}

// NewPlanRepository creates a new PostgreSQL plan repository
func NewPlanRepository(db *database.GormDB) repository.PlanRepository {
	return &planRepository{db: db}
}

// Create creates a new plan
func (r *planRepository) Create(ctx context.Context, plan *entity.Plan) error {
	// Generate UUID if not provided
	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}

	// Set default status if not provided
	if plan.Status == "" {
		plan.Status = entity.PlanStatusDRAFT
	}

	result := r.db.WithContext(ctx).Create(plan)
	if result.Error != nil {
		return fmt.Errorf("failed to create plan: %w", result.Error)
	}

	return nil
}

// GetByID retrieves a plan by ID
func (r *planRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plan, error) {
	var plan entity.Plan

	result := r.db.WithContext(ctx).Preload("Task").First(&plan, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("plan not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get plan: %w", result.Error)
	}

	return &plan, nil
}

// GetByTaskID retrieves the plan for a specific task
func (r *planRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.Plan, error) {
	var plan entity.Plan

	result := r.db.WithContext(ctx).Preload("Task").Where("task_id = ?", taskID).First(&plan)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("plan not found for task %s", taskID)
		}
		return nil, fmt.Errorf("failed to get plan by task ID: %w", result.Error)
	}

	return &plan, nil
}

// Update updates an existing plan
func (r *planRepository) Update(ctx context.Context, plan *entity.Plan) error {
	// First check if plan exists
	var existingPlan entity.Plan
	result := r.db.WithContext(ctx).First(&existingPlan, "id = ?", plan.ID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return fmt.Errorf("plan not found with id %s", plan.ID)
		}
		return fmt.Errorf("failed to check plan existence: %w", result.Error)
	}

	// Update the plan
	result = r.db.WithContext(ctx).Save(plan)
	if result.Error != nil {
		return fmt.Errorf("failed to update plan: %w", result.Error)
	}

	return nil
}

// Delete deletes a plan by ID (soft delete)
func (r *planRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Plan{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete plan: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("plan not found with id %s", id)
	}

	return nil
}

// ListByStatus retrieves all plans with a specific status
func (r *planRepository) ListByStatus(ctx context.Context, status entity.PlanStatus) ([]*entity.Plan, error) {
	var plans []entity.Plan

	result := r.db.WithContext(ctx).Preload("Task").Where("status = ?", status).Order("created_at DESC").Find(&plans)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get plans by status: %w", result.Error)
	}

	// Convert to slice of pointers
	planPtrs := make([]*entity.Plan, len(plans))
	for i := range plans {
		planPtrs[i] = &plans[i]
	}

	return planPtrs, nil
}

// UpdateStatus updates the status of a plan
func (r *planRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.PlanStatus) error {
	result := r.db.WithContext(ctx).Model(&entity.Plan{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update plan status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("plan not found with id %s", id)
	}

	return nil
}

// ListByProjectID retrieves all plans for a specific project
func (r *planRepository) ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Plan, error) {
	var plans []entity.Plan

	result := r.db.WithContext(ctx).
		Preload("Task").
		Joins("JOIN tasks ON plans.task_id = tasks.id").
		Where("tasks.project_id = ?", projectID).
		Order("plans.created_at DESC").
		Find(&plans)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get plans by project ID: %w", result.Error)
	}

	// Convert to slice of pointers
	planPtrs := make([]*entity.Plan, len(plans))
	for i := range plans {
		planPtrs[i] = &plans[i]
	}

	return planPtrs, nil
}

// ListByTaskIDs retrieves plans for specific task IDs
func (r *planRepository) ListByTaskIDs(ctx context.Context, taskIDs []uuid.UUID) ([]*entity.Plan, error) {
	var plans []entity.Plan

	result := r.db.WithContext(ctx).Preload("Task").Where("task_id IN ?", taskIDs).Order("created_at DESC").Find(&plans)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get plans by task IDs: %w", result.Error)
	}

	// Convert to slice of pointers
	planPtrs := make([]*entity.Plan, len(plans))
	for i := range plans {
		planPtrs[i] = &plans[i]
	}

	return planPtrs, nil
}

// GetLatestByTaskID retrieves the most recent plan for a task
func (r *planRepository) GetLatestByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.Plan, error) {
	var plan entity.Plan

	result := r.db.WithContext(ctx).
		Preload("Task").
		Where("task_id = ?", taskID).
		Order("created_at DESC").
		First(&plan)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no plan found for task %s", taskID)
		}
		return nil, fmt.Errorf("failed to get latest plan by task ID: %w", result.Error)
	}

	return &plan, nil
}

// UpdateContent updates the content of a plan
func (r *planRepository) UpdateContent(ctx context.Context, id uuid.UUID, content string) error {
	result := r.db.WithContext(ctx).Model(&entity.Plan{}).Where("id = ?", id).Update("content", content)
	if result.Error != nil {
		return fmt.Errorf("failed to update plan content: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("plan not found with id %s", id)
	}

	return nil
}

// SearchByContent performs full-text search on plan content
func (r *planRepository) SearchByContent(ctx context.Context, query string, projectID *uuid.UUID) ([]*entity.Plan, error) {
	searchQuery := r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Preload("Task").
		Where("to_tsvector('english', content) @@ plainto_tsquery('english', ?)", query)

	if projectID != nil {
		searchQuery = searchQuery.
			Joins("JOIN tasks ON plans.task_id = tasks.id").
			Where("tasks.project_id = ?", *projectID)
	}

	var plans []entity.Plan
	result := searchQuery.Order("created_at DESC").Find(&plans)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to search plans by content: %w", result.Error)
	}

	// Convert to slice of pointers
	planPtrs := make([]*entity.Plan, len(plans))
	for i := range plans {
		planPtrs[i] = &plans[i]
	}

	return planPtrs, nil
}

// CreateVersion creates a new version of a plan
func (r *planRepository) CreateVersion(ctx context.Context, planID uuid.UUID, content string, createdBy string) (*entity.PlanVersion, error) {
	var resultVersion *entity.PlanVersion
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get the current max version number for this plan
		var maxVersion int
		result := tx.Model(&entity.PlanVersion{}).
			Where("plan_id = ?", planID).
			Select("COALESCE(MAX(version), 0)").
			Scan(&maxVersion)

		if result.Error != nil {
			return fmt.Errorf("failed to get max version: %w", result.Error)
		}

		// Create new version
		version := &entity.PlanVersion{
			ID:        uuid.New(),
			PlanID:    planID,
			Version:   maxVersion + 1,
			Content:   content,
			CreatedBy: createdBy,
		}

		if err := tx.Create(version).Error; err != nil {
			return fmt.Errorf("failed to create plan version: %w", err)
		}

		resultVersion = version
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resultVersion, nil
}

// GetVersions retrieves all versions of a plan
func (r *planRepository) GetVersions(ctx context.Context, planID uuid.UUID) ([]*entity.PlanVersion, error) {
	var versions []entity.PlanVersion

	result := r.db.WithContext(ctx).Where("plan_id = ?", planID).Order("version ASC").Find(&versions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get plan versions: %w", result.Error)
	}

	// Convert to slice of pointers
	versionPtrs := make([]*entity.PlanVersion, len(versions))
	for i := range versions {
		versionPtrs[i] = &versions[i]
	}

	return versionPtrs, nil
}

// GetVersion retrieves a specific version of a plan
func (r *planRepository) GetVersion(ctx context.Context, planID uuid.UUID, version int) (*entity.PlanVersion, error) {
	var planVersion entity.PlanVersion

	result := r.db.WithContext(ctx).Where("plan_id = ? AND version = ?", planID, version).First(&planVersion)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("plan version %d not found for plan %s", version, planID)
		}
		return nil, fmt.Errorf("failed to get plan version: %w", result.Error)
	}

	return &planVersion, nil
}

// RestoreVersion restores a plan to a specific version
func (r *planRepository) RestoreVersion(ctx context.Context, planID uuid.UUID, version int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get the version content
		var planVersion entity.PlanVersion
		result := tx.Where("plan_id = ? AND version = ?", planID, version).First(&planVersion)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return fmt.Errorf("plan version %d not found for plan %s", version, planID)
			}
			return fmt.Errorf("failed to get plan version: %w", result.Error)
		}

		// Update the plan content
		result = tx.Model(&entity.Plan{}).Where("id = ?", planID).Update("content", planVersion.Content)
		if result.Error != nil {
			return fmt.Errorf("failed to restore plan content: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("plan not found with id %s", planID)
		}

		return nil
	})
}

// CompareVersions compares two versions of a plan
func (r *planRepository) CompareVersions(ctx context.Context, planID uuid.UUID, fromVersion, toVersion int) (*entity.PlanVersionComparison, error) {
	// Get both versions
	fromV, err := r.GetVersion(ctx, planID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get from version: %w", err)
	}

	toV, err := r.GetVersion(ctx, planID, toVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get to version: %w", err)
	}

	// Simple line-by-line comparison (in a real implementation, you might use a proper diff algorithm)
	fromLines := strings.Split(fromV.Content, "\n")
	toLines := strings.Split(toV.Content, "\n")

	var differences []string
	maxLines := len(fromLines)
	if len(toLines) > maxLines {
		maxLines = len(toLines)
	}

	for i := 0; i < maxLines; i++ {
		var fromLine, toLine string
		if i < len(fromLines) {
			fromLine = fromLines[i]
		}
		if i < len(toLines) {
			toLine = toLines[i]
		}

		if fromLine != toLine {
			differences = append(differences, fmt.Sprintf("Line %d: '%s' -> '%s'", i+1, fromLine, toLine))
		}
	}

	return &entity.PlanVersionComparison{
		PlanID:      planID,
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Differences: differences,
		ChangedAt:   time.Now(),
	}, nil
}

// BulkUpdateStatus updates status for multiple plans
func (r *planRepository) BulkUpdateStatus(ctx context.Context, planIDs []uuid.UUID, status entity.PlanStatus) error {
	result := r.db.WithContext(ctx).Model(&entity.Plan{}).Where("id IN ?", planIDs).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk update plan status: %w", result.Error)
	}

	return nil
}

// BulkDelete deletes multiple plans
func (r *planRepository) BulkDelete(ctx context.Context, planIDs []uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id IN ?", planIDs).Delete(&entity.Plan{})
	if result.Error != nil {
		return fmt.Errorf("failed to bulk delete plans: %w", result.Error)
	}

	return nil
}

// GetPlanStatistics retrieves comprehensive plan statistics for a project
func (r *planRepository) GetPlanStatistics(ctx context.Context, projectID uuid.UUID) (*entity.PlanStatistics, error) {
	stats := &entity.PlanStatistics{
		ProjectID:          projectID,
		StatusDistribution: make(map[entity.PlanStatus]int),
		GeneratedAt:        time.Now(),
	}

	// Get total plans for the project
	var totalPlans int64
	result := r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Joins("JOIN tasks ON plans.task_id = tasks.id").
		Where("tasks.project_id = ?", projectID).
		Count(&totalPlans)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to count total plans: %w", result.Error)
	}
	stats.TotalPlans = int(totalPlans)

	// Get status distribution
	var statusStats []struct {
		Status entity.PlanStatus
		Count  int
	}

	result = r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Select("status, count(*) as count").
		Joins("JOIN tasks ON plans.task_id = tasks.id").
		Where("tasks.project_id = ? AND plans.deleted_at IS NULL", projectID).
		Group("status").
		Scan(&statusStats)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get status distribution: %w", result.Error)
	}

	for _, stat := range statusStats {
		stats.StatusDistribution[stat.Status] = stat.Count
	}

	// Get average content length
	var avgLength float64
	result = r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Select("AVG(LENGTH(content))").
		Joins("JOIN tasks ON plans.task_id = tasks.id").
		Where("tasks.project_id = ? AND plans.deleted_at IS NULL", projectID).
		Scan(&avgLength)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get average content length: %w", result.Error)
	}
	stats.AverageContentLength = avgLength

	// Get plans with versions count
	var plansWithVersions int64
	result = r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Joins("JOIN tasks ON plans.task_id = tasks.id").
		Joins("JOIN plan_versions ON plans.id = plan_versions.plan_id").
		Where("tasks.project_id = ? AND plans.deleted_at IS NULL AND plan_versions.deleted_at IS NULL", projectID).
		Distinct("plans.id").
		Count(&plansWithVersions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to count plans with versions: %w", result.Error)
	}
	stats.PlansWithVersions = int(plansWithVersions)

	// Get most active task (task with most plan versions)
	var mostActiveTaskID uuid.UUID
	result = r.db.WithContext(ctx).
		Model(&entity.PlanVersion{}).
		Select("plans.task_id").
		Joins("JOIN plans ON plan_versions.plan_id = plans.id").
		Joins("JOIN tasks ON plans.task_id = tasks.id").
		Where("tasks.project_id = ? AND plan_versions.deleted_at IS NULL", projectID).
		Group("plans.task_id").
		Order("COUNT(*) DESC").
		Limit(1).
		Scan(&mostActiveTaskID)

	if result.Error == nil && mostActiveTaskID != uuid.Nil {
		stats.MostActiveTask = &mostActiveTaskID
	}

	return stats, nil
}

// GetStatusDistribution retrieves the distribution of plan statuses
func (r *planRepository) GetStatusDistribution(ctx context.Context, projectID *uuid.UUID) (map[entity.PlanStatus]int, error) {
	distribution := make(map[entity.PlanStatus]int)

	query := r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Select("status, count(*) as count").
		Where("deleted_at IS NULL").
		Group("status")

	if projectID != nil {
		query = query.
			Joins("JOIN tasks ON plans.task_id = tasks.id").
			Where("tasks.project_id = ?", *projectID)
	}

	var statusStats []struct {
		Status entity.PlanStatus
		Count  int
	}

	result := query.Scan(&statusStats)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get status distribution: %w", result.Error)
	}

	for _, stat := range statusStats {
		distribution[stat.Status] = stat.Count
	}

	return distribution, nil
}

// ValidatePlanExists checks if a plan exists
func (r *planRepository) ValidatePlanExists(ctx context.Context, planID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Plan{}).Where("id = ? AND deleted_at IS NULL", planID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to validate plan exists: %w", err)
	}

	return count > 0, nil
}

// CheckDuplicatePlanForTask checks if a plan already exists for a task
func (r *planRepository) CheckDuplicatePlanForTask(ctx context.Context, taskID uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&entity.Plan{}).Where("task_id = ? AND deleted_at IS NULL", taskID)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check duplicate plan for task: %w", err)
	}

	return count > 0, nil
}