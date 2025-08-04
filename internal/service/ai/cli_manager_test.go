package ai

import (
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCLIManager(t *testing.T) {
	t.Run("with valid config", func(t *testing.T) {
		config := DefaultCLIConfig()

		manager, err := NewCLIManager(config)

		assert.NoError(t, err)
		assert.NotNil(t, manager)
		assert.Equal(t, config.CLICommand, manager.config.CLICommand)
	})

	t.Run("with nil config uses default", func(t *testing.T) {
		// Nil config will not causes any failure, because it will use the default config
		manager, err := NewCLIManager(nil)

		assert.NoError(t, err)
		assert.NotNil(t, manager)
		assert.Equal(t, DefaultCLIConfig().CLICommand, manager.config.CLICommand)
	})

	t.Run("with invalid config", func(t *testing.T) {
		config := &CLIConfig{
			CLICommand: "",
			Timeout:    30 * time.Minute,
		}

		manager, err := NewCLIManager(config)

		assert.Error(t, err)
		assert.Nil(t, manager)
		assert.Contains(t, err.Error(), "CLI command is required")
	})
}

func TestCLIManager_ComposePrompt(t *testing.T) {
	config := DefaultCLIConfig()
	manager, err := NewCLIManager(config)
	require.NoError(t, err)

	task := entity.Task{
		ID:          uuid.New(),
		Title:       "Test Task",
		Description: "This is a test task",
		Priority:    entity.TaskPriorityHigh,
	}

	plan := &Plan{
		ID:          "plan-123",
		TaskID:      task.ID.String(),
		Description: "Test plan",
		Steps: []PlanStep{
			{
				ID:          "step-1",
				Description: "First step",
				Action:      "create",
				Order:       1,
			},
			{
				ID:          "step-2",
				Description: "Second step",
				Action:      "update",
				Order:       2,
			},
		},
	}

	t.Run("planning prompt", func(t *testing.T) {
		task.Status = entity.TaskStatusPLANNING

		prompt, err := manager.ComposePrompt(task, plan)

		assert.NoError(t, err)
		assert.Contains(t, prompt, "Test Task")
		assert.Contains(t, prompt, "This is a test task")
		assert.Contains(t, prompt, "implementation plan")
		assert.Contains(t, prompt, "Technical approach")
	})

	t.Run("implementation prompt", func(t *testing.T) {
		task.Status = entity.TaskStatusIMPLEMENTING

		prompt, err := manager.ComposePrompt(task, plan)

		assert.NoError(t, err)
		assert.Contains(t, prompt, "Test Task")
		assert.Contains(t, prompt, "This is a test task")
		assert.Contains(t, prompt, "Implementation Plan:")
		assert.Contains(t, prompt, "1. First step")
		assert.Contains(t, prompt, "2. Second step")
		assert.Contains(t, prompt, "production-ready code")
	})

	t.Run("unsupported status", func(t *testing.T) {
		task.Status = entity.TaskStatusDONE

		prompt, err := manager.ComposePrompt(task, plan)

		assert.Error(t, err)
		assert.Empty(t, prompt)
		assert.Contains(t, err.Error(), "unsupported task status")
	})
}

func TestCLIManager_GetEnvironmentVars(t *testing.T) {
	config := &CLIConfig{
		CLICommand:    "claude",
		Timeout:       30 * time.Minute,
		EnableLogging: true,
	}

	manager := &CLIManager{config: config}

	envVars := manager.GetEnvironmentVars()

	assert.Equal(t, "info", envVars["CLAUDE_LOG_LEVEL"])
}

func TestCLIManager_GetEnvironmentVarsWithLoggingDisabled(t *testing.T) {
	config := &CLIConfig{
		CLICommand:    "claude",
		EnableLogging: false,
	}

	manager := &CLIManager{config: config}

	envVars := manager.GetEnvironmentVars()

	assert.Equal(t, "error", envVars["CLAUDE_LOG_LEVEL"])
}

func TestCLIManager_GetConfig(t *testing.T) {
	config := &CLIConfig{
		CLICommand:    "claude",
		Timeout:       30 * time.Minute,
		EnableLogging: true,
		RetryAttempts: 3,
		RetryDelay:    5 * time.Second,
	}

	manager := &CLIManager{config: config}

	retrievedConfig := manager.GetConfig()

	// Should be a copy with same values
	assert.Equal(t, config.CLICommand, retrievedConfig.CLICommand)
	assert.Equal(t, config.Timeout, retrievedConfig.Timeout)
	assert.Equal(t, config.EnableLogging, retrievedConfig.EnableLogging)
	assert.Equal(t, config.RetryAttempts, retrievedConfig.RetryAttempts)
	assert.Equal(t, config.RetryDelay, retrievedConfig.RetryDelay)
}

func TestCLIResult_String(t *testing.T) {
	t.Run("successful result", func(t *testing.T) {
		result := &CLIResult{
			Command:    "claude --version",
			Output:     "claude 1.0.0",
			Duration:   2 * time.Second,
			ExitCode:   0,
			Success:    true,
			ExecutedAt: time.Now(),
		}

		str := result.String()
		assert.Contains(t, str, "SUCCESS")
		assert.Contains(t, str, "2s")
		assert.Contains(t, str, "ExitCode: 0")
	})

	t.Run("failed result", func(t *testing.T) {
		result := &CLIResult{
			Command:    "claude --invalid",
			Output:     "",
			Error:      "unknown flag: --invalid",
			Duration:   1 * time.Second,
			ExitCode:   1,
			Success:    false,
			ExecutedAt: time.Now(),
		}

		str := result.String()
		assert.Contains(t, str, "FAILED")
		assert.Contains(t, str, "1s")
		assert.Contains(t, str, "ExitCode: 1")
	})
}

func TestCLIResult_ToJSON(t *testing.T) {
	result := &CLIResult{
		Command:    "claude --version",
		Output:     "claude 1.0.0",
		Duration:   2 * time.Second,
		ExitCode:   0,
		Success:    true,
		ExecutedAt: time.Now(),
	}

	jsonStr, err := result.ToJSON()

	assert.NoError(t, err)
	assert.Contains(t, jsonStr, "\"command\"")
	assert.Contains(t, jsonStr, "\"output\"")
	assert.Contains(t, jsonStr, "\"success\"")
	assert.Contains(t, jsonStr, "claude --version")
}

func TestPlan_Struct(t *testing.T) {
	// Test the Plan struct definition and basic functionality
	plan := &Plan{
		ID:          "plan-123",
		TaskID:      "task-456",
		Description: "Test plan description",
		Steps: []PlanStep{
			{
				ID:          "step-1",
				Description: "First step",
				Action:      "create",
				Parameters:  map[string]string{"file": "test.go"},
				Order:       1,
			},
			{
				ID:          "step-2",
				Description: "Second step",
				Action:      "update",
				Parameters:  map[string]string{"method": "testMethod"},
				Order:       2,
			},
		},
		Context:   map[string]string{"language": "go", "framework": "gin"},
		CreatedAt: time.Now(),
	}

	assert.Equal(t, "plan-123", plan.ID)
	assert.Equal(t, "task-456", plan.TaskID)
	assert.Equal(t, "Test plan description", plan.Description)
	assert.Len(t, plan.Steps, 2)
	assert.Equal(t, "First step", plan.Steps[0].Description)
	assert.Equal(t, "create", plan.Steps[0].Action)
	assert.Equal(t, "test.go", plan.Steps[0].Parameters["file"])
	assert.Equal(t, 1, plan.Steps[0].Order)
	assert.Equal(t, "go", plan.Context["language"])
}

func TestCLIManager_composePlanningPrompt(t *testing.T) {
	config := DefaultCLIConfig()
	manager, err := NewCLIManager(config)
	require.NoError(t, err)

	task := entity.Task{
		ID:          uuid.New(),
		Title:       "Implement user authentication",
		Description: "Add JWT-based authentication to the API",
		Priority:    entity.TaskPriorityHigh,
	}

	prompt := manager.composePlanningPrompt(task, nil)

	assert.Contains(t, prompt, "Implement user authentication")
	assert.Contains(t, prompt, "Add JWT-based authentication to the API")
	assert.Contains(t, prompt, "HIGH")
	assert.Contains(t, prompt, "implementation plan")
	assert.Contains(t, prompt, "Technical approach")
}

func TestCLIManager_composeImplementationPrompt(t *testing.T) {
	config := DefaultCLIConfig()
	manager, err := NewCLIManager(config)
	require.NoError(t, err)

	task := entity.Task{
		ID:          uuid.New(),
		Title:       "Implement user authentication",
		Description: "Add JWT-based authentication to the API",
		Priority:    entity.TaskPriorityHigh,
	}

	plan := &Plan{
		Steps: []PlanStep{
			{Description: "Create user model"},
			{Description: "Implement JWT middleware"},
		},
	}

	prompt := manager.composeImplementationPrompt(task, plan)

	assert.Contains(t, prompt, "Implement user authentication")
	assert.Contains(t, prompt, "Implementation Plan:")
	assert.Contains(t, prompt, "1. Create user model")
	assert.Contains(t, prompt, "2. Implement JWT middleware")
	assert.Contains(t, prompt, "production-ready code")
}
