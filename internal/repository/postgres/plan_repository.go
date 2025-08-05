package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
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
		plan.Status = entity.PlanStatusDraft
	}

	// Set initial version
	if plan.Version == 0 {
		plan.Version = 1
	}

	// Begin transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	// Create the plan
	if err := tx.Create(plan).Error; err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	// Create initial version
	version := plan.CreateVersion("Initial plan creation", plan.CreatedBy)
	if err := tx.Create(version).Error; err != nil {
		return fmt.Errorf("failed to create initial plan version: %w", err)
	}

	return tx.Commit().Error
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

// GetByTaskID retrieves a plan by task ID
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
	// Begin transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	// Get existing plan to compare for versioning
	var existingPlan entity.Plan
	if err := tx.First(&existingPlan, "id = ?", plan.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("plan not found with id %s", plan.ID)
		}
		return fmt.Errorf("failed to get existing plan: %w", err)
	}

	// Check if significant changes were made that warrant a new version
	shouldCreateVersion := r.shouldCreateNewVersion(&existingPlan, plan)
	
	if shouldCreateVersion {
		// Create new version before updating
		version := existingPlan.CreateVersion("Plan updated", plan.CreatedBy)
		if err := tx.Create(version).Error; err != nil {
			return fmt.Errorf("failed to create plan version: %w", err)
		}

		// Increment version number
		plan.Version = existingPlan.Version + 1
	}

	// Update the plan
	if err := tx.Save(plan).Error; err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	return tx.Commit().Error
}

// Delete soft deletes a plan by ID
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

// ListByStatus retrieves plans by status
func (r *planRepository) ListByStatus(ctx context.Context, status entity.PlanStatus) ([]*entity.Plan, error) {
	var plans []entity.Plan

	result := r.db.WithContext(ctx).
		Preload("Task").
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&plans)
	
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
	// Begin transaction to update status and create version if needed
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	// Get current plan
	var plan entity.Plan
	if err := tx.First(&plan, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("plan not found with id %s", id)
		}
		return fmt.Errorf("failed to get plan: %w", err)
	}

	// Update status with appropriate metadata
	updates := map[string]interface{}{
		"status": status,
	}

	now := time.Now()
	switch status {
	case entity.PlanStatusApproved:
		updates["approved_at"] = &now
	case entity.PlanStatusRejected:
		updates["rejected_at"] = &now
	}

	if err := tx.Model(&plan).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update plan status: %w", err)
	}

	// Create version for status changes
	version := plan.CreateVersion(fmt.Sprintf("Status changed to %s", status), nil)
	if err := tx.Create(version).Error; err != nil {
		return fmt.Errorf("failed to create version for status change: %w", err)
	}

	return tx.Commit().Error
}

// GetPlansWithFilters retrieves plans with advanced filtering
func (r *planRepository) GetPlansWithFilters(ctx context.Context, filters entity.PlanFilters) ([]*entity.Plan, error) {
	query := r.db.WithContext(ctx).Preload("Task")

	// Apply filters
	if filters.TaskID != nil {
		query = query.Where("task_id = ?", *filters.TaskID)
	}

	if len(filters.Statuses) > 0 {
		query = query.Where("status IN ?", filters.Statuses)
	}

	if filters.CreatedBy != nil {
		query = query.Where("created_by = ?", *filters.CreatedBy)
	}

	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
	}

	if filters.SearchTerm != nil && *filters.SearchTerm != "" {
		searchTerm := "%" + *filters.SearchTerm + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}

	// Apply ordering
	orderBy := "created_at"
	if filters.OrderBy != nil {
		orderBy = *filters.OrderBy
	}

	orderDir := "DESC"
	if filters.OrderDir != nil {
		orderDir = strings.ToUpper(*filters.OrderDir)
	}

	query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))

	// Apply pagination
	if filters.Offset != nil {
		query = query.Offset(*filters.Offset)
	}

	if filters.Limit != nil {
		query = query.Limit(*filters.Limit)
	}

	var plans []entity.Plan
	if err := query.Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("failed to get plans with filters: %w", err)
	}

	// Convert to slice of pointers
	planPtrs := make([]*entity.Plan, len(plans))
	for i := range plans {
		planPtrs[i] = &plans[i]
	}

	return planPtrs, nil
}

// CreateVersion creates a new version of a plan
func (r *planRepository) CreateVersion(ctx context.Context, planID uuid.UUID, changeLog string, createdBy *string) (*entity.PlanVersion, error) {
	// Get current plan
	var plan entity.Plan
	if err := r.db.WithContext(ctx).First(&plan, "id = ?", planID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("plan not found with id %s", planID)
		}
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	// Create new version
	version := plan.CreateVersion(changeLog, createdBy)
	
	if err := r.db.WithContext(ctx).Create(version).Error; err != nil {
		return nil, fmt.Errorf("failed to create plan version: %w", err)
	}

	return version, nil
}

// GetVersions retrieves all versions of a plan
func (r *planRepository) GetVersions(ctx context.Context, planID uuid.UUID) ([]*entity.PlanVersion, error) {
	var versions []entity.PlanVersion

	result := r.db.WithContext(ctx).
		Where("plan_id = ?", planID).
		Order("version DESC").
		Find(&versions)
	
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

// GetVersionByNumber retrieves a specific version of a plan
func (r *planRepository) GetVersionByNumber(ctx context.Context, planID uuid.UUID, version int) (*entity.PlanVersion, error) {
	var planVersion entity.PlanVersion

	result := r.db.WithContext(ctx).
		Where("plan_id = ? AND version = ?", planID, version).
		First(&planVersion)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("plan version %d not found for plan %s", version, planID)
		}
		return nil, fmt.Errorf("failed to get plan version: %w", result.Error)
	}

	return &planVersion, nil
}

// RollbackToVersion rolls back a plan to a specific version
func (r *planRepository) RollbackToVersion(ctx context.Context, planID uuid.UUID, version int, createdBy *string) error {
	// Begin transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	// Get target version
	var targetVersion entity.PlanVersion
	if err := tx.Where("plan_id = ? AND version = ?", planID, version).First(&targetVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("plan version %d not found for plan %s", version, planID)
		}
		return fmt.Errorf("failed to get target version: %w", err)
	}

	// Get current plan
	var currentPlan entity.Plan
	if err := tx.First(&currentPlan, "id = ?", planID).Error; err != nil {
		return fmt.Errorf("failed to get current plan: %w", err)
	}

	// Create a version from current plan before rollback
	rollbackVersion := currentPlan.CreateVersion(fmt.Sprintf("Rollback to version %d", version), createdBy)
	if err := tx.Create(rollbackVersion).Error; err != nil {
		return fmt.Errorf("failed to create rollback version: %w", err)
	}

	// Update plan with target version data
	updates := map[string]interface{}{
		"title":       targetVersion.Title,
		"description": targetVersion.Description,
		"status":      targetVersion.Status,
		"steps":       targetVersion.StepsJSON,
		"context":     targetVersion.ContextJSON,
		"version":     currentPlan.Version + 1,
	}

	if err := tx.Model(&currentPlan).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to rollback plan: %w", err)
	}

	return tx.Commit().Error
}

// CompareVersions compares two versions of a plan
func (r *planRepository) CompareVersions(ctx context.Context, planID uuid.UUID, fromVersion, toVersion int) (*repository.PlanVersionComparison, error) {
	// Get both versions
	var fromVer, toVer entity.PlanVersion
	
	if err := r.db.WithContext(ctx).Where("plan_id = ? AND version = ?", planID, fromVersion).First(&fromVer).Error; err != nil {
		return nil, fmt.Errorf("failed to get from version: %w", err)
	}

	if err := r.db.WithContext(ctx).Where("plan_id = ? AND version = ?", planID, toVersion).First(&toVer).Error; err != nil {
		return nil, fmt.Errorf("failed to get to version: %w", err)
	}

	// Compare and generate changes
	changes := r.compareVersionData(&fromVer, &toVer)
	
	// Generate summary
	summary := r.generateChangeSummary(changes)

	return &repository.PlanVersionComparison{
		PlanID:      planID,
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Changes:     changes,
		Summary:     summary,
	}, nil
}

// UpdateStepStatus updates the completion status of a specific step
func (r *planRepository) UpdateStepStatus(ctx context.Context, planID uuid.UUID, stepID string, completed bool) error {
	// Get current plan
	var plan entity.Plan
	if err := r.db.WithContext(ctx).First(&plan, "id = ?", planID).Error; err != nil {
		return fmt.Errorf("failed to get plan: %w", err)
	}

	// Update step status
	updated := false
	if completed {
		updated = plan.MarkStepCompleted(stepID)
	} else {
		step := plan.GetStepByID(stepID)
		if step != nil {
			step.Completed = false
			step.CompletedAt = nil
			updated = true
		}
	}

	if !updated {
		return fmt.Errorf("step not found with id %s", stepID)
	}

	// Save updated plan
	if err := r.db.WithContext(ctx).Save(&plan).Error; err != nil {
		return fmt.Errorf("failed to update step status: %w", err)
	}

	return nil
}

// GetPlanProgress retrieves the current progress of a plan
func (r *planRepository) GetPlanProgress(ctx context.Context, planID uuid.UUID) (*repository.PlanProgress, error) {
	var plan entity.Plan
	if err := r.db.WithContext(ctx).First(&plan, "id = ?", planID).Error; err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	completedSteps := 0
	var currentStep *string
	
	for _, step := range plan.Steps {
		if step.Completed {
			completedSteps++
		} else if currentStep == nil {
			// First incomplete step is the current step
			currentStep = &step.Description
		}
	}

	return &repository.PlanProgress{
		PlanID:              planID,
		TotalSteps:          len(plan.Steps),
		CompletedSteps:      completedSteps,
		CompletionPercentage: plan.GetCompletionPercentage(),
		CurrentStep:         currentStep,
	}, nil
}

// BulkUpdateStatus updates the status of multiple plans
func (r *planRepository) BulkUpdateStatus(ctx context.Context, planIDs []uuid.UUID, status entity.PlanStatus) error {
	if len(planIDs) == 0 {
		return nil
	}

	updates := map[string]interface{}{
		"status": status,
	}

	now := time.Now()
	switch status {
	case entity.PlanStatusApproved:
		updates["approved_at"] = &now
	case entity.PlanStatusRejected:
		updates["rejected_at"] = &now
	}

	result := r.db.WithContext(ctx).Model(&entity.Plan{}).Where("id IN ?", planIDs).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk update plan status: %w", result.Error)
	}

	return nil
}

// BulkDelete soft deletes multiple plans
func (r *planRepository) BulkDelete(ctx context.Context, planIDs []uuid.UUID) error {
	if len(planIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).Delete(&entity.Plan{}, "id IN ?", planIDs)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk delete plans: %w", result.Error)
	}

	return nil
}

// GetPlanStatistics retrieves comprehensive statistics for plans
func (r *planRepository) GetPlanStatistics(ctx context.Context, taskID *uuid.UUID) (*repository.PlanStatistics, error) {
	query := r.db.WithContext(ctx).Model(&entity.Plan{})
	
	if taskID != nil {
		query = query.Where("task_id = ?", *taskID)
	}

	// Get total count
	var totalPlans int64
	if err := query.Count(&totalPlans).Error; err != nil {
		return nil, fmt.Errorf("failed to count plans: %w", err)
	}

	// Get status distribution
	statusDist, err := r.GetStatusDistribution(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get status distribution: %w", err)
	}

	// Calculate average steps per plan
	var avgSteps float64
	if totalPlans > 0 {
		var plans []entity.Plan
		if err := query.Find(&plans).Error; err != nil {
			return nil, fmt.Errorf("failed to get plans for statistics: %w", err)
		}

		totalSteps := 0
		for _, plan := range plans {
			totalSteps += len(plan.Steps)
		}
		avgSteps = float64(totalSteps) / float64(totalPlans)
	}

	// Get version statistics
	versionStats, err := r.getVersionStatistics(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version statistics: %w", err)
	}

	return &repository.PlanStatistics{
		TaskID:                taskID,
		TotalPlans:            int(totalPlans),
		StatusDistribution:    statusDist,
		AverageStepsPerPlan:   avgSteps,
		VersionStatistics:     *versionStats,
	}, nil
}

// GetStatusDistribution retrieves the distribution of plans by status
func (r *planRepository) GetStatusDistribution(ctx context.Context) (map[entity.PlanStatus]int, error) {
	var results []struct {
		Status entity.PlanStatus
		Count  int
	}

	err := r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get status distribution: %w", err)
	}

	distribution := make(map[entity.PlanStatus]int)
	for _, result := range results {
		distribution[result.Status] = result.Count
	}

	return distribution, nil
}

// ValidatePlanExists checks if a plan exists
func (r *planRepository) ValidatePlanExists(ctx context.Context, planID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Plan{}).Where("id = ?", planID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to validate plan existence: %w", err)
	}
	return count > 0, nil
}

// CheckDuplicateTitle checks if a plan title already exists for the task
func (r *planRepository) CheckDuplicateTitle(ctx context.Context, taskID uuid.UUID, title string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&entity.Plan{}).Where("task_id = ? AND title = ?", taskID, title)
	
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check duplicate title: %w", err)
	}

	return count > 0, nil
}

// Helper methods

// shouldCreateNewVersion determines if changes warrant a new version
func (r *planRepository) shouldCreateNewVersion(old, new *entity.Plan) bool {
	// Check if title, description, or steps changed
	if old.Title != new.Title || old.Description != new.Description {
		return true
	}

	// Check if steps changed
	if !r.stepsEqual(old.Steps, new.Steps) {
		return true
	}

	// Check if context changed significantly
	if !r.contextEqual(old.Context, new.Context) {
		return true
	}

	return false
}

// stepsEqual compares two step slices for equality
func (r *planRepository) stepsEqual(a, b []entity.PlanStep) bool {
	if len(a) != len(b) {
		return true // Different length means changes
	}

	for i := range a {
		if a[i].Description != b[i].Description || 
		   a[i].Action != b[i].Action || 
		   a[i].Order != b[i].Order {
			return false
		}
		
		// Check parameters
		if !reflect.DeepEqual(a[i].Parameters, b[i].Parameters) {
			return false
		}
	}

	return true
}

// contextEqual compares two context maps for equality
func (r *planRepository) contextEqual(a, b map[string]string) bool {
	return reflect.DeepEqual(a, b)
}

// compareVersionData compares two plan versions and returns changes
func (r *planRepository) compareVersionData(from, to *entity.PlanVersion) []repository.PlanVersionChange {
	var changes []repository.PlanVersionChange

	// Compare title
	if from.Title != to.Title {
		changes = append(changes, repository.PlanVersionChange{
			Field:    "title",
			Type:     "modified",
			OldValue: from.Title,
			NewValue: to.Title,
		})
	}

	// Compare description
	if from.Description != to.Description {
		changes = append(changes, repository.PlanVersionChange{
			Field:    "description",
			Type:     "modified",
			OldValue: from.Description,
			NewValue: to.Description,
		})
	}

	// Compare status
	if from.Status != to.Status {
		changes = append(changes, repository.PlanVersionChange{
			Field:    "status",
			Type:     "modified",
			OldValue: from.Status,
			NewValue: to.Status,
		})
	}

	// Compare steps
	stepChanges := r.compareSteps(from.Steps, to.Steps)
	changes = append(changes, stepChanges...)

	// Compare context
	contextChanges := r.compareContext(from.Context, to.Context)
	changes = append(changes, contextChanges...)

	return changes
}

// compareSteps compares steps between versions
func (r *planRepository) compareSteps(fromSteps, toSteps []entity.PlanStep) []repository.PlanVersionChange {
	var changes []repository.PlanVersionChange

	// Create maps for easier comparison
	fromMap := make(map[string]entity.PlanStep)
	toMap := make(map[string]entity.PlanStep)

	for _, step := range fromSteps {
		fromMap[step.ID] = step
	}

	for _, step := range toSteps {
		toMap[step.ID] = step
	}

	// Find added steps
	for id, step := range toMap {
		if _, exists := fromMap[id]; !exists {
			changes = append(changes, repository.PlanVersionChange{
				Field:    "steps",
				Type:     "added",
				NewValue: step,
			})
		}
	}

	// Find removed and modified steps
	for id, fromStep := range fromMap {
		if toStep, exists := toMap[id]; !exists {
			// Step removed
			changes = append(changes, repository.PlanVersionChange{
				Field:    "steps",
				Type:     "removed",
				OldValue: fromStep,
			})
		} else if !reflect.DeepEqual(fromStep, toStep) {
			// Step modified
			changes = append(changes, repository.PlanVersionChange{
				Field:    "steps",
				Type:     "modified",
				OldValue: fromStep,
				NewValue: toStep,
			})
		}
	}

	return changes
}

// compareContext compares context between versions
func (r *planRepository) compareContext(fromContext, toContext map[string]string) []repository.PlanVersionChange {
	var changes []repository.PlanVersionChange

	// Find added and modified context
	for key, toValue := range toContext {
		if fromValue, exists := fromContext[key]; !exists {
			changes = append(changes, repository.PlanVersionChange{
				Field:    "context",
				Type:     "added",
				NewValue: map[string]string{key: toValue},
			})
		} else if fromValue != toValue {
			changes = append(changes, repository.PlanVersionChange{
				Field:    "context",
				Type:     "modified",
				OldValue: map[string]string{key: fromValue},
				NewValue: map[string]string{key: toValue},
			})
		}
	}

	// Find removed context
	for key, fromValue := range fromContext {
		if _, exists := toContext[key]; !exists {
			changes = append(changes, repository.PlanVersionChange{
				Field:    "context",
				Type:     "removed",
				OldValue: map[string]string{key: fromValue},
			})
		}
	}

	return changes
}

// generateChangeSummary generates a summary from changes
func (r *planRepository) generateChangeSummary(changes []repository.PlanVersionChange) repository.PlanVersionChangeSummary {
	summary := repository.PlanVersionChangeSummary{
		TotalChanges: len(changes),
	}

	for _, change := range changes {
		switch change.Field {
		case "steps":
			switch change.Type {
			case "added":
				summary.StepsAdded++
			case "removed":
				summary.StepsRemoved++
			case "modified":
				summary.StepsModified++
			}
		default:
			summary.MetaDataChanges++
		}
	}

	return summary
}

// getVersionStatistics calculates version-related statistics
func (r *planRepository) getVersionStatistics(ctx context.Context, taskID *uuid.UUID) (*repository.PlanVersionStatistics, error) {
	query := r.db.WithContext(ctx).Model(&entity.PlanVersion{})
	
	if taskID != nil {
		query = query.Joins("JOIN plans ON plan_versions.plan_id = plans.id").
			Where("plans.task_id = ?", *taskID)
	}

	// Get total versions
	var totalVersions int64
	if err := query.Count(&totalVersions).Error; err != nil {
		return nil, fmt.Errorf("failed to count versions: %w", err)
	}

	// Get total plans for average calculation
	planQuery := r.db.WithContext(ctx).Model(&entity.Plan{})
	if taskID != nil {
		planQuery = planQuery.Where("task_id = ?", *taskID)
	}

	var totalPlans int64
	if err := planQuery.Count(&totalPlans).Error; err != nil {
		return nil, fmt.Errorf("failed to count plans: %w", err)
	}

	avgVersions := float64(0)
	if totalPlans > 0 {
		avgVersions = float64(totalVersions) / float64(totalPlans)
	}

	return &repository.PlanVersionStatistics{
		TotalVersions:         int(totalVersions),
		AverageVersionsPerPlan: avgVersions,
	}, nil
}