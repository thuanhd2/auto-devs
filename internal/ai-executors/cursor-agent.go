package aiexecutors

import (
    "context"
    "encoding/json"
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
    logs := make([]*entity.ExecutionLog, 0, len(lines))
    for i, line := range lines {
        if strings.TrimSpace(line) == "" {
            continue
        }
        logItem := &entity.ExecutionLog{
            Message: line,
            Level:   entity.LogLevelInfo,
            Source:  "stdout",
            Line:    i,
        }

        var generic map[string]interface{}
        if err := json.Unmarshal([]byte(line), &generic); err == nil {
            if t, ok := generic["type"].(string); ok {
                logItem.LogType = t
            }
            if msg, ok := generic["message"].(map[string]interface{}); ok {
                if content, ok := msg["content"].([]interface{}); ok && len(content) > 0 {
                    logItem.ParsedContent = entity.JSONB{"content": content}
                    for _, c := range content {
                        if m, ok := c.(map[string]interface{}); ok {
                            if ct, _ := m["type"].(string); ct == "tool_use" {
                                if id, _ := m["id"].(string); id != "" {
                                    logItem.ToolUseID = id
                                }
                                if name, _ := m["name"].(string); name != "" {
                                    logItem.ToolName = name
                                }
                            }
                        }
                    }
                }
            }
            if logItem.ParsedContent == nil {
                logItem.ParsedContent = entity.JSONB(generic)
            }
        }
        logs = append(logs, logItem)
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
