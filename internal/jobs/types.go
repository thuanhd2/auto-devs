package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Job type constants
const (
	TypeTaskPlanning       = "task:planning"
	TypeTaskImplementation = "task:implementation"
	TypePRStatusSync       = "pr:status_sync"
	TypeWorktreeCleanup    = "worktree:cleanup"
)

// TaskPlanningPayload represents the payload for task planning jobs
type TaskPlanningPayload struct {
	TaskID     uuid.UUID `json:"task_id"`
	BranchName string    `json:"branch_name"`
	ProjectID  uuid.UUID `json:"project_id"`
	AIType     string    `json:"ai_type"`
}

// TaskImplementationPayload represents the payload for task implementation jobs
type TaskImplementationPayload struct {
	TaskID    uuid.UUID `json:"task_id"`
	ProjectID uuid.UUID `json:"project_id"`
	AIType    string    `json:"ai_type"`
}

// PRStatusSyncPayload represents the payload for PR status sync jobs
type PRStatusSyncPayload struct {
	// Empty payload since this job checks all open PRs
}

// WorktreeCleanupPayload represents the payload for worktree cleanup jobs
type WorktreeCleanupPayload struct {
	// Empty payload since this job processes all eligible tasks
}

// NewTaskPlanningJob creates a new task planning job
func NewTaskPlanningJob(taskID uuid.UUID, branchName string, projectID uuid.UUID, aiType string) (*asynq.Task, error) {
	payload := TaskPlanningPayload{
		TaskID:     taskID,
		BranchName: branchName,
		ProjectID:  projectID,
		AIType:     aiType,
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
func NewTaskImplementationJob(taskID uuid.UUID, projectID uuid.UUID, aiType string) (*asynq.Task, error) {
	payload := TaskImplementationPayload{
		TaskID:    taskID,
		ProjectID: projectID,
		AIType:    aiType,
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

// NewWorktreeCleanupJob creates a new worktree cleanup job
func NewWorktreeCleanupJob() (*asynq.Task, error) {
	payload := WorktreeCleanupPayload{}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal worktree cleanup payload: %w", err)
	}

	return asynq.NewTask(TypeWorktreeCleanup, data), nil
}

// ParseWorktreeCleanupPayload parses the worktree cleanup payload from asynq task
func ParseWorktreeCleanupPayload(task *asynq.Task) (*WorktreeCleanupPayload, error) {
	var payload WorktreeCleanupPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal worktree cleanup payload: %w", err)
	}
	return &payload, nil
}
