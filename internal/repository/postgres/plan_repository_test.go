package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestProject(t *testing.T, ctx context.Context, db *database.GormDB) *entity.Project {
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepositoryURL: "https://github.com/test/repo.git",
	}
	
	result := db.WithContext(ctx).Create(project)
	require.NoError(t, result.Error)
	return project
}

func createTestTask(t *testing.T, ctx context.Context, db *database.GormDB, projectID uuid.UUID) *entity.Task {
	task := &entity.Task{
		ProjectID:   projectID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}
	
	result := db.WithContext(ctx).Create(task)
	require.NoError(t, result.Error)
	return task
}

func TestPlanRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewPlanRepository(db)
	ctx := context.Background()

	// Create test project and task
	project := createTestProject(t, ctx, db)
	task := createTestTask(t, ctx, db, project.ID)

	// Test creating a plan
	plan := &entity.Plan{
		TaskID:  task.ID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Test Plan

This is a test plan.",
	}

	err := repo.Create(ctx, plan)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, plan.ID)
	assert.Equal(t, entity.PlanStatusDRAFT, plan.Status)
}

func TestPlanRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)

	repo := NewPlanRepository(db)
	ctx := context.Background()

	// Create test data
	projectID := uuid.New()
	taskID := uuid.New()
	planID := uuid.New()

	// Insert test project
	_, err := db.ExecContext(ctx, `
		INSERT INTO projects (id, name, description, repo_url) 
		VALUES ($1, 'Test Project', 'Test Description', 'https://github.com/test/repo')`,
		projectID)
	require.NoError(t, err)

	// Insert test task
	_, err = db.ExecContext(ctx, `
		INSERT INTO tasks (id, project_id, title, description, status) 
		VALUES ($1, $2, 'Test Task', 'Test Description', 'TODO')`,
		taskID, projectID)
	require.NoError(t, err)

	// Insert test plan
	_, err = db.ExecContext(ctx, `
		INSERT INTO plans (id, task_id, status, content, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		planID, taskID, entity.PlanStatusDRAFT, "Test content", time.Now(), time.Now())
	require.NoError(t, err)

	// Test getting plan by ID
	plan, err := repo.GetByID(ctx, planID)
	require.NoError(t, err)
	assert.Equal(t, planID, plan.ID)
	assert.Equal(t, taskID, plan.TaskID)
	assert.Equal(t, entity.PlanStatusDRAFT, plan.Status)
	assert.Equal(t, "Test content", plan.Content)
}

func TestPlanRepository_GetByTaskID(t *testing.T) {
	db := SetupTestDB(t)

	repo := NewPlanRepository(db)
	ctx := context.Background()

	// Create test data
	projectID := uuid.New()
	taskID := uuid.New()
	planID := uuid.New()

	// Insert test project
	_, err := db.ExecContext(ctx, `
		INSERT INTO projects (id, name, description, repo_url) 
		VALUES ($1, 'Test Project', 'Test Description', 'https://github.com/test/repo')`,
		projectID)
	require.NoError(t, err)

	// Insert test task
	_, err = db.ExecContext(ctx, `
		INSERT INTO tasks (id, project_id, title, description, status) 
		VALUES ($1, $2, 'Test Task', 'Test Description', 'TODO')`,
		taskID, projectID)
	require.NoError(t, err)

	// Insert test plan
	_, err = db.ExecContext(ctx, `
		INSERT INTO plans (id, task_id, status, content, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		planID, taskID, entity.PlanStatusDRAFT, "Test content", time.Now(), time.Now())
	require.NoError(t, err)

	// Test getting plan by task ID
	plan, err := repo.GetByTaskID(ctx, taskID)
	require.NoError(t, err)
	assert.Equal(t, planID, plan.ID)
	assert.Equal(t, taskID, plan.TaskID)
	assert.Equal(t, entity.PlanStatusDRAFT, plan.Status)
	assert.Equal(t, "Test content", plan.Content)
}

func TestPlanRepository_Update(t *testing.T) {
	db := SetupTestDB(t)

	repo := NewPlanRepository(db)
	ctx := context.Background()

	// Create test data
	projectID := uuid.New()
	taskID := uuid.New()

	// Insert test project
	_, err := db.ExecContext(ctx, `
		INSERT INTO projects (id, name, description, repo_url) 
		VALUES ($1, 'Test Project', 'Test Description', 'https://github.com/test/repo')`,
		projectID)
	require.NoError(t, err)

	// Insert test task
	_, err = db.ExecContext(ctx, `
		INSERT INTO tasks (id, project_id, title, description, status) 
		VALUES ($1, $2, 'Test Task', 'Test Description', 'TODO')`,
		taskID, projectID)
	require.NoError(t, err)

	// Create and insert plan
	plan := &entity.Plan{
		TaskID:  taskID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Initial Plan",
	}

	err = repo.Create(ctx, plan)
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

	repo := NewPlanRepository(db)
	ctx := context.Background()

	// Create test data
	projectID := uuid.New()
	taskID := uuid.New()

	// Insert test project
	_, err := db.ExecContext(ctx, `
		INSERT INTO projects (id, name, description, repo_url) 
		VALUES ($1, 'Test Project', 'Test Description', 'https://github.com/test/repo')`,
		projectID)
	require.NoError(t, err)

	// Insert test task
	_, err = db.ExecContext(ctx, `
		INSERT INTO tasks (id, project_id, title, description, status) 
		VALUES ($1, $2, 'Test Task', 'Test Description', 'TODO')`,
		taskID, projectID)
	require.NoError(t, err)

	// Create and insert plan
	plan := &entity.Plan{
		TaskID:  taskID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Test Plan",
	}

	err = repo.Create(ctx, plan)
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

	repo := NewPlanRepository(db)
	ctx := context.Background()

	// Create test data
	projectID := uuid.New()
	taskID1 := uuid.New()
	taskID2 := uuid.New()

	// Insert test project
	_, err := db.ExecContext(ctx, `
		INSERT INTO projects (id, name, description, repo_url) 
		VALUES ($1, 'Test Project', 'Test Description', 'https://github.com/test/repo')`,
		projectID)
	require.NoError(t, err)

	// Insert test tasks
	_, err = db.ExecContext(ctx, `
		INSERT INTO tasks (id, project_id, title, description, status) 
		VALUES ($1, $2, 'Test Task 1', 'Test Description', 'TODO')`,
		taskID1, projectID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO tasks (id, project_id, title, description, status) 
		VALUES ($1, $2, 'Test Task 2', 'Test Description', 'TODO')`,
		taskID2, projectID)
	require.NoError(t, err)

	// Create plans with different statuses
	plan1 := &entity.Plan{
		TaskID:  taskID1,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Draft Plan",
	}

	plan2 := &entity.Plan{
		TaskID:  taskID2,
		Status:  entity.PlanStatusREVIEWING,
		Content: "# Reviewing Plan",
	}

	err = repo.Create(ctx, plan1)
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

	repo := NewPlanRepository(db)
	ctx := context.Background()

	// Create test data
	projectID := uuid.New()
	taskID := uuid.New()

	// Insert test project
	_, err := db.ExecContext(ctx, `
		INSERT INTO projects (id, name, description, repo_url) 
		VALUES ($1, 'Test Project', 'Test Description', 'https://github.com/test/repo')`,
		projectID)
	require.NoError(t, err)

	// Insert test task
	_, err = db.ExecContext(ctx, `
		INSERT INTO tasks (id, project_id, title, description, status) 
		VALUES ($1, $2, 'Test Task', 'Test Description', 'TODO')`,
		taskID, projectID)
	require.NoError(t, err)

	// Create plan
	plan := &entity.Plan{
		TaskID:  taskID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Initial Plan",
	}

	err = repo.Create(ctx, plan)
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

	repo := NewPlanRepository(db)
	ctx := context.Background()

	// Create test data
	projectID := uuid.New()
	taskID := uuid.New()

	// Insert test project
	_, err := db.ExecContext(ctx, `
		INSERT INTO projects (id, name, description, repo_url) 
		VALUES ($1, 'Test Project', 'Test Description', 'https://github.com/test/repo')`,
		projectID)
	require.NoError(t, err)

	// Insert test task
	_, err = db.ExecContext(ctx, `
		INSERT INTO tasks (id, project_id, title, description, status) 
		VALUES ($1, $2, 'Test Task', 'Test Description', 'TODO')`,
		taskID, projectID)
	require.NoError(t, err)

	// Create plan
	plan := &entity.Plan{
		TaskID:  taskID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Initial Plan",
	}

	err = repo.Create(ctx, plan)
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

	repo := NewPlanRepository(db)
	ctx := context.Background()

	// Create test data
	projectID := uuid.New()
	taskID := uuid.New()

	// Insert test project
	_, err := db.ExecContext(ctx, `
		INSERT INTO projects (id, name, description, repo_url) 
		VALUES ($1, 'Test Project', 'Test Description', 'https://github.com/test/repo')`,
		projectID)
	require.NoError(t, err)

	// Insert test task
	_, err = db.ExecContext(ctx, `
		INSERT INTO tasks (id, project_id, title, description, status) 
		VALUES ($1, $2, 'Test Task', 'Test Description', 'TODO')`,
		taskID, projectID)
	require.NoError(t, err)

	// Create plan
	plan := &entity.Plan{
		TaskID:  taskID,
		Status:  entity.PlanStatusDRAFT,
		Content: "# Test Plan",
	}

	err = repo.Create(ctx, plan)
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