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

// PRSyncWorkerConfig holds configuration for the PR synchronization worker
type PRSyncWorkerConfig struct {
	SyncInterval       time.Duration // How often to sync PRs from GitHub
	BatchSize          int           // Number of PRs to process in each batch
	MaxConcurrentSyncs int           // Maximum number of concurrent sync operations
	SyncTimeout        time.Duration // Timeout for individual sync operations
	RetryAttempts      int           // Number of retry attempts for failed syncs
	RetryDelay         time.Duration // Delay between retry attempts
}

// DefaultPRSyncWorkerConfig returns default configuration
func DefaultPRSyncWorkerConfig() *PRSyncWorkerConfig {
	return &PRSyncWorkerConfig{
		SyncInterval:       1 * time.Minute,
		BatchSize:          10,
		MaxConcurrentSyncs: 5,
		SyncTimeout:        30 * time.Second,
		RetryAttempts:      3,
		RetryDelay:         10 * time.Second,
	}
}

// PRSyncWorker handles periodic synchronization of PR status from GitHub
type PRSyncWorker struct {
	githubService   GitHubServiceInterface
	prRepo          PRRepository
	taskRepo        TaskRepository
	worktreeService WorktreeService
	websocketSvc    WebSocketServiceInterface
	config          *PRSyncWorkerConfig
	logger          *slog.Logger

	// Worker state
	running    bool
	stopCh     chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex
	lastSyncAt time.Time
	syncCount  int64
	errorCount int64
}

// SyncResult represents the result of a PR synchronization operation
type SyncResult struct {
	PRID         uuid.UUID
	PRNumber     int
	Repository   string
	Success      bool
	StatusChange bool
	OldStatus    entity.PullRequestStatus
	NewStatus    entity.PullRequestStatus
	Error        error
	SyncedAt     time.Time
}

// NewPRSyncWorker creates a new PR synchronization worker
func NewPRSyncWorker(
	githubService GitHubServiceInterface,
	prRepo PRRepository,
	taskRepo TaskRepository,
	worktreeService WorktreeService,
	websocketSvc WebSocketServiceInterface,
	config *PRSyncWorkerConfig,
	logger *slog.Logger,
) *PRSyncWorker {
	if config == nil {
		config = DefaultPRSyncWorkerConfig()
	}

	return &PRSyncWorker{
		githubService:   githubService,
		prRepo:          prRepo,
		taskRepo:        taskRepo,
		worktreeService: worktreeService,
		websocketSvc:    websocketSvc,
		config:          config,
		logger:          logger.With("component", "pr_sync_worker"),
		stopCh:          make(chan struct{}),
	}
}

// Start starts the PR synchronization worker
func (w *PRSyncWorker) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return fmt.Errorf("PR sync worker is already running")
	}

	w.running = true
	w.wg.Add(1)

	go w.workerLoop(ctx)

	w.logger.Info("PR sync worker started",
		"sync_interval", w.config.SyncInterval,
		"batch_size", w.config.BatchSize,
		"max_concurrent", w.config.MaxConcurrentSyncs,
	)

	return nil
}

// Stop stops the PR synchronization worker
func (w *PRSyncWorker) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return fmt.Errorf("PR sync worker is not running")
	}

	close(w.stopCh)
	w.wg.Wait()
	w.running = false

	w.logger.Info("PR sync worker stopped",
		"total_syncs", w.syncCount,
		"total_errors", w.errorCount,
	)

	return nil
}

// IsRunning returns true if the worker is currently running
func (w *PRSyncWorker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// GetStats returns worker statistics
func (w *PRSyncWorker) GetStats() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return map[string]interface{}{
		"running":      w.running,
		"last_sync_at": w.lastSyncAt,
		"sync_count":   w.syncCount,
		"error_count":  w.errorCount,
		"config": map[string]interface{}{
			"sync_interval":        w.config.SyncInterval,
			"batch_size":          w.config.BatchSize,
			"max_concurrent_syncs": w.config.MaxConcurrentSyncs,
			"sync_timeout":        w.config.SyncTimeout,
		},
	}
}

// workerLoop runs the main worker loop
func (w *PRSyncWorker) workerLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.SyncInterval)
	defer ticker.Stop()

	w.logger.Info("Starting PR sync worker loop")

	// Run initial sync
	w.performSync(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("PR sync worker stopped due to context cancellation")
			return
		case <-w.stopCh:
			w.logger.Info("PR sync worker stopped")
			return
		case <-ticker.C:
			w.performSync(ctx)
		}
	}
}

// performSync performs a single synchronization cycle
func (w *PRSyncWorker) performSync(ctx context.Context) {
	syncStart := time.Now()
	w.logger.Info("Starting PR synchronization cycle")

	// Create timeout context for this sync cycle
	syncCtx, cancel := context.WithTimeout(ctx, w.config.SyncTimeout*2) // Allow extra time for the full cycle
	defer cancel()

	// Get all PRs that need synchronization
	prs, err := w.prRepo.GetActiveMonitoringPRs(syncCtx)
	if err != nil {
		w.incrementErrorCount()
		w.logger.Error("Failed to get PRs for synchronization", "error", err)
		return
	}

	if len(prs) == 0 {
		w.logger.Debug("No PRs found for synchronization")
		w.updateLastSyncTime()
		return
	}

	w.logger.Info("Starting batch synchronization", 
		"total_prs", len(prs),
		"batch_size", w.config.BatchSize,
	)

	// Process PRs in batches
	results := w.syncPRsBatch(syncCtx, prs)

	// Process results
	w.processSyncResults(syncCtx, results)

	syncDuration := time.Since(syncStart)
	w.updateLastSyncTime()
	w.incrementSyncCount()

	w.logger.Info("PR synchronization cycle completed",
		"duration", syncDuration,
		"total_prs", len(prs),
		"successful_syncs", w.countSuccessfulSyncs(results),
		"failed_syncs", w.countFailedSyncs(results),
	)
}

// syncPRsBatch synchronizes PRs in batches with concurrency control
func (w *PRSyncWorker) syncPRsBatch(ctx context.Context, prs []*entity.PullRequest) []*SyncResult {
	var allResults []*SyncResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, w.config.MaxConcurrentSyncs)

	// Process PRs in batches
	for i := 0; i < len(prs); i += w.config.BatchSize {
		end := i + w.config.BatchSize
		if end > len(prs) {
			end = len(prs)
		}

		batch := prs[i:end]
		w.logger.Debug("Processing PR batch", 
			"batch_start", i,
			"batch_end", end,
			"batch_size", len(batch),
		)

		// Process each PR in the batch concurrently
		for _, pr := range batch {
			wg.Add(1)
			go func(pr *entity.PullRequest) {
				defer wg.Done()

				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				result := w.syncSinglePR(ctx, pr)
				
				mu.Lock()
				allResults = append(allResults, result)
				mu.Unlock()
			}(pr)
		}
	}

	// Wait for all goroutines to complete
	wg.Wait()

	return allResults
}

// syncSinglePR synchronizes a single PR with GitHub
func (w *PRSyncWorker) syncSinglePR(ctx context.Context, pr *entity.PullRequest) *SyncResult {
	result := &SyncResult{
		PRID:       pr.ID,
		PRNumber:   pr.GitHubPRNumber,
		Repository: pr.Repository,
		SyncedAt:   time.Now(),
	}

	w.logger.Debug("Syncing PR", 
		"pr_id", pr.ID,
		"pr_number", pr.GitHubPRNumber,
		"repository", pr.Repository,
		"current_status", pr.Status,
	)

	// Create timeout context for this individual sync
	syncCtx, cancel := context.WithTimeout(ctx, w.config.SyncTimeout)
	defer cancel()

	// Fetch PR from GitHub with retry logic
	githubPR, err := w.fetchPRWithRetry(syncCtx, pr.Repository, pr.GitHubPRNumber)
	if err != nil {
		result.Success = false
		result.Error = err
		w.logger.Error("Failed to fetch PR from GitHub",
			"pr_id", pr.ID,
			"pr_number", pr.GitHubPRNumber,
			"repository", pr.Repository,
			"error", err,
		)
		return result
	}

	// Check for status changes
	result.OldStatus = pr.Status
	result.NewStatus = githubPR.Status

	if pr.Status != githubPR.Status {
		result.StatusChange = true
		w.logger.Info("PR status changed",
			"pr_id", pr.ID,
			"pr_number", pr.GitHubPRNumber,
			"old_status", pr.Status,
			"new_status", githubPR.Status,
		)

		// Update PR with new data from GitHub
		pr.Status = githubPR.Status
		pr.MergedAt = githubPR.MergedAt
		pr.ClosedAt = githubPR.ClosedAt
		pr.MergeCommitSHA = githubPR.MergeCommitSHA
		pr.MergedBy = githubPR.MergedBy
		pr.Mergeable = githubPR.Mergeable
		pr.MergeableState = githubPR.MergeableState

		// Save to database
		if err := w.prRepo.Update(syncCtx, pr); err != nil {
			result.Success = false
			result.Error = fmt.Errorf("failed to update PR in database: %w", err)
			w.logger.Error("Failed to update PR in database",
				"pr_id", pr.ID,
				"error", err,
			)
			return result
		}

		// Handle status change (similar to PR monitor)
		if err := w.handleStatusChange(syncCtx, pr, string(result.OldStatus), string(result.NewStatus)); err != nil {
			w.logger.Error("Failed to handle PR status change",
				"pr_id", pr.ID,
				"old_status", result.OldStatus,
				"new_status", result.NewStatus,
				"error", err,
			)
			// Don't fail the sync for notification errors
		}
	} else {
		w.logger.Debug("PR status unchanged",
			"pr_id", pr.ID,
			"status", pr.Status,
		)
	}

	result.Success = true
	return result
}

// fetchPRWithRetry fetches a PR from GitHub with retry logic
func (w *PRSyncWorker) fetchPRWithRetry(ctx context.Context, repo string, prNumber int) (*entity.PullRequest, error) {
	var lastErr error

	for attempt := 0; attempt <= w.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(w.config.RetryDelay):
				// Continue to retry
			}

			w.logger.Debug("Retrying GitHub API call",
				"repository", repo,
				"pr_number", prNumber,
				"attempt", attempt,
			)
		}

		pr, err := w.githubService.GetPullRequest(ctx, repo, prNumber)
		if err == nil {
			return pr, nil
		}

		lastErr = err
		w.logger.Warn("GitHub API call failed",
			"repository", repo,
			"pr_number", prNumber,
			"attempt", attempt,
			"error", err,
		)
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", w.config.RetryAttempts, lastErr)
}

// handleStatusChange handles PR status changes (similar to PR monitor)
func (w *PRSyncWorker) handleStatusChange(ctx context.Context, pr *entity.PullRequest, oldStatus, newStatus string) error {
	// Get the associated task
	task, err := w.taskRepo.GetByID(ctx, pr.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	oldTaskStatus := task.Status
	var newTaskStatus entity.TaskStatus

	// Map PR status to task status
	switch entity.PullRequestStatus(newStatus) {
	case entity.PullRequestStatusOpen:
		newTaskStatus = entity.TaskStatusCODEREVIEWING
	case entity.PullRequestStatusMerged:
		newTaskStatus = entity.TaskStatusDONE
	case entity.PullRequestStatusClosed:
		if task.Status == entity.TaskStatusCODEREVIEWING {
			newTaskStatus = entity.TaskStatusCANCELLED
		} else {
			newTaskStatus = task.Status
		}
	}

	// Update task status if it changed
	if newTaskStatus != oldTaskStatus {
		if err := w.taskRepo.UpdateStatus(ctx, task.ID, newTaskStatus); err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}

		w.logger.Info("Updated task status due to PR sync",
			"task_id", task.ID,
			"pr_id", pr.ID,
			"old_task_status", oldTaskStatus,
			"new_task_status", newTaskStatus,
			"pr_status", newStatus,
		)

		// Send WebSocket notification for task status change
		if err := w.websocketSvc.NotifyStatusChanged(
			task.ID,
			task.ProjectID,
			"task",
			string(oldTaskStatus),
			string(newTaskStatus),
		); err != nil {
			w.logger.Error("Failed to send task status notification",
				"task_id", task.ID,
				"error", err,
			)
		}
	}

	// Send PR status change notification
	if err := w.sendStatusChangeNotification(ctx, pr, task, oldStatus, newStatus); err != nil {
		w.logger.Error("Failed to send PR status notification",
			"pr_id", pr.ID,
			"error", err,
		)
	}

	// Handle merge completion if PR was merged
	if entity.PullRequestStatus(newStatus) == entity.PullRequestStatusMerged && pr.MergedAt != nil {
		if err := w.handleMergeCompletion(ctx, pr, task); err != nil {
			w.logger.Error("Failed to handle merge completion",
				"pr_id", pr.ID,
				"error", err,
			)
		}
	}

	return nil
}

// sendStatusChangeNotification sends a WebSocket notification for PR status changes
func (w *PRSyncWorker) sendStatusChangeNotification(ctx context.Context, pr *entity.PullRequest, task *entity.Task, oldStatus, newStatus string) error {
	notification := map[string]interface{}{
		"type":         "pr_status_sync",
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
		"sync_source":  "worker",
		"timestamp":    time.Now(),
	}

	return w.websocketSvc.SendProjectMessage(
		task.ProjectID,
		websocket.MessageTypePRUpdate,
		notification,
	)
}

// handleMergeCompletion handles PR merge completion
func (w *PRSyncWorker) handleMergeCompletion(ctx context.Context, pr *entity.PullRequest, task *entity.Task) error {
	w.logger.Info("Handling PR merge completion from sync",
		"pr_id", pr.ID,
		"pr_number", pr.GitHubPRNumber,
		"task_id", pr.TaskID,
	)

	// Trigger worktree cleanup if needed
	if w.worktreeService != nil {
		worktree, err := w.worktreeService.GetWorktreeByTaskID(ctx, pr.TaskID)
		if err != nil {
			w.logger.Warn("Failed to get worktree for cleanup",
				"task_id", pr.TaskID,
				"error", err,
			)
		} else if worktree != nil {
			if err := w.worktreeService.CleanupTaskWorktree(ctx, pr.TaskID, task.ProjectID); err != nil {
				w.logger.Error("Failed to cleanup worktree",
					"task_id", pr.TaskID,
					"worktree_id", worktree.ID,
					"error", err,
				)
			} else {
				w.logger.Info("Worktree cleanup triggered from PR sync",
					"task_id", pr.TaskID,
					"worktree_id", worktree.ID,
				)
			}
		}
	}

	// Send merge completion notification
	notification := map[string]interface{}{
		"type":              "pr_merged_sync",
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
		"cleanup_triggered": w.worktreeService != nil,
		"sync_source":       "worker",
		"timestamp":         time.Now(),
	}

	return w.websocketSvc.SendProjectMessage(
		task.ProjectID,
		websocket.MessageTypePRUpdate,
		notification,
	)
}

// processSyncResults processes the results of a sync batch
func (w *PRSyncWorker) processSyncResults(ctx context.Context, results []*SyncResult) {
	successCount := w.countSuccessfulSyncs(results)
	failureCount := w.countFailedSyncs(results)
	statusChangeCount := w.countStatusChanges(results)

	w.logger.Info("Sync batch results",
		"total", len(results),
		"successful", successCount,
		"failed", failureCount,
		"status_changes", statusChangeCount,
	)

	// Log any failures
	for _, result := range results {
		if !result.Success {
			w.logger.Warn("PR sync failed",
				"pr_id", result.PRID,
				"pr_number", result.PRNumber,
				"repository", result.Repository,
				"error", result.Error,
			)
		} else if result.StatusChange {
			w.logger.Info("PR status synchronized",
				"pr_id", result.PRID,
				"pr_number", result.PRNumber,
				"old_status", result.OldStatus,
				"new_status", result.NewStatus,
			)
		}
	}
}

// Helper methods for counting results
func (w *PRSyncWorker) countSuccessfulSyncs(results []*SyncResult) int {
	count := 0
	for _, result := range results {
		if result.Success {
			count++
		}
	}
	return count
}

func (w *PRSyncWorker) countFailedSyncs(results []*SyncResult) int {
	count := 0
	for _, result := range results {
		if !result.Success {
			count++
		}
	}
	return count
}

func (w *PRSyncWorker) countStatusChanges(results []*SyncResult) int {
	count := 0
	for _, result := range results {
		if result.StatusChange {
			count++
		}
	}
	return count
}

// Helper methods for updating worker state
func (w *PRSyncWorker) updateLastSyncTime() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lastSyncAt = time.Now()
}

func (w *PRSyncWorker) incrementSyncCount() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.syncCount++
}

func (w *PRSyncWorker) incrementErrorCount() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.errorCount++
}

// ForcSync triggers an immediate synchronization (useful for testing or manual triggers)
func (w *PRSyncWorker) ForceSync(ctx context.Context) error {
	if !w.IsRunning() {
		return fmt.Errorf("worker is not running")
	}

	w.logger.Info("Forcing immediate PR synchronization")
	go w.performSync(ctx)
	return nil
}