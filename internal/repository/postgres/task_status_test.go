package postgres

import (
	"context"
	"testing"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskRepository_UpdateStatusWithHistory(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := CreateTestProject(t, projectRepo, ctx)

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
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := CreateTestProject(t, projectRepo, ctx)

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
		Title:       "Task 1",
		Description: "Description 1",
		Status:      entity.TaskStatusTODO,
	}
	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 2",
		Description: "Description 2",
		Status:      entity.TaskStatusPLANNING,
	}
	task3 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 3",
		Description: "Description 3",
		Status:      entity.TaskStatusIMPLEMENTING,
	}

	err := taskRepo.Create(ctx, task1)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task2)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task3)
	require.NoError(t, err)

	// Get tasks by multiple statuses
	statuses := []entity.TaskStatus{entity.TaskStatusTODO, entity.TaskStatusPLANNING}
	tasks, err := taskRepo.GetByStatuses(ctx, statuses)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)

	// Verify we got the expected tasks
	taskIDs := make(map[uuid.UUID]bool)
	for _, task := range tasks {
		taskIDs[task.ID] = true
	}
	assert.True(t, taskIDs[task1.ID])
	assert.True(t, taskIDs[task2.ID])
	assert.False(t, taskIDs[task3.ID])
}

func TestTaskRepository_BulkUpdateStatus(t *testing.T) {
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
		Title:       "Task 1",
		Description: "Description 1",
		Status:      entity.TaskStatusTODO,
	}
	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 2",
		Description: "Description 2",
		Status:      entity.TaskStatusTODO,
	}
	task3 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 3",
		Description: "Description 3",
		Status:      entity.TaskStatusTODO,
	}

	err := taskRepo.Create(ctx, task1)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task2)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task3)
	require.NoError(t, err)

	// Bulk update status
	taskIDs := []uuid.UUID{task1.ID, task2.ID, task3.ID}
	changedBy := "admin"
	err = taskRepo.BulkUpdateStatus(ctx, taskIDs, entity.TaskStatusPLANNING, &changedBy)
	require.NoError(t, err)

	// Verify all tasks were updated
	for _, taskID := range taskIDs {
		updatedTask, err := taskRepo.GetByID(ctx, taskID)
		require.NoError(t, err)
		assert.Equal(t, entity.TaskStatusPLANNING, updatedTask.Status)
	}

	// Verify history was created for all tasks
	for _, taskID := range taskIDs {
		history, err := taskRepo.GetStatusHistory(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, history, 1)
		assert.Equal(t, entity.TaskStatusTODO, *history[0].FromStatus)
		assert.Equal(t, entity.TaskStatusPLANNING, history[0].ToStatus)
		assert.Equal(t, changedBy, *history[0].ChangedBy)
	}
}

func TestTaskRepository_BulkUpdateStatus_PartialInvalidTransitions(t *testing.T) {
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
		Title:       "Task 1",
		Description: "Description 1",
		Status:      entity.TaskStatusTODO,
	}
	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 2",
		Description: "Description 2",
		Status:      entity.TaskStatusIMPLEMENTING,
	}

	err := taskRepo.Create(ctx, task1)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task2)
	require.NoError(t, err)

	// Try bulk update with invalid transition for one task
	taskIDs := []uuid.UUID{task1.ID, task2.ID}
	changedBy := "admin"
	err = taskRepo.BulkUpdateStatus(ctx, taskIDs, entity.TaskStatusDONE, &changedBy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")

	// Verify no tasks were updated
	unchangedTask1, err := taskRepo.GetByID(ctx, task1.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusTODO, unchangedTask1.Status)

	unchangedTask2, err := taskRepo.GetByID(ctx, task2.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusIMPLEMENTING, unchangedTask2.Status)
}

func TestTaskRepository_GetStatusHistory(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := CreateTestProject(t, projectRepo, ctx)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}
	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Update status multiple times
	changedBy := "user123"
	reason1 := "Starting planning"
	err = taskRepo.UpdateStatusWithHistory(ctx, task.ID, entity.TaskStatusPLANNING, &changedBy, &reason1)
	require.NoError(t, err)

	reason11 := "Finish planning, wait for planning review"
	err = taskRepo.UpdateStatusWithHistory(ctx, task.ID, entity.TaskStatusPLANREVIEWING, &changedBy, &reason11)
	require.NoError(t, err)

	reason2 := "Plan approved, starting implementation"
	err = taskRepo.UpdateStatusWithHistory(ctx, task.ID, entity.TaskStatusIMPLEMENTING, &changedBy, &reason2)
	require.NoError(t, err)

	// Get status history
	history, err := taskRepo.GetStatusHistory(ctx, task.ID)
	require.NoError(t, err)
	assert.Len(t, history, 3)

	// Verify history order (newest first)
	assert.Equal(t, entity.TaskStatusPLANNING, history[0].ToStatus)
	assert.Equal(t, entity.TaskStatusTODO, *history[0].FromStatus)
	assert.Equal(t, reason1, *history[0].Reason)

	assert.Equal(t, entity.TaskStatusPLANREVIEWING, history[1].ToStatus)
	assert.Equal(t, entity.TaskStatusPLANNING, *history[1].FromStatus)
	assert.Equal(t, reason11, *history[1].Reason)

	assert.Equal(t, entity.TaskStatusIMPLEMENTING, history[2].ToStatus)
	assert.Equal(t, entity.TaskStatusPLANREVIEWING, *history[2].FromStatus)
	assert.Equal(t, reason2, *history[2].Reason)
}

func TestTaskRepository_GetStatusAnalytics(t *testing.T) {
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
		Title:       "Task 1",
		Description: "Description 1",
		Status:      entity.TaskStatusTODO,
	}
	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 2",
		Description: "Description 2",
		Status:      entity.TaskStatusPLANNING,
	}
	task3 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 3",
		Description: "Description 3",
		Status:      entity.TaskStatusIMPLEMENTING,
	}

	err := taskRepo.Create(ctx, task1)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task2)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task3)
	require.NoError(t, err)

	// Get status analytics
	analytics, err := taskRepo.GetStatusAnalytics(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, project.ID, analytics.ProjectID)
	assert.Equal(t, 3, analytics.TotalTasks)
	assert.Equal(t, 0, analytics.CompletedTasks)

	// Verify status distribution
	statusCounts := make(map[entity.TaskStatus]int)
	for _, stat := range analytics.StatusDistribution {
		statusCounts[stat.Status] = stat.Count
	}
	assert.Equal(t, 1, statusCounts[entity.TaskStatusTODO])
	assert.Equal(t, 1, statusCounts[entity.TaskStatusPLANNING])
	assert.Equal(t, 1, statusCounts[entity.TaskStatusIMPLEMENTING])
}

func TestTaskRepository_GetTasksWithFilters(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create test project
	project := CreateTestProject(t, projectRepo, ctx)

	// Create tasks with different attributes
	task1 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "High Priority Task",
		Description: "Important task",
		Status:      entity.TaskStatusTODO,
		Priority:    entity.TaskPriorityHigh,
	}
	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Low Priority Task",
		Description: "Less important task",
		Status:      entity.TaskStatusPLANNING,
		Priority:    entity.TaskPriorityLow,
	}

	err := taskRepo.Create(ctx, task1)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task2)
	require.NoError(t, err)

	// Test filtering by status
	filters := entity.TaskFilters{
		ProjectID: &project.ID,
		Statuses:  []entity.TaskStatus{entity.TaskStatusTODO},
	}
	tasks, err := taskRepo.GetTasksWithFilters(ctx, filters)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, task1.ID, tasks[0].ID)

	// Test filtering by priority
	filters = entity.TaskFilters{
		ProjectID:  &project.ID,
		Priorities: []entity.TaskPriority{entity.TaskPriorityHigh},
	}
	tasks, err = taskRepo.GetTasksWithFilters(ctx, filters)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.Equal(t, task2.ID, tasks[0].ID)
	assert.Equal(t, task1.ID, tasks[1].ID)

	// Test filtering by multiple criteria
	filters = entity.TaskFilters{
		ProjectID:  &project.ID,
		Statuses:   []entity.TaskStatus{entity.TaskStatusTODO, entity.TaskStatusPLANNING},
		Priorities: []entity.TaskPriority{entity.TaskPriorityHigh, entity.TaskPriorityLow},
	}
	tasks, err = taskRepo.GetTasksWithFilters(ctx, filters)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}
