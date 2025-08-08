package aiexecutors

import (
	"context"
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
	fakeCliPath := filepath.Join(projectRootPath, "fake-cli", "fake-planning.sh")
	prompt, err := e.generatePlanningPrompt(*task)
	if err != nil {
		return "", "", err
	}
	return fakeCliPath, prompt, nil
}

func (e *FakeCodeExecutor) GetImplementationCommand(ctx context.Context, task *entity.Task) (string, string, error) {
	projectPath, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	projectRootPath := projectPath
	fakeCliPath := filepath.Join(projectRootPath, "fake-cli", "fake.sh")
	prompt, err := e.getImplementationPrompt(ctx, task)
	if err != nil {
		return "", "", err
	}
	return fakeCliPath, prompt, nil
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
