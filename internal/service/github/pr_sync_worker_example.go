package github

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/auto-devs/auto-devs/internal/repository/postgres"
	"github.com/auto-devs/auto-devs/internal/websocket"
	"gorm.io/gorm"
)

// ExamplePRSyncWorkerUsage demonstrates how to integrate and use the PR sync worker
func ExamplePRSyncWorkerUsage() {
	// This example shows how to set up and run the PR sync worker
	// in your application alongside the existing PR monitor

	// 1. Setup dependencies (normally injected via DI)
	var db *gorm.DB // Your database connection
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create repositories
	_ = postgres.NewPullRequestRepository(db) // prRepo would be used
	// taskRepo := postgres.NewTaskRepository(db) // Would need proper implementation

	// Create services
	_ = &GitHubService{}         // githubService would be used
	_ = &websocket.Service{}     // websocketSvc would be used
	_ = "worktreeService"        // worktreeService would be used (optional)

	// 2. Configure PR sync worker
	_ = &PRSyncWorkerConfig{ // config would be used
		SyncInterval:       1 * time.Minute, // Sync every minute as requested
		BatchSize:          20,              // Process 20 PRs at a time
		MaxConcurrentSyncs: 5,               // Max 5 concurrent GitHub API calls
		SyncTimeout:        30 * time.Second, // Timeout per sync operation
		RetryAttempts:      3,               // Retry failed requests 3 times
		RetryDelay:         5 * time.Second,  // Wait 5s between retries
	}

	// 3. Create PR sync worker
	// syncWorker := NewPRSyncWorker(
	// 	githubService,
	// 	prRepo,
	// 	taskRepo,
	// 	worktreeService,
	// 	websocketSvc,
	// 	config,
	// 	logger,
	// )

	// 4. Start the worker
	_ = context.Background() // ctx would be used
	// if err := syncWorker.Start(ctx); err != nil {
	// 	log.Printf("Failed to start PR sync worker: %v", err)
	// 	return
	// }

	// 5. Setup graceful shutdown
	setupGracefulShutdown := func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigChan
			logger.Info("Received shutdown signal, stopping PR sync worker...")
			
			// if err := syncWorker.Stop(); err != nil {
			// 	logger.Error("Failed to stop PR sync worker", "error", err)
			// }
			
			logger.Info("PR sync worker stopped gracefully")
			os.Exit(0)
		}()
	}

	setupGracefulShutdown()

	// 6. Monitor worker status (optional)
	// go func() {
	// 	ticker := time.NewTicker(5 * time.Minute)
	// 	defer ticker.Stop()

	// 	for {
	// 		select {
	// 		case <-ticker.C:
	// 			// stats := syncWorker.GetStats()
	// 			// logger.Info("PR sync worker stats", "stats", stats)
	// 		case <-ctx.Done():
	// 			return
	// 		}
	// 	}
	// }()

	// Keep main goroutine alive
	select {}
}

// ExampleIntegrationWithExistingServices shows how to integrate with existing services
func ExampleIntegrationWithExistingServices() {
	// If you already have a PR monitor service, you can run both together
	// The sync worker provides regular background synchronization
	// The PR monitor provides real-time event handling

	var (
		// githubService GitHubServiceInterface
		// prRepo        PRRepository  
		// taskRepo      TaskRepository
		// worktreeService WorktreeService
		// websocketSvc    WebSocketServiceInterface
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	)

	// Create both services
	// prMonitor := NewPRMonitor(
	// 	githubService,
	// 	prRepo,
	// 	taskRepo,
	// 	worktreeService,
	// 	websocketSvc,
	// 	DefaultPRMonitorConfig(),
	// 	logger,
	// )

	// syncWorker := NewPRSyncWorker(
	// 	githubService,
	// 	prRepo,
	// 	taskRepo,
	// 	worktreeService,
	// 	websocketSvc,
	// 	DefaultPRSyncWorkerConfig(),
	// 	logger,
	// )

	_ = context.Background() // ctx would be used

	// Start both services
	// if err := prMonitor.StartMonitoring(ctx); err != nil {
	// 	log.Printf("Failed to start PR monitor: %v", err)
	// }

	// if err := syncWorker.Start(ctx); err != nil {
	// 	log.Printf("Failed to start PR sync worker: %v", err)
	// }

	logger.Info("Both PR monitor and sync worker are running")

	// The sync worker will:
	// 1. Run every minute to fetch all active PRs from GitHub
	// 2. Compare with database status and update if different
	// 3. Handle status changes and send notifications
	// 4. Provide a fallback for any PRs that the monitor might miss

	// The PR monitor will:
	// 1. Monitor specific PRs in real-time
	// 2. Handle immediate status changes
	// 3. Process webhooks (if implemented)
	// 4. Provide faster response times for active PRs
}

// ExampleCustomSyncScheduling shows how to customize sync scheduling
func ExampleCustomSyncScheduling() {
	// You can customize the sync scheduling based on your needs

	configs := map[string]*PRSyncWorkerConfig{
		"high_frequency": {
			SyncInterval:       30 * time.Second, // Every 30 seconds
			BatchSize:          5,
			MaxConcurrentSyncs: 3,
			SyncTimeout:        15 * time.Second,
			RetryAttempts:      2,
			RetryDelay:         3 * time.Second,
		},
		"standard": {
			SyncInterval:       1 * time.Minute, // Every minute (as requested)
			BatchSize:          20,
			MaxConcurrentSyncs: 5,
			SyncTimeout:        30 * time.Second,
			RetryAttempts:      3,
			RetryDelay:         5 * time.Second,
		},
		"low_frequency": {
			SyncInterval:       5 * time.Minute, // Every 5 minutes
			BatchSize:          50,
			MaxConcurrentSyncs: 10,
			SyncTimeout:        45 * time.Second,
			RetryAttempts:      3,
			RetryDelay:         10 * time.Second,
		},
	}

	// Choose configuration based on your needs
	// For high-activity repositories: use "high_frequency"
	// For normal usage: use "standard" (1 minute as requested)
	// For low-activity or rate-limit concerns: use "low_frequency"

	selectedConfig := configs["standard"] // 1 minute sync as requested
	_ = selectedConfig

	// You can also dynamically adjust the configuration
	// based on repository activity, time of day, etc.
}

// ExampleMonitoringAndAlerting shows how to monitor the sync worker
func ExampleMonitoringAndAlerting() {
	// var syncWorker *PRSyncWorker
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create a monitoring goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C

			// stats := syncWorker.GetStats()
			
			// Check if worker is healthy
			// if !stats["running"].(bool) {
			// 	logger.Error("PR sync worker is not running!")
			// 	// Send alert to monitoring system
			// }

			// Check error rate
			// errorCount := stats["error_count"].(int64)
			// syncCount := stats["sync_count"].(int64)
			
			// if syncCount > 0 {
			// 	errorRate := float64(errorCount) / float64(syncCount)
			// 	if errorRate > 0.1 { // More than 10% error rate
			// 		logger.Warn("High error rate in PR sync worker",
			// 			"error_rate", errorRate,
			// 			"error_count", errorCount,
			// 			"sync_count", syncCount,
			// 		)
			// 		// Send alert
			// 	}
			// }

			// Check last sync time
			// lastSyncAt := stats["last_sync_at"].(time.Time)
			// if !lastSyncAt.IsZero() && time.Since(lastSyncAt) > 5*time.Minute {
			// 	logger.Error("PR sync worker hasn't synced in a while",
			// 		"last_sync", lastSyncAt,
			// 		"minutes_ago", time.Since(lastSyncAt).Minutes(),
			// 	)
			// 	// Send alert
			// }

			logger.Info("PR sync worker health check completed")
		}
	}()
}

// ExampleManualSync shows how to trigger manual synchronization
func ExampleManualSync() {
	// var syncWorker *PRSyncWorker

	// You can force an immediate sync (useful for webhooks or manual triggers)
	_ = context.Background() // ctx would be used
	
	// if err := syncWorker.ForceSync(ctx); err != nil {
	// 	log.Printf("Failed to force sync: %v", err)
	// } else {
	// 	log.Printf("Manual sync triggered successfully")
	// }

	// This is useful when:
	// 1. Receiving GitHub webhooks
	// 2. User requests immediate refresh
	// 3. After system maintenance
	// 4. Testing purposes
}

// ExampleHealthCheck shows how to implement health checks
func ExampleHealthCheck() {
	// var syncWorker *PRSyncWorker

	healthCheck := func() map[string]interface{} {
		// stats := syncWorker.GetStats()
		
		status := "healthy"
		// if !stats["running"].(bool) {
		// 	status = "unhealthy"
		// }

		return map[string]interface{}{
			"service":    "pr_sync_worker",
			"status":     status,
			"timestamp":  time.Now(),
			// "stats":      stats,
		}
	}

	// Use in your health check endpoint
	health := healthCheck()
	_ = health
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(health)
}