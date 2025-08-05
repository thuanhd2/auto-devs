package postgres

import (
	"context"
	"testing"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Using CreateTestProject from test_helper.go

func TestTaskRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := CreateTestProject(t, projectRepo, ctx)

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
	project := CreateTestProject(t, projectRepo, ctx)

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
	project := CreateTestProject(t, projectRepo, ctx)

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
	assert.Equal(t, task.Title, retrieved.Title)
	assert.Equal(t, task.Description, retrieved.Description)
	assert.Equal(t, task.Status, retrieved.Status)
	assert.Equal(t, task.ProjectID, retrieved.ProjectID)
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
	project := CreateTestProject(t, projectRepo, ctx)

	// Create multiple tasks
	task1 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task 1",
		Description: "Test Description 1",
		Status:      entity.TaskStatusTODO,
	}

	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task 2",
		Description: "Test Description 2",
		Status:      entity.TaskStatusIMPLEMENTING,
	}

	err := taskRepo.Create(ctx, task1)
	require.NoError(t, err)

	err = taskRepo.Create(ctx, task2)
	require.NoError(t, err)

	// Get tasks by project ID
	tasks, err := taskRepo.GetByProjectID(ctx, project.ID)
	require.NoError(t, err)

	assert.Len(t, tasks, 2)
	// Check that both tasks are returned
	taskIDs := make(map[uuid.UUID]bool)
	for _, task := range tasks {
		taskIDs[task.ID] = true
	}
	assert.True(t, taskIDs[task1.ID])
	assert.True(t, taskIDs[task2.ID])
}

func TestTaskRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := CreateTestProject(t, projectRepo, ctx)

	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}

	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Update task
	task.Title = "Updated Task"
	task.Description = "Updated Description"
	task.Status = entity.TaskStatusIMPLEMENTING

	err = taskRepo.Update(ctx, task)
	require.NoError(t, err)

	// Verify the update
	updated, err := taskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)

	assert.Equal(t, "Updated Task", updated.Title)
	assert.Equal(t, "Updated Description", updated.Description)
	assert.Equal(t, entity.TaskStatusIMPLEMENTING, updated.Status)
}

func TestTaskRepository_Update_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	task := &entity.Task{
		ID:          uuid.New(),
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}

	err := taskRepo.Update(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := CreateTestProject(t, projectRepo, ctx)

	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}

	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Delete task
	err = taskRepo.Delete(ctx, task.ID)
	require.NoError(t, err)

	// Verify the task is deleted (soft delete)
	_, err = taskRepo.GetByID(ctx, task.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
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
	project := CreateTestProject(t, projectRepo, ctx)

	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}

	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Update status
	err = taskRepo.UpdateStatus(ctx, task.ID, entity.TaskStatusIMPLEMENTING)
	require.NoError(t, err)

	// Verify the status update
	updated, err := taskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusIMPLEMENTING, updated.Status)
}

func TestTaskRepository_UpdateStatus_InvalidStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := CreateTestProject(t, projectRepo, ctx)

	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}

	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Try to update with invalid status
	err = taskRepo.UpdateStatus(ctx, task.ID, "INVALID_STATUS")
	assert.Error(t, err)
}

func TestTaskRepository_UpdateStatus_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	err := taskRepo.UpdateStatus(ctx, uuid.New(), entity.TaskStatusIMPLEMENTING)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskRepository_GetByStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := CreateTestProject(t, projectRepo, ctx)

	// Create tasks with different statuses
	task1 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task 1",
		Description: "Test Description 1",
		Status:      entity.TaskStatusTODO,
	}

	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task 2",
		Description: "Test Description 2",
		Status:      entity.TaskStatusIMPLEMENTING,
	}

	task3 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task 3",
		Description: "Test Description 3",
		Status:      entity.TaskStatusTODO,
	}

	err := taskRepo.Create(ctx, task1)
	require.NoError(t, err)

	err = taskRepo.Create(ctx, task2)
	require.NoError(t, err)

	err = taskRepo.Create(ctx, task3)
	require.NoError(t, err)

	// Get tasks by status
	todoTasks, err := taskRepo.GetByStatus(ctx, entity.TaskStatusTODO)
	require.NoError(t, err)
	assert.Len(t, todoTasks, 2)

	inProgressTasks, err := taskRepo.GetByStatus(ctx, entity.TaskStatusIMPLEMENTING)
	require.NoError(t, err)
	assert.Len(t, inProgressTasks, 1)

	doneTasks, err := taskRepo.GetByStatus(ctx, entity.TaskStatusDONE)
	require.NoError(t, err)
	assert.Len(t, doneTasks, 0)
}

func TestTaskRepository_WithNullableFields(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := CreateTestProject(t, projectRepo, ctx)

	// Create task with nullable fields
	task := &entity.Task{
		ProjectID: project.ID,
		Title:     "Test Task",
		// Description is nullable, so we can omit it
		Status: entity.TaskStatusTODO,
	}

	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Verify the task was created
	assert.NotEqual(t, uuid.Nil, task.ID)
	assert.Equal(t, "Test Task", task.Title)
	assert.Empty(t, task.Description) // Should be empty string, not nil
	assert.Equal(t, entity.TaskStatusTODO, task.Status)
}
