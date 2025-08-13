package jobs

import (
	"time"

	"github.com/auto-devs/auto-devs/internal/usecase"
)

// ClientInterface defines the interface for job client operations
type ClientInterface interface {
	EnqueueTaskPlanningString(payload *TaskPlanningPayload, delay time.Duration) (string, error)
	EnqueueTaskImplementationString(payload *TaskImplementationPayload, delay time.Duration) (string, error)
	Close() error
}

// JobClientAdapter adapts the actual job client to the usecase interface
type JobClientAdapter struct {
	client ClientInterface
}

// NewJobClientAdapter creates a new job client adapter
func NewJobClientAdapter(client ClientInterface) usecase.JobClientInterface {
	return &JobClientAdapter{
		client: client,
	}
}

// EnqueueTaskPlanning enqueues a task planning job
func (a *JobClientAdapter) EnqueueTaskPlanning(payload *usecase.TaskPlanningPayload, delay time.Duration) (string, error) {
	// Convert usecase payload to jobs package payload
	jobPayload := &TaskPlanningPayload{
		TaskID:     payload.TaskID,
		BranchName: payload.BranchName,
		ProjectID:  payload.ProjectID,
		AIType:     payload.AIType,
	}

	// Enqueue the job
	jobID, err := a.client.EnqueueTaskPlanningString(jobPayload, delay)
	if err != nil {
		return "", err
	}

	return jobID, nil
}

// EnqueueTaskImplementation enqueues a task implementation job
func (a *JobClientAdapter) EnqueueTaskImplementation(payload *usecase.TaskImplementationPayload, delay time.Duration) (string, error) {
	// Convert usecase payload to jobs package payload
	jobPayload := &TaskImplementationPayload{
		TaskID:    payload.TaskID,
		ProjectID: payload.ProjectID,
		AIType:    payload.AIType,
	}

	// Enqueue the job
	jobID, err := a.client.EnqueueTaskImplementationString(jobPayload, delay)
	if err != nil {
		return "", err
	}

	return jobID, nil
}
