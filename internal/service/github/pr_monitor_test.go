package github

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/websocket"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type MockPRRepository struct {
	mock.Mock
}

func (m *MockPRRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.PullRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PullRequest), args.Error(1)
}

func (m *MockPRRepository) Update(ctx context.Context, pr *entity.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *MockPRRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.PullRequest, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PullRequest), args.Error(1)
}

func (m *MockPRRepository) GetActiveMonitoringPRs(ctx context.Context) ([]*entity.PullRequest, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.PullRequest), args.Error(1)
}

type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *entity.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TaskStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockWorktreeService struct {
	mock.Mock
}

func (m *MockWorktreeService) CleanupTaskWorktree(ctx context.Context, taskID uuid.UUID, projectID uuid.UUID) error {
	args := m.Called(ctx, taskID, projectID)
	return args.Error(0)
}

func (m *MockWorktreeService) GetWorktreeByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.Worktree, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Worktree), args.Error(1)
}

// GitHubServiceInterface for testing
type GitHubServiceInterface interface {
	GetPullRequest(ctx context.Context, repo string, prNumber int) (*entity.PullRequest, error)
}

type MockGitHubServiceForPR struct {
	mock.Mock
}

func (m *MockGitHubServiceForPR) GetPullRequest(ctx context.Context, repo string, prNumber int) (*entity.PullRequest, error) {
	args := m.Called(ctx, repo, prNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PullRequest), args.Error(1)
}

type MockWebSocketService struct {
	mock.Mock
}

func (m *MockWebSocketService) SendProjectMessage(projectID uuid.UUID, msgType websocket.MessageType, data interface{}) error {
	args := m.Called(projectID, msgType, data)
	return args.Error(0)
}

func (m *MockWebSocketService) NotifyStatusChanged(entityID, projectID uuid.UUID, entityType, oldStatus, newStatus string) error {
	args := m.Called(entityID, projectID, entityType, oldStatus, newStatus)
	return args.Error(0)
}

// Test fixtures

func createTestPR() *entity.PullRequest {
	return &entity.PullRequest{
		ID:             uuid.New(),
		TaskID:         uuid.New(),
		GitHubPRNumber: 123,
		Repository:     "owner/repo",
		Title:          "Test PR",
		Status:         entity.PullRequestStatusOpen,
		HeadBranch:     "feature-branch",
		BaseBranch:     "main",
		GitHubURL:      "https://github.com/owner/repo/pull/123",
	}
}

func createTestTask() *entity.Task {
	return &entity.Task{
		ID:        uuid.New(),
		ProjectID: uuid.New(),
		Title:     "Test Task",
		Status:    entity.TaskStatusIMPLEMENTING,
	}
}

func createTestConfig() *PRMonitorConfig {
	return &PRMonitorConfig{
		PollInterval:        100 * time.Millisecond, // Faster for testing
		MaxRetries:          2,
		RetryDelay:          10 * time.Millisecond,
		ConcurrentMonitors:  2,
		NotificationTimeout: 5 * time.Second,
	}
}

func createPRMonitor(t *testing.T) (
	*PRMonitor,
	*MockGitHubServiceForPR,
	*MockPRRepository,
	*MockTaskRepository,
	*MockWorktreeService,
	*MockWebSocketService,
) {
	githubSvc := &MockGitHubServiceForPR{}
	prRepo := &MockPRRepository{}
	taskRepo := &MockTaskRepository{}
	worktreeSvc := &MockWorktreeService{}
	websocketSvc := &MockWebSocketService{}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := createTestConfig()

	monitor := NewPRMonitor(
		githubSvc,
		prRepo,
		taskRepo,
		worktreeSvc,
		websocketSvc,
		config,
		logger,
	)

	return monitor, githubSvc, prRepo, taskRepo, worktreeSvc, websocketSvc
}

// Tests

func TestNewPRMonitor(t *testing.T) {
	monitor, _, _, _, _, _ := createPRMonitor(t)

	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.config)
	assert.NotNil(t, monitor.logger)
	assert.NotNil(t, monitor.activeMonitors)
	assert.NotNil(t, monitor.stopCh)
}

func TestMonitorPR(t *testing.T) {
	monitor, _, _, taskRepo, _, _ := createPRMonitor(t)

	pr := createTestPR()
	task := createTestTask()
	task.ID = pr.TaskID

	taskRepo.On("GetByID", mock.Anything, pr.TaskID).Return(task, nil)

	err := monitor.MonitorPR(pr)
	require.NoError(t, err)

	// Check if PR is being monitored
	assert.True(t, monitor.IsMonitoring(pr.ID))

	// Check monitoring stats
	stats := monitor.GetMonitoringStats()
	assert.Equal(t, 1, stats["active_monitors"])

	// Stop monitoring
	err = monitor.StopMonitoring(pr.ID)
	require.NoError(t, err)
	assert.False(t, monitor.IsMonitoring(pr.ID))

	taskRepo.AssertExpectations(t)
}

func TestHandlePRStatusChange(t *testing.T) {
	tests := []struct {
		name           string
		oldPRStatus    entity.PullRequestStatus
		newPRStatus    entity.PullRequestStatus
		oldTaskStatus  entity.TaskStatus
		expectedTaskStatus entity.TaskStatus
		shouldUpdateTask bool
	}{
		{
			name:           "PR opened should set task to code reviewing",
			oldPRStatus:    entity.PullRequestStatusOpen,
			newPRStatus:    entity.PullRequestStatusOpen,
			oldTaskStatus:  entity.TaskStatusIMPLEMENTING,
			expectedTaskStatus: entity.TaskStatusCODEREVIEWING,
			shouldUpdateTask: true,
		},
		{
			name:           "PR merged should set task to done",
			oldPRStatus:    entity.PullRequestStatusOpen,
			newPRStatus:    entity.PullRequestStatusMerged,
			oldTaskStatus:  entity.TaskStatusCODEREVIEWING,
			expectedTaskStatus: entity.TaskStatusDONE,
			shouldUpdateTask: true,
		},
		{
			name:           "PR closed should set task to cancelled",
			oldPRStatus:    entity.PullRequestStatusOpen,
			newPRStatus:    entity.PullRequestStatusClosed,
			oldTaskStatus:  entity.TaskStatusCODEREVIEWING,
			expectedTaskStatus: entity.TaskStatusCANCELLED,
			shouldUpdateTask: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor, _, _, taskRepo, _, websocketSvc := createPRMonitor(t)

			pr := createTestPR()
			pr.Status = tt.newPRStatus

			task := createTestTask()
			task.ID = pr.TaskID
			task.Status = tt.oldTaskStatus

			taskRepo.On("GetByID", mock.Anything, pr.TaskID).Return(task, nil)

			if tt.shouldUpdateTask {
				taskRepo.On("UpdateStatus", mock.Anything, task.ID, tt.expectedTaskStatus).Return(nil)
				websocketSvc.On("NotifyStatusChanged", task.ID, task.ProjectID, "task", string(tt.oldTaskStatus), string(tt.expectedTaskStatus)).Return(nil)
			}

			// Mock the notification sending
			websocketSvc.On("SendProjectMessage", task.ProjectID, websocket.MessageTypePRUpdate, mock.Anything).Return(nil)

			err := monitor.HandlePRStatusChange(pr, string(tt.newPRStatus))
			require.NoError(t, err)

			taskRepo.AssertExpectations(t)
			websocketSvc.AssertExpectations(t)
		})
	}
}

func TestHandlePRMerge(t *testing.T) {
	monitor, _, _, taskRepo, worktreeSvc, websocketSvc := createPRMonitor(t)

	pr := createTestPR()
	pr.Status = entity.PullRequestStatusMerged
	mergedAt := time.Now()
	pr.MergedAt = &mergedAt

	task := createTestTask()
	task.ID = pr.TaskID
	task.Status = entity.TaskStatusCODEREVIEWING

	worktree := &entity.Worktree{
		ID:     uuid.New(),
		TaskID: task.ID,
	}

	taskRepo.On("GetByID", mock.Anything, pr.TaskID).Return(task, nil)
	taskRepo.On("UpdateStatus", mock.Anything, task.ID, entity.TaskStatusDONE).Return(nil)

	worktreeSvc.On("GetWorktreeByTaskID", mock.Anything, pr.TaskID).Return(worktree, nil)
	worktreeSvc.On("CleanupTaskWorktree", mock.Anything, pr.TaskID, task.ProjectID).Return(nil)

	websocketSvc.On("NotifyStatusChanged", task.ID, task.ProjectID, "task", string(entity.TaskStatusCODEREVIEWING), string(entity.TaskStatusDONE)).Return(nil)
	websocketSvc.On("SendProjectMessage", task.ProjectID, websocket.MessageTypePRUpdate, mock.Anything).Return(nil)

	err := monitor.HandlePRMerge(pr)
	require.NoError(t, err)

	taskRepo.AssertExpectations(t)
	worktreeSvc.AssertExpectations(t)
	websocketSvc.AssertExpectations(t)
}

func TestMonitorAllActivePRs(t *testing.T) {
	monitor, _, _, taskRepo, _, _ := createPRMonitor(t)

	pr1 := createTestPR()
	pr2 := createTestPR()
	pr2.ID = uuid.New()

	task1 := createTestTask()
	task1.ID = pr1.TaskID
	task2 := createTestTask()
	task2.ID = pr2.TaskID

	prs := []*entity.PullRequest{pr1, pr2}

	prRepo.On("GetActiveMonitoringPRs", mock.Anything).Return(prs, nil)
	taskRepo.On("GetByID", mock.Anything, pr1.TaskID).Return(task1, nil)
	taskRepo.On("GetByID", mock.Anything, pr2.TaskID).Return(task2, nil)

	err := monitor.MonitorAllActivePRs(context.Background())
	require.NoError(t, err)

	// Check both PRs are being monitored
	assert.True(t, monitor.IsMonitoring(pr1.ID))
	assert.True(t, monitor.IsMonitoring(pr2.ID))

	stats := monitor.GetMonitoringStats()
	assert.Equal(t, 2, stats["active_monitors"])

	prRepo.AssertExpectations(t)
	taskRepo.AssertExpectations(t)
}

func TestRefreshPR(t *testing.T) {
	monitor, githubSvc, prRepo, taskRepo, _, websocketSvc := createPRMonitor(t)

	pr := createTestPR()
	task := createTestTask()
	task.ID = pr.TaskID

	// Start monitoring first
	taskRepo.On("GetByID", mock.Anything, pr.TaskID).Return(task, nil).Times(2)

	err := monitor.MonitorPR(pr)
	require.NoError(t, err)

	// Mock GitHub API call for refresh
	updatedPR := createTestPR()
	updatedPR.ID = pr.ID
	updatedPR.Status = entity.PullRequestStatusMerged

	githubSvc.On("GetPullRequest", mock.Anything, pr.Repository, pr.GitHubPRNumber).Return(updatedPR, nil)
	prRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.PullRequest) bool {
		return p.ID == pr.ID && p.Status == entity.PullRequestStatusMerged
	})).Return(nil)

	// Mock the status change handling
	taskRepo.On("UpdateStatus", mock.Anything, task.ID, entity.TaskStatusDONE).Return(nil)
	websocketSvc.On("NotifyStatusChanged", task.ID, task.ProjectID, "task", string(entity.TaskStatusIMPLEMENTING), string(entity.TaskStatusDONE)).Return(nil)
	websocketSvc.On("SendProjectMessage", task.ProjectID, websocket.MessageTypePRUpdate, mock.Anything).Return(nil)

	err = monitor.RefreshPR(pr.ID)
	require.NoError(t, err)

	githubSvc.AssertExpectations(t)
	prRepo.AssertExpectations(t)
	taskRepo.AssertExpectations(t)
	websocketSvc.AssertExpectations(t)
}

func TestGetMonitoringStats(t *testing.T) {
	monitor, _, _, taskRepo, _, _ := createPRMonitor(t)

	// Initially no monitors
	stats := monitor.GetMonitoringStats()
	assert.Equal(t, 0, stats["active_monitors"])

	// Add a monitor
	pr := createTestPR()
	task := createTestTask()
	task.ID = pr.TaskID

	taskRepo.On("GetByID", mock.Anything, pr.TaskID).Return(task, nil)

	err := monitor.MonitorPR(pr)
	require.NoError(t, err)

	// Check stats
	stats = monitor.GetMonitoringStats()
	assert.Equal(t, 1, stats["active_monitors"])

	monitors := stats["monitors"].([]map[string]interface{})
	assert.Len(t, monitors, 1)
	assert.Equal(t, pr.ID, monitors[0]["pr_id"])
	assert.Equal(t, pr.GitHubPRNumber, monitors[0]["pr_number"])
	assert.Equal(t, task.ID, monitors[0]["task_id"])

	taskRepo.AssertExpectations(t)
}