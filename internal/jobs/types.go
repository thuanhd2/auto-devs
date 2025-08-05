package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Job type constants
const (
	TypeTaskPlanning = "task:planning"
)

// TaskPlanningPayload represents the payload for task planning jobs
type TaskPlanningPayload struct {
	TaskID     uuid.UUID `json:"task_id"`
	BranchName string    `json:"branch_name"`
	ProjectID  uuid.UUID `json:"project_id"`
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