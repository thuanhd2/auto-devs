package testutil

import (
	"context"
	"errors"
	"fmt"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockProjectRepository is a mock implementation of repository.ProjectRepository
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Project), args.Error(1)
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

func (m *MockProjectRepository) GetAllWithParams(ctx context.Context, params repository.GetProjectsParams) ([]*entity.Project, int, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entity.Project), args.Int(1), args.Error(2)
}

func (m *MockProjectRepository) CheckNameExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockProjectRepository) GetTaskStatistics(ctx context.Context, projectID uuid.UUID) (map[entity.TaskStatus]int, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[entity.TaskStatus]int), args.Error(1)
}

func (m *MockProjectRepository) Archive(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) GetLastActivityAt(ctx context.Context, projectID uuid.UUID) (*entity.Time, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Time), args.Error(1)
}

// MockTaskRepository is a mock implementation of repository.TaskRepository
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) GetByProjectIDWithParams(ctx context.Context, projectID uuid.UUID, params repository.GetTasksParams) ([]*entity.Task, int, error) {
	args := m.Called(ctx, projectID, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entity.Task), args.Int(1), args.Error(2)
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

func (m *MockTaskRepository) GetByStatus(ctx context.Context, status entity.TaskStatus) ([]*entity.Task, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) GetStatsByProjectID(ctx context.Context, projectID uuid.UUID) (map[entity.TaskStatus]int, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[entity.TaskStatus]int), args.Error(1)
}

func (m *MockTaskRepository) Archive(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockAuditRepository is a mock implementation of repository.AuditRepository
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(ctx context.Context, auditLog *entity.AuditLog) error {
	args := m.Called(ctx, auditLog)
	return args.Error(0)
}

func (m *MockAuditRepository) GetByEntityID(ctx context.Context, entityType string, entityID uuid.UUID) ([]*entity.AuditLog, error) {
	args := m.Called(ctx, entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) GetByUserID(ctx context.Context, userID string) ([]*entity.AuditLog, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) GetRecentActivity(ctx context.Context, limit int) ([]*entity.AuditLog, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AuditLog), args.Error(1)
}

// MockErrorRepository is a special mock that can simulate various error conditions
type MockErrorRepository struct {
	mock.Mock
	shouldFailOnCreate bool
	shouldFailOnUpdate bool
	shouldFailOnDelete bool
	shouldFailOnGet    bool
}

// NewMockErrorRepository creates a new MockErrorRepository
func NewMockErrorRepository() *MockErrorRepository {
	return &MockErrorRepository{}
}

// SetFailOnCreate makes the next Create call fail
func (m *MockErrorRepository) SetFailOnCreate(fail bool) {
	m.shouldFailOnCreate = fail
}

// SetFailOnUpdate makes the next Update call fail
func (m *MockErrorRepository) SetFailOnUpdate(fail bool) {
	m.shouldFailOnUpdate = fail
}

// SetFailOnDelete makes the next Delete call fail
func (m *MockErrorRepository) SetFailOnDelete(fail bool) {
	m.shouldFailOnDelete = fail
}

// SetFailOnGet makes the next Get call fail
func (m *MockErrorRepository) SetFailOnGet(fail bool) {
	m.shouldFailOnGet = fail
}

// SimulateDBError returns a database error
func (m *MockErrorRepository) SimulateDBError(operation string) error {
	return fmt.Errorf("database error during %s operation", operation)
}

// SimulateNotFoundError returns a not found error
func (m *MockErrorRepository) SimulateNotFoundError(entityType string, id uuid.UUID) error {
	return fmt.Errorf("%s with id %s not found", entityType, id.String())
}

// SimulateConstraintError returns a constraint violation error
func (m *MockErrorRepository) SimulateConstraintError(constraint string) error {
	return fmt.Errorf("constraint violation: %s", constraint)
}

// MockWebSocketService provides a mock WebSocket service for testing
type MockWebSocketService struct {
	mock.Mock
	broadcastCalls []BroadcastCall
}

type BroadcastCall struct {
	MessageType string
	Data        interface{}
	ProjectID   *uuid.UUID
}

func NewMockWebSocketService() *MockWebSocketService {
	return &MockWebSocketService{
		broadcastCalls: make([]BroadcastCall, 0),
	}
}

func (m *MockWebSocketService) BroadcastToProject(messageType string, data interface{}, projectID uuid.UUID) {
	m.broadcastCalls = append(m.broadcastCalls, BroadcastCall{
		MessageType: messageType,
		Data:        data,
		ProjectID:   &projectID,
	})
	m.Called(messageType, data, projectID)
}

func (m *MockWebSocketService) BroadcastToAll(messageType string, data interface{}) {
	m.broadcastCalls = append(m.broadcastCalls, BroadcastCall{
		MessageType: messageType,
		Data:        data,
		ProjectID:   nil,
	})
	m.Called(messageType, data)
}

func (m *MockWebSocketService) GetBroadcastCalls() []BroadcastCall {
	return m.broadcastCalls
}

func (m *MockWebSocketService) ClearBroadcastCalls() {
	m.broadcastCalls = make([]BroadcastCall, 0)
}

// Helper functions for common test scenarios

// SetupMockProjectRepositorySuccess configures a mock project repository for successful operations
func SetupMockProjectRepositorySuccess(mock *MockProjectRepository, project *entity.Project) {
	mock.On("Create", mock.Anything, mock.Anything).Return(nil)
	mock.On("GetByID", mock.Anything, project.ID).Return(project, nil)
	mock.On("Update", mock.Anything, mock.Anything).Return(nil)
	mock.On("Delete", mock.Anything, project.ID).Return(nil)
}

// SetupMockProjectRepositoryErrors configures a mock project repository to return errors
func SetupMockProjectRepositoryErrors(mock *MockProjectRepository) {
	mock.On("Create", mock.Anything, mock.Anything).Return(errors.New("create failed"))
	mock.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("project not found"))
	mock.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))
	mock.On("Delete", mock.Anything, mock.Anything).Return(errors.New("delete failed"))
}

// SetupMockTaskRepositorySuccess configures a mock task repository for successful operations
func SetupMockTaskRepositorySuccess(mock *MockTaskRepository, task *entity.Task) {
	mock.On("Create", mock.Anything, mock.Anything).Return(nil)
	mock.On("GetByID", mock.Anything, task.ID).Return(task, nil)
	mock.On("Update", mock.Anything, mock.Anything).Return(nil)
	mock.On("Delete", mock.Anything, task.ID).Return(nil)
}

// SetupMockTaskRepositoryErrors configures a mock task repository to return errors
func SetupMockTaskRepositoryErrors(mock *MockTaskRepository) {
	mock.On("Create", mock.Anything, mock.Anything).Return(errors.New("create failed"))
	mock.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("task not found"))
	mock.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))
	mock.On("Delete", mock.Anything, mock.Anything).Return(errors.New("delete failed"))
}