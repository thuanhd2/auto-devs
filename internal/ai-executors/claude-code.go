package aiexecutors

import (
	"context"
	"fmt"

	"github.com/auto-devs/auto-devs/internal/entity"
)

type ClaudeCodeExecutor struct{}

func NewClaudeCodeExecutor() *ClaudeCodeExecutor {
	return &ClaudeCodeExecutor{}
}

func (e *ClaudeCodeExecutor) GetPlanningCommand(ctx context.Context, task *entity.Task) (string, string, error) {
	return "", "", nil
}

func (e *ClaudeCodeExecutor) GetImplementationCommand(ctx context.Context, task *entity.Task) (string, string, error) {
	return "npx -y @anthropic-ai/claude-code@latest -p --dangerously-skip-permissions --verbose --output-format=stream-json", "", nil
}

func (e *ClaudeCodeExecutor) GetImplementationPrompt(ctx context.Context, task *entity.Task) (string, error) {
	if len(task.Plans) == 0 {
		return "", fmt.Errorf("no plan found for task")
	}

	prompt := fmt.Sprintf(`
	Task: %s
	Task Description: %s
	Plan: %s
	`, task.Title, task.Description, task.Plans[0].Content)
	return prompt, nil
}
