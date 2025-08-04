package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTaskUsecase_Create(t *testing.T) {
	ctx := context.Background()
	
	t.Run("successful creation", func(t *testing.T) {
		// Setup mocks
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		notificationUsecase := &MockNotificationUsecase{}
		
		projectFactory := testutil.NewProjectFactory()
		project := projectFactory.CreateProject()
		
		// Mock expectations
		taskRepo.On("Create", ctx, mock.AnythingOfType("*entity.Task")).
			Run(func(args mock.Arguments) {
				task := args.Get(1).(*entity.Task)
				// Simulate what the repository would do
				task.CreatedAt = time.Now()
				task.UpdatedAt = time.Now()
			}).Return(nil)
		
		projectRepo.On("GetByID", ctx, project.ID).Return(project, nil)
		notificationUsecase.On("SendTaskCreatedNotification", ctx, mock.AnythingOfType("*entity.Task"), project).Return(nil)
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, notificationUsecase)
		
		// Execute
		req := CreateTaskRequest{
			ProjectID:   project.ID,
			Title:       "Test Task",
			Description: "Test Description",
		}
		
		result, err := usecase.Create(ctx, req)
		
		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEqual(t, uuid.Nil, result.ID)
		assert.Equal(t, req.ProjectID, result.ProjectID)
		assert.Equal(t, req.Title, result.Title)
		assert.Equal(t, req.Description, result.Description)
		assert.Equal(t, entity.TaskStatusTODO, result.Status)
		assert.NotZero(t, result.CreatedAt)
		assert.NotZero(t, result.UpdatedAt)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
		projectRepo.AssertExpectations(t)
		notificationUsecase.AssertExpectations(t)
	})
	
	t.Run("repository creation failure", func(t *testing.T) {
		// Setup mocks
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		notificationUsecase := &MockNotificationUsecase{}
		
		projectID := uuid.New()
		
		// Mock expectations
		taskRepo.On("Create", ctx, mock.AnythingOfType("*entity.Task")).Return(errors.New("database error"))
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, notificationUsecase)
		
		// Execute
		req := CreateTaskRequest{
			ProjectID:   projectID,
			Title:       "Test Task",
			Description: "Test Description",
		}
		
		result, err := usecase.Create(ctx, req)
		
		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database error")
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
	
	t.Run("notification failure doesn't affect task creation", func(t *testing.T) {
		// Setup mocks
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		notificationUsecase := &MockNotificationUsecase{}
		
		projectFactory := testutil.NewProjectFactory()
		project := projectFactory.CreateProject()
		
		// Mock expectations
		taskRepo.On("Create", ctx, mock.AnythingOfType("*entity.Task")).Return(nil)
		projectRepo.On("GetByID", ctx, project.ID).Return(project, nil)
		notificationUsecase.On("SendTaskCreatedNotification", ctx, mock.AnythingOfType("*entity.Task"), project).
			Return(errors.New("notification service down"))
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, notificationUsecase)
		
		// Execute
		req := CreateTaskRequest{
			ProjectID:   project.ID,
			Title:       "Test Task",
			Description: "Test Description",
		}
		
		result, err := usecase.Create(ctx, req)
		
		// Assertions - task creation should still succeed
		require.NoError(t, err)
		assert.NotNil(t, result)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
		projectRepo.AssertExpectations(t)
		notificationUsecase.AssertExpectations(t)
	})
	
	t.Run("nil notification usecase", func(t *testing.T) {
		// Setup mocks
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		projectID := uuid.New()
		
		// Mock expectations
		taskRepo.On("Create", ctx, mock.AnythingOfType("*entity.Task")).Return(nil)
		
		// Create usecase with nil notification usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		req := CreateTaskRequest{
			ProjectID:   projectID,
			Title:       "Test Task",
			Description: "Test Description",
		}
		
		result, err := usecase.Create(ctx, req)
		
		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, result)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
}

func TestTaskUsecase_GetByID(t *testing.T) {
	ctx := context.Background()
	
	t.Run("successful retrieval", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskFactory := testutil.NewTaskFactory()
		task := taskFactory.CreateTask()
		
		// Mock expectations
		taskRepo.On("GetByID", ctx, task.ID).Return(task, nil)
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		result, err := usecase.GetByID(ctx, task.ID)
		
		// Assertions
		require.NoError(t, err)
		assert.Equal(t, task, result)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
	
	t.Run("task not found", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskID := uuid.New()
		
		// Mock expectations
		taskRepo.On("GetByID", ctx, taskID).Return(nil, errors.New("task not found"))
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		result, err := usecase.GetByID(ctx, taskID)
		
		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "task not found")
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
}

func TestTaskUsecase_Update(t *testing.T) {
	ctx := context.Background()
	
	t.Run("successful update all fields", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskFactory := testutil.NewTaskFactory()
		originalTask := taskFactory.CreateTask()
		
		// Mock expectations
		taskRepo.On("GetByID", ctx, originalTask.ID).Return(originalTask, nil)
		taskRepo.On("Update", ctx, mock.AnythingOfType("*entity.Task")).
			Run(func(args mock.Arguments) {
				task := args.Get(1).(*entity.Task)
				assert.Equal(t, "Updated Title", task.Title)
				assert.Equal(t, "Updated Description", task.Description)
				assert.NotNil(t, task.BranchName)
				assert.Equal(t, "feature/updated", *task.BranchName)
				assert.NotNil(t, task.PullRequest)
				assert.Equal(t, "PR-123", *task.PullRequest)
			}).Return(nil)
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		req := UpdateTaskRequest{
			Title:       "Updated Title",
			Description: "Updated Description",
			BranchName:  "feature/updated",
			PullRequest: "PR-123",
		}
		
		result, err := usecase.Update(ctx, originalTask.ID, req)
		
		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, result)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
	
	t.Run("partial update", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskFactory := testutil.NewTaskFactory()
		originalTask := taskFactory.CreateTask()
		
		// Mock expectations
		taskRepo.On("GetByID", ctx, originalTask.ID).Return(originalTask, nil)
		taskRepo.On("Update", ctx, mock.AnythingOfType("*entity.Task")).
			Run(func(args mock.Arguments) {
				task := args.Get(1).(*entity.Task)
				assert.Equal(t, "Updated Title", task.Title)
				// Description should remain unchanged
				assert.Equal(t, originalTask.Description, task.Description)
			}).Return(nil)
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute - only update title
		req := UpdateTaskRequest{
			Title: "Updated Title",
			// Other fields empty, should not be updated
		}
		
		result, err := usecase.Update(ctx, originalTask.ID, req)
		
		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, result)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
	
	t.Run("task not found", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskID := uuid.New()
		
		// Mock expectations
		taskRepo.On("GetByID", ctx, taskID).Return(nil, errors.New("task not found"))
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		req := UpdateTaskRequest{
			Title: "Updated Title",
		}
		
		result, err := usecase.Update(ctx, taskID, req)
		
		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "task not found")
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
	
	t.Run("update fails", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskFactory := testutil.NewTaskFactory()
		task := taskFactory.CreateTask()
		
		// Mock expectations
		taskRepo.On("GetByID", ctx, task.ID).Return(task, nil)
		taskRepo.On("Update", ctx, mock.AnythingOfType("*entity.Task")).Return(errors.New("database error"))
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		req := UpdateTaskRequest{
			Title: "Updated Title",
		}
		
		result, err := usecase.Update(ctx, task.ID, req)
		
		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database error")
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
}

func TestTaskUsecase_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	
	t.Run("successful status update", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskFactory := testutil.NewTaskFactory()
		task := taskFactory.CreateTask(func(t *entity.Task) {
			t.Status = entity.TaskStatusIMPLEMENTING
		})
		
		// Mock expectations
		taskRepo.On("UpdateStatus", ctx, task.ID, entity.TaskStatusIMPLEMENTING).Return(nil)
		taskRepo.On("GetByID", ctx, task.ID).Return(task, nil)
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		result, err := usecase.UpdateStatus(ctx, task.ID, entity.TaskStatusIMPLEMENTING)
		
		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, entity.TaskStatusIMPLEMENTING, result.Status)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
	
	t.Run("status update fails", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskID := uuid.New()
		
		// Mock expectations
		taskRepo.On("UpdateStatus", ctx, taskID, entity.TaskStatusDONE).Return(errors.New("status update failed"))
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		result, err := usecase.UpdateStatus(ctx, taskID, entity.TaskStatusDONE)
		
		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "status update failed")
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
}

func TestTaskUsecase_UpdateStatusWithHistory(t *testing.T) {
	ctx := context.Background()
	
	t.Run("successful status update with history", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		notificationUsecase := &MockNotificationUsecase{}
		
		taskFactory := testutil.NewTaskFactory()
		projectFactory := testutil.NewProjectFactory()
		
		task := taskFactory.CreateTask(func(t *entity.Task) {
			t.Status = entity.TaskStatusTODO
		})
		project := projectFactory.CreateProject(func(p *entity.Project) {
			p.ID = task.ProjectID
		})
		
		updatedTask := taskFactory.CreateTask(func(t *entity.Task) {
			t.ID = task.ID
			t.ProjectID = task.ProjectID
			t.Status = entity.TaskStatusIMPLEMENTING
		})
		
		changedBy := "user-123"
		reason := "Starting implementation"
		
		// Mock expectations
		taskRepo.On("GetByID", ctx, task.ID).Return(task, nil)
		taskRepo.On("UpdateStatusWithHistory", ctx, task.ID, entity.TaskStatusIMPLEMENTING, &changedBy, &reason).Return(nil)
		taskRepo.On("GetByID", ctx, task.ID).Return(updatedTask, nil)
		projectRepo.On("GetByID", ctx, task.ProjectID).Return(project, nil)
		notificationUsecase.On("SendTaskStatusChangeNotification", ctx, mock.AnythingOfType("entity.TaskStatusChangeNotificationData")).Return(nil)
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, notificationUsecase)
		
		// Execute
		req := UpdateStatusRequest{
			TaskID:    task.ID,
			Status:    entity.TaskStatusIMPLEMENTING,
			ChangedBy: &changedBy,
			Reason:    &reason,
		}
		
		result, err := usecase.UpdateStatusWithHistory(ctx, req)
		
		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, entity.TaskStatusIMPLEMENTING, result.Status)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
		projectRepo.AssertExpectations(t)
		notificationUsecase.AssertExpectations(t)
	})
	
	t.Run("invalid status transition", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskFactory := testutil.NewTaskFactory()
		task := taskFactory.CreateTask(func(t *entity.Task) {
			t.Status = entity.TaskStatusDONE
		})
		
		// Mock expectations
		taskRepo.On("GetByID", ctx, task.ID).Return(task, nil)
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute - try to go from DONE back to TODO (invalid transition)
		req := UpdateStatusRequest{
			TaskID: task.ID,
			Status: entity.TaskStatusTODO,
		}
		
		result, err := usecase.UpdateStatusWithHistory(ctx, req)
		
		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
}

func TestTaskUsecase_GetByStatuses(t *testing.T) {
	ctx := context.Background()
	
	t.Run("successful retrieval with valid statuses", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskFactory := testutil.NewTaskFactory()
		tasks := []*entity.Task{
			taskFactory.CreateTask(func(t *entity.Task) { t.Status = entity.TaskStatusTODO }),
			taskFactory.CreateTask(func(t *entity.Task) { t.Status = entity.TaskStatusDONE }),
		}
		
		statuses := []entity.TaskStatus{entity.TaskStatusTODO, entity.TaskStatusDONE}
		
		// Mock expectations
		taskRepo.On("GetByStatuses", ctx, statuses).Return(tasks, nil)
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		result, err := usecase.GetByStatuses(ctx, statuses)
		
		// Assertions
		require.NoError(t, err)
		assert.Len(t, result, 2)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
	
	t.Run("invalid status", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		statuses := []entity.TaskStatus{"INVALID_STATUS"}
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		result, err := usecase.GetByStatuses(ctx, statuses)
		
		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid status")
	})
}

func TestTaskUsecase_BulkUpdateStatus(t *testing.T) {
	ctx := context.Background()
	
	t.Run("successful bulk update", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskIDs := []uuid.UUID{uuid.New(), uuid.New()}
		changedBy := "user-123"
		
		// Mock expectations
		taskRepo.On("BulkUpdateStatus", ctx, taskIDs, entity.TaskStatusDONE, &changedBy).Return(nil)
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		req := BulkUpdateStatusRequest{
			TaskIDs:   taskIDs,
			Status:    entity.TaskStatusDONE,
			ChangedBy: &changedBy,
		}
		
		err := usecase.BulkUpdateStatus(ctx, req)
		
		// Assertions
		require.NoError(t, err)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
	
	t.Run("empty task IDs", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		req := BulkUpdateStatusRequest{
			TaskIDs: []uuid.UUID{},
			Status:  entity.TaskStatusDONE,
		}
		
		err := usecase.BulkUpdateStatus(ctx, req)
		
		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no task IDs provided")
	})
	
	t.Run("invalid status", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		req := BulkUpdateStatusRequest{
			TaskIDs: []uuid.UUID{uuid.New()},
			Status:  "INVALID_STATUS",
		}
		
		err := usecase.BulkUpdateStatus(ctx, req)
		
		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid target status")
	})
}

func TestTaskUsecase_ValidateStatusTransition(t *testing.T) {
	ctx := context.Background()
	
	t.Run("valid transition", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskFactory := testutil.NewTaskFactory()
		task := taskFactory.CreateTask(func(t *entity.Task) {
			t.Status = entity.TaskStatusTODO
		})
		
		// Mock expectations
		taskRepo.On("GetByID", ctx, task.ID).Return(task, nil)
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		err := usecase.ValidateStatusTransition(ctx, task.ID, entity.TaskStatusPLANNING)
		
		// Assertions
		require.NoError(t, err)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
	
	t.Run("invalid transition", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskFactory := testutil.NewTaskFactory()
		task := taskFactory.CreateTask(func(t *entity.Task) {
			t.Status = entity.TaskStatusDONE
		})
		
		// Mock expectations
		taskRepo.On("GetByID", ctx, task.ID).Return(task, nil)
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		err := usecase.ValidateStatusTransition(ctx, task.ID, entity.TaskStatusTODO)
		
		// Assertions
		assert.Error(t, err)
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
	
	t.Run("task not found", func(t *testing.T) {
		// Setup
		taskRepo := &testutil.MockTaskRepository{}
		projectRepo := &testutil.MockProjectRepository{}
		
		taskID := uuid.New()
		
		// Mock expectations
		taskRepo.On("GetByID", ctx, taskID).Return(nil, errors.New("task not found"))
		
		// Create usecase
		usecase := NewTaskUsecase(taskRepo, projectRepo, nil)
		
		// Execute
		err := usecase.ValidateStatusTransition(ctx, taskID, entity.TaskStatusDONE)
		
		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get task")
		
		// Verify mocks
		taskRepo.AssertExpectations(t)
	})
}

// MockNotificationUsecase is a mock implementation for testing
type MockNotificationUsecase struct {
	mock.Mock
}

func (m *MockNotificationUsecase) SendTaskCreatedNotification(ctx context.Context, task *entity.Task, project *entity.Project) error {
	args := m.Called(ctx, task, project)
	return args.Error(0)
}

func (m *MockNotificationUsecase) SendTaskStatusChangeNotification(ctx context.Context, data entity.TaskStatusChangeNotificationData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}