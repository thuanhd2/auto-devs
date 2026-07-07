package aiexecutors

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/auto-devs/auto-devs/internal/entity"
)

/*
Deep Seek is essentially claude-code with ENVs forced-set to use the Deep Seek API.
See more at: https://api-docs.deepseek.com/guides/agent_integrations/claude_code

To run Deep Seek CLI, you need to assign the following ENVs to AUTODEVS:
- AUTODEVS_DEEPSEEK_BASE_URL: https://api.deepseek.com/anthropic
- AUTODEVS_DEEPSEEK_AUTH_TOKEN: Deep Seek token
- AUTODEVS_DEEPSEEK_MODEL: deepseek-v4-pro
- AUTODEVS_DEEPSEEK_DEFAULT_OPUS_MODEL: deepseek-v4-pro
- AUTODEVS_DEEPSEEK_DEFAULT_SONNET_MODEL: deepseek-v4-pro
- AUTODEVS_DEEPSEEK_DEFAULT_HAIKU_MODEL: deepseek-v4-flash
*/

type DeepSeekExecutor struct{}

func NewDeepSeekExecutor() *DeepSeekExecutor {
	return &DeepSeekExecutor{}
}

const ENV_PREFIX = "AUTODEVS_"

func getEnv(key, defaultValue string) string {
	keyWithPrefix := ENV_PREFIX + key
	if value := os.Getenv(keyWithPrefix); value != "" {
		return value
	}
	return defaultValue
}

func (e *DeepSeekExecutor) getEnvVars() map[string]string {
	return map[string]string{
		"ANTHROPIC_BASE_URL":             getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/anthropic"),
		"ANTHROPIC_AUTH_TOKEN":           getEnv("DEEPSEEK_AUTH_TOKEN", ""),
		"ANTHROPIC_MODEL":                getEnv("DEEPSEEK_MODEL", "deepseek-v4-pro"),
		"ANTHROPIC_DEFAULT_OPUS_MODEL":   getEnv("DEEPSEEK_DEFAULT_OPUS_MODEL", "deepseek-v4-pro"),
		"ANTHROPIC_DEFAULT_SONNET_MODEL": getEnv("DEEPSEEK_DEFAULT_SONNET_MODEL", "deepseek-v4-pro"),
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":  getEnv("DEEPSEEK_DEFAULT_HAIKU_MODEL", "deepseek-v4-flash"),
	}
}

func (e *DeepSeekExecutor) GetPlanningCommand(ctx context.Context, task *entity.Task) (string, string, map[string]string, error) {
	command := "npx -y @anthropic-ai/claude-code@2.1.119 -p --permission-mode=plan --verbose --output-format=stream-json"
	prompt, err := e.generatePlanningPrompt(*task)
	if err != nil {
		return "", "", nil, err
	}
	return command, prompt, e.getEnvVars(), nil
}

func (e *DeepSeekExecutor) GetImplementationCommand(ctx context.Context, task *entity.Task) (string, string, map[string]string, error) {
	command := "npx -y @anthropic-ai/claude-code@2.1.119 -p --dangerously-skip-permissions --verbose --output-format=stream-json"
	prompt, err := e.getImplementationPrompt(ctx, task)
	if err != nil {
		return "", "", nil, err
	}
	return command, prompt, e.getEnvVars(), nil
}

func (e *DeepSeekExecutor) ParseOutputToLogs(output string) []*entity.ExecutionLog {
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

		// Attempt to parse structured stream-json from Claude Code
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

func (e *DeepSeekExecutor) getImplementationPrompt(_ context.Context, task *entity.Task) (string, error) {
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
func (e *DeepSeekExecutor) generatePlanningPrompt(task entity.Task) (string, error) {
	prompt := fmt.Sprintf(`
	Plan for bellow task, only output the plan, no other text:
	Task: %s
	Task Description: %s
	`, task.Title, task.Description)
	return prompt, nil
}

func (e *DeepSeekExecutor) ParseOutputToPlan(output string) (string, error) {
	lines := strings.Split(output, "\n")
	// find the line that contains "name":"ExitPlanMode" and "type": "tool_use"
	planResultLine := ""
	for _, line := range lines {
		if strings.Contains(line, "name\":\"ExitPlanMode\"") && strings.Contains(line, "type\":\"tool_use\"") {
			planResultLine = line
			break
		}
	}
	/*Line example:

	{"type":"assistant","message":{"id":"msg_01PDQXBzHtQLA3ruAvhawZh2","type":"message","role":"assistant","model":"claude-sonnet-4-20250514","content":[{"type":"tool_use","id":"toolu_01NuDYFV4iF6kp3bdo48cXgE","name":"ExitPlanMode","input":{"plan":"# Plan: Add Code Changes Tab to Task Details Modal\n\nBased on my research, I have a clear understanding of how to implement the Code Changes tab. Here's my implementation plan:\n\n## Current State Analysis\n- Task details modal currently has 3 tabs: Plan Review, Executions, and Metadata\n- Task model already has pr_url field and TaskGitInfo with pull request information\n- Existing usePullRequestByTask hook can fetch PR data by task ID\n- Rich PRDetail component already exists for displaying pull request information\n\n## Implementation Plan\n\n### 1. Update Task Detail Sheet Component (frontend/src/components/kanban/task-detail-sheet.tsx)\n- Add \"Code Changes\" as the 4th tab (after Plan Review)\n- Update TabsList grid from grid-cols-3 to grid-cols-4\n- Add new TabsTrigger for \"code-changes\" \n- Add new TabsContent with CodeChanges component\n\n### 2. Create CodeChanges Component\n- Use usePullRequestByTask hook to fetch PR data for the task\n- Display different states:\n - No PR yet: Show message \"No pull request created yet\"\n - PR exists: Show PR link button with external link icon\n - Loading: Show skeleton loader\n- Keep it simple and focused per user requirements - just show the PR link\n\n### 3. Implementation Details\n- Add the tab after \"Plan Review\" but before \"Executions\" and \"Metadata\"\n- Use existing UI components (Button, ExternalLink icon, etc.)\n- Handle loading and error states gracefully\n- Make PR link open in new tab when clicked\n\n## Files to Modify\n1. frontend/src/components/kanban/task-detail-sheet.tsx - Add new tab and component\n2. No new files needed - will create inline component for simplicity\n\nThis implementation will be clean, simple, and follows the existing patterns in the codebase."}}],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":0,"cache_creation_input_tokens":397,"cache_read_input_tokens":62637,"output_tokens":499,"service_tier":"standard"}},"parent_tool_use_id":null,"session_id":"9d3ac8dd-5572-4bdc-ae86-ff1071e369e7"}

	*/

	var planOutput PlanOutput
	err := json.Unmarshal([]byte(planResultLine), &planOutput)
	if err != nil {
		return "", err
	}
	planContent := planOutput.Message.Content[0].Input.Plan
	return planContent, nil
}
