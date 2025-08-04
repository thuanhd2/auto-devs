package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// Plan represents a task execution plan (temporary definition)
// TODO: Move this to proper entity package when planning system is implemented
type Plan struct {
	ID          string            `json:"id"`
	TaskID      string            `json:"task_id"`
	Description string            `json:"description"`
	Steps       []PlanStep        `json:"steps"`
	Context     map[string]string `json:"context"`
	CreatedAt   time.Time         `json:"created_at"`
}

// PlanStep represents a single step in a plan
type PlanStep struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Action      string            `json:"action"`
	Parameters  map[string]string `json:"parameters"`
	Order       int               `json:"order"`
}

// CLIManager manages interactions with Claude Code CLI
type CLIManager struct {
	config *CLIConfig
	logger *slog.Logger
}

// NewCLIManager creates a new CLIManager instance
func NewCLIManager(config *CLIConfig) (*CLIManager, error) {
	if config == nil {
		config = DefaultCLIConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid CLI configuration: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return &CLIManager{
		config: config,
		logger: logger,
	}, nil
}

// ComposePrompt generates a prompt for the given task and plan
func (cm *CLIManager) ComposePrompt(task entity.Task, plan *Plan) (string, error) {
	if task.ID == uuid.Nil {
		return "", fmt.Errorf("task ID is required")
	}

	var prompt strings.Builder

	// Determine prompt type based on task status
	switch task.Status {
	case entity.TaskStatusPLANNING:
		prompt.WriteString(cm.composePlanningPrompt(task, plan))
	case entity.TaskStatusIMPLEMENTING:
		prompt.WriteString(cm.composeImplementationPrompt(task, plan))
	default:
		return "", fmt.Errorf("unsupported task status for prompt composition: %s", task.Status)
	}

	composedPrompt := prompt.String()

	cm.logger.Debug("Composed prompt",
		slog.String("task_id", task.ID.String()),
		slog.String("status", string(task.Status)),
		slog.String("prompt_length", fmt.Sprintf("%d", len(composedPrompt))))

	return composedPrompt, nil
}

// composePlanningPrompt creates prompt for planning tasks
func (cm *CLIManager) composePlanningPrompt(task entity.Task, _ *Plan) string {
	return fmt.Sprintf(`You are tasked with creating a detailed implementation plan for the following task:

Title: %s
Description: %s
Priority: %s

Please analyze the requirements and create a step-by-step implementation plan. Include:
1. Technical approach and architecture decisions
2. Implementation steps in logical order  
3. Potential challenges and solutions
4. Testing strategy
5. Definition of done criteria

Focus on being thorough and actionable.`,
		task.Title,
		task.Description,
		task.Priority)
}

// composeImplementationPrompt creates prompt for implementation tasks
func (cm *CLIManager) composeImplementationPrompt(task entity.Task, plan *Plan) string {
	var prompt strings.Builder
	prompt.WriteString(fmt.Sprintf(`You are tasked with implementing the following:

Title: %s
Description: %s
Priority: %s

`, task.Title, task.Description, task.Priority))

	// Add plan context if available
	if plan != nil && len(plan.Steps) > 0 {
		prompt.WriteString("Implementation Plan:\n")
		for i, step := range plan.Steps {
			prompt.WriteString(fmt.Sprintf("%d. %s\n", i+1, step.Description))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString(`Please implement the solution following these guidelines:
1. Follow existing code patterns and architecture
2. Write clean, maintainable, and well-commented code
3. Include appropriate error handling
4. Add tests where applicable
5. Ensure the implementation meets all requirements

Focus on producing production-ready code.`)

	return prompt.String()
}

// GetEnvironmentVars returns environment variables needed for Claude CLI
func (cm *CLIManager) GetEnvironmentVars() map[string]string {
	envVars := make(map[string]string)

	// Add logging configuration
	if cm.config.EnableLogging {
		envVars["CLAUDE_LOG_LEVEL"] = "info"
	} else {
		envVars["CLAUDE_LOG_LEVEL"] = "error"
	}

	return envVars
}

// getEnvironmentVars returns environment variables as a slice for exec.Cmd
func (cm *CLIManager) getEnvironmentVars() []string {
	envVars := cm.GetEnvironmentVars()
	var env []string

	for key, value := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env
}

// ExecuteCommand executes a CLI command with prompt via stdin
func (cm *CLIManager) ExecuteCommand(ctx context.Context, prompt string) (*CLIResult, error) {
	var lastErr error

	for attempt := 0; attempt <= cm.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			cm.logger.Warn("Retrying CLI command",
				slog.Int("attempt", attempt),
				slog.String("cli_command", cm.config.CLICommand))

			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(cm.config.RetryDelay):
			}
		}

		result, err := cm.executeCommandOnce(ctx, cm.config.CLICommand, prompt)
		if err == nil {
			return result, nil
		}

		lastErr = err
		cm.logger.Error("CLI command failed",
			slog.Int("attempt", attempt+1),
			slog.String("error", err.Error()))
	}

	return nil, fmt.Errorf("command failed after %d attempts: %w", cm.config.RetryAttempts+1, lastErr)
}

// ExecuteTask composes a prompt for the given task and executes the CLI command
func (cm *CLIManager) ExecuteTask(ctx context.Context, task entity.Task, plan *Plan) (*CLIResult, error) {
	// Compose prompt for the task
	prompt, err := cm.ComposePrompt(task, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to compose prompt: %w", err)
	}

	// Execute the command with the composed prompt
	return cm.ExecuteCommand(ctx, prompt)
}

// executeCommandOnce executes a CLI command once with prompt via stdin
func (cm *CLIManager) executeCommandOnce(ctx context.Context, cliCommand string, prompt string) (*CLIResult, error) {
	// Parse command into parts
	parts := strings.Fields(cliCommand)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty CLI command")
	}

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, cm.config.Timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, parts[0], parts[1:]...)
	cmd.Env = append(os.Environ(), cm.getEnvironmentVars()...)

	// Set working directory if specified
	if cm.config.WorkingDirectory != "" {
		cmd.Dir = cm.config.WorkingDirectory
	}

	// Set stdin to the prompt
	cmd.Stdin = strings.NewReader(prompt)

	startTime := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	result := &CLIResult{
		Command:    cliCommand,
		Output:     string(output),
		Duration:   duration,
		ExitCode:   0,
		Success:    err == nil,
		ExecutedAt: startTime,
	}

	if err != nil {
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		return result, fmt.Errorf("command execution failed: %w", err)
	}

	cm.logger.Info("CLI command executed successfully",
		slog.String("command", cliCommand),
		slog.Duration("duration", duration))

	return result, nil
}

// GetConfig returns a copy of the current configuration
func (cm *CLIManager) GetConfig() *CLIConfig {
	return cm.config.Clone()
}

// CLIResult represents the result of a CLI command execution
type CLIResult struct {
	Command    string        `json:"command"`
	Output     string        `json:"output"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
	ExitCode   int           `json:"exit_code"`
	Success    bool          `json:"success"`
	ExecutedAt time.Time     `json:"executed_at"`
}

// String returns a string representation of the CLI result
func (r *CLIResult) String() string {
	status := "SUCCESS"
	if !r.Success {
		status = "FAILED"
	}

	return fmt.Sprintf("CLIResult{Status: %s, Duration: %v, ExitCode: %d}",
		status, r.Duration, r.ExitCode)
}

// ToJSON converts the result to JSON format
func (r *CLIResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result to JSON: %w", err)
	}
	return string(data), nil
}
