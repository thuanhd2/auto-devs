package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTaskRepository implements TaskRepository interface for testing
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, task *entity.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *entity.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TaskStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockTaskRepository) UpdateStatusWithHistory(ctx context.Context, id uuid.UUID, status entity.TaskStatus, changedBy *string, reason *string) error {
	args := m.Called(ctx, id, status, changedBy, reason)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByStatus(ctx context.Context, status entity.TaskStatus) ([]*entity.Task, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) GetByStatuses(ctx context.Context, statuses []entity.TaskStatus) ([]*entity.Task, error) {
	args := m.Called(ctx, statuses)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.TaskStatus, changedBy *string) error {
	args := m.Called(ctx, ids, status, changedBy)
	return args.Error(0)
}

func (m *MockTaskRepository) GetStatusHistory(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskStatusHistory, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).([]*entity.TaskStatusHistory), args.Error(1)
}

func (m *MockTaskRepository) GetStatusAnalytics(ctx context.Context, projectID uuid.UUID) (*entity.TaskStatusAnalytics, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TaskStatusAnalytics), args.Error(1)
}

func (m *MockTaskRepository) GetTasksWithFilters(ctx context.Context, filters repository.TaskFilters) ([]*entity.Task, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

// MockProjectRepository implements ProjectRepository interface for testing
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(ctx context.Context, project *entity.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Project), args.Error(1)
}

func (m *MockProjectRepository) GetAll(ctx context.Context) ([]*entity.Project, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Project), args.Error(1)
}

func (m *MockProjectRepository) GetAllWithParams(ctx context.Context, params repository.GetProjectsParams) ([]*entity.Project, int, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*entity.Project), args.Get(1).(int), args.Error(2)
}

func (m *MockProjectRepository) Update(ctx context.Context, project *entity.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) GetWithTaskCount(ctx context.Context, id uuid.UUID) (*repository.ProjectWithTaskCount, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ProjectWithTaskCount), args.Error(1)
}

func (m *MockProjectRepository) GetTaskStatistics(ctx context.Context, projectID uuid.UUID) (map[entity.TaskStatus]int, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(map[entity.TaskStatus]int), args.Error(1)
}

func (m *MockProjectRepository) GetLastActivityAt(ctx context.Context, projectID uuid.UUID) (*time.Time, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

func (m *MockProjectRepository) Archive(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) CheckNameExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockProjectRepository) GetSettings(ctx context.Context, projectID uuid.UUID) (*entity.ProjectSettings, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ProjectSettings), args.Error(1)
}

func (m *MockProjectRepository) CreateSettings(ctx context.Context, settings *entity.ProjectSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

func (m *MockProjectRepository) UpdateSettings(ctx context.Context, settings *entity.ProjectSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

// MockNotificationUsecase implements NotificationUsecase interface for testing
type MockNotificationUsecase struct {
	mock.Mock
}

func (m *MockNotificationUsecase) SendTaskStatusChangeNotification(ctx context.Context, data entity.TaskStatusChangeNotificationData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockNotificationUsecase) SendTaskCreatedNotification(ctx context.Context, task *entity.Task, project *entity.Project) error {
	args := m.Called(ctx, task, project)
	return args.Error(0)
}

func (m *MockNotificationUsecase) RegisterHandler(notificationType entity.NotificationType, handler entity.NotificationHandler) error {
	args := m.Called(notificationType, handler)
	return args.Error(0)
}

func (m *MockNotificationUsecase) UnregisterHandler(notificationType entity.NotificationType) error {
	args := m.Called(notificationType)
	return args.Error(0)
}

func TestTaskUsecase_UpdateStatusWithHistory(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)
	ctx := context.Background()

	taskID := uuid.New()
	projectID := uuid.New()
	changedBy := "user123"
	reason := "Testing"

	// Mock current task
	currentTask := &entity.Task{
		ID:        taskID,
		ProjectID: projectID,
		Title:     "Test Task",
		Status:    entity.TaskStatusTODO,
	}

	// Mock updated task
	updatedTask := &entity.Task{
		ID:        taskID,
		ProjectID: projectID,
		Title:     "Test Task",
		Status:    entity.TaskStatusPLANNING,
	}

	// Mock project
	project := &entity.Project{
		ID:   projectID,
		Name: "Test Project",
	}

	req := UpdateStatusRequest{
		TaskID:    taskID,
		Status:    entity.TaskStatusPLANNING,
		ChangedBy: &changedBy,
		Reason:    &reason,
	}

	// Set up expectations
	mockTaskRepo.On("GetByID", ctx, taskID).Return(currentTask, nil).Once() // First call for validation
	mockTaskRepo.On("UpdateStatusWithHistory", ctx, taskID, entity.TaskStatusPLANNING, &changedBy, &reason).Return(nil)
	mockTaskRepo.On("GetByID", ctx, taskID).Return(updatedTask, nil).Once() // Second call after update
	mockProjectRepo.On("GetByID", ctx, projectID).Return(project, nil)
	mockNotificationUsecase.On("SendTaskStatusChangeNotification", ctx, mock.AnythingOfType("entity.TaskStatusChangeNotificationData")).Return(nil)

	// Execute
	result, err := usecase.UpdateStatusWithHistory(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, updatedTask, result)

	// Verify all expectations were met
	mockTaskRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
	mockNotificationUsecase.AssertExpectations(t)
}

func TestTaskUsecase_UpdateStatusWithHistory_InvalidTransition(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)
	ctx := context.Background()

	taskID := uuid.New()

	// Mock current task
	currentTask := &entity.Task{
		ID:     taskID,
		Title:  "Test Task",
		Status: entity.TaskStatusTODO,
	}

	req := UpdateStatusRequest{
		TaskID: taskID,
		Status: entity.TaskStatusDONE, // Invalid transition from TODO
	}

	// Set up expectations
	mockTaskRepo.On("GetByID", ctx, taskID).Return(currentTask, nil)

	// Execute
	result, err := usecase.UpdateStatusWithHistory(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid status transition")

	mockTaskRepo.AssertExpectations(t)
}

func TestTaskUsecase_GetByStatuses(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)
	ctx := context.Background()

	statuses := []entity.TaskStatus{entity.TaskStatusTODO, entity.TaskStatusDONE}
	expectedTasks := []*entity.Task{
		{ID: uuid.New(), Status: entity.TaskStatusTODO},
		{ID: uuid.New(), Status: entity.TaskStatusDONE},
	}

	// Set up expectations
	mockTaskRepo.On("GetByStatuses", ctx, statuses).Return(expectedTasks, nil)

	// Execute
	result, err := usecase.GetByStatuses(ctx, statuses)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedTasks, result)

	mockTaskRepo.AssertExpectations(t)
}

func TestTaskUsecase_GetByStatuses_InvalidStatus(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)
	ctx := context.Background()

	statuses := []entity.TaskStatus{entity.TaskStatusTODO, "INVALID_STATUS"}

	// Execute
	result, err := usecase.GetByStatuses(ctx, statuses)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestTaskUsecase_BulkUpdateStatus(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)
	ctx := context.Background()

	taskIDs := []uuid.UUID{uuid.New(), uuid.New()}
	changedBy := "admin"

	req := BulkUpdateStatusRequest{
		TaskIDs:   taskIDs,
		Status:    entity.TaskStatusPLANNING,
		ChangedBy: &changedBy,
	}

	// Set up expectations
	mockTaskRepo.On("BulkUpdateStatus", ctx, taskIDs, entity.TaskStatusPLANNING, &changedBy).Return(nil)

	// Execute
	err := usecase.BulkUpdateStatus(ctx, req)

	// Assert
	require.NoError(t, err)

	mockTaskRepo.AssertExpectations(t)
}

func TestTaskUsecase_BulkUpdateStatus_EmptyTaskList(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)
	ctx := context.Background()

	req := BulkUpdateStatusRequest{
		TaskIDs: []uuid.UUID{}, // Empty list
		Status:  entity.TaskStatusPLANNING,
	}

	// Execute
	err := usecase.BulkUpdateStatus(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no task IDs provided")
}

func TestTaskUsecase_BulkUpdateStatus_InvalidStatus(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)
	ctx := context.Background()

	req := BulkUpdateStatusRequest{
		TaskIDs: []uuid.UUID{uuid.New()},
		Status:  "INVALID_STATUS",
	}

	// Execute
	err := usecase.BulkUpdateStatus(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid target status")
}

func TestTaskUsecase_GetStatusHistory(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)
	ctx := context.Background()

	taskID := uuid.New()
	task := &entity.Task{ID: taskID, Title: "Test Task"}

	expectedHistory := []*entity.TaskStatusHistory{
		{
			ID:       uuid.New(),
			TaskID:   taskID,
			ToStatus: entity.TaskStatusPLANNING,
		},
	}

	// Set up expectations
	mockTaskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	mockTaskRepo.On("GetStatusHistory", ctx, taskID).Return(expectedHistory, nil)

	// Execute
	result, err := usecase.GetStatusHistory(ctx, taskID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedHistory, result)

	mockTaskRepo.AssertExpectations(t)
}

func TestTaskUsecase_GetStatusHistory_TaskNotFound(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)
	ctx := context.Background()

	taskID := uuid.New()

	// Set up expectations
	mockTaskRepo.On("GetByID", ctx, taskID).Return(nil, assert.AnError)

	// Execute
	result, err := usecase.GetStatusHistory(ctx, taskID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "task not found")

	mockTaskRepo.AssertExpectations(t)
}

func TestTaskUsecase_ValidateStatusTransition(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)
	ctx := context.Background()

	taskID := uuid.New()
	task := &entity.Task{
		ID:     taskID,
		Status: entity.TaskStatusTODO,
	}

	// Set up expectations
	mockTaskRepo.On("GetByID", ctx, taskID).Return(task, nil)

	// Test valid transition
	err := usecase.ValidateStatusTransition(ctx, taskID, entity.TaskStatusPLANNING)
	assert.NoError(t, err)

	// Test invalid transition
	err = usecase.ValidateStatusTransition(ctx, taskID, entity.TaskStatusDONE)
	assert.Error(t, err)

	mockTaskRepo.AssertExpectations(t)
}
