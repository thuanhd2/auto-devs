package main

import (
	"fmt"
	"log"
	"time"

	"github.com/auto-devs/auto-devs/internal/jobs"
	"github.com/google/uuid"
)

// ExampleTestWorker demonstrates how to test the job worker
func main() {
	fmt.Println("=== Job Worker Test Example ===")

	// 1. Create job client
	client := jobs.NewClient("localhost:6379", "", 0)
	defer client.Close()

	fmt.Println("✓ Job client created")

	// 2. Create test payload
	taskID := uuid.New()
	projectID := uuid.New()
	branchName := "feature/test-worker"

	payload := &jobs.TaskPlanningPayload{
		TaskID:     taskID,
		BranchName: branchName,
		ProjectID:  projectID,
	}

	fmt.Printf("✓ Test payload created - TaskID: %s, ProjectID: %s, Branch: %s\n",
		taskID.String()[:8], projectID.String()[:8], branchName)

	// 3. Enqueue test job
	fmt.Println("Enqueueing test job...")
	taskInfo, err := client.EnqueueTaskPlanning(payload, 0)
	if err != nil {
		log.Fatalf("Failed to enqueue job: %v", err)
	}

	fmt.Printf("✓ Job enqueued successfully - JobID: %s\n", taskInfo.ID)

	// 4. Wait a bit to see if job gets processed
	fmt.Println("Waiting for job processing...")
	time.Sleep(5 * time.Second)

	// 5. Check job status (this would require asynq.Inspector in real app)
	fmt.Println("Job processing test completed!")
	fmt.Println("")
	fmt.Println("To see the job being processed:")
	fmt.Println("1. Start the worker: ./scripts/run-worker.sh")
	fmt.Println("2. Watch the logs for job processing")
	fmt.Println("3. Check Redis queue: redis-cli LLEN asynq:queues:planning")
}

// ExampleEnqueueMultipleJobs demonstrates enqueueing multiple jobs
func ExampleEnqueueMultipleJobs() {
	client := jobs.NewClient("localhost:6379", "", 0)
	defer client.Close()

	// Enqueue multiple jobs
	for i := 1; i <= 5; i++ {
		payload := &jobs.TaskPlanningPayload{
			TaskID:     uuid.New(),
			BranchName: fmt.Sprintf("feature/test-%d", i),
			ProjectID:  uuid.New(),
		}

		taskInfo, err := client.EnqueueTaskPlanning(payload, 0)
		if err != nil {
			log.Printf("Failed to enqueue job %d: %v", i, err)
			continue
		}

		fmt.Printf("Job %d enqueued: %s\n", i, taskInfo.ID)
	}
}

// ExampleEnqueueWithDelay demonstrates enqueueing with delay
func ExampleEnqueueWithDelay() {
	client := jobs.NewClient("localhost:6379", "", 0)
	defer client.Close()

	payload := &jobs.TaskPlanningPayload{
		TaskID:     uuid.New(),
		BranchName: "feature/delayed-test",
		ProjectID:  uuid.New(),
	}

	// Enqueue with 10 second delay
	delay := 10 * time.Second
	taskInfo, err := client.EnqueueTaskPlanning(payload, delay)
	if err != nil {
		log.Fatalf("Failed to enqueue delayed job: %v", err)
	}

	fmt.Printf("Delayed job enqueued: %s (will be processed in %v)\n", taskInfo.ID, delay)
}
