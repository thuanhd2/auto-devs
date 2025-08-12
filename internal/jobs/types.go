package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Job type constants
const (
	TypeTaskPlanning      = "task:planning"
	TypeTaskImplementation = "task:implementation"
	TypePRStatusSync      = "pr:status_sync"
)

// TaskPlanningPayload represents the payload for task planning jobs
type TaskPlanningPayload struct {
	TaskID     uuid.UUID `json:"task_id"`
	BranchName string    `json:"branch_name"`
	ProjectID  uuid.UUID `json:"project_id"`
}

// TaskImplementationPayload represents the payload for task implementation jobs
type TaskImplementationPayload struct {
	TaskID    uuid.UUID `json:"task_id"`
	ProjectID uuid.UUID `json:"project_id"`
}

// PRStatusSyncPayload represents the payload for PR status sync jobs
type PRStatusSyncPayload struct {
	// Empty payload since this job checks all open PRs
}

// NewTaskPlanningJob creates a new task planning job
func NewTaskPlanningJob(taskID uuid.UUID, branchName string, projectID uuid.UUID) (*asynq.Task, error) {
	payload := TaskPlanningPayload{
		TaskID:     taskID,
		BranchName: branchName,
		ProjectID:  projectID,
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task planning payload: %w", err)
	}
	
	return asynq.NewTask(TypeTaskPlanning, data), nil
}

// ParseTaskPlanningPayload parses the task planning payload from asynq task
func ParseTaskPlanningPayload(task *asynq.Task) (*TaskPlanningPayload, error) {
	var payload TaskPlanningPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task planning payload: %w", err)
	}
	return &payload, nil
}

// NewTaskImplementationJob creates a new task implementation job
func NewTaskImplementationJob(taskID uuid.UUID, projectID uuid.UUID) (*asynq.Task, error) {
	payload := TaskImplementationPayload{
		TaskID:    taskID,
		ProjectID: projectID,
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task implementation payload: %w", err)
	}
	
	return asynq.NewTask(TypeTaskImplementation, data), nil
}

// ParseTaskImplementationPayload parses the task implementation payload from asynq task
func ParseTaskImplementationPayload(task *asynq.Task) (*TaskImplementationPayload, error) {
	var payload TaskImplementationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task implementation payload: %w", err)
	}
	return &payload, nil
}

// NewPRStatusSyncJob creates a new PR status sync job
func NewPRStatusSyncJob() (*asynq.Task, error) {
	payload := PRStatusSyncPayload{}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PR status sync payload: %w", err)
	}
	
	return asynq.NewTask(TypePRStatusSync, data), nil
}

// ParsePRStatusSyncPayload parses the PR status sync payload from asynq task
func ParsePRStatusSyncPayload(task *asynq.Task) (*PRStatusSyncPayload, error) {
	var payload PRStatusSyncPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PR status sync payload: %w", err)
	}
	return &payload, nil
}
