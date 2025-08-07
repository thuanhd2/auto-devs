package main

import (
	"log/slog"
	"time"

	"github.com/auto-devs/auto-devs/internal/jobs"
	"github.com/auto-devs/auto-devs/internal/websocket"
	"github.com/google/uuid"
)

// ExampleRedisBrokerUsage demonstrates how to use Redis broker for cross-process messaging
func ExampleRedisBrokerUsage() {
	// Example 1: Server setup with Redis broker
	exampleServerSetup()

	// Example 2: Worker setup with Redis broker client
	exampleWorkerSetup()

	// Example 3: Cross-process messaging
	exampleCrossProcessMessaging()
}

// exampleServerSetup shows how to setup server with Redis broker
func exampleServerSetup() {
	// Create WebSocket service with Redis broker
	wsService := websocket.NewServiceWithRedisBroker("localhost:6379", "", 0)

	// Start Redis broker
	if err := wsService.StartRedisBroker(); err != nil {
		slog.Error("Failed to start Redis broker", "error", err)
		return
	}
	defer wsService.StopRedisBroker()

	slog.Info("Server started with Redis broker")

	// The server will now:
	// 1. Listen for WebSocket connections
	// 2. Listen for Redis messages from workers
	// 3. Broadcast messages to connected clients
}

// exampleWorkerSetup shows how to setup worker with Redis broker client
func exampleWorkerSetup() {
	// Create Redis broker client
	redisClient := jobs.NewRedisBrokerClient("localhost:6379", "", 0)
	defer redisClient.Close()

	// Test connection
	if err := redisClient.TestConnection(); err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		return
	}

	slog.Info("Worker connected to Redis broker")

	// Worker can now publish messages to Redis
	// These messages will be received by the server and broadcast to WebSocket clients
}

// exampleCrossProcessMessaging shows cross-process messaging
func exampleCrossProcessMessaging() {
	// Simulate worker publishing a message
	redisClient := jobs.NewRedisBrokerClient("localhost:6379", "", 0)
	defer redisClient.Close()

	taskID := uuid.New()
	projectID := uuid.New()

	changes := map[string]interface{}{
		"status": map[string]interface{}{
			"old": "TODO",
			"new": "IN_PROGRESS",
		},
	}

	taskResponse := map[string]interface{}{
		"id":         taskID.String(),
		"project_id": projectID.String(),
		"title":      "Example task",
		"status":     "IN_PROGRESS",
		"updated_at": time.Now(),
	}

	// Publish task update
	if err := redisClient.PublishTaskUpdated(taskID, projectID, changes, taskResponse); err != nil {
		slog.Error("Failed to publish task update", "error", err)
		return
	}

	// Publish status change
	if err := redisClient.PublishStatusChanged(taskID, projectID, "task", "TODO", "IN_PROGRESS"); err != nil {
		slog.Error("Failed to publish status change", "error", err)
		return
	}

	slog.Info("Published messages to Redis broker", "task_id", taskID)
}

// ExampleIntegrationWithServer shows how to integrate Redis broker with server
func ExampleIntegrationWithServer() {
	// In your server main function:

	// 1. Create WebSocket service with Redis broker
	wsService := websocket.NewServiceWithRedisBroker("localhost:6379", "", 0)

	// 2. Start Redis broker
	if err := wsService.StartRedisBroker(); err != nil {
		slog.Error("Failed to start Redis broker", "error", err)
		return
	}

	// 3. Setup routes
	// handler.SetupRoutes(router, wsService)

	// 4. Start server
	// server.ListenAndServe()

	// 5. Graceful shutdown
	defer func() {
		wsService.StopRedisBroker()
	}()

	slog.Info("Server running with Redis broker integration")
}

// ExampleIntegrationWithWorker shows how to integrate Redis broker with worker
func ExampleIntegrationWithWorker() {
	// In your worker main function:

	// 1. Create Redis broker client
	redisClient := jobs.NewRedisBrokerClient("localhost:6379", "", 0)
	defer redisClient.Close()

	// 2. Test connection
	if err := redisClient.TestConnection(); err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		return
	}

	// 3. Create processor with Redis broker
	// processor := jobs.NewProcessorWithRedisBroker(
	//     taskUsecase,
	//     projectUsecase,
	//     worktreeUsecase,
	//     planningService,
	//     executionService,
	//     planRepo,
	//     wsService,
	//     redisClient,
	// )

	// 4. Start job processing
	// server.Start()

	slog.Info("Worker running with Redis broker integration")
}
