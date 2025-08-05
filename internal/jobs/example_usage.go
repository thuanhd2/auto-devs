package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/google/uuid"
)

// ExampleStartPlanningJob demonstrates how to use the job client integration
func ExampleStartPlanningJob() {
	// 1. Setup Redis connection
	client := NewClient("localhost:6379", "", 0)
	defer client.Close()

	// 2. Create adapter for usecase
	adapter := NewJobClientAdapter(client)

	// 3. Create task planning payload
	taskID := uuid.New()
	projectID := uuid.New()
	branchName := "feature/example-branch"

	payload := &usecase.TaskPlanningPayload{
		TaskID:     taskID,
		BranchName: branchName,
		ProjectID:  projectID,
	}

	// 4. Enqueue the job
	jobID, err := adapter.EnqueueTaskPlanning(payload, 0)
	if err != nil {
		log.Fatalf("Failed to enqueue planning job: %v", err)
	}

	fmt.Printf("Planning job enqueued successfully with ID: %s\n", jobID)
	fmt.Printf("Task ID: %s\n", taskID)
	fmt.Printf("Project ID: %s\n", projectID)
	fmt.Printf("Branch: %s\n", branchName)
}

// ExampleStartPlanningJobWithDelay demonstrates enqueueing with delay
func ExampleStartPlanningJobWithDelay() {
	// Setup
	client := NewClient("localhost:6379", "", 0)
	defer client.Close()

	adapter := NewJobClientAdapter(client)

	// Create payload
	taskID := uuid.New()
	projectID := uuid.New()
	branchName := "feature/delayed-branch"

	payload := &usecase.TaskPlanningPayload{
		TaskID:     taskID,
		BranchName: branchName,
		ProjectID:  projectID,
	}

	// Enqueue with 5 minute delay
	delay := 5 * time.Minute
	jobID, err := adapter.EnqueueTaskPlanning(payload, delay)
	if err != nil {
		log.Fatalf("Failed to enqueue delayed planning job: %v", err)
	}

	fmt.Printf("Delayed planning job enqueued successfully with ID: %s\n", jobID)
	fmt.Printf("Job will be processed in %v\n", delay)
}

// ExampleJobProcessing demonstrates how jobs are processed
func ExampleJobProcessing() {
	// This would typically be in a separate worker process
	// or run as a background service

	// 1. Setup Redis connection
	redisAddr := "localhost:6379"
	redisPassword := ""
	redisDB := 0

	// 2. Create processor with dependencies
	// Note: In real application, these would be injected via DI
	processor := &Processor{
		// taskUsecase: taskUsecase,
		// projectUsecase: projectUsecase,
		// gitManager: gitManager,
		// worktreeManager: worktreeManager,
	}

	// 3. Create server
	server := NewServer(redisAddr, redisPassword, redisDB, processor)

	// 4. Register handlers
	server.RegisterHandlers()

	// 5. Start processing (this would block)
	fmt.Println("Starting job server...")
	// err := server.Start()
	// if err != nil {
	//     log.Fatalf("Failed to start job server: %v", err)
	// }
}

// ExampleTaskUsecaseIntegration demonstrates integration with TaskUsecase
func ExampleTaskUsecaseIntegration() {
	// This shows how the job client is integrated into TaskUsecase

	// 1. Setup dependencies (in real app, this would be done via DI)
	client := NewClient("localhost:6379", "", 0)
	defer client.Close()

	_ = NewJobClientAdapter(client) // adapter would be used in real implementation

	// 2. Create task usecase with job client
	// taskUsecase := &taskUsecase{
	//     jobClient: adapter,
	//     // other dependencies...
	// }

	// 3. Use the StartPlanning method
	_ = context.Background() // ctx would be used in real implementation
	taskID := uuid.New()
	branchName := "feature/integration-test"

	// jobID, err := taskUsecase.StartPlanning(ctx, taskID, branchName)
	// if err != nil {
	//     log.Fatalf("Failed to start planning: %v", err)
	// }

	// fmt.Printf("Planning started for task %s with job ID: %s\n", taskID, jobID)
	fmt.Printf("Example: Planning would be started for task %s with branch %s\n", taskID, branchName)
}

// ExampleErrorHandling demonstrates error handling patterns
func ExampleErrorHandling() {
	client := NewClient("localhost:6379", "", 0)
	defer client.Close()

	adapter := NewJobClientAdapter(client)

	// Example payload
	payload := &usecase.TaskPlanningPayload{
		TaskID:     uuid.New(),
		BranchName: "feature/error-handling",
		ProjectID:  uuid.New(),
	}

	// Enqueue with error handling
	jobID, err := adapter.EnqueueTaskPlanning(payload, 0)
	if err != nil {
		// Handle different types of errors
		switch {
		case err.Error() == "redis connection failed":
			fmt.Println("Redis connection error - check Redis server")
		case err.Error() == "queue full":
			fmt.Println("Queue is full - retry later")
		default:
			fmt.Printf("Unexpected error: %v\n", err)
		}
		return
	}

	fmt.Printf("Job enqueued successfully: %s\n", jobID)
}

// ExampleMonitoring demonstrates how to monitor job status
func ExampleMonitoring() {
	// In a real application, you would use asynq.Inspector
	// to monitor job status and queue metrics

	fmt.Println("Example monitoring commands:")
	fmt.Println("1. Check queue status:")
	fmt.Println("   redis-cli LLEN asynq:queues:planning")

	fmt.Println("2. Check failed jobs:")
	fmt.Println("   redis-cli LLEN asynq:failed")

	fmt.Println("3. Check processing jobs:")
	fmt.Println("   redis-cli LLEN asynq:processing")

	fmt.Println("4. Use asynq.Inspector in Go code:")
	fmt.Println("   inspector := asynq.NewInspector(redisOpt)")
	fmt.Println("   queues, err := inspector.Queues()")
}
