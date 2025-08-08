# GitHub Service - PR Monitoring and Synchronization

This package provides comprehensive GitHub Pull Request monitoring and synchronization capabilities for the Auto-Devs system.

## Components

### 1. PR Monitor Service (`pr_monitor.go`)
Real-time monitoring service that tracks specific PRs and responds to changes immediately.

**Key Features:**
- Real-time PR status monitoring
- WebSocket notifications for status changes
- Task status synchronization
- Automatic worktree cleanup on PR merge
- Concurrent monitoring of multiple PRs

### 2. PR Sync Worker (`pr_sync_worker.go`) 
Background worker that runs every minute to synchronize all active PRs with GitHub.

**Key Features:**
- Periodic synchronization (configurable interval, default: 1 minute)
- Batch processing with concurrency control
- Retry logic for failed API calls
- Comprehensive error handling and logging
- Status change detection and notification

### 3. PR Creator Service (`pr_creator.go`)
Service for creating pull requests automatically from task implementations.

## Quick Start

### Setting up the PR Sync Worker

```go
// 1. Create dependencies
githubService := NewGitHubService(githubConfig)
prRepo := postgres.NewPullRequestRepository(db)
taskRepo := postgres.NewTaskRepository(db)
worktreeService := worktree.NewService(...)
websocketService := websocket.NewService()

// 2. Configure the worker
config := &PRSyncWorkerConfig{
    SyncInterval:       1 * time.Minute,  // Sync every minute
    BatchSize:          20,               // Process 20 PRs per batch
    MaxConcurrentSyncs: 5,                // Max 5 concurrent GitHub API calls
    SyncTimeout:        30 * time.Second, // Timeout per operation
    RetryAttempts:      3,                // Retry failed requests 3 times
    RetryDelay:         5 * time.Second,  // Wait between retries
}

// 3. Create and start the worker
syncWorker := NewPRSyncWorker(
    githubService,
    prRepo,
    taskRepo,
    worktreeService,
    websocketService,
    config,
    logger,
)

ctx := context.Background()
if err := syncWorker.Start(ctx); err != nil {
    log.Fatal("Failed to start PR sync worker:", err)
}

// 4. Setup graceful shutdown
defer syncWorker.Stop()
```

### Using Both Monitor and Worker Together

```go
// The PR Monitor provides real-time monitoring
prMonitor := NewPRMonitor(...)
if err := prMonitor.StartMonitoring(ctx); err != nil {
    log.Fatal("Failed to start PR monitor:", err)
}

// The PR Sync Worker provides periodic backup synchronization
syncWorker := NewPRSyncWorker(...)
if err := syncWorker.Start(ctx); err != nil {
    log.Fatal("Failed to start sync worker:", err)
}

// Both services work together:
// - Monitor: Real-time events, immediate responses
// - Worker: Periodic sync, catches any missed changes
```

## Architecture

### How It Works

1. **PR Creation**: When a task moves to IMPLEMENTING and code is pushed, a PR is created
2. **Monitoring Setup**: The PR Monitor starts tracking the PR for real-time changes
3. **Background Sync**: The Sync Worker runs every minute to ensure consistency
4. **Status Updates**: Both services detect PR status changes and update:
   - PR status in database
   - Associated task status 
   - Send WebSocket notifications
5. **Merge Handling**: When PR is merged:
   - Task status → DONE
   - Worktree cleanup triggered
   - Notifications sent

### PR Status → Task Status Mapping

| PR Status | Task Status | Action |
|-----------|-------------|---------|
| OPEN | CODE_REVIEWING | PR is ready for review |
| MERGED | DONE | Task completed, trigger cleanup |
| CLOSED | CANCELLED | PR closed without merge |

## Configuration

### PR Sync Worker Configuration

```go
type PRSyncWorkerConfig struct {
    SyncInterval       time.Duration // How often to sync (default: 1 minute)
    BatchSize          int          // PRs per batch (default: 10)
    MaxConcurrentSyncs int          // Concurrent API calls (default: 5)
    SyncTimeout        time.Duration // Per-operation timeout (default: 30s)
    RetryAttempts      int          // Retry failed calls (default: 3)
    RetryDelay         time.Duration // Delay between retries (default: 10s)
}
```

### PR Monitor Configuration

```go
type PRMonitorConfig struct {
    PollInterval        time.Duration // Real-time polling interval (default: 5min)
    MaxRetries          int          // Max retries for failed calls (default: 3)
    RetryDelay          time.Duration // Retry delay (default: 30s)
    ConcurrentMonitors  int          // Max concurrent monitors (default: 5)
    NotificationTimeout time.Duration // Notification timeout (default: 10s)
}
```

## Monitoring and Health Checks

### Worker Statistics

```go
stats := syncWorker.GetStats()
// Returns:
// {
//     "running": true,
//     "last_sync_at": "2023-12-07T10:30:00Z",
//     "sync_count": 42,
//     "error_count": 1,
//     "config": {...}
// }
```

### Health Check Example

```go
func healthCheck(syncWorker *PRSyncWorker) map[string]interface{} {
    stats := syncWorker.GetStats()
    
    status := "healthy"
    if !stats["running"].(bool) {
        status = "unhealthy"
    }
    
    return map[string]interface{}{
        "service": "pr_sync_worker",
        "status": status,
        "stats": stats,
    }
}
```

## Manual Operations

### Force Immediate Sync

```go
// Trigger immediate sync (useful for webhooks or manual refresh)
if err := syncWorker.ForceSync(ctx); err != nil {
    log.Printf("Failed to force sync: %v", err)
}
```

### Monitor Specific PR

```go
// Start monitoring a specific PR
pr := &entity.PullRequest{...}
if err := prMonitor.MonitorPR(pr); err != nil {
    log.Printf("Failed to monitor PR: %v", err)
}
```

## Error Handling

The services provide comprehensive error handling:

- **Retry Logic**: Failed GitHub API calls are retried automatically
- **Timeout Protection**: All operations have configurable timeouts  
- **Graceful Degradation**: Individual PR sync failures don't stop the batch
- **Detailed Logging**: All operations are logged with context
- **Health Monitoring**: Services expose metrics for monitoring

## Integration with Existing Services

### WebSocket Notifications

Both services send real-time notifications via WebSocket:

```go
// PR status change notification
{
    "type": "pr_status_change",
    "pr_id": "uuid",
    "pr_number": 123,
    "task_id": "uuid", 
    "old_status": "OPEN",
    "new_status": "MERGED",
    "sync_source": "worker", // or "monitor"
    "timestamp": "2023-12-07T10:30:00Z"
}
```

### Task Status Integration

PR changes automatically update associated task status:

```go
// Task status change notification  
{
    "type": "task_status_change",
    "task_id": "uuid",
    "old_status": "CODE_REVIEWING", 
    "new_status": "DONE",
    "pr_trigger": true,
    "timestamp": "2023-12-07T10:30:00Z"
}
```

## Best Practices

1. **Run Both Services**: Use PR Monitor for real-time + Sync Worker for consistency
2. **Configure Intervals**: 1-minute sync is good for most cases, adjust based on activity
3. **Monitor Health**: Set up alerts on worker health and error rates
4. **Handle Rate Limits**: GitHub API has rate limits, configure concurrency accordingly
5. **Graceful Shutdown**: Always stop workers cleanly on application shutdown

## Testing

Comprehensive tests are provided for all components:

```bash
# Run all GitHub service tests
go test ./internal/service/github/ -v

# Run specific worker tests  
go test ./internal/service/github/ -v -run="TestPRSync"

# Run with coverage
go test ./internal/service/github/ -cover
```

## Troubleshooting

### Worker Not Syncing
- Check if worker is running: `syncWorker.IsRunning()`
- Verify GitHub token and permissions
- Check rate limit status
- Review error logs for API failures

### Missing Status Updates
- Ensure WebSocket service is running
- Check database connectivity
- Verify PR-Task associations in database
- Review notification logs

### High API Usage
- Reduce sync frequency (`SyncInterval`)
- Lower concurrent connections (`MaxConcurrentSyncs`) 
- Implement webhook-based updates instead of polling