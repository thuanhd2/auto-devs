package postgres

import (
	"context"
	"testing"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewPlanRepository(db)
	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project and task
	project := CreateTestProject(t, projectRepo, ctx)
	task := CreateTestTask(t, taskRepo, project.ID, ctx)

	// Test creating a plan
	plan := &entity.Plan{
		TaskID:  task.ID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Test Plan\nThis is a test plan.",
	}

	err := repo.Create(ctx, plan)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, plan.ID)
	assert.Equal(t, entity.PlanStatusDRAFT, plan.Status)
}

func TestPlanRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewPlanRepository(db)
	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test data
	project := CreateTestProject(t, projectRepo, ctx)
	task := CreateTestTask(t, taskRepo, project.ID, ctx)

	// Create test plan
	plan := &entity.Plan{
		TaskID:  task.ID,
		Status:  entity.PlanStatusDRAFT,
		Content: "Test content",
	}

	err := repo.Create(ctx, plan)
	require.NoError(t, err)

	// Test getting plan by ID
	retrievedPlan, err := repo.GetByID(ctx, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, plan.ID, retrievedPlan.ID)
	assert.Equal(t, task.ID, retrievedPlan.TaskID)
	assert.Equal(t, entity.PlanStatusDRAFT, retrievedPlan.Status)
	assert.Equal(t, "Test content", retrievedPlan.Content)
}

func TestPlanRepository_GetByTaskID(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewPlanRepository(db)
	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test data
	project := CreateTestProject(t, projectRepo, ctx)
	task := CreateTestTask(t, taskRepo, project.ID, ctx)

	// Create test plan
	plan := &entity.Plan{
		TaskID:  task.ID,
		Status:  entity.PlanStatusDRAFT,
		Content: "Test content",
	}

	err := repo.Create(ctx, plan)
	require.NoError(t, err)

	// Test getting plan by task ID
	retrievedPlan, err := repo.GetByTaskID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, plan.ID, retrievedPlan.ID)
	assert.Equal(t, task.ID, retrievedPlan.TaskID)
	assert.Equal(t, entity.PlanStatusDRAFT, retrievedPlan.Status)
	assert.Equal(t, "Test content", retrievedPlan.Content)
}

func TestPlanRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewPlanRepository(db)
	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test data
	project := CreateTestProject(t, projectRepo, ctx)
	task := CreateTestTask(t, taskRepo, project.ID, ctx)

	// Create and insert plan
	plan := &entity.Plan{
		TaskID:  task.ID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Initial Plan",
	}

	err := repo.Create(ctx, plan)
	require.NoError(t, err)

	// Update the plan
	plan.Status = entity.PlanStatusREVIEWING
	plan.Content = "# Updated Plan"

	err = repo.Update(ctx, plan)
	require.NoError(t, err)

	// Verify the update
	updatedPlan, err := repo.GetByID(ctx, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.PlanStatusREVIEWING, updatedPlan.Status)
	assert.Equal(t, "# Updated Plan", updatedPlan.Content)
}

func TestPlanRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewPlanRepository(db)
	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test data
	project := CreateTestProject(t, projectRepo, ctx)
	task := CreateTestTask(t, taskRepo, project.ID, ctx)

	// Create and insert plan
	plan := &entity.Plan{
		TaskID:  task.ID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Test Plan",
	}

	err := repo.Create(ctx, plan)
	require.NoError(t, err)

	// Delete the plan
	err = repo.Delete(ctx, plan.ID)
	require.NoError(t, err)

	// Verify the plan is deleted (soft delete)
	_, err = repo.GetByID(ctx, plan.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan not found")
}

func TestPlanRepository_ListByStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewPlanRepository(db)
	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test data
	project := CreateTestProject(t, projectRepo, ctx)
	task1 := CreateTestTask(t, taskRepo, project.ID, ctx)
	task2 := CreateTestTask(t, taskRepo, project.ID, ctx)

	// Create plans with different statuses
	plan1 := &entity.Plan{
		TaskID:  task1.ID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Draft Plan",
	}

	plan2 := &entity.Plan{
		TaskID:  task2.ID,
		Status:  entity.PlanStatusREVIEWING,
		Content: "# Reviewing Plan",
	}

	err := repo.Create(ctx, plan1)
	require.NoError(t, err)

	err = repo.Create(ctx, plan2)
	require.NoError(t, err)

	// Test listing by status
	draftPlans, err := repo.ListByStatus(ctx, entity.PlanStatusDRAFT)
	require.NoError(t, err)
	assert.Len(t, draftPlans, 1)
	assert.Equal(t, entity.PlanStatusDRAFT, draftPlans[0].Status)

	reviewingPlans, err := repo.ListByStatus(ctx, entity.PlanStatusREVIEWING)
	require.NoError(t, err)
	assert.Len(t, reviewingPlans, 1)
	assert.Equal(t, entity.PlanStatusREVIEWING, reviewingPlans[0].Status)
}

func TestPlanRepository_CreateVersion(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewPlanRepository(db)
	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test data
	project := CreateTestProject(t, projectRepo, ctx)
	task := CreateTestTask(t, taskRepo, project.ID, ctx)

	// Create plan
	plan := &entity.Plan{
		TaskID:  task.ID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Initial Plan",
	}

	err := repo.Create(ctx, plan)
	require.NoError(t, err)

	// Create a version
	version, err := repo.CreateVersion(ctx, plan.ID, "# Updated Plan Content", "test-user")
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, version.ID)
	assert.Equal(t, plan.ID, version.PlanID)
	assert.Equal(t, 2, version.Version) // Should be version 2 (version 1 created automatically on plan creation)
	assert.Equal(t, "# Updated Plan Content", version.Content)
	assert.Equal(t, "test-user", version.CreatedBy)
}

func TestPlanRepository_GetVersions(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewPlanRepository(db)
	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test data
	project := CreateTestProject(t, projectRepo, ctx)
	task := CreateTestTask(t, taskRepo, project.ID, ctx)

	// Create plan
	plan := &entity.Plan{
		TaskID:  task.ID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Initial Plan",
	}

	err := repo.Create(ctx, plan)
	require.NoError(t, err)

	// Create additional versions
	_, err = repo.CreateVersion(ctx, plan.ID, "# Version 2", "user1")
	require.NoError(t, err)

	_, err = repo.CreateVersion(ctx, plan.ID, "# Version 3", "user2")
	require.NoError(t, err)

	// Get all versions
	versions, err := repo.GetVersions(ctx, plan.ID)
	require.NoError(t, err)
	assert.Len(t, versions, 3) // Initial version + 2 created versions
	assert.Equal(t, 1, versions[0].Version)
	assert.Equal(t, 2, versions[1].Version)
	assert.Equal(t, 3, versions[2].Version)
}

func TestPlanRepository_ValidatePlanExists(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewPlanRepository(db)
	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test data
	project := CreateTestProject(t, projectRepo, ctx)
	task := CreateTestTask(t, taskRepo, project.ID, ctx)

	// Create plan
	plan := &entity.Plan{
		TaskID:  task.ID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Test Plan",
	}

	err := repo.Create(ctx, plan)
	require.NoError(t, err)

	// Test validation
	exists, err := repo.ValidatePlanExists(ctx, plan.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test with non-existent plan
	nonExistentID := uuid.New()
	exists, err = repo.ValidatePlanExists(ctx, nonExistentID)
	require.NoError(t, err)
	assert.False(t, exists)
}
