package jobs

import (
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// Client wraps asynq.Client for job enqueueing
type Client struct {
	client *asynq.Client
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)

// NewClient creates a new job client
func NewClient(redisAddr, redisPassword string, redisDB int) *Client {
	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	}

	return &Client{
		client: asynq.NewClient(redisOpt),
	}
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.client.Close()
}

// EnqueueTaskPlanning enqueues a task planning job
func (c *Client) EnqueueTaskPlanning(payload *TaskPlanningPayload, delay time.Duration) (*asynq.TaskInfo, error) {
	task, err := NewTaskPlanningJob(payload.TaskID, payload.BranchName, payload.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create task planning job: %w", err)
	}

	// Set task options
	opts := []asynq.Option{
		asynq.MaxRetry(1),
		asynq.Timeout(30 * time.Minute), // Planning can take a while
		asynq.Queue("planning"),         // Use dedicated queue for planning jobs
	}

	if delay > 0 {
		opts = append(opts, asynq.ProcessIn(delay))
	}

	taskInfo, err := c.client.Enqueue(task, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task planning job: %w", err)
	}

	return taskInfo, nil
}

// EnqueueTaskPlanningString enqueues a task planning job and returns job ID as string
func (c *Client) EnqueueTaskPlanningString(payload *TaskPlanningPayload, delay time.Duration) (string, error) {
	taskInfo, err := c.EnqueueTaskPlanning(payload, delay)
	if err != nil {
		return "", err
	}
	return taskInfo.ID, nil
}

// EnqueueTaskImplementation enqueues a task implementation job
func (c *Client) EnqueueTaskImplementation(payload *TaskImplementationPayload, delay time.Duration) (*asynq.TaskInfo, error) {
	task, err := NewTaskImplementationJob(payload.TaskID, payload.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create task implementation job: %w", err)
	}

	// Set task options
	opts := []asynq.Option{
		asynq.MaxRetry(1),
		asynq.Timeout(60 * time.Minute), // Implementation can take longer than planning
		asynq.Queue("implementation"),   // Use dedicated queue for implementation jobs
	}

	if delay > 0 {
		opts = append(opts, asynq.ProcessIn(delay))
	}

	taskInfo, err := c.client.Enqueue(task, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task implementation job: %w", err)
	}

	return taskInfo, nil
}

// EnqueueTaskImplementationString enqueues a task implementation job and returns job ID as string
func (c *Client) EnqueueTaskImplementationString(payload *TaskImplementationPayload, delay time.Duration) (string, error) {
	taskInfo, err := c.EnqueueTaskImplementation(payload, delay)
	if err != nil {
		return "", err
	}
	return taskInfo.ID, nil
}

// GetTaskInfo retrieves information about a task
func (c *Client) GetTaskInfo(queue, taskID string) (*asynq.TaskInfo, error) {
	// Note: asynq.Client doesn't have GetTaskInfo method
	// This would typically be handled by asynq.Inspector
	return nil, fmt.Errorf("task info retrieval not implemented")
}
