package github

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/websocket"
	"github.com/google/uuid"
)

// PRMonitorConfig holds configuration for PR monitoring
type PRMonitorConfig struct {
	PollInterval        time.Duration
	MaxRetries          int
	RetryDelay          time.Duration
	ConcurrentMonitors  int
	NotificationTimeout time.Duration
}

// DefaultPRMonitorConfig returns default configuration
func DefaultPRMonitorConfig() *PRMonitorConfig {
	return &PRMonitorConfig{
		PollInterval:        5 * time.Minute,
		MaxRetries:          3,
		RetryDelay:          30 * time.Second,
		ConcurrentMonitors:  5,
		NotificationTimeout: 10 * time.Second,
	}
}

// PRRepository interface for PR data operations
type PRRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PullRequest, error)
	Update(ctx context.Context, pr *entity.PullRequest) error
	GetByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.PullRequest, error)
	GetActiveMonitoringPRs(ctx context.Context) ([]*entity.PullRequest, error)
}

// TaskRepository interface for task data operations
type TaskRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Task, error)
	Update(ctx context.Context, task *entity.Task) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TaskStatus) error
}

// WorktreeService interface for worktree operations
type WorktreeService interface {
	CleanupTaskWorktree(ctx context.Context, taskID uuid.UUID, projectID uuid.UUID) error
	GetWorktreeByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.Worktree, error)
}

// Extend the existing GitHubServiceInterface with monitoring operations
// Note: GitHubServiceInterface is defined in pr_creator.go

// WebSocketServiceInterface defines the WebSocket operations needed for notifications
type WebSocketServiceInterface interface {
	SendProjectMessage(projectID uuid.UUID, msgType websocket.MessageType, data interface{}) error
	NotifyStatusChanged(entityID, projectID uuid.UUID, entityType, oldStatus, newStatus string) error
}

// PRMonitor handles monitoring GitHub PR status and managing state changes
type PRMonitor struct {
	githubService   GitHubServiceInterface
	prRepo          PRRepository
	taskRepo        TaskRepository
	worktreeService WorktreeService
	websocketSvc    WebSocketServiceInterface
	config          *PRMonitorConfig
	logger          *slog.Logger
	
	// Monitoring state
	activeMonitors map[uuid.UUID]*monitorSession
	mu             sync.RWMutex
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

// monitorSession represents an active monitoring session for a PR
type monitorSession struct {
	pr       *entity.PullRequest
	task     *entity.Task
	stopCh   chan struct{}
	lastPoll time.Time
	retries  int
}

// NewPRMonitor creates a new PR monitoring service
func NewPRMonitor(
	githubService GitHubServiceInterface,
	prRepo PRRepository,
	taskRepo TaskRepository,
	worktreeService WorktreeService,
	websocketSvc WebSocketServiceInterface,
	config *PRMonitorConfig,
	logger *slog.Logger,
) *PRMonitor {
	if config == nil {
		config = DefaultPRMonitorConfig()
	}

	return &PRMonitor{
		githubService:   githubService,
		prRepo:          prRepo,
		taskRepo:        taskRepo,
		worktreeService: worktreeService,
		websocketSvc:    websocketSvc,
		config:          config,
		logger:          logger.With("component", "pr_monitor"),
		activeMonitors:  make(map[uuid.UUID]*monitorSession),
		stopCh:          make(chan struct{}),
	}
}

// MonitorPR starts monitoring a pull request for status changes
func (pm *PRMonitor) MonitorPR(pr *entity.PullRequest) error {
	if pr == nil {
		return fmt.Errorf("pull request cannot be nil")
	}

	// Get associated task
	task, err := pm.taskRepo.GetByID(context.Background(), pr.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if already monitoring
	if _, exists := pm.activeMonitors[pr.ID]; exists {
		pm.logger.Info("PR already being monitored", "pr_id", pr.ID, "pr_number", pr.GitHubPRNumber)
		return nil
	}

	// Create monitoring session
	session := &monitorSession{
		pr:     pr,
		task:   task,
		stopCh: make(chan struct{}),
	}

	pm.activeMonitors[pr.ID] = session
	pm.wg.Add(1)

	// Start monitoring goroutine
	go pm.monitorLoop(session)

	pm.logger.Info("Started monitoring PR", 
		"pr_id", pr.ID, 
		"pr_number", pr.GitHubPRNumber,
		"task_id", pr.TaskID,
		"repository", pr.Repository,
	)

	return nil
}

// StopMonitoring stops monitoring a specific PR
func (pm *PRMonitor) StopMonitoring(prID uuid.UUID) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	session, exists := pm.activeMonitors[prID]
	if !exists {
		return fmt.Errorf("PR %s is not being monitored", prID)
	}

	close(session.stopCh)
	delete(pm.activeMonitors, prID)

	pm.logger.Info("Stopped monitoring PR", "pr_id", prID)
	return nil
}

// StartMonitoring starts the PR monitoring service
func (pm *PRMonitor) StartMonitoring(ctx context.Context) error {
	pm.logger.Info("Starting PR monitoring service")

	// Get all active PRs that need monitoring
	prs, err := pm.prRepo.GetActiveMonitoringPRs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active PRs: %w", err)
	}

	// Start monitoring each PR
	for _, pr := range prs {
		if err := pm.MonitorPR(pr); err != nil {
			pm.logger.Error("Failed to start monitoring PR", 
				"pr_id", pr.ID, 
				"error", err,
			)
		}
	}

	pm.logger.Info("PR monitoring service started", "active_monitors", len(prs))
	return nil
}

// StopMonitoring stops the PR monitoring service
func (pm *PRMonitor) Stop() error {
	pm.logger.Info("Stopping PR monitoring service")

	close(pm.stopCh)
	
	pm.mu.Lock()
	// Stop all active monitors
	for prID, session := range pm.activeMonitors {
		close(session.stopCh)
		delete(pm.activeMonitors, prID)
	}
	pm.mu.Unlock()

	// Wait for all goroutines to finish
	pm.wg.Wait()

	pm.logger.Info("PR monitoring service stopped")
	return nil
}

// monitorLoop runs the monitoring loop for a specific PR
func (pm *PRMonitor) monitorLoop(session *monitorSession) {
	defer pm.wg.Done()
	
	ticker := time.NewTicker(pm.config.PollInterval)
	defer ticker.Stop()

	pm.logger.Info("Starting monitor loop", 
		"pr_id", session.pr.ID,
		"pr_number", session.pr.GitHubPRNumber,
	)

	for {
		select {
		case <-session.stopCh:
			pm.logger.Info("Monitor loop stopped", "pr_id", session.pr.ID)
			return
		case <-pm.stopCh:
			pm.logger.Info("Monitor loop stopped by service", "pr_id", session.pr.ID)
			return
		case <-ticker.C:
			if err := pm.pollPRStatus(session); err != nil {
				pm.logger.Error("Failed to poll PR status", 
					"pr_id", session.pr.ID,
					"error", err,
				)
			}
		}
	}
}

// pollPRStatus polls the GitHub API for PR status updates
func (pm *PRMonitor) pollPRStatus(session *monitorSession) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get latest PR data from GitHub
	updatedPR, err := pm.githubService.GetPullRequest(ctx, session.pr.Repository, session.pr.GitHubPRNumber)
	if err != nil {
		session.retries++
		if session.retries >= pm.config.MaxRetries {
			pm.logger.Error("Max retries reached for PR monitoring", 
				"pr_id", session.pr.ID,
				"retries", session.retries,
			)
			// Stop monitoring this PR
			pm.mu.Lock()
			delete(pm.activeMonitors, session.pr.ID)
			pm.mu.Unlock()
			return fmt.Errorf("max retries reached: %w", err)
		}
		return fmt.Errorf("failed to get PR: %w", err)
	}

	// Reset retry counter on successful call
	session.retries = 0
	session.lastPoll = time.Now()

	// Check for status changes
	if err := pm.handleStatusChange(session, updatedPR); err != nil {
		return fmt.Errorf("failed to handle status change: %w", err)
	}

	return nil
}

// handleStatusChange handles PR status changes
func (pm *PRMonitor) handleStatusChange(session *monitorSession, updatedPR *entity.PullRequest) error {
	oldStatus := session.pr.Status
	newStatus := updatedPR.Status

	// Update session PR with latest data
	session.pr.Status = updatedPR.Status
	session.pr.MergedAt = updatedPR.MergedAt
	session.pr.ClosedAt = updatedPR.ClosedAt
	session.pr.MergeCommitSHA = updatedPR.MergeCommitSHA
	session.pr.MergedBy = updatedPR.MergedBy
	session.pr.Mergeable = updatedPR.Mergeable
	session.pr.MergeableState = updatedPR.MergeableState

	// Save to database
	if err := pm.prRepo.Update(context.Background(), session.pr); err != nil {
		return fmt.Errorf("failed to update PR: %w", err)
	}

	// Handle status-specific changes
	if oldStatus != newStatus {
		pm.logger.Info("PR status changed", 
			"pr_id", session.pr.ID,
			"pr_number", session.pr.GitHubPRNumber,
			"old_status", oldStatus,
			"new_status", newStatus,
		)

		return pm.HandlePRStatusChange(session.pr, string(newStatus))
	}

	// Check for merge event even if status didn't change
	if newStatus == entity.PullRequestStatusMerged && session.pr.MergedAt != nil {
		return pm.HandlePRMerge(session.pr)
	}

	return nil
}

// HandlePRStatusChange handles PR status changes and updates task status accordingly
func (pm *PRMonitor) HandlePRStatusChange(pr *entity.PullRequest, newStatus string) error {
	ctx := context.Background()

	// Get the associated task
	task, err := pm.taskRepo.GetByID(ctx, pr.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	oldTaskStatus := task.Status
	var newTaskStatus entity.TaskStatus

	// Map PR status to task status
	switch entity.PullRequestStatus(newStatus) {
	case entity.PullRequestStatusOpen:
		// PR is open, task should be in code reviewing
		newTaskStatus = entity.TaskStatusCODEREVIEWING
		
	case entity.PullRequestStatusMerged:
		// PR is merged, task is done
		newTaskStatus = entity.TaskStatusDONE
		
	case entity.PullRequestStatusClosed:
		// PR is closed without merge, check if task should be cancelled
		if task.Status == entity.TaskStatusCODEREVIEWING {
			newTaskStatus = entity.TaskStatusCANCELLED
		} else {
			newTaskStatus = task.Status // Keep current status
		}
	}

	// Update task status if it changed
	if newTaskStatus != oldTaskStatus {
		if err := pm.taskRepo.UpdateStatus(ctx, task.ID, newTaskStatus); err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}

		pm.logger.Info("Updated task status due to PR change", 
			"task_id", task.ID,
			"pr_id", pr.ID,
			"old_task_status", oldTaskStatus,
			"new_task_status", newTaskStatus,
			"pr_status", newStatus,
		)

		// Send WebSocket notification for task status change
		if err := pm.websocketSvc.NotifyStatusChanged(
			task.ID,
			task.ProjectID,
			"task",
			string(oldTaskStatus),
			string(newTaskStatus),
		); err != nil {
			pm.logger.Error("Failed to send task status notification", 
				"task_id", task.ID,
				"error", err,
			)
		}
	}

	// Send PR status change notification
	if err := pm.sendPRStatusNotification(pr, string(oldTaskStatus), newStatus); err != nil {
		pm.logger.Error("Failed to send PR status notification", 
			"pr_id", pr.ID,
			"error", err,
		)
	}

	return nil
}

// HandlePRMerge handles PR merge completion
func (pm *PRMonitor) HandlePRMerge(pr *entity.PullRequest) error {
	ctx := context.Background()

	pm.logger.Info("Handling PR merge", 
		"pr_id", pr.ID,
		"pr_number", pr.GitHubPRNumber,
		"task_id", pr.TaskID,
	)

	// Get the associated task
	task, err := pm.taskRepo.GetByID(ctx, pr.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Update task status to DONE if not already
	if task.Status != entity.TaskStatusDONE {
		if err := pm.taskRepo.UpdateStatus(ctx, task.ID, entity.TaskStatusDONE); err != nil {
			return fmt.Errorf("failed to update task status to DONE: %w", err)
		}

		// Send task status notification
		if err := pm.websocketSvc.NotifyStatusChanged(
			task.ID,
			task.ProjectID,
			"task",
			string(task.Status),
			string(entity.TaskStatusDONE),
		); err != nil {
			pm.logger.Error("Failed to send task completion notification", 
				"task_id", task.ID,
				"error", err,
			)
		}
	}

	// Trigger worktree cleanup
	if err := pm.triggerWorktreeCleanup(ctx, pr.TaskID, task.ProjectID); err != nil {
		pm.logger.Error("Failed to trigger worktree cleanup", 
			"task_id", pr.TaskID,
			"error", err,
		)
		// Don't return error as cleanup failure shouldn't fail the merge handling
	}

	// Send merge completion notification
	if err := pm.sendMergeNotification(pr, task); err != nil {
		pm.logger.Error("Failed to send merge notification", 
			"pr_id", pr.ID,
			"error", err,
		)
	}

	// Stop monitoring this PR since it's complete
	if err := pm.StopMonitoring(pr.ID); err != nil {
		pm.logger.Error("Failed to stop monitoring merged PR", 
			"pr_id", pr.ID,
			"error", err,
		)
	}

	return nil
}

// HandlePRReview handles PR review events
func (pm *PRMonitor) HandlePRReview(pr *entity.PullRequest, review *entity.PullRequestReview) error {
	pm.logger.Info("Handling PR review", 
		"pr_id", pr.ID,
		"pr_number", pr.GitHubPRNumber,
		"reviewer", review.Reviewer,
		"state", review.State,
	)

	// Send review notification
	if err := pm.sendReviewNotification(pr, review); err != nil {
		pm.logger.Error("Failed to send review notification", 
			"pr_id", pr.ID,
			"review_id", review.ID,
			"error", err,
		)
	}

	return nil
}

// triggerWorktreeCleanup triggers cleanup of the task's worktree
func (pm *PRMonitor) triggerWorktreeCleanup(ctx context.Context, taskID uuid.UUID, projectID uuid.UUID) error {
	// Get worktree information
	worktree, err := pm.worktreeService.GetWorktreeByTaskID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	if worktree == nil {
		pm.logger.Info("No worktree found for cleanup", "task_id", taskID)
		return nil
	}

	// Trigger cleanup
	if err := pm.worktreeService.CleanupTaskWorktree(ctx, taskID, projectID); err != nil {
		return fmt.Errorf("failed to cleanup worktree: %w", err)
	}

	pm.logger.Info("Worktree cleanup triggered", 
		"task_id", taskID,
		"worktree_id", worktree.ID,
	)

	return nil
}

// GetMonitoringStats returns monitoring statistics
func (pm *PRMonitor) GetMonitoringStats() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := map[string]interface{}{
		"active_monitors": len(pm.activeMonitors),
		"monitors":        make([]map[string]interface{}, 0, len(pm.activeMonitors)),
	}

	for prID, session := range pm.activeMonitors {
		monitorInfo := map[string]interface{}{
			"pr_id":       prID,
			"pr_number":   session.pr.GitHubPRNumber,
			"task_id":     session.task.ID,
			"repository":  session.pr.Repository,
			"status":      session.pr.Status,
			"last_poll":   session.lastPoll,
			"retries":     session.retries,
		}
		stats["monitors"] = append(stats["monitors"].([]map[string]interface{}), monitorInfo)
	}

	return stats
}

// IsMonitoring returns true if the PR is currently being monitored
func (pm *PRMonitor) IsMonitoring(prID uuid.UUID) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	_, exists := pm.activeMonitors[prID]
	return exists
}

// sendPRStatusNotification sends a WebSocket notification for PR status changes
func (pm *PRMonitor) sendPRStatusNotification(pr *entity.PullRequest, oldStatus, newStatus string) error {
	ctx, cancel := context.WithTimeout(context.Background(), pm.config.NotificationTimeout)
	defer cancel()

	// Get task to include in notification
	task, err := pm.taskRepo.GetByID(ctx, pr.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task for notification: %w", err)
	}

	notification := map[string]interface{}{
		"type":         "pr_status_change",
		"pr_id":        pr.ID,
		"pr_number":    pr.GitHubPRNumber,
		"task_id":      pr.TaskID,
		"task_title":   task.Title,
		"repository":   pr.Repository,
		"old_status":   oldStatus,
		"new_status":   newStatus,
		"github_url":   pr.GitHubURL,
		"merged_at":    pr.MergedAt,
		"closed_at":    pr.ClosedAt,
		"merged_by":    pr.MergedBy,
		"timestamp":    time.Now(),
	}

	// Send to project channel
	if err := pm.websocketSvc.SendProjectMessage(
		task.ProjectID,
		websocket.MessageTypePRUpdate,
		notification,
	); err != nil {
		return fmt.Errorf("failed to send project notification: %w", err)
	}

	return nil
}

// sendMergeNotification sends a WebSocket notification for PR merge completion
func (pm *PRMonitor) sendMergeNotification(pr *entity.PullRequest, task *entity.Task) error {

	notification := map[string]interface{}{
		"type":              "pr_merged",
		"pr_id":             pr.ID,
		"pr_number":         pr.GitHubPRNumber,
		"task_id":           pr.TaskID,
		"task_title":        task.Title,
		"repository":        pr.Repository,
		"github_url":        pr.GitHubURL,
		"merge_commit_sha":  pr.MergeCommitSHA,
		"merged_at":         pr.MergedAt,
		"merged_by":         pr.MergedBy,
		"task_completed":    true,
		"cleanup_triggered": true,
		"timestamp":         time.Now(),
	}

	// Send to project channel
	if err := pm.websocketSvc.SendProjectMessage(
		task.ProjectID,
		websocket.MessageTypePRUpdate,
		notification,
	); err != nil {
		return fmt.Errorf("failed to send merge notification: %w", err)
	}

	pm.logger.Info("Sent merge notification", 
		"pr_id", pr.ID,
		"task_id", pr.TaskID,
		"project_id", task.ProjectID,
	)

	return nil
}

// sendReviewNotification sends a WebSocket notification for PR reviews
func (pm *PRMonitor) sendReviewNotification(pr *entity.PullRequest, review *entity.PullRequestReview) error {
	ctx, cancel := context.WithTimeout(context.Background(), pm.config.NotificationTimeout)
	defer cancel()

	// Get task information
	task, err := pm.taskRepo.GetByID(ctx, pr.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task for review notification: %w", err)
	}

	notification := map[string]interface{}{
		"type":         "pr_review",
		"pr_id":        pr.ID,
		"pr_number":    pr.GitHubPRNumber,
		"task_id":      pr.TaskID,
		"task_title":   task.Title,
		"repository":   pr.Repository,
		"github_url":   pr.GitHubURL,
		"review_id":    review.ID,
		"reviewer":     review.Reviewer,
		"state":        review.State,
		"body":         review.Body,
		"submitted_at": review.SubmittedAt,
		"timestamp":    time.Now(),
	}

	// Send to project channel
	if err := pm.websocketSvc.SendProjectMessage(
		task.ProjectID,
		websocket.MessageTypePRUpdate,
		notification,
	); err != nil {
		return fmt.Errorf("failed to send review notification: %w", err)
	}

	pm.logger.Info("Sent review notification", 
		"pr_id", pr.ID,
		"review_id", review.ID,
		"reviewer", review.Reviewer,
		"state", review.State,
	)

	return nil
}

// sendErrorNotification sends a WebSocket notification for monitoring errors
func (pm *PRMonitor) sendErrorNotification(pr *entity.PullRequest, err error) error {
	ctx, cancel := context.WithTimeout(context.Background(), pm.config.NotificationTimeout)
	defer cancel()

	// Get task information
	task, err := pm.taskRepo.GetByID(ctx, pr.TaskID)
	if err != nil {
		pm.logger.Error("Failed to get task for error notification", 
			"pr_id", pr.ID,
			"error", err,
		)
		return nil // Don't fail on notification errors
	}

	notification := map[string]interface{}{
		"type":       "pr_monitor_error",
		"pr_id":      pr.ID,
		"pr_number":  pr.GitHubPRNumber,
		"task_id":    pr.TaskID,
		"task_title": task.Title,
		"repository": pr.Repository,
		"error":      err.Error(),
		"timestamp":  time.Now(),
	}

	// Send to project channel
	if sendErr := pm.websocketSvc.SendProjectMessage(
		task.ProjectID,
		websocket.Error,
		notification,
	); sendErr != nil {
		pm.logger.Error("Failed to send error notification", 
			"pr_id", pr.ID,
			"notification_error", sendErr,
			"original_error", err,
		)
		return sendErr
	}

	return nil
}

// MonitorAllActivePRs starts monitoring all active PRs
func (pm *PRMonitor) MonitorAllActivePRs(ctx context.Context) error {
	prs, err := pm.prRepo.GetActiveMonitoringPRs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active PRs: %w", err)
	}

	pm.logger.Info("Starting monitoring for all active PRs", "count", len(prs))

	for _, pr := range prs {
		if err := pm.MonitorPR(pr); err != nil {
			pm.logger.Error("Failed to start monitoring PR", 
				"pr_id", pr.ID,
				"pr_number", pr.GitHubPRNumber,
				"error", err,
			)
			continue
		}
	}

	pm.logger.Info("Monitoring started for active PRs", "active_monitors", len(pm.activeMonitors))
	return nil
}

// RefreshPR forces a refresh of PR data from GitHub
func (pm *PRMonitor) RefreshPR(prID uuid.UUID) error {
	pm.mu.RLock()
	session, exists := pm.activeMonitors[prID]
	pm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("PR %s is not being monitored", prID)
	}

	// Force a poll
	if err := pm.pollPRStatus(session); err != nil {
		return fmt.Errorf("failed to refresh PR: %w", err)
	}

	pm.logger.Info("PR refreshed manually", "pr_id", prID)
	return nil
}