package aiexecutors

import (
	"context"
	"fmt"
	"strings"

	"github.com/auto-devs/auto-devs/internal/entity"
)

type CursorAgentExecutor struct{}

func NewCursorAgentExecutor() *CursorAgentExecutor {
	return &CursorAgentExecutor{}
}

func (e *CursorAgentExecutor) GetPlanningCommand(ctx context.Context, task *entity.Task) (string, string, error) {
	return "", "", fmt.Errorf(NOT_SUPPORT_PLANNING)
}

func (e *CursorAgentExecutor) GetImplementationCommand(ctx context.Context, task *entity.Task) (string, string, error) {
	command := "cursor-agent -p --output-format=stream-json --force"
	prompt, err := e.getImplementationPrompt(ctx, task)
	if err != nil {
		return "", "", err
	}
	return command, prompt, nil
}

func (e *CursorAgentExecutor) ParseOutputToLogs(output string) []*entity.ExecutionLog {
	lines := strings.Split(output, "\n")
	logs := make([]*entity.ExecutionLog, len(lines))
	for i, line := range lines {
		logs[i] = &entity.ExecutionLog{
			Message: line,
			Level:   entity.LogLevelInfo,
			Source:  "stdout",
			Line:    i,
		}
	}
	return logs
}

func (e *CursorAgentExecutor) getImplementationPrompt(_ context.Context, task *entity.Task) (string, error) {
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

func (e *CursorAgentExecutor) ParseOutputToPlan(output string) (string, error) {
	return "", fmt.Errorf(NOT_SUPPORT_PLANNING)
}
