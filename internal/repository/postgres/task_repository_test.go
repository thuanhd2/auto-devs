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

func createTestProject(t *testing.T, projectRepo *projectRepository, ctx context.Context) *entity.Project {
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
	db, cleanup := setupTestDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db).(*projectRepository)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTodo,
	}

	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Verify the task was created
	assert.NotEqual(t, uuid.Nil, task.ID)
	assert.NotZero(t, task.CreatedAt)
	assert.NotZero(t, task.UpdatedAt)
	assert.Equal(t, entity.TaskStatusTodo, task.Status)
}

func TestTaskRepository_CreateWithDefaultStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db).(*projectRepository)
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

	assert.Equal(t, entity.TaskStatusTodo, task.Status)
}

func TestTaskRepository_CreateWithInvalidProjectID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	task := &entity.Task{
		ProjectID:   uuid.New(), // Non-existent project
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTodo,
	}

	err := taskRepo.Create(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

func TestTaskRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db).(*projectRepository)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTodo,
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
	db, cleanup := setupTestDB(t)
	defer cleanup()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	_, err := taskRepo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskRepository_GetByProjectID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db).(*projectRepository)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create multiple tasks
	task1 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 1",
		Description: "Description 1",
		Status:      entity.TaskStatusTodo,
	}
	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 2",
		Description: "Description 2",
		Status:      entity.TaskStatusDone,
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
	db, cleanup := setupTestDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db).(*projectRepository)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Original Title",
		Description: "Original Description",
		Status:      entity.TaskStatusTodo,
	}
	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	originalUpdatedAt := task.UpdatedAt

	// Update task
	time.Sleep(10 * time.Millisecond) // Ensure different timestamp
	task.Title = "Updated Title"
	task.Description = "Updated Description"
	task.Status = entity.TaskStatusImplementing
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
	assert.Equal(t, entity.TaskStatusImplementing, retrieved.Status)
	assert.NotNil(t, retrieved.BranchName)
	assert.Equal(t, "feature/updated-task", *retrieved.BranchName)
}

func TestTaskRepository_Update_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	task := &entity.Task{
		ID:          uuid.New(),
		ProjectID:   uuid.New(),
		Title:       "Non-existent",
		Description: "Description",
		Status:      entity.TaskStatusTodo,
	}

	err := taskRepo.Update(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db).(*projectRepository)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTodo,
	}
	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Delete task
	err = taskRepo.Delete(ctx, task.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = taskRepo.GetByID(ctx, task.ID)
	assert.Error(t, err)
}

func TestTaskRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	err := taskRepo.Delete(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskRepository_UpdateStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db).(*projectRepository)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTodo,
	}
	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	originalUpdatedAt := task.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure different timestamp

	// Update status
	err = taskRepo.UpdateStatus(ctx, task.ID, entity.TaskStatusImplementing)
	require.NoError(t, err)

	// Verify status update
	retrieved, err := taskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)

	assert.Equal(t, entity.TaskStatusImplementing, retrieved.Status)
	assert.True(t, retrieved.UpdatedAt.After(originalUpdatedAt))
}

func TestTaskRepository_UpdateStatus_InvalidStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db).(*projectRepository)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTodo,
	}
	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Try to update with invalid status
	err = taskRepo.UpdateStatus(ctx, task.ID, "INVALID_STATUS")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid task status")
}

func TestTaskRepository_UpdateStatus_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	err := taskRepo.UpdateStatus(ctx, uuid.New(), entity.TaskStatusDone)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskRepository_GetByStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db).(*projectRepository)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create multiple tasks with different statuses
	task1 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Todo Task",
		Description: "Description 1",
		Status:      entity.TaskStatusTodo,
	}
	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Done Task",
		Description: "Description 2",
		Status:      entity.TaskStatusDone,
	}
	task3 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Another Todo Task",
		Description: "Description 3",
		Status:      entity.TaskStatusTodo,
	}

	err := taskRepo.Create(ctx, task1)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task2)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task3)
	require.NoError(t, err)

	// Get tasks by status TODO
	todoTasks, err := taskRepo.GetByStatus(ctx, entity.TaskStatusTodo)
	require.NoError(t, err)

	assert.Len(t, todoTasks, 2)
	// Tasks should be ordered by created_at DESC (newest first)
	assert.Equal(t, task3.ID, todoTasks[0].ID)
	assert.Equal(t, task1.ID, todoTasks[1].ID)

	// Get tasks by status DONE
	doneTasks, err := taskRepo.GetByStatus(ctx, entity.TaskStatusDone)
	require.NoError(t, err)

	assert.Len(t, doneTasks, 1)
	assert.Equal(t, task2.ID, doneTasks[0].ID)

	// Get tasks by status that doesn't exist
	planningTasks, err := taskRepo.GetByStatus(ctx, entity.TaskStatusPlanning)
	require.NoError(t, err)
	assert.Len(t, planningTasks, 0)
}

func TestTaskRepository_WithNullableFields(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db).(*projectRepository)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task with nullable fields set to nil
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task with Nulls",
		Description: "Test Description",
		Status:      entity.TaskStatusTodo,
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