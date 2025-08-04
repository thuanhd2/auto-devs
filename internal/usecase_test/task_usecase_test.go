package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test setup function
func setupTaskUsecaseTest() (usecase.TaskUsecase, *repository.TaskRepositoryMock, *repository.ProjectRepositoryMock, *usecase.NotificationUsecaseMock, *usecase.WorktreeUsecaseMock) {
	mockTaskRepo := &repository.TaskRepositoryMock{}
	mockProjectRepo := &repository.ProjectRepositoryMock{}
	mockNotificationUsecase := &usecase.NotificationUsecaseMock{}
	mockWorktreeUsecase := &usecase.WorktreeUsecaseMock{}

	// Tạo usecase instance
	taskUsecase := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase, mockWorktreeUsecase)

	return taskUsecase, mockTaskRepo, mockProjectRepo, mockNotificationUsecase, mockWorktreeUsecase
}

func TestTaskUsecase_Create(t *testing.T) {
	t.Run("successful task creation", func(t *testing.T) {
		taskUsecase, mockTaskRepo, mockProjectRepo, mockNotificationUsecase, _ := setupTaskUsecaseTest()
		ctx := context.Background()

		// Arrange
		projectID := uuid.New()
		req := usecase.CreateTaskRequest{
			ProjectID:   projectID,
			Title:       "Test Task",
			Description: "Test Description",
			Priority:    entity.TaskPriorityMedium,
		}

		// Mock project validation
		mockTaskRepo.On("ValidateProjectExists", ctx, projectID).Return(true, nil)

		// Mock duplicate title check
		mockTaskRepo.On("CheckDuplicateTitle", ctx, projectID, req.Title, (*uuid.UUID)(nil)).Return(false, nil)

		// Mock task creation
		mockTaskRepo.On("Create", ctx, mock.AnythingOfType("*entity.Task")).Return(nil)

		// Mock project retrieval for notification
		mockProjectRepo.On("GetByID", ctx, projectID).Return(&entity.Project{
			ID:   projectID,
			Name: "Test Project",
		}, nil)

		// Mock notification
		mockNotificationUsecase.On("SendTaskCreatedNotification", ctx, mock.AnythingOfType("*entity.Task"), mock.AnythingOfType("*entity.Project")).Return(nil)

		// Act
		result, err := taskUsecase.Create(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Title, result.Title)
		assert.Equal(t, req.Description, result.Description)
		assert.Equal(t, req.Priority, result.Priority)
		assert.Equal(t, entity.TaskStatusTODO, result.Status)

		// Verify mocks
		mockTaskRepo.AssertExpectations(t)
		mockProjectRepo.AssertExpectations(t)
	})

	t.Run("project not found", func(t *testing.T) {
		taskUsecase, mockTaskRepo, _, _, _ := setupTaskUsecaseTest()
		ctx := context.Background()

		// Arrange
		projectID := uuid.New()
		req := usecase.CreateTaskRequest{
			ProjectID: projectID,
			Title:     "Test Task",
		}

		// Mock project validation failure
		mockTaskRepo.On("ValidateProjectExists", ctx, projectID).Return(false, nil)

		// Act
		result, err := taskUsecase.Create(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "project not found")

		// Verify mocks
		mockTaskRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		taskUsecase, mockTaskRepo, _, _, _ := setupTaskUsecaseTest()
		ctx := context.Background()

		// Arrange
		projectID := uuid.New()
		req := usecase.CreateTaskRequest{
			ProjectID: projectID,
			Title:     "Test Task",
		}

		// Mock project validation
		mockTaskRepo.On("ValidateProjectExists", ctx, projectID).Return(true, nil)

		// Mock duplicate title check
		mockTaskRepo.On("CheckDuplicateTitle", ctx, projectID, req.Title, (*uuid.UUID)(nil)).Return(false, nil)

		// Mock repository error
		mockTaskRepo.On("Create", ctx, mock.AnythingOfType("*entity.Task")).Return(errors.New("database error"))

		// Act
		result, err := taskUsecase.Create(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database error")

		// Verify mocks
		mockTaskRepo.AssertExpectations(t)
	})
}

func TestTaskUsecase_GetByID(t *testing.T) {
	taskUsecase, mockTaskRepo, _, _, _ := setupTaskUsecaseTest()
	ctx := context.Background()

	t.Run("successful get task by ID", func(t *testing.T) {
		// Arrange
		taskID := uuid.New()
		expectedTask := &entity.Task{
			ID:    taskID,
			Title: "Test Task",
		}

		// Mock task retrieval
		mockTaskRepo.On("GetByID", ctx, taskID).Return(expectedTask, nil)

		// Act
		result, err := taskUsecase.GetByID(ctx, taskID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedTask, result)

		// Verify mocks
		mockTaskRepo.AssertExpectations(t)
	})

	t.Run("task not found", func(t *testing.T) {
		// Arrange
		taskID := uuid.New()

		// Mock task not found
		mockTaskRepo.On("GetByID", ctx, taskID).Return((*entity.Task)(nil), errors.New("task not found"))

		// Act
		result, err := taskUsecase.GetByID(ctx, taskID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)

		// Verify mocks
		mockTaskRepo.AssertExpectations(t)
	})
}

func TestTaskUsecase_UpdateStatus(t *testing.T) {
	t.Run("successful status update", func(t *testing.T) {
		taskUsecase, mockTaskRepo, _, _, _ := setupTaskUsecaseTest()
		ctx := context.Background()

		// Arrange
		taskID := uuid.New()
		newStatus := entity.TaskStatusIMPLEMENTING

		// Mock status update
		mockTaskRepo.On("UpdateStatus", ctx, taskID, newStatus).Return(nil)
		mockTaskRepo.On("GetByID", ctx, taskID).Return(&entity.Task{
			ID:    taskID,
			Title: "Test Task",
		}, nil)

		// Act
		result, err := taskUsecase.UpdateStatus(ctx, taskID, newStatus)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify mocks
		mockTaskRepo.AssertExpectations(t)
	})

	t.Run("update status error", func(t *testing.T) {
		taskUsecase, mockTaskRepo, _, _, _ := setupTaskUsecaseTest()
		ctx := context.Background()

		// Arrange
		taskID := uuid.New()
		newStatus := entity.TaskStatusIMPLEMENTING

		// Mock update error
		mockTaskRepo.On("UpdateStatus", ctx, taskID, newStatus).Return(errors.New("update failed"))

		// Act
		result, err := taskUsecase.UpdateStatus(ctx, taskID, newStatus)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)

		// Verify mocks
		mockTaskRepo.AssertExpectations(t)
	})
}

func TestTaskUsecase_Delete(t *testing.T) {
	t.Run("successful task deletion", func(t *testing.T) {
		taskUsecase, mockTaskRepo, _, _, _ := setupTaskUsecaseTest()
		ctx := context.Background()

		// Arrange
		taskID := uuid.New()

		// Mock task deletion
		mockTaskRepo.On("Delete", ctx, taskID).Return(nil)

		// Act
		err := taskUsecase.Delete(ctx, taskID)

		// Assert
		assert.NoError(t, err)

		// Verify mocks
		mockTaskRepo.AssertExpectations(t)
	})

	t.Run("delete error", func(t *testing.T) {
		taskUsecase, mockTaskRepo, _, _, _ := setupTaskUsecaseTest()
		ctx := context.Background()

		// Arrange
		taskID := uuid.New()

		// Mock delete error
		mockTaskRepo.On("Delete", ctx, taskID).Return(errors.New("delete failed"))

		// Act
		err := taskUsecase.Delete(ctx, taskID)

		// Assert
		assert.Error(t, err)

		// Verify mocks
		mockTaskRepo.AssertExpectations(t)
	})
}

func TestTaskUsecase_GetByProjectID(t *testing.T) {
	t.Run("successful get tasks by project ID", func(t *testing.T) {
		taskUsecase, mockTaskRepo, _, _, _ := setupTaskUsecaseTest()
		ctx := context.Background()

		// Arrange
		projectID := uuid.New()
		expectedTasks := []*entity.Task{
			{ID: uuid.New(), Title: "Task 1", ProjectID: projectID},
			{ID: uuid.New(), Title: "Task 2", ProjectID: projectID},
		}

		// Mock task retrieval
		mockTaskRepo.On("GetByProjectID", ctx, projectID).Return(expectedTasks, nil)

		// Act
		result, err := taskUsecase.GetByProjectID(ctx, projectID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedTasks, result)

		// Verify mocks
		mockTaskRepo.AssertExpectations(t)
	})

	t.Run("get tasks error", func(t *testing.T) {
		taskUsecase, mockTaskRepo, _, _, _ := setupTaskUsecaseTest()
		ctx := context.Background()

		// Arrange
		projectID := uuid.New()

		// Mock retrieval error
		mockTaskRepo.On("GetByProjectID", ctx, projectID).Return(([]*entity.Task)(nil), errors.New("database error"))

		// Act
		result, err := taskUsecase.GetByProjectID(ctx, projectID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)

		// Verify mocks
		mockTaskRepo.AssertExpectations(t)
	})
}
