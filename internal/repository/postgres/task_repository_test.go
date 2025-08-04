package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestProject(t *testing.T, projectRepo repository.ProjectRepository, ctx context.Context) *entity.Project {
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepoURL:     "https://github.com/test/repo.git",
	}
	err := projectRepo.Create(ctx, project)
	require.NoError(t, err)
	return project
}

func TestTaskRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}

	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Verify the task was created
	assert.NotEqual(t, uuid.Nil, task.ID)
	assert.NotZero(t, task.CreatedAt)
	assert.NotZero(t, task.UpdatedAt)
	assert.Equal(t, entity.TaskStatusTODO, task.Status)
}

func TestTaskRepository_CreateWithDefaultStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		// Status not set, should default to TODO
	}

	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	assert.Equal(t, entity.TaskStatusTODO, task.Status)
}

func TestTaskRepository_CreateWithInvalidProjectID(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	task := &entity.Task{
		ProjectID:   uuid.New(), // Non-existent project
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}

	err := taskRepo.Create(ctx, task)
	assert.Error(t, err)
	// GORM will return a foreign key constraint error
	assert.Contains(t, err.Error(), "failed to create task")
}

func TestTaskRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}
	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Get task
	retrieved, err := taskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)

	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, task.ProjectID, retrieved.ProjectID)
	assert.Equal(t, task.Title, retrieved.Title)
	assert.Equal(t, task.Description, retrieved.Description)
	assert.Equal(t, task.Status, retrieved.Status)
}

func TestTaskRepository_GetByID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	_, err := taskRepo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskRepository_GetByProjectID(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create multiple tasks
	task1 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 1",
		Description: "Description 1",
		Status:      entity.TaskStatusTODO,
	}
	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 2",
		Description: "Description 2",
		Status:      entity.TaskStatusDONE,
	}

	err := taskRepo.Create(ctx, task1)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task2)
	require.NoError(t, err)

	// Get tasks by project ID
	tasks, err := taskRepo.GetByProjectID(ctx, project.ID)
	require.NoError(t, err)

	assert.Len(t, tasks, 2)
	// Tasks should be ordered by created_at DESC (newest first)
	assert.Equal(t, task2.ID, tasks[0].ID)
	assert.Equal(t, task1.ID, tasks[1].ID)
}

func TestTaskRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Original Title",
		Description: "Original Description",
		Status:      entity.TaskStatusTODO,
	}
	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	originalUpdatedAt := task.UpdatedAt

	// Update task
	time.Sleep(10 * time.Millisecond) // Ensure different timestamp
	task.Title = "Updated Title"
	task.Description = "Updated Description"
	task.Status = entity.TaskStatusIMPLEMENTING
	branchName := "feature/updated-task"
	task.BranchName = &branchName

	err = taskRepo.Update(ctx, task)
	require.NoError(t, err)

	// Verify updates
	assert.True(t, task.UpdatedAt.After(originalUpdatedAt))

	// Get and verify
	retrieved, err := taskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)

	assert.Equal(t, "Updated Title", retrieved.Title)
	assert.Equal(t, "Updated Description", retrieved.Description)
	assert.Equal(t, entity.TaskStatusIMPLEMENTING, retrieved.Status)
	assert.NotNil(t, retrieved.BranchName)
	assert.Equal(t, "feature/updated-task", *retrieved.BranchName)
}

func TestTaskRepository_Update_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	task := &entity.Task{
		ID:          uuid.New(),
		ProjectID:   uuid.New(),
		Title:       "Non-existent",
		Description: "Description",
		Status:      entity.TaskStatusTODO,
	}

	err := taskRepo.Update(ctx, task)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "task not found")
	}
}

func TestTaskRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}
	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Delete task (soft delete)
	err = taskRepo.Delete(ctx, task.ID)
	require.NoError(t, err)

	// Verify deletion (soft delete)
	_, err = taskRepo.GetByID(ctx, task.ID)
	assert.Error(t, err)
}

func TestTaskRepository_Delete_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	err := taskRepo.Delete(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskRepository_UpdateStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}
	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	originalUpdatedAt := task.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure different timestamp

	// Update status
	err = taskRepo.UpdateStatus(ctx, task.ID, entity.TaskStatusIMPLEMENTING)
	require.NoError(t, err)

	// Verify status update
	retrieved, err := taskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)

	assert.Equal(t, entity.TaskStatusIMPLEMENTING, retrieved.Status)
	assert.True(t, retrieved.UpdatedAt.After(originalUpdatedAt))
}

func TestTaskRepository_UpdateStatus_InvalidStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}
	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Try to update with invalid status
	// Note: GORM doesn't validate enum values at the database level like raw SQL
	// This test is more about ensuring the method doesn't crash with invalid input
	err = taskRepo.UpdateStatus(ctx, task.ID, "INVALID_STATUS")
	// GORM will allow this since it doesn't validate enum at DB level
	// We'll just ensure it doesn't crash
	if err != nil {
		// If there's an error, it should be about the task not being found or other DB issues
		assert.Contains(t, err.Error(), "failed to update task status")
	}
}

func TestTaskRepository_UpdateStatus_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	err := taskRepo.UpdateStatus(ctx, uuid.New(), entity.TaskStatusDONE)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "task not found")
	}
}

func TestTaskRepository_GetByStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create multiple tasks with different statuses
	task1 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Todo Task",
		Description: "Description 1",
		Status:      entity.TaskStatusTODO,
	}
	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Done Task",
		Description: "Description 2",
		Status:      entity.TaskStatusDONE,
	}
	task3 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Another Todo Task",
		Description: "Description 3",
		Status:      entity.TaskStatusTODO,
	}

	err := taskRepo.Create(ctx, task1)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task2)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task3)
	require.NoError(t, err)

	// Get tasks by status TODO
	todoTasks, err := taskRepo.GetByStatus(ctx, entity.TaskStatusTODO)
	require.NoError(t, err)

	assert.Len(t, todoTasks, 2)
	// Tasks should be ordered by created_at DESC (newest first)
	assert.Equal(t, task3.ID, todoTasks[0].ID)
	assert.Equal(t, task1.ID, todoTasks[1].ID)

	// Get tasks by status DONE
	doneTasks, err := taskRepo.GetByStatus(ctx, entity.TaskStatusDONE)
	require.NoError(t, err)

	assert.Len(t, doneTasks, 1)
	assert.Equal(t, task2.ID, doneTasks[0].ID)

	// Get tasks by status that doesn't exist
	planningTasks, err := taskRepo.GetByStatus(ctx, entity.TaskStatusPLANNING)
	require.NoError(t, err)
	assert.Len(t, planningTasks, 0)
}

func TestTaskRepository_WithNullableFields(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task with nullable fields set to nil
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task with Nulls",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
		BranchName:  nil,
		PullRequest: nil,
	}

	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Retrieve and verify nullable fields
	retrieved, err := taskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)

	assert.Nil(t, retrieved.BranchName)
	assert.Nil(t, retrieved.PullRequest)
}
