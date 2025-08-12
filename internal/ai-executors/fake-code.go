package aiexecutors

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/auto-devs/auto-devs/internal/entity"
)

type FakeCodeExecutor struct{}

func NewFakeCodeExecutor() *FakeCodeExecutor {
	return &FakeCodeExecutor{}
}

func (e *FakeCodeExecutor) GetPlanningCommand(ctx context.Context, task *entity.Task) (string, string, error) {
	projectPath, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	projectRootPath := projectPath
	fakeCliPath := filepath.Join(projectRootPath, "fake-cli", "fake-planning-claude.js")
	command := fmt.Sprintf("node %s", fakeCliPath)
	prompt, err := e.generatePlanningPrompt(*task)
	if err != nil {
		return "", "", err
	}
	return command, prompt, nil
}

func (e *FakeCodeExecutor) GetImplementationCommand(ctx context.Context, task *entity.Task) (string, string, error) {
	projectPath, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	projectRootPath := projectPath
	fakeCliPath := filepath.Join(projectRootPath, "fake-cli", "fake-claude.js")
	command := fmt.Sprintf("node %s", fakeCliPath)
	prompt, err := e.getImplementationPrompt(ctx, task)
	if err != nil {
		return "", "", err
	}
	return command, prompt, nil
}

func (e *FakeCodeExecutor) ParseOutputToLogs(output string) []*entity.ExecutionLog {
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

func (e *FakeCodeExecutor) getImplementationPrompt(_ context.Context, task *entity.Task) (string, error) {
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

// generatePlanningPrompt creates a structured prompt for AI planning phase
func (e *FakeCodeExecutor) generatePlanningPrompt(task entity.Task) (string, error) {
	var promptBuilder strings.Builder

	promptBuilder.WriteString("# Task Implementation Planning\n\n")
	promptBuilder.WriteString("You are an expert software developer tasked with creating a detailed implementation plan.\n\n")

	// Task Information
	promptBuilder.WriteString("## Task Details\n")
	promptBuilder.WriteString(fmt.Sprintf("**Title:** %s\n", task.Title))
	promptBuilder.WriteString(fmt.Sprintf("**Description:** %s\n", task.Description))
	promptBuilder.WriteString(fmt.Sprintf("**Priority:** %s\n", task.Priority))

	if task.EstimatedHours != nil {
		promptBuilder.WriteString(fmt.Sprintf("**Estimated Hours:** %.2f\n", *task.EstimatedHours))
	}

	if len(task.Tags) > 0 {
		promptBuilder.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(task.Tags, ", ")))
	}

	promptBuilder.WriteString("\n## Requirements\n")
	promptBuilder.WriteString("Please create a comprehensive implementation plan that includes:\n\n")
	promptBuilder.WriteString("1. **Analysis Phase**\n")
	promptBuilder.WriteString("   - Understanding the requirements\n")
	promptBuilder.WriteString("   - Identifying key components and dependencies\n")
	promptBuilder.WriteString("   - Risk assessment\n\n")

	promptBuilder.WriteString("2. **Design Phase**\n")
	promptBuilder.WriteString("   - Architecture decisions\n")
	promptBuilder.WriteString("   - Interface definitions\n")
	promptBuilder.WriteString("   - Database schema changes (if applicable)\n\n")

	promptBuilder.WriteString("3. **Implementation Phase**\n")
	promptBuilder.WriteString("   - Step-by-step implementation tasks\n")
	promptBuilder.WriteString("   - File modifications and creations\n")
	promptBuilder.WriteString("   - Code structure and patterns\n\n")

	promptBuilder.WriteString("4. **Testing Phase**\n")
	promptBuilder.WriteString("   - Unit test requirements\n")
	promptBuilder.WriteString("   - Integration test scenarios\n")
	promptBuilder.WriteString("   - Manual testing steps\n\n")

	promptBuilder.WriteString("5. **Validation Phase**\n")
	promptBuilder.WriteString("   - Acceptance criteria verification\n")
	promptBuilder.WriteString("   - Code review checklist\n")
	promptBuilder.WriteString("   - Documentation updates\n\n")

	promptBuilder.WriteString("## Output Format\n")
	promptBuilder.WriteString("Please provide the plan as structured markdown with clear sections and actionable steps.\n")
	promptBuilder.WriteString("Each step should be specific, measurable, and include estimated time if possible.\n")
	promptBuilder.WriteString("Include any assumptions, dependencies, or potential risks.\n\n")

	promptBuilder.WriteString("## Context\n")
	promptBuilder.WriteString("This is a Go-based web application with Clean Architecture pattern.\n")
	promptBuilder.WriteString("The codebase uses Gin framework, GORM for database, and follows standard Go practices.\n")

	return promptBuilder.String(), nil
}

func (e *FakeCodeExecutor) ParseOutputToPlan(output string) (string, error) {
	lines := strings.Split(output, "\n")
	// find the line that contains "name":"ExitPlanMode"
	planResultLine := ""
	for _, line := range lines {
		if strings.Contains(line, "name\":\"ExitPlanMode\"") {
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

type PlanOutput struct {
	Type            string      `json:"type"`
	Message         PlanMessage `json:"message"`
	ParentToolUseID string      `json:"parent_tool_use_id"`
	SessionID       string      `json:"session_id"`
}

type PlanMessage struct {
	ID      string        `json:"id"`
	Type    string        `json:"type"`
	Role    string        `json:"role"`
	Model   string        `json:"model"`
	Content []PlanContent `json:"content"`
}

type PlanContent struct {
	Type  string           `json:"type"`
	ID    string           `json:"id"`
	Role  string           `json:"role"`
	Model string           `json:"model"`
	Input PlanContentInput `json:"input"`
}

type PlanContentInput struct {
	Plan string `json:"plan"`
}
