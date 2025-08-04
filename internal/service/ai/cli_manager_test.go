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
		config.APIKey = "test-api-key"
		
		manager, err := NewCLIManager(config)
		
		assert.NoError(t, err)
		assert.NotNil(t, manager)
		assert.Equal(t, config.CLIPath, manager.config.CLIPath)
		assert.Equal(t, config.APIKey, manager.config.APIKey)
	})

	t.Run("with nil config uses default", func(t *testing.T) {
		// Default config will fail validation because API key is empty
		manager, err := NewCLIManager(nil)
		
		assert.Error(t, err)
		assert.Nil(t, manager)
		assert.Contains(t, err.Error(), "API key is required")
	})

	t.Run("with invalid config", func(t *testing.T) {
		config := &CLIConfig{
			CLIPath:   "",
			APIKey:    "test-key",
			Model:     "claude-3.5-sonnet",
			MaxTokens: 4000,
			Timeout:   30 * time.Minute,
		}
		
		manager, err := NewCLIManager(config)
		
		assert.Error(t, err)
		assert.Nil(t, manager)
		assert.Contains(t, err.Error(), "CLI path is required")
	})
}

func TestCLIManager_ComposeCommand(t *testing.T) {
	config := DefaultCLIConfig()
	config.APIKey = "test-api-key"
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

	t.Run("planning command", func(t *testing.T) {
		task.Status = entity.TaskStatusPLANNING
		
		command, err := manager.ComposeCommand(task, plan)
		
		assert.NoError(t, err)
		assert.Contains(t, command, "claude")
		assert.Contains(t, command, "--model")
		assert.Contains(t, command, "claude-3.5-sonnet")
		assert.Contains(t, command, "--max-tokens")
		assert.Contains(t, command, "4000")
		assert.Contains(t, command, "--prompt")
		assert.Contains(t, command, "Test Task")
		assert.Contains(t, command, "This is a test task")
	})

	t.Run("implementation command", func(t *testing.T) {
		task.Status = entity.TaskStatusIMPLEMENTING
		
		command, err := manager.ComposeCommand(task, plan)
		
		assert.NoError(t, err)
		assert.Contains(t, command, "claude")
		assert.Contains(t, command, "--model")
		assert.Contains(t, command, "claude-3.5-sonnet")
		assert.Contains(t, command, "--prompt")
		assert.Contains(t, command, "Test Task")
		assert.Contains(t, command, "First step")
		assert.Contains(t, command, "Second step")
	})

	t.Run("unsupported task status", func(t *testing.T) {
		task.Status = entity.TaskStatusDONE
		
		command, err := manager.ComposeCommand(task, plan)
		
		assert.Error(t, err)
		assert.Empty(t, command)
		assert.Contains(t, err.Error(), "unsupported task status")
	})

	t.Run("empty task ID", func(t *testing.T) {
		task.ID = uuid.Nil
		task.Status = entity.TaskStatusPLANNING
		
		command, err := manager.ComposeCommand(task, plan)
		
		assert.Error(t, err)
		assert.Empty(t, command)
		assert.Contains(t, err.Error(), "task ID is required")
	})
}

func TestCLIManager_GetEnvironmentVars(t *testing.T) {
	config := &CLIConfig{
		CLIPath:          "claude",
		APIKey:           "test-api-key",
		Model:            "claude-3.5-sonnet",
		MaxTokens:        4000,
		Timeout:          30 * time.Minute,
		WorkingDirectory: "/tmp",
		EnableLogging:    true,
		RetryAttempts:    3,
	}
	
	manager, err := NewCLIManager(config)
	require.NoError(t, err)

	envVars := manager.GetEnvironmentVars()

	assert.Equal(t, "test-api-key", envVars["ANTHROPIC_API_KEY"])
	assert.Equal(t, "claude-3.5-sonnet", envVars["CLAUDE_MODEL"])
	assert.Equal(t, "/tmp", envVars["CLAUDE_WORKING_DIR"])
	assert.Equal(t, "info", envVars["CLAUDE_LOG_LEVEL"])
}

func TestCLIManager_GetEnvironmentVarsWithLoggingDisabled(t *testing.T) {
	config := DefaultCLIConfig()
	config.APIKey = "test-api-key"
	config.EnableLogging = false
	
	manager, err := NewCLIManager(config)
	require.NoError(t, err)

	envVars := manager.GetEnvironmentVars()

	assert.Equal(t, "error", envVars["CLAUDE_LOG_LEVEL"])
}

func TestCLIManager_SetWorkingDirectory(t *testing.T) {
	config := DefaultCLIConfig()
	config.APIKey = "test-api-key"
	manager, err := NewCLIManager(config)
	require.NoError(t, err)

	t.Run("set valid directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		err := manager.SetWorkingDirectory(tmpDir)
		
		assert.NoError(t, err)
		assert.Equal(t, tmpDir, manager.config.WorkingDirectory)
	})

	t.Run("set empty directory", func(t *testing.T) {
		err := manager.SetWorkingDirectory("")
		
		assert.NoError(t, err)
		assert.Equal(t, "", manager.config.WorkingDirectory)
	})

	t.Run("set non-existent directory", func(t *testing.T) {
		nonExistentDir := "/non/existent/directory"
		
		err := manager.SetWorkingDirectory(nonExistentDir)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})
}

func TestCLIManager_GetConfig(t *testing.T) {
	config := DefaultCLIConfig()
	config.APIKey = "test-api-key"
	manager, err := NewCLIManager(config)
	require.NoError(t, err)

	retrievedConfig := manager.GetConfig()

	// Should be a copy with same values
	assert.Equal(t, config.CLIPath, retrievedConfig.CLIPath)
	assert.Equal(t, config.APIKey, retrievedConfig.APIKey)
	assert.Equal(t, config.Model, retrievedConfig.Model)

	// Should be different instances
	assert.NotSame(t, manager.config, retrievedConfig)
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

func TestCLIManager_composePlanningCommand(t *testing.T) {
	config := DefaultCLIConfig()
	config.APIKey = "test-api-key"
	manager, err := NewCLIManager(config)
	require.NoError(t, err)

	task := entity.Task{
		ID:          uuid.New(),
		Title:       "Implement user authentication",
		Description: "Add JWT-based authentication to the API",
		Priority:    entity.TaskPriorityHigh,
	}

	args := manager.composePlanningCommand(task, nil)

	assert.Contains(t, args, "--prompt")
	
	// Find the prompt argument
	var prompt string
	for i, arg := range args {
		if arg == "--prompt" && i+1 < len(args) {
			prompt = args[i+1]
			break
		}
	}
	
	assert.Contains(t, prompt, "Implement user authentication")
	assert.Contains(t, prompt, "Add JWT-based authentication to the API")
	assert.Contains(t, prompt, "HIGH")
	assert.Contains(t, prompt, "implementation plan")
}

func TestCLIManager_composeImplementationCommand(t *testing.T) {
	config := DefaultCLIConfig()
	config.APIKey = "test-api-key"
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

	args := manager.composeImplementationCommand(task, plan)

	assert.Contains(t, args, "--prompt")
	
	// Find the prompt argument
	var prompt string
	for i, arg := range args {
		if arg == "--prompt" && i+1 < len(args) {
			prompt = args[i+1]
			break
		}
	}
	
	assert.Contains(t, prompt, "Implement user authentication")
	assert.Contains(t, prompt, "Create user model")
	assert.Contains(t, prompt, "Implement JWT middleware")
	assert.Contains(t, prompt, "production-ready code")
}