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

// Test fixtures for sync worker

func createTestSyncWorkerConfig() *PRSyncWorkerConfig {
	return &PRSyncWorkerConfig{
		SyncInterval:       100 * time.Millisecond, // Fast for testing
		BatchSize:          5,
		MaxConcurrentSyncs: 2,
		SyncTimeout:        5 * time.Second,
		RetryAttempts:      2,
		RetryDelay:         10 * time.Millisecond,
	}
}

func createSyncWorker(t *testing.T) (
	*PRSyncWorker,
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
	config := createTestSyncWorkerConfig()

	worker := NewPRSyncWorker(
		githubSvc,
		prRepo,
		taskRepo,
		worktreeSvc,
		websocketSvc,
		config,
		logger,
	)

	return worker, githubSvc, prRepo, taskRepo, worktreeSvc, websocketSvc
}

// Tests

func TestNewPRSyncWorker(t *testing.T) {
	worker, _, _, _, _, _ := createSyncWorker(t)

	assert.NotNil(t, worker)
	assert.NotNil(t, worker.config)
	assert.NotNil(t, worker.logger)
	assert.NotNil(t, worker.stopCh)
	assert.False(t, worker.IsRunning())
}

func TestPRSyncWorkerStartStop(t *testing.T) {
	worker, _, prRepo, _, _, _ := createSyncWorker(t)

	// Mock empty PR list for sync cycles
	prRepo.On("GetActiveMonitoringPRs", mock.Anything).Return([]*entity.PullRequest{}, nil)

	// Start worker
	ctx := context.Background()
	err := worker.Start(ctx)
	require.NoError(t, err)
	assert.True(t, worker.IsRunning())

	// Try to start again (should fail)
	err = worker.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Stop worker
	err = worker.Stop()
	require.NoError(t, err)
	assert.False(t, worker.IsRunning())

	// Try to stop again (should fail)
	err = worker.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	prRepo.AssertExpectations(t)
}

func TestPRSyncWorkerGetStats(t *testing.T) {
	worker, _, _, _, _, _ := createSyncWorker(t)

	stats := worker.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, false, stats["running"])
	assert.Equal(t, int64(0), stats["sync_count"])
	assert.Equal(t, int64(0), stats["error_count"])
	assert.NotNil(t, stats["config"])

	// Check config in stats
	config := stats["config"].(map[string]interface{})
	assert.Equal(t, worker.config.SyncInterval, config["sync_interval"])
	assert.Equal(t, worker.config.BatchSize, config["batch_size"])
}

func TestSyncSinglePRNoStatusChange(t *testing.T) {
	worker, githubSvc, prRepo, _, _, _ := createSyncWorker(t)

	pr := createTestPR()
	pr.Status = entity.PullRequestStatusOpen

	// Mock GitHub API to return same status
	githubPR := createTestPR()
	githubPR.ID = pr.ID
	githubPR.Status = entity.PullRequestStatusOpen

	githubSvc.On("GetPullRequest", mock.Anything, pr.Repository, pr.GitHubPRNumber).Return(githubPR, nil)

	result := worker.syncSinglePR(context.Background(), pr)

	assert.True(t, result.Success)
	assert.False(t, result.StatusChange)
	assert.Equal(t, entity.PullRequestStatusOpen, result.OldStatus)
	assert.Equal(t, entity.PullRequestStatusOpen, result.NewStatus)
	assert.NoError(t, result.Error)

	githubSvc.AssertExpectations(t)
	prRepo.AssertExpectations(t)
}

func TestSyncSinglePRStatusChange(t *testing.T) {
	worker, githubSvc, prRepo, taskRepo, _, websocketSvc := createSyncWorker(t)

	pr := createTestPR()
	pr.Status = entity.PullRequestStatusOpen

	task := createTestTask()
	task.ID = pr.TaskID
	task.Status = entity.TaskStatusCODEREVIEWING

	// Mock GitHub API to return merged status
	githubPR := createTestPR()
	githubPR.ID = pr.ID
	githubPR.Status = entity.PullRequestStatusMerged
	mergedAt := time.Now()
	githubPR.MergedAt = &mergedAt

	githubSvc.On("GetPullRequest", mock.Anything, pr.Repository, pr.GitHubPRNumber).Return(githubPR, nil)
	prRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.PullRequest) bool {
		return p.ID == pr.ID && p.Status == entity.PullRequestStatusMerged
	})).Return(nil)

	taskRepo.On("GetByID", mock.Anything, pr.TaskID).Return(task, nil)
	taskRepo.On("UpdateStatus", mock.Anything, task.ID, entity.TaskStatusDONE).Return(nil)

	websocketSvc.On("NotifyStatusChanged", task.ID, task.ProjectID, "task", string(entity.TaskStatusCODEREVIEWING), string(entity.TaskStatusDONE)).Return(nil)
	websocketSvc.On("SendProjectMessage", task.ProjectID, websocket.MessageTypePRUpdate, mock.Anything).Return(nil)

	result := worker.syncSinglePR(context.Background(), pr)

	assert.True(t, result.Success)
	assert.True(t, result.StatusChange)
	assert.Equal(t, entity.PullRequestStatusOpen, result.OldStatus)
	assert.Equal(t, entity.PullRequestStatusMerged, result.NewStatus)
	assert.NoError(t, result.Error)

	githubSvc.AssertExpectations(t)
	prRepo.AssertExpectations(t)
	taskRepo.AssertExpectations(t)
	websocketSvc.AssertExpectations(t)
}

func TestSyncSinglePRWithRetry(t *testing.T) {
	worker, githubSvc, _, _, _, _ := createSyncWorker(t)

	pr := createTestPR()

	// Mock GitHub API to fail first time, succeed second time
	githubSvc.On("GetPullRequest", mock.Anything, pr.Repository, pr.GitHubPRNumber).
		Return(nil, assert.AnError).Once()

	githubPR := createTestPR()
	githubPR.ID = pr.ID
	githubSvc.On("GetPullRequest", mock.Anything, pr.Repository, pr.GitHubPRNumber).
		Return(githubPR, nil).Once()

	result := worker.syncSinglePR(context.Background(), pr)

	assert.True(t, result.Success)
	assert.NoError(t, result.Error)

	githubSvc.AssertExpectations(t)
}

func TestSyncSinglePRMaxRetriesExceeded(t *testing.T) {
	worker, githubSvc, _, _, _, _ := createSyncWorker(t)

	pr := createTestPR()

	// Mock GitHub API to always fail
	githubSvc.On("GetPullRequest", mock.Anything, pr.Repository, pr.GitHubPRNumber).
		Return(nil, assert.AnError).Times(worker.config.RetryAttempts + 1)

	result := worker.syncSinglePR(context.Background(), pr)

	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed after")

	githubSvc.AssertExpectations(t)
}

func TestPerformSyncBatch(t *testing.T) {
	worker, githubSvc, prRepo, _, _, _ := createSyncWorker(t)

	// Create test PRs
	pr1 := createTestPR()
	pr2 := createTestPR()
	pr2.ID = uuid.New()
	pr2.GitHubPRNumber = 456

	prs := []*entity.PullRequest{pr1, pr2}

	// Mock repository call
	prRepo.On("GetActiveMonitoringPRs", mock.Anything).Return(prs, nil)

	// Mock GitHub API calls
	githubPR1 := createTestPR()
	githubPR1.ID = pr1.ID
	githubSvc.On("GetPullRequest", mock.Anything, pr1.Repository, pr1.GitHubPRNumber).Return(githubPR1, nil)

	githubPR2 := createTestPR()
	githubPR2.ID = pr2.ID
	githubPR2.GitHubPRNumber = 456
	githubSvc.On("GetPullRequest", mock.Anything, pr2.Repository, pr2.GitHubPRNumber).Return(githubPR2, nil)

	worker.performSync(context.Background())

	// Verify calls were made
	prRepo.AssertExpectations(t)
	githubSvc.AssertExpectations(t)

	// Check stats were updated
	stats := worker.GetStats()
	assert.Equal(t, int64(1), stats["sync_count"])
}

func TestForceSyncWhenNotRunning(t *testing.T) {
	worker, _, _, _, _, _ := createSyncWorker(t)

	err := worker.ForceSync(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestForceSyncWhenRunning(t *testing.T) {
	worker, _, prRepo, _, _, _ := createSyncWorker(t)

	// Mock empty PR list
	prRepo.On("GetActiveMonitoringPRs", mock.Anything).Return([]*entity.PullRequest{}, nil)

	// Start worker
	ctx := context.Background()
	err := worker.Start(ctx)
	require.NoError(t, err)
	defer worker.Stop()

	// Force sync should work
	err = worker.ForceSync(ctx)
	assert.NoError(t, err)

	// Give some time for the sync to be processed
	time.Sleep(50 * time.Millisecond)

	prRepo.AssertExpectations(t)
}

func TestSyncResultCounters(t *testing.T) {
	worker, _, _, _, _, _ := createSyncWorker(t)

	results := []*SyncResult{
		{Success: true, StatusChange: false},
		{Success: true, StatusChange: true},
		{Success: false, StatusChange: false},
		{Success: true, StatusChange: true},
		{Success: false, StatusChange: false},
	}

	assert.Equal(t, 3, worker.countSuccessfulSyncs(results))
	assert.Equal(t, 2, worker.countFailedSyncs(results))
	assert.Equal(t, 2, worker.countStatusChanges(results))
}

func TestHandleMergeCompletion(t *testing.T) {
	worker, _, _, _, worktreeSvc, websocketSvc := createSyncWorker(t)

	pr := createTestPR()
	pr.Status = entity.PullRequestStatusMerged
	mergedAt := time.Now()
	pr.MergedAt = &mergedAt

	task := createTestTask()
	task.ID = pr.TaskID

	worktree := &entity.Worktree{
		ID:     uuid.New(),
		TaskID: task.ID,
	}

	worktreeSvc.On("GetWorktreeByTaskID", mock.Anything, pr.TaskID).Return(worktree, nil)
	worktreeSvc.On("CleanupTaskWorktree", mock.Anything, pr.TaskID, task.ProjectID).Return(nil)
	websocketSvc.On("SendProjectMessage", task.ProjectID, websocket.MessageTypePRUpdate, mock.Anything).Return(nil)

	err := worker.handleMergeCompletion(context.Background(), pr, task)
	require.NoError(t, err)

	worktreeSvc.AssertExpectations(t)
	websocketSvc.AssertExpectations(t)
}

func TestDefaultPRSyncWorkerConfig(t *testing.T) {
	config := DefaultPRSyncWorkerConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 1*time.Minute, config.SyncInterval)
	assert.Equal(t, 10, config.BatchSize)
	assert.Equal(t, 5, config.MaxConcurrentSyncs)
	assert.Equal(t, 30*time.Second, config.SyncTimeout)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, 10*time.Second, config.RetryDelay)
}