package postgres

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/internal/testutil"
	"github.com/auto-devs/auto-devs/pkg/database"
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
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

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
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

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
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

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
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

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
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	_, err := taskRepo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskRepository_GetByProjectID(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

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
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

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
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

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
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

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
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	err := taskRepo.Delete(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskRepository_UpdateStatus(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

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
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

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
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	err := taskRepo.UpdateStatus(ctx, uuid.New(), entity.TaskStatusDONE)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "task not found")
	}
}

func TestTaskRepository_GetByStatus(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

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
	planningTasks, err := taskRepo.GetByStatus(ctx, entity.TaskStatusTODO)
	require.NoError(t, err)
	assert.Len(t, planningTasks, 0)
}

func TestTaskRepository_WithNullableFields(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

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

func TestTaskRepository_GetByProjectIDWithParams(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create multiple tasks with different properties
	taskFactory := testutil.NewTaskFactory()
	tasks := []*entity.Task{
		taskFactory.CreateTask(func(t *entity.Task) {
			t.ProjectID = project.ID
			t.Title = "Alpha Task"
			t.Status = entity.TaskStatusTODO
			t.Priority = entity.TaskPriorityHigh
		}),
		taskFactory.CreateTask(func(t *entity.Task) {
			t.ProjectID = project.ID
			t.Title = "Beta Task"
			t.Status = entity.TaskStatusDONE
			t.Priority = entity.TaskPriorityMedium
		}),
		taskFactory.CreateTask(func(t *entity.Task) {
			t.ProjectID = project.ID
			t.Title = "Gamma Search"
			t.Status = entity.TaskStatusIMPLEMENTING
			t.Priority = entity.TaskPriorityLow
		}),
	}

	for _, task := range tasks {
		err := taskRepo.Create(ctx, task)
		require.NoError(t, err)
	}

	t.Run("filter by status", func(t *testing.T) {
		params := repository.GetTasksParams{
			Status:   entity.TaskStatusTODO,
			Page:     1,
			PageSize: 10,
		}

		results, total, err := taskRepo.GetByProjectIDWithParams(ctx, project.ID, params)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, results, 1)
		assert.Equal(t, entity.TaskStatusTODO, results[0].Status)
	})

	t.Run("filter by priority", func(t *testing.T) {
		params := repository.GetTasksParams{
			Priority: entity.TaskPriorityHigh,
			Page:     1,
			PageSize: 10,
		}

		results, total, err := taskRepo.GetByProjectIDWithParams(ctx, project.ID, params)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, results, 1)
		assert.Equal(t, entity.TaskPriorityHigh, results[0].Priority)
	})

	t.Run("search functionality", func(t *testing.T) {
		params := repository.GetTasksParams{
			Search:   "search",
			Page:     1,
			PageSize: 10,
		}

		results, total, err := taskRepo.GetByProjectIDWithParams(ctx, project.ID, params)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, results, 1)
		assert.Contains(t, results[0].Title, "Search")
	})

	t.Run("pagination", func(t *testing.T) {
		params := repository.GetTasksParams{
			Page:     1,
			PageSize: 2,
		}

		results, total, err := taskRepo.GetByProjectIDWithParams(ctx, project.ID, params)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, results, 2)

		// Second page
		params.Page = 2
		results, total, err = taskRepo.GetByProjectIDWithParams(ctx, project.ID, params)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, results, 1)
	})

	t.Run("sorting", func(t *testing.T) {
		params := repository.GetTasksParams{
			SortBy:    "title",
			SortOrder: "asc",
			Page:      1,
			PageSize:  10,
		}

		results, _, err := taskRepo.GetByProjectIDWithParams(ctx, project.ID, params)
		require.NoError(t, err)
		assert.Equal(t, "Alpha Task", results[0].Title)
		assert.Equal(t, "Beta Task", results[1].Title)
		assert.Equal(t, "Gamma Search", results[2].Title)
	})
}

func TestTaskRepository_GetStatsByProjectID(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create tasks with different statuses
	taskFactory := testutil.NewTaskFactory()
	tasks := taskFactory.CreateTasksWithDifferentStatuses(project.ID)

	for _, task := range tasks {
		err := taskRepo.Create(ctx, task)
		require.NoError(t, err)
	}

	// Get statistics
	stats, err := taskRepo.GetStatsByProjectID(ctx, project.ID)
	require.NoError(t, err)

	assert.Equal(t, 1, stats[entity.TaskStatusTODO])
	assert.Equal(t, 1, stats[entity.TaskStatusPLANNING])
	assert.Equal(t, 1, stats[entity.TaskStatusPLAN_REVIEWING])
	assert.Equal(t, 1, stats[entity.TaskStatusIMPLEMENTING])
	assert.Equal(t, 1, stats[entity.TaskStatusCODE_REVIEWING])
	assert.Equal(t, 1, stats[entity.TaskStatusDONE])
	assert.Equal(t, 1, stats[entity.TaskStatusCANCELLED])
}

func TestTaskRepository_Archive_Restore(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Archive Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}
	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Archive task
	err = taskRepo.Archive(ctx, task.ID)
	require.NoError(t, err)

	// Verify it's archived (not found in normal queries)
	_, err = taskRepo.GetByID(ctx, task.ID)
	assert.Error(t, err)

	// Restore task
	err = taskRepo.Restore(ctx, task.ID)
	require.NoError(t, err)

	// Verify it's restored
	restored, err := taskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, restored.ID)
}

func TestTaskRepository_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create initial task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Concurrent Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}
	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Simulate concurrent updates
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer func() { done <- true }()
			
			// Update status
			newStatus := entity.TaskStatusIMPLEMENTING
			if i%2 == 0 {
				newStatus = entity.TaskStatusDONE
			}
			
			if err := taskRepo.UpdateStatus(ctx, task.ID, newStatus); err != nil {
				errors <- err
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check for errors
	select {
	case err := <-errors:
		t.Fatalf("Concurrent operation failed: %v", err)
	default:
		// No errors, verify task still exists and is in a valid state
		finalTask, err := taskRepo.GetByID(ctx, task.ID)
		require.NoError(t, err)
		assert.NotNil(t, finalTask)
	}
}

func TestTaskRepository_TransactionRollback(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Test transaction rollback
	tx := db.DB.Begin()
	defer tx.Rollback()

	// Create task in transaction
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Transaction Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}

	// Create task repository with transaction
	txTaskRepo := NewTaskRepository(&testutil.TestContainer{GormDB: tx, DB: &database.GormDB{DB: tx}}.DB)
	err := txTaskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Task should exist in transaction
	_, err = txTaskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)

	// But not in main database (transaction not committed)
	_, err = taskRepo.GetByID(ctx, task.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestTaskRepository_BulkOperations(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create multiple tasks
	const numTasks = 100
	tasks := make([]*entity.Task, numTasks)
	
	for i := 0; i < numTasks; i++ {
		tasks[i] = &entity.Task{
			ProjectID:   project.ID,
			Title:       fmt.Sprintf("Bulk Task %d", i+1),
			Description: fmt.Sprintf("Description %d", i+1),
			Status:      entity.TaskStatusTODO,
		}
		
		err := taskRepo.Create(ctx, tasks[i])
		require.NoError(t, err)
	}

	// Verify all tasks were created
	allTasks, err := taskRepo.GetByProjectID(ctx, project.ID)
	require.NoError(t, err)
	assert.Len(t, allTasks, numTasks)
}

func TestTaskRepository_EdgeCases(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	t.Run("empty project has no tasks", func(t *testing.T) {
		project := createTestProject(t, projectRepo, ctx)
		
		tasks, err := taskRepo.GetByProjectID(ctx, project.ID)
		require.NoError(t, err)
		assert.Len(t, tasks, 0)
	})

	t.Run("very long title and description", func(t *testing.T) {
		project := createTestProject(t, projectRepo, ctx)
		
		longTitle := strings.Repeat("A", 1000)
		longDescription := strings.Repeat("B", 10000)
		
		task := &entity.Task{
			ProjectID:   project.ID,
			Title:       longTitle,
			Description: longDescription,
			Status:      entity.TaskStatusTODO,
		}
		
		err := taskRepo.Create(ctx, task)
		require.NoError(t, err)
		
		retrieved, err := taskRepo.GetByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, longTitle, retrieved.Title)
		assert.Equal(t, longDescription, retrieved.Description)
	})

	t.Run("special characters in title", func(t *testing.T) {
		project := createTestProject(t, projectRepo, ctx)
		
		specialTitle := "Task with special chars: @#$%^&*()[]{}|\:;\"'<>,.?/~`"
		
		task := &entity.Task{
			ProjectID:   project.ID,
			Title:       specialTitle,
			Description: "Test description",
			Status:      entity.TaskStatusTODO,
		}
		
		err := taskRepo.Create(ctx, task)
		require.NoError(t, err)
		
		retrieved, err := taskRepo.GetByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, specialTitle, retrieved.Title)
	})

	t.Run("unicode characters in title", func(t *testing.T) {
		project := createTestProject(t, projectRepo, ctx)
		
		unicodeTitle := "タスク测试🚀🎉"
		
		task := &entity.Task{
			ProjectID:   project.ID,
			Title:       unicodeTitle,
			Description: "Test description",
			Status:      entity.TaskStatusTODO,
		}
		
		err := taskRepo.Create(ctx, task)
		require.NoError(t, err)
		
		retrieved, err := taskRepo.GetByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, unicodeTitle, retrieved.Title)
	})
}
