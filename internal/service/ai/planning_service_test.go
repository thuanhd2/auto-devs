package ai

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/auto-devs/auto-devs/internal/entity"
)

func TestPlanningService_GeneratePlan(t *testing.T) {
	// Setup
	cliManager, err := NewCLIManager(&CLIConfig{
		CLICommand:       "claude-code",
		Timeout:          300 * time.Second,
		WorkingDirectory: "",
		EnableLogging:    true,
		RetryAttempts:    3,
		RetryDelay:       5 * time.Second,
	})
	require.NoError(t, err)
	processManager := NewProcessManager()
	executionService := NewExecutionService(cliManager, processManager)
	planningService := NewPlanningService(executionService, cliManager)

	// Create test task
	task := entity.Task{
		ID:          uuid.New(),
		Title:       "Implement user authentication",
		Description: "Add JWT-based authentication system with login, register, and password reset functionality",
		Status:      entity.TaskStatusTODO,
		Priority:    entity.TaskPriorityHigh,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test plan generation
	plan, err := planningService.GeneratePlan(task)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.NotEmpty(t, plan.ID)
	assert.Equal(t, task.ID.String(), plan.TaskID)
	assert.Contains(t, plan.Description, task.Title)
	assert.NotEmpty(t, plan.Steps)
	assert.Equal(t, 5, len(plan.Steps)) // Should have 5 phases
	assert.NotEmpty(t, plan.Context)

	// Verify context contains expected data
	assert.Equal(t, task.Title, plan.Context["task_title"])
	assert.Equal(t, task.Description, plan.Context["task_description"])
	assert.Equal(t, string(task.Priority), plan.Context["task_priority"])
	assert.Equal(t, string(task.Status), plan.Context["task_status"])
	assert.NotEmpty(t, plan.Context["prompt"])

	// Verify steps are properly ordered
	for i, step := range plan.Steps {
		assert.Equal(t, i+1, step.Order)
		assert.NotEmpty(t, step.ID)
		assert.NotEmpty(t, step.Description)
		assert.NotEmpty(t, step.Action)
	}

	// Verify specific step actions
	expectedActions := []string{"analysis", "design", "implement", "test", "validate"}
	for i, step := range plan.Steps {
		assert.Equal(t, expectedActions[i], step.Action)
	}
}

func TestPlanningService_GeneratePlanningPrompt(t *testing.T) {
	// Setup
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)
	executionService := NewExecutionService(cliManager, NewProcessManager())
	planningService := NewPlanningService(executionService, cliManager)

	// Create test task with all fields
	estimatedHours := 16.5
	task := entity.Task{
		ID:             uuid.New(),
		Title:          "Create API endpoints",
		Description:    "Implement REST API endpoints for user management",
		Status:         entity.TaskStatusTODO,
		Priority:       entity.TaskPriorityMedium,
		EstimatedHours: &estimatedHours,
		Tags:           []string{"backend", "api", "user-management"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Test prompt generation
	prompt, err := planningService.generatePlanningPrompt(task)

	// Assertions
	require.NoError(t, err)
	assert.NotEmpty(t, prompt)

	// Verify prompt contains task information
	assert.Contains(t, prompt, task.Title)
	assert.Contains(t, prompt, task.Description)
	assert.Contains(t, prompt, string(task.Priority))
	assert.Contains(t, prompt, "16.5")
	assert.Contains(t, prompt, "backend, api, user-management")

	// Verify prompt structure
	assert.Contains(t, prompt, "# Task Implementation Planning")
	assert.Contains(t, prompt, "## Task Details")
	assert.Contains(t, prompt, "## Requirements")
	assert.Contains(t, prompt, "## Output Format")
	assert.Contains(t, prompt, "## Context")

	// Verify all phases are mentioned
	assert.Contains(t, prompt, "Analysis Phase")
	assert.Contains(t, prompt, "Design Phase")
	assert.Contains(t, prompt, "Implementation Phase")
	assert.Contains(t, prompt, "Testing Phase")
	assert.Contains(t, prompt, "Validation Phase")
}

func TestPlanningService_GetPlanAsMarkdown(t *testing.T) {
	// Setup
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)
	executionService := NewExecutionService(cliManager, NewProcessManager())
	planningService := NewPlanningService(executionService, cliManager)

	// Create test plan
	plan := &Plan{
		ID:          uuid.New().String(),
		TaskID:      uuid.New().String(),
		Description: "Test implementation plan",
		Steps: []PlanStep{
			{
				ID:          uuid.New().String(),
				Description: "Analyze requirements",
				Action:      "analysis",
				Parameters:  map[string]string{"type": "requirements"},
				Order:       1,
			},
			{
				ID:          uuid.New().String(),
				Description: "Design solution",
				Action:      "design",
				Parameters:  map[string]string{"type": "technical"},
				Order:       2,
			},
		},
		Context: map[string]string{
			"task_title":    "Test Task",
			"task_priority": "HIGH",
			"task_status":   "TODO",
		},
		CreatedAt: time.Now(),
	}

	// Test markdown generation
	markdown := planningService.GetPlanAsMarkdown(plan)

	// Assertions
	assert.NotEmpty(t, markdown)

	// Verify markdown structure
	assert.Contains(t, markdown, "# Implementation Plan: Test Task")
	assert.Contains(t, markdown, "## Overview")
	assert.Contains(t, markdown, "## Task Details")
	assert.Contains(t, markdown, "## Implementation Steps")
	assert.Contains(t, markdown, "## Context Information")

	// Verify plan details are included
	assert.Contains(t, markdown, plan.ID)
	assert.Contains(t, markdown, plan.TaskID)
	assert.Contains(t, markdown, plan.Description)

	// Verify steps are included
	assert.Contains(t, markdown, "Step 1: Analyze requirements")
	assert.Contains(t, markdown, "Step 2: Design solution")

	// Verify context is included
	assert.Contains(t, markdown, "Test Task")
	assert.Contains(t, markdown, "HIGH")
	assert.Contains(t, markdown, "TODO")
}

func TestPlanningService_GenerateInitialPlanSteps(t *testing.T) {
	// Setup
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)
	executionService := NewExecutionService(cliManager, NewProcessManager())
	planningService := NewPlanningService(executionService, cliManager)

	// Create test task
	task := entity.Task{
		ID:        uuid.New(),
		Title:     "Test Task",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test step generation
	steps := planningService.generateInitialPlanSteps(task)

	// Assertions
	require.NotNil(t, steps)
	assert.Equal(t, 5, len(steps))

	// Verify each step has required fields
	for i, step := range steps {
		assert.NotEmpty(t, step.ID)
		assert.NotEmpty(t, step.Description)
		assert.NotEmpty(t, step.Action)
		assert.Equal(t, i+1, step.Order)
		assert.NotNil(t, step.Parameters)
		assert.Equal(t, task.ID.String(), step.Parameters["task_id"])
	}

	// Verify specific actions
	expectedActions := []string{"analysis", "design", "implement", "test", "validate"}
	for i, step := range steps {
		assert.Equal(t, expectedActions[i], step.Action)
	}
}