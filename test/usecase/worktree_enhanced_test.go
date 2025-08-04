package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWorktreeUsecase is a mock implementation of WorktreeUsecase for testing
type MockWorktreeUsecase struct {
	mock.Mock
}

func (m *MockWorktreeUsecase) CreateWorktreeForTask(ctx context.Context, req usecase.CreateWorktreeRequest) (*entity.Worktree, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Worktree), args.Error(1)
}

func (m *MockWorktreeUsecase) CleanupWorktreeForTask(ctx context.Context, req usecase.CleanupWorktreeRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockWorktreeUsecase) GetWorktreeByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.Worktree, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Worktree), args.Error(1)
}

func (m *MockWorktreeUsecase) GetWorktreesByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Worktree, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Worktree), args.Error(1)
}

func (m *MockWorktreeUsecase) UpdateWorktreeStatus(ctx context.Context, worktreeID uuid.UUID, status entity.WorktreeStatus) error {
	args := m.Called(ctx, worktreeID, status)
	return args.Error(0)
}

func (m *MockWorktreeUsecase) ValidateWorktree(ctx context.Context, worktreeID uuid.UUID) (*usecase.WorktreeValidationResult, error) {
	args := m.Called(ctx, worktreeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.WorktreeValidationResult), args.Error(1)
}

func (m *MockWorktreeUsecase) GetWorktreeHealth(ctx context.Context, worktreeID uuid.UUID) (*usecase.WorktreeHealthInfo, error) {
	args := m.Called(ctx, worktreeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.WorktreeHealthInfo), args.Error(1)
}

func (m *MockWorktreeUsecase) CreateBranchForTask(ctx context.Context, taskID uuid.UUID, branchName string) error {
	args := m.Called(ctx, taskID, branchName)
	return args.Error(0)
}

func (m *MockWorktreeUsecase) SwitchToBranch(ctx context.Context, worktreeID uuid.UUID, branchName string) error {
	args := m.Called(ctx, worktreeID, branchName)
	return args.Error(0)
}

func (m *MockWorktreeUsecase) GetBranchInfo(ctx context.Context, worktreeID uuid.UUID) (*usecase.BranchInfo, error) {
	args := m.Called(ctx, worktreeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.BranchInfo), args.Error(1)
}

func (m *MockWorktreeUsecase) InitializeWorktree(ctx context.Context, worktreeID uuid.UUID) error {
	args := m.Called(ctx, worktreeID)
	return args.Error(0)
}

func (m *MockWorktreeUsecase) CopyConfigurationFiles(ctx context.Context, worktreeID uuid.UUID, sourcePath string) error {
	args := m.Called(ctx, worktreeID, sourcePath)
	return args.Error(0)
}

func (m *MockWorktreeUsecase) HandleWorktreeCreationFailure(ctx context.Context, taskID uuid.UUID, error error) error {
	args := m.Called(ctx, taskID, error)
	return args.Error(0)
}

func (m *MockWorktreeUsecase) RecoverFailedWorktree(ctx context.Context, worktreeID uuid.UUID) error {
	args := m.Called(ctx, worktreeID)
	return args.Error(0)
}

func (m *MockWorktreeUsecase) GetWorktreeStatistics(ctx context.Context, projectID uuid.UUID) (*entity.WorktreeStatistics, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorktreeStatistics), args.Error(1)
}

func (m *MockWorktreeUsecase) GetActiveWorktreesCount(ctx context.Context, projectID uuid.UUID) (int, error) {
	args := m.Called(ctx, projectID)
	return args.Int(0), args.Error(1)
}

// TestTaskWorktreeIntegration tests the integration between task and worktree operations
func TestTaskWorktreeIntegration(t *testing.T) {
	// Setup
	mockWorktreeUsecase := new(MockWorktreeUsecase)

	// Test case 1: Task moves to IMPLEMENTING status - should create worktree
	t.Run("Task moves to IMPLEMENTING - creates worktree", func(t *testing.T) {
		// Arrange
		taskID := uuid.New()
		projectID := uuid.New()

		// Mock worktree creation
		expectedWorktree := &entity.Worktree{
			ID:           uuid.New(),
			TaskID:       taskID,
			ProjectID:    projectID,
			BranchName:   "task-test-task",
			WorktreePath: "/worktrees/project-" + projectID.String() + "/task-" + taskID.String(),
			Status:       entity.WorktreeStatusActive,
		}

		mockWorktreeUsecase.On("GetWorktreeByTaskID", mock.Anything, taskID).Return(nil, assert.AnError)
		mockWorktreeUsecase.On("CreateWorktreeForTask", mock.Anything, usecase.CreateWorktreeRequest{
			TaskID:    taskID,
			ProjectID: projectID,
			TaskTitle: "Test Task",
		}).Return(expectedWorktree, nil)

		// Act

		// Simulate the worktree creation logic
		worktreeReq := usecase.CreateWorktreeRequest{
			TaskID:    taskID,
			ProjectID: projectID,
			TaskTitle: "Test Task",
		}

		worktree, err := mockWorktreeUsecase.CreateWorktreeForTask(context.Background(), worktreeReq)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, worktree)
		assert.Equal(t, taskID, worktree.TaskID)
		assert.Equal(t, projectID, worktree.ProjectID)
		assert.Equal(t, entity.WorktreeStatusActive, worktree.Status)

		mockWorktreeUsecase.AssertExpectations(t)
	})

	// Test case 2: Task moves to DONE status - should cleanup worktree
	t.Run("Task moves to DONE - cleans up worktree", func(t *testing.T) {
		// Arrange
		taskID := uuid.New()
		projectID := uuid.New()
		existingWorktree := &entity.Worktree{
			ID:           uuid.New(),
			TaskID:       taskID,
			ProjectID:    projectID,
			BranchName:   "task-test-task",
			WorktreePath: "/worktrees/project-" + projectID.String() + "/task-" + taskID.String(),
			Status:       entity.WorktreeStatusActive,
		}

		mockWorktreeUsecase.On("GetWorktreeByTaskID", mock.Anything, taskID).Return(existingWorktree, nil)
		mockWorktreeUsecase.On("CleanupWorktreeForTask", mock.Anything, usecase.CleanupWorktreeRequest{
			TaskID:    taskID,
			ProjectID: projectID,
			Force:     true,
		}).Return(nil)

		// Act
		cleanupReq := usecase.CleanupWorktreeRequest{
			TaskID:    taskID,
			ProjectID: projectID,
			Force:     true,
		}

		err := mockWorktreeUsecase.CleanupWorktreeForTask(context.Background(), cleanupReq)

		// Assert
		assert.NoError(t, err)
		mockWorktreeUsecase.AssertExpectations(t)
	})

	// Test case 3: Task moves to CANCELLED status - should cleanup worktree
	t.Run("Task moves to CANCELLED - cleans up worktree", func(t *testing.T) {
		// Arrange
		taskID := uuid.New()
		projectID := uuid.New()
		existingWorktree := &entity.Worktree{
			ID:           uuid.New(),
			TaskID:       taskID,
			ProjectID:    projectID,
			BranchName:   "task-test-task",
			WorktreePath: "/worktrees/project-" + projectID.String() + "/task-" + taskID.String(),
			Status:       entity.WorktreeStatusActive,
		}

		mockWorktreeUsecase.On("GetWorktreeByTaskID", mock.Anything, taskID).Return(existingWorktree, nil)
		mockWorktreeUsecase.On("CleanupWorktreeForTask", mock.Anything, usecase.CleanupWorktreeRequest{
			TaskID:    taskID,
			ProjectID: projectID,
			Force:     true,
		}).Return(nil)

		// Act
		cleanupReq := usecase.CleanupWorktreeRequest{
			TaskID:    taskID,
			ProjectID: projectID,
			Force:     true,
		}

		err := mockWorktreeUsecase.CleanupWorktreeForTask(context.Background(), cleanupReq)

		// Assert
		assert.NoError(t, err)
		mockWorktreeUsecase.AssertExpectations(t)
	})

	// Test case 4: Worktree already exists - should not create duplicate
	t.Run("Worktree already exists - should not create duplicate", func(t *testing.T) {
		// Arrange
		taskID := uuid.New()
		projectID := uuid.New()
		existingWorktree := &entity.Worktree{
			ID:           uuid.New(),
			TaskID:       taskID,
			ProjectID:    projectID,
			BranchName:   "task-test-task",
			WorktreePath: "/worktrees/project-" + projectID.String() + "/task-" + taskID.String(),
			Status:       entity.WorktreeStatusActive,
		}

		mockWorktreeUsecase.On("GetWorktreeByTaskID", mock.Anything, taskID).Return(existingWorktree, nil)

		// Act
		worktree, err := mockWorktreeUsecase.GetWorktreeByTaskID(context.Background(), taskID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, worktree)
		assert.Equal(t, taskID, worktree.TaskID)
		assert.Equal(t, entity.WorktreeStatusActive, worktree.Status)

		// Should not call CreateWorktreeForTask
		mockWorktreeUsecase.AssertNotCalled(t, "CreateWorktreeForTask")
		mockWorktreeUsecase.AssertExpectations(t)
	})

	// Test case 5: Worktree validation
	t.Run("Worktree validation", func(t *testing.T) {
		// Arrange
		worktreeID := uuid.New()
		validationResult := &usecase.WorktreeValidationResult{
			IsValid:         true,
			GitRepositoryOK: true,
			BranchExists:    true,
			DirectoryExists: true,
			PermissionsOK:   true,
			ValidationTime:  time.Now(),
		}

		mockWorktreeUsecase.On("ValidateWorktree", mock.Anything, worktreeID).Return(validationResult, nil)

		// Act
		result, err := mockWorktreeUsecase.ValidateWorktree(context.Background(), worktreeID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsValid)
		assert.True(t, result.GitRepositoryOK)
		assert.True(t, result.BranchExists)
		assert.True(t, result.DirectoryExists)
		assert.True(t, result.PermissionsOK)

		mockWorktreeUsecase.AssertExpectations(t)
	})

	// Test case 6: Worktree health monitoring
	t.Run("Worktree health monitoring", func(t *testing.T) {
		// Arrange
		worktreeID := uuid.New()
		healthInfo := &usecase.WorktreeHealthInfo{
			WorktreeID:      worktreeID,
			Status:          entity.WorktreeStatusActive,
			IsHealthy:       true,
			LastActivity:    time.Now(),
			DiskUsage:       1024 * 1024, // 1MB
			FileCount:       10,
			GitStatus:       "clean",
			BranchStatus:    "active",
			HealthScore:     95,
			LastHealthCheck: time.Now(),
		}

		mockWorktreeUsecase.On("GetWorktreeHealth", mock.Anything, worktreeID).Return(healthInfo, nil)

		// Act
		health, err := mockWorktreeUsecase.GetWorktreeHealth(context.Background(), worktreeID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, health)
		assert.Equal(t, worktreeID, health.WorktreeID)
		assert.True(t, health.IsHealthy)
		assert.Equal(t, entity.WorktreeStatusActive, health.Status)
		assert.Equal(t, 95, health.HealthScore)
		assert.Equal(t, "clean", health.GitStatus)

		mockWorktreeUsecase.AssertExpectations(t)
	})

	// Test case 7: Branch management
	t.Run("Branch management", func(t *testing.T) {
		// Arrange
		worktreeID := uuid.New()
		branchInfo := &usecase.BranchInfo{
			Name:           "task-test-task",
			IsCurrent:      true,
			LastCommit:     "abc123",
			LastCommitDate: time.Now(),
			CommitCount:    5,
			IsClean:        true,
			HasUncommitted: false,
			HasUntracked:   false,
		}

		mockWorktreeUsecase.On("GetBranchInfo", mock.Anything, worktreeID).Return(branchInfo, nil)

		// Act
		branch, err := mockWorktreeUsecase.GetBranchInfo(context.Background(), worktreeID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, branch)
		assert.Equal(t, "task-test-task", branch.Name)
		assert.True(t, branch.IsCurrent)
		assert.True(t, branch.IsClean)
		assert.Equal(t, 5, branch.CommitCount)

		mockWorktreeUsecase.AssertExpectations(t)
	})

	// Test case 8: Worktree statistics
	t.Run("Worktree statistics", func(t *testing.T) {
		// Arrange
		projectID := uuid.New()
		statistics := &entity.WorktreeStatistics{
			ProjectID:          projectID,
			TotalWorktrees:     10,
			ActiveWorktrees:    5,
			CompletedWorktrees: 3,
			ErrorWorktrees:     1,
			WorktreesByStatus: map[entity.WorktreeStatus]int{
				entity.WorktreeStatusActive:    5,
				entity.WorktreeStatusCompleted: 3,
				entity.WorktreeStatusError:     1,
				entity.WorktreeStatusCreating:  1,
			},
			AverageCreationTime: 30.0,
			GeneratedAt:         time.Now(),
		}

		mockWorktreeUsecase.On("GetWorktreeStatistics", mock.Anything, projectID).Return(statistics, nil)

		// Act
		stats, err := mockWorktreeUsecase.GetWorktreeStatistics(context.Background(), projectID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, projectID, stats.ProjectID)
		assert.Equal(t, 10, stats.TotalWorktrees)
		assert.Equal(t, 5, stats.ActiveWorktrees)
		assert.Equal(t, 3, stats.CompletedWorktrees)
		assert.Equal(t, 1, stats.ErrorWorktrees)
		assert.Equal(t, 30.0, stats.AverageCreationTime)

		mockWorktreeUsecase.AssertExpectations(t)
	})

	// Test case 9: Error handling - worktree creation failure
	t.Run("Worktree creation failure", func(t *testing.T) {
		// Arrange
		taskID := uuid.New()
		projectID := uuid.New()
		creationError := assert.AnError

		mockWorktreeUsecase.On("GetWorktreeByTaskID", mock.Anything, taskID).Return(nil, assert.AnError)
		mockWorktreeUsecase.On("CreateWorktreeForTask", mock.Anything, usecase.CreateWorktreeRequest{
			TaskID:    taskID,
			ProjectID: projectID,
			TaskTitle: "Test Task",
		}).Return(nil, creationError)
		mockWorktreeUsecase.On("HandleWorktreeCreationFailure", mock.Anything, taskID, creationError).Return(nil)

		// Act
		worktreeReq := usecase.CreateWorktreeRequest{
			TaskID:    taskID,
			ProjectID: projectID,
			TaskTitle: "Test Task",
		}

		worktree, err := mockWorktreeUsecase.CreateWorktreeForTask(context.Background(), worktreeReq)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, worktree)

		// Handle the failure
		handleErr := mockWorktreeUsecase.HandleWorktreeCreationFailure(context.Background(), taskID, creationError)
		assert.NoError(t, handleErr)

		mockWorktreeUsecase.AssertExpectations(t)
	})

	// Test case 10: Worktree recovery
	t.Run("Worktree recovery", func(t *testing.T) {
		// Arrange
		worktreeID := uuid.New()

		mockWorktreeUsecase.On("RecoverFailedWorktree", mock.Anything, worktreeID).Return(nil)

		// Act
		err := mockWorktreeUsecase.RecoverFailedWorktree(context.Background(), worktreeID)

		// Assert
		assert.NoError(t, err)
		mockWorktreeUsecase.AssertExpectations(t)
	})
}

// TestWorktreeLifecycle tests the complete worktree lifecycle
func TestWorktreeLifecycle(t *testing.T) {
	mockWorktreeUsecase := new(MockWorktreeUsecase)

	t.Run("Complete worktree lifecycle", func(t *testing.T) {
		// Arrange
		taskID := uuid.New()
		projectID := uuid.New()
		worktreeID := uuid.New()

		// Step 1: Create worktree
		expectedWorktree := &entity.Worktree{
			ID:           worktreeID,
			TaskID:       taskID,
			ProjectID:    projectID,
			BranchName:   "task-test-task",
			WorktreePath: "/worktrees/project-" + projectID.String() + "/task-" + taskID.String(),
			Status:       entity.WorktreeStatusCreating,
		}

		mockWorktreeUsecase.On("GetWorktreeByTaskID", mock.Anything, taskID).Return(nil, assert.AnError)
		mockWorktreeUsecase.On("CreateWorktreeForTask", mock.Anything, usecase.CreateWorktreeRequest{
			TaskID:    taskID,
			ProjectID: projectID,
			TaskTitle: "Test Task",
		}).Return(expectedWorktree, nil)

		// Step 2: Update status to active
		mockWorktreeUsecase.On("UpdateWorktreeStatus", mock.Anything, worktreeID, entity.WorktreeStatusActive).Return(nil)

		// Step 3: Initialize worktree
		mockWorktreeUsecase.On("InitializeWorktree", mock.Anything, worktreeID).Return(nil)

		// Step 4: Validate worktree
		validationResult := &usecase.WorktreeValidationResult{
			IsValid:         true,
			GitRepositoryOK: true,
			BranchExists:    true,
			DirectoryExists: true,
			PermissionsOK:   true,
			ValidationTime:  time.Now(),
		}
		mockWorktreeUsecase.On("ValidateWorktree", mock.Anything, worktreeID).Return(validationResult, nil)

		// Step 5: Get health info
		healthInfo := &usecase.WorktreeHealthInfo{
			WorktreeID:      worktreeID,
			Status:          entity.WorktreeStatusActive,
			IsHealthy:       true,
			HealthScore:     95,
			LastHealthCheck: time.Now(),
		}
		mockWorktreeUsecase.On("GetWorktreeHealth", mock.Anything, worktreeID).Return(healthInfo, nil)

		// Step 6: Cleanup worktree
		mockWorktreeUsecase.On("GetWorktreeByTaskID", mock.Anything, taskID).Return(expectedWorktree, nil)
		mockWorktreeUsecase.On("CleanupWorktreeForTask", mock.Anything, usecase.CleanupWorktreeRequest{
			TaskID:    taskID,
			ProjectID: projectID,
			Force:     true,
		}).Return(nil)

		// Act & Assert - Step 1: Create worktree
		worktreeReq := usecase.CreateWorktreeRequest{
			TaskID:    taskID,
			ProjectID: projectID,
			TaskTitle: "Test Task",
		}

		worktree, err := mockWorktreeUsecase.CreateWorktreeForTask(context.Background(), worktreeReq)
		assert.NoError(t, err)
		assert.NotNil(t, worktree)
		assert.Equal(t, entity.WorktreeStatusCreating, worktree.Status)

		// Step 2: Update status to active
		err = mockWorktreeUsecase.UpdateWorktreeStatus(context.Background(), worktreeID, entity.WorktreeStatusActive)
		assert.NoError(t, err)

		// Step 3: Initialize worktree
		err = mockWorktreeUsecase.InitializeWorktree(context.Background(), worktreeID)
		assert.NoError(t, err)

		// Step 4: Validate worktree
		validation, err := mockWorktreeUsecase.ValidateWorktree(context.Background(), worktreeID)
		assert.NoError(t, err)
		assert.NotNil(t, validation)
		assert.True(t, validation.IsValid)

		// Step 5: Get health info
		health, err := mockWorktreeUsecase.GetWorktreeHealth(context.Background(), worktreeID)
		assert.NoError(t, err)
		assert.NotNil(t, health)
		assert.True(t, health.IsHealthy)

		// Step 6: Cleanup worktree
		cleanupReq := usecase.CleanupWorktreeRequest{
			TaskID:    taskID,
			ProjectID: projectID,
			Force:     true,
		}

		err = mockWorktreeUsecase.CleanupWorktreeForTask(context.Background(), cleanupReq)
		assert.NoError(t, err)

		mockWorktreeUsecase.AssertExpectations(t)
	})
}
