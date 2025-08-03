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

func TestTaskRepository_UpdateStatusWithHistory(t *testing.T) {
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

	changedBy := "user123"
	reason := "Testing status update"

	// Update status with history
	err = taskRepo.UpdateStatusWithHistory(ctx, task.ID, entity.TaskStatusPLANNING, &changedBy, &reason)
	require.NoError(t, err)

	// Verify task status was updated
	updatedTask, err := taskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusPLANNING, updatedTask.Status)

	// Verify history was created
	history, err := taskRepo.GetStatusHistory(ctx, task.ID)
	require.NoError(t, err)
	require.Len(t, history, 1)

	historyEntry := history[0]
	assert.Equal(t, task.ID, historyEntry.TaskID)
	assert.Equal(t, entity.TaskStatusTODO, *historyEntry.FromStatus)
	assert.Equal(t, entity.TaskStatusPLANNING, historyEntry.ToStatus)
	assert.Equal(t, changedBy, *historyEntry.ChangedBy)
	assert.Equal(t, reason, *historyEntry.Reason)
}

func TestTaskRepository_UpdateStatusWithHistory_InvalidTransition(t *testing.T) {
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

	// Try invalid status transition
	err = taskRepo.UpdateStatusWithHistory(ctx, task.ID, entity.TaskStatusDONE, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")

	// Verify task status was not updated
	unchangedTask, err := taskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusTODO, unchangedTask.Status)
}

func TestTaskRepository_GetByStatuses(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create tasks with different statuses
	tasks := []*entity.Task{
		{ProjectID: project.ID, Title: "Task 1", Status: entity.TaskStatusTODO},
		{ProjectID: project.ID, Title: "Task 2", Status: entity.TaskStatusPLANNING},
		{ProjectID: project.ID, Title: "Task 3", Status: entity.TaskStatusIMPLEMENTING},
		{ProjectID: project.ID, Title: "Task 4", Status: entity.TaskStatusDONE},
		{ProjectID: project.ID, Title: "Task 5", Status: entity.TaskStatusTODO},
	}

	for _, task := range tasks {
		err := taskRepo.Create(ctx, task)
		require.NoError(t, err)
	}

	// Get tasks with multiple statuses
	statuses := []entity.TaskStatus{entity.TaskStatusTODO, entity.TaskStatusDONE}
	results, err := taskRepo.GetByStatuses(ctx, statuses)
	require.NoError(t, err)

	assert.Len(t, results, 3) // 2 TODO + 1 DONE

	// Verify all returned tasks have the requested statuses
	for _, task := range results {
		assert.Contains(t, statuses, task.Status)
	}
}

func TestTaskRepository_BulkUpdateStatus(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create multiple tasks
	tasks := []*entity.Task{
		{ProjectID: project.ID, Title: "Task 1", Status: entity.TaskStatusTODO},
		{ProjectID: project.ID, Title: "Task 2", Status: entity.TaskStatusTODO},
		{ProjectID: project.ID, Title: "Task 3", Status: entity.TaskStatusTODO},
	}

	var taskIDs []uuid.UUID
	for _, task := range tasks {
		err := taskRepo.Create(ctx, task)
		require.NoError(t, err)
		taskIDs = append(taskIDs, task.ID)
	}

	changedBy := "admin"

	// Bulk update status
	err := taskRepo.BulkUpdateStatus(ctx, taskIDs, entity.TaskStatusPLANNING, &changedBy)
	require.NoError(t, err)

	// Verify all tasks were updated
	for _, taskID := range taskIDs {
		task, err := taskRepo.GetByID(ctx, taskID)
		require.NoError(t, err)
		assert.Equal(t, entity.TaskStatusPLANNING, task.Status)

		// Verify history was created
		history, err := taskRepo.GetStatusHistory(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, history, 1)
		assert.Equal(t, changedBy, *history[0].ChangedBy)
	}
}

func TestTaskRepository_BulkUpdateStatus_PartialInvalidTransitions(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create tasks with different statuses
	tasks := []*entity.Task{
		{ProjectID: project.ID, Title: "Task 1", Status: entity.TaskStatusTODO},
		{ProjectID: project.ID, Title: "Task 2", Status: entity.TaskStatusIMPLEMENTING}, // Can't go to PLANNING
	}

	var taskIDs []uuid.UUID
	for _, task := range tasks {
		err := taskRepo.Create(ctx, task)
		require.NoError(t, err)
		taskIDs = append(taskIDs, task.ID)
	}

	// Try bulk update with invalid transition for one task
	err := taskRepo.BulkUpdateStatus(ctx, taskIDs, entity.TaskStatusPLANNING, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")

	// Verify no tasks were updated (transaction rollback)
	for i, taskID := range taskIDs {
		task, err := taskRepo.GetByID(ctx, taskID)
		require.NoError(t, err)
		assert.Equal(t, tasks[i].Status, task.Status, "Task status should remain unchanged")
	}
}

func TestTaskRepository_GetStatusHistory(t *testing.T) {
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

	// Make several status changes
	statusChanges := []entity.TaskStatus{
		entity.TaskStatusPLANNING,
		entity.TaskStatusPLANREVIEWING,
		entity.TaskStatusIMPLEMENTING,
	}

	for _, status := range statusChanges {
		changedBy := "user123"
		reason := "Progress update"
		err = taskRepo.UpdateStatusWithHistory(ctx, task.ID, status, &changedBy, &reason)
		require.NoError(t, err)
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	// Get status history
	history, err := taskRepo.GetStatusHistory(ctx, task.ID)
	require.NoError(t, err)
	require.Len(t, history, 3)

	// Verify history is ordered by created_at ASC
	expectedStatuses := statusChanges
	for i, entry := range history {
		assert.Equal(t, expectedStatuses[i], entry.ToStatus)
		if i == 0 {
			assert.Equal(t, entity.TaskStatusTODO, *entry.FromStatus)
		} else {
			assert.Equal(t, expectedStatuses[i-1], *entry.FromStatus)
		}
	}
}

func TestTaskRepository_GetStatusAnalytics(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := createTestProject(t, projectRepo, ctx)

	// Create tasks with various statuses
	tasks := []*entity.Task{
		{ProjectID: project.ID, Title: "Task 1", Status: entity.TaskStatusTODO},
		{ProjectID: project.ID, Title: "Task 2", Status: entity.TaskStatusTODO},
		{ProjectID: project.ID, Title: "Task 3", Status: entity.TaskStatusPLANNING},
		{ProjectID: project.ID, Title: "Task 4", Status: entity.TaskStatusIMPLEMENTING},
		{ProjectID: project.ID, Title: "Task 5", Status: entity.TaskStatusDONE},
		{ProjectID: project.ID, Title: "Task 6", Status: entity.TaskStatusDONE},
	}

	for _, task := range tasks {
		err := taskRepo.Create(ctx, task)
		require.NoError(t, err)
	}

	// Make some status transitions to generate transition data
	changedBy := "user123"
	err := taskRepo.UpdateStatusWithHistory(ctx, tasks[0].ID, entity.TaskStatusPLANNING, &changedBy, nil)
	require.NoError(t, err)

	// Get analytics
	analytics, err := taskRepo.GetStatusAnalytics(ctx, project.ID)
	require.NoError(t, err)

	assert.Equal(t, project.ID, analytics.ProjectID)
	assert.Equal(t, 6, analytics.TotalTasks)
	assert.Equal(t, 2, analytics.CompletedTasks)
	assert.InDelta(t, 33.33, analytics.CompletionRate, 0.1)

	// Check status distribution
	statusCounts := make(map[entity.TaskStatus]int)
	for _, stat := range analytics.StatusDistribution {
		statusCounts[stat.Status] = stat.Count
	}

	assert.Equal(t, 1, statusCounts[entity.TaskStatusTODO]) // One was moved to PLANNING
	assert.Equal(t, 2, statusCounts[entity.TaskStatusPLANNING]) // Original + moved from TODO
	assert.Equal(t, 1, statusCounts[entity.TaskStatusIMPLEMENTING])
	assert.Equal(t, 2, statusCounts[entity.TaskStatusDONE])

	// Check transition counts
	assert.Contains(t, analytics.TransitionCount, "TODO->PLANNING")
	assert.Equal(t, 1, analytics.TransitionCount["TODO->PLANNING"])
}

func TestTaskRepository_GetTasksWithFilters(t *testing.T) {
	db, cleanup := setupTestGormDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test projects
	project1 := createTestProject(t, projectRepo, ctx)
	project2 := &entity.Project{
		Name:        "Test Project 2",
		Description: "Test Description 2",
		RepoURL:     "https://github.com/test/repo2.git",
	}
	err := projectRepo.Create(ctx, project2)
	require.NoError(t, err)

	// Create tasks across different projects
	tasks := []*entity.Task{
		{ProjectID: project1.ID, Title: "Authentication Task", Status: entity.TaskStatusTODO},
		{ProjectID: project1.ID, Title: "Database Task", Status: entity.TaskStatusPLANNING},
		{ProjectID: project1.ID, Title: "API Task", Status: entity.TaskStatusDONE},
		{ProjectID: project2.ID, Title: "Frontend Task", Status: entity.TaskStatusTODO},
		{ProjectID: project2.ID, Title: "Testing Task", Status: entity.TaskStatusIMPLEMENTING},
	}

	for _, task := range tasks {
		err := taskRepo.Create(ctx, task)
		require.NoError(t, err)
	}

	// Test filtering by project ID
	t.Run("FilterByProjectID", func(t *testing.T) {
		filters := repository.TaskFilters{
			ProjectID: &project1.ID,
		}
		results, err := taskRepo.GetTasksWithFilters(ctx, filters)
		require.NoError(t, err)
		assert.Len(t, results, 3)
		for _, task := range results {
			assert.Equal(t, project1.ID, task.ProjectID)
		}
	})

	// Test filtering by statuses
	t.Run("FilterByStatuses", func(t *testing.T) {
		filters := repository.TaskFilters{
			Statuses: []entity.TaskStatus{entity.TaskStatusTODO, entity.TaskStatusDONE},
		}
		results, err := taskRepo.GetTasksWithFilters(ctx, filters)
		require.NoError(t, err)
		assert.Len(t, results, 3) // 2 TODO + 1 DONE
		for _, task := range results {
			assert.Contains(t, []entity.TaskStatus{entity.TaskStatusTODO, entity.TaskStatusDONE}, task.Status)
		}
	})

	// Test search term
	t.Run("FilterBySearchTerm", func(t *testing.T) {
		searchTerm := "auth"
		filters := repository.TaskFilters{
			SearchTerm: &searchTerm,
		}
		results, err := taskRepo.GetTasksWithFilters(ctx, filters)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Contains(t, results[0].Title, "Authentication")
	})

	// Test ordering
	t.Run("FilterWithOrdering", func(t *testing.T) {
		orderBy := "title"
		orderDir := "ASC"
		filters := repository.TaskFilters{
			ProjectID: &project1.ID,
			OrderBy:   &orderBy,
			OrderDir:  &orderDir,
		}
		results, err := taskRepo.GetTasksWithFilters(ctx, filters)
		require.NoError(t, err)
		assert.Len(t, results, 3)
		
		// Should be ordered by title ASC
		titles := []string{results[0].Title, results[1].Title, results[2].Title}
		assert.Equal(t, "API Task", titles[0])
		assert.Equal(t, "Authentication Task", titles[1])
		assert.Equal(t, "Database Task", titles[2])
	})

	// Test pagination
	t.Run("FilterWithPagination", func(t *testing.T) {
		limit := 2
		offset := 0
		filters := repository.TaskFilters{
			Limit:  &limit,
			Offset: &offset,
		}
		results, err := taskRepo.GetTasksWithFilters(ctx, filters)
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	// Test combined filters
	t.Run("FilterCombined", func(t *testing.T) {
		searchTerm := "task"
		limit := 10
		filters := repository.TaskFilters{
			ProjectID:  &project1.ID,
			Statuses:   []entity.TaskStatus{entity.TaskStatusTODO, entity.TaskStatusPLANNING},
			SearchTerm: &searchTerm,
			Limit:      &limit,
		}
		results, err := taskRepo.GetTasksWithFilters(ctx, filters)
		require.NoError(t, err)
		assert.Len(t, results, 2) // Authentication + Database tasks
		for _, task := range results {
			assert.Equal(t, project1.ID, task.ProjectID)
			assert.Contains(t, []entity.TaskStatus{entity.TaskStatusTODO, entity.TaskStatusPLANNING}, task.Status)
		}
	})
}