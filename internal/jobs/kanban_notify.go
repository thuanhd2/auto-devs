package jobs

import (
	"context"
	"fmt"
	"strings"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/hibiken/asynq"
)

// ProcessKanbanNotify posts a machine-readable comment on the linked Hermes
// kanban card and unblocks it so the Hermes dispatcher picks it up. Any error
// is returned so asynq retries the job (exponential backoff, max retry set at
// enqueue time). Duplicate comments from retries are acceptable — the Hermes
// worker reads the latest comment.
func (p *Processor) ProcessKanbanNotify(ctx context.Context, task *asynq.Task) error {
	payload, err := ParseKanbanNotifyPayload(task)
	if err != nil {
		return fmt.Errorf("failed to parse kanban notify payload: %w", err)
	}

	p.logger.Info("Processing kanban notify job",
		"task_id", payload.TaskID,
		"kanban_task_id", payload.KanbanTaskID,
		"new_status", payload.NewStatus,
	)

	if p.kanbanClient == nil || !p.kanbanClient.Enabled() {
		p.logger.Warn("Kanban client not enabled, skipping notify job", "task_id", payload.TaskID)
		return nil
	}

	// Load the task for the freshest title/PR/error info
	taskEntity, err := p.taskUsecase.GetByID(ctx, payload.TaskID)
	if err != nil {
		return fmt.Errorf("failed to load task %s: %w", payload.TaskID, err)
	}

	// Asynq retries can deliver callbacks out of order (e.g. a retried
	// PLAN_REVIEWING landing after CODE_REVIEWING already fired). A stale
	// callback would unblock the card and spawn a worker acting on old
	// state — skip it; the callback for the current status covers the card.
	if taskEntity.Status != payload.NewStatus {
		p.logger.Warn("Skipping stale kanban notify",
			"task_id", payload.TaskID,
			"payload_status", payload.NewStatus,
			"current_status", taskEntity.Status,
		)
		return nil
	}

	comment := buildKanbanComment(taskEntity, payload)

	if err := p.kanbanClient.CommentTask(ctx, payload.KanbanTaskID, comment); err != nil {
		return fmt.Errorf("failed to comment kanban task %s: %w", payload.KanbanTaskID, err)
	}

	if err := p.kanbanClient.UnblockTask(ctx, payload.KanbanTaskID); err != nil {
		return fmt.Errorf("failed to unblock kanban task %s: %w", payload.KanbanTaskID, err)
	}

	p.logger.Info("Kanban notify completed",
		"task_id", payload.TaskID,
		"kanban_task_id", payload.KanbanTaskID,
	)
	return nil
}

// buildKanbanComment renders the machine-readable comment the Hermes worker
// parses. Keep the format stable — it is part of the Auto-Devs ↔ Hermes
// contract.
func buildKanbanComment(task *entity.Task, payload *KanbanNotifyPayload) string {
	pr := "none"
	if task.PullRequest != nil && *task.PullRequest != "" {
		pr = *task.PullRequest
	}

	errorInfo := "none"
	if payload.NewStatus == entity.TaskStatusCANCELLED && len(task.ErrorLogEntries) > 0 {
		errorInfo = task.ErrorLogEntries[len(task.ErrorLogEntries)-1]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "[auto-devs] status=%s\n", payload.NewStatus)
	fmt.Fprintf(&b, "task: %s — %s\n", task.ID, task.Title)
	fmt.Fprintf(&b, "old_status: %s\n", payload.OldStatus)
	fmt.Fprintf(&b, "plans: GET /api/v1/tasks/%s/plans\n", task.ID)
	fmt.Fprintf(&b, "pr: %s\n", pr)
	fmt.Fprintf(&b, "error: %s", errorInfo)
	return b.String()
}
