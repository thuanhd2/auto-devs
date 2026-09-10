package aiexecutors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/auto-devs/auto-devs/internal/entity"
)

type ReasonixExecutor struct{}

func NewReasonixExecutor() *ReasonixExecutor {
	return &ReasonixExecutor{}
}

// GetPlanningCommand returns the Reasonix CLI command for the planning phase.
// The prompt is piped to the CLI via stdin, and the final plan comes back in
// the stream-json result object ({"type":"result","result":"..."}).
func (e *ReasonixExecutor) GetPlanningCommand(ctx context.Context, task *entity.Task) (string, string, map[string]string, error) {
	command := "reasonix run -p --permission-mode=auto --output-format=stream-json"
	prompt, err := e.generatePlanningPrompt(*task)
	if err != nil {
		return "", "", nil, err
	}
	return command, prompt, nil, nil
}

// GetImplementationCommand returns the Reasonix CLI command for the
// implementation phase. bypassPermissions is the Reasonix equivalent of
// Claude Code's --dangerously-skip-permissions for unattended runs.
func (e *ReasonixExecutor) GetImplementationCommand(ctx context.Context, task *entity.Task) (string, string, map[string]string, error) {
	command := "reasonix run -p --permission-mode=bypassPermissions --output-format=stream-json"
	prompt, err := e.getImplementationPrompt(ctx, task)
	if err != nil {
		return "", "", nil, err
	}
	return command, prompt, nil, nil
}

func (e *ReasonixExecutor) ParseOutputToLogs(output string) []*entity.ExecutionLog {
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

		// Attempt to parse structured stream-json from Reasonix
		var generic map[string]interface{}
		if err := json.Unmarshal([]byte(line), &generic); err == nil {
			// Extract type and message fields if present
			if t, ok := generic["type"].(string); ok {
				logItem.LogType = t
			}
			if msg, ok := generic["message"].(map[string]interface{}); ok {
				// Look for tool use content
				if content, ok := msg["content"].([]interface{}); ok && len(content) > 0 {
					// We only keep structured content as parsed_content
					logItem.ParsedContent = entity.JSONB{"content": content}
					// try to find tool_use info
					for _, c := range content {
						if m, ok := c.(map[string]interface{}); ok {
							typeVal, _ := m["type"].(string)
							if typeVal == "tool_use" {
								if id, _ := m["id"].(string); id != "" {
									logItem.ToolUseID = id
								}
								if name, _ := m["name"].(string); name != "" {
									logItem.ToolName = name
								}
							} else if typeVal == "tool_result" {
								t := false
								logItem.IsError = &t
							}
						}
					}
				}
			}

			// Also propagate the entire parsed JSON as parsed_content if nothing else
			if logItem.ParsedContent == nil {
				logItem.ParsedContent = entity.JSONB(generic)
			}
		}

		logs = append(logs, logItem)
	}
	return logs
}

func (e *ReasonixExecutor) getImplementationPrompt(_ context.Context, task *entity.Task) (string, error) {
	var prompt string
	if len(task.Plans) > 0 {
		prompt = fmt.Sprintf(`
		Task: %s
		Task Description: %s
		Plan: %s
		`, task.Title, task.Description, task.Plans[0].Content)
	} else {
		prompt = fmt.Sprintf(`
		Task: %s
		Task Description: %s
		`, task.Title, task.Description)
	}
	return prompt, nil
}

// generatePlanningPrompt creates a structured prompt for AI planning phase
func (e *ReasonixExecutor) generatePlanningPrompt(task entity.Task) (string, error) {
	prompt := fmt.Sprintf(`
	Plan for bellow task, only output the plan, no other text:
	Task: %s
	Task Description: %s
	`, task.Title, task.Description)
	return prompt, nil
}

// ParseOutputToPlan extracts the plan from the Reasonix stream-json output.
// Unlike Claude Code, Reasonix has no ExitPlanMode tool_use; the final answer
// is carried in the last result object: {"type":"result","result":"..."}.
func (e *ReasonixExecutor) ParseOutputToPlan(output string) (string, error) {
	lines := strings.Split(output, "\n")
	// find the line that contains "type":"result" and a "result" field
	planResultLine := ""
	for _, line := range lines {
		if strings.Contains(line, "\"type\":\"result\"") && strings.Contains(line, "\"result\":") {
			planResultLine = line
			break
		}
	}

	if planResultLine == "" {
		return "", fmt.Errorf("no plan result found in output")
	}

	var result struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal([]byte(planResultLine), &result); err != nil {
		return "", err
	}
	if strings.TrimSpace(result.Result) == "" {
		return "", fmt.Errorf("plan result is empty")
	}
	return result.Result, nil
}
