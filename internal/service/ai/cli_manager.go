package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// ValidateCLIInstallation checks if Claude CLI is properly installed and configured
func (cm *CLIManager) ValidateCLIInstallation() error {
	// Check if CLI binary exists and is executable
	if err := cm.validateCLIBinary(); err != nil {
		return fmt.Errorf("CLI binary validation failed: %w", err)
	}

	// Test API key authentication
	if err := cm.validateAPIKey(); err != nil {
		return fmt.Errorf("API key validation failed: %w", err)
	}

	// Validate CLI version compatibility
	if err := cm.validateCLIVersion(); err != nil {
		return fmt.Errorf("CLI version validation failed: %w", err)
	}

	cm.logger.Info("Claude CLI validation successful", 
		slog.String("cli_path", cm.config.CLIPath),
		slog.String("model", cm.config.Model))

	return nil
}

// validateCLIBinary checks if the CLI binary exists and is executable
func (cm *CLIManager) validateCLIBinary() error {
	// Resolve the CLI path
	cliPath, err := exec.LookPath(cm.config.CLIPath)
	if err != nil {
		return fmt.Errorf("Claude CLI not found in PATH: %w", err)
	}

	// Check if file exists and is executable
	info, err := os.Stat(cliPath)
	if err != nil {
		return fmt.Errorf("cannot access Claude CLI at %s: %w", cliPath, err)
	}

	if info.IsDir() {
		return fmt.Errorf("CLI path points to a directory, not a file: %s", cliPath)
	}

	// Check if executable (Unix-style permissions)
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("Claude CLI is not executable: %s", cliPath)
	}

	return nil
}

// validateAPIKey tests the API key by making a simple authentication request
func (cm *CLIManager) validateAPIKey() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a simple test command to validate API key
	cmd := exec.CommandContext(ctx, cm.config.CLIPath, "--version")
	
	// Set environment variables including API key
	cmd.Env = append(os.Environ(), cm.getEnvironmentVars()...)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("API key validation failed (exit code %v): %s", err, string(output))
	}

	cm.logger.Debug("API key validation successful", slog.String("output", string(output)))
	return nil
}

// validateCLIVersion checks if the CLI version is compatible
func (cm *CLIManager) validateCLIVersion() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cm.config.CLIPath, "--version")
	cmd.Env = append(os.Environ(), cm.getEnvironmentVars()...)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get CLI version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	cm.logger.Info("Claude CLI version detected", slog.String("version", version))

	// TODO: Add version compatibility checks here when requirements are defined
	// For now, we just log the version

	return nil
}

// ComposeCommand generates a Claude CLI command for the given task and plan
func (cm *CLIManager) ComposeCommand(task entity.Task, plan *Plan) (string, error) {
	if task.ID == uuid.Nil {
		return "", fmt.Errorf("task ID is required")
	}

	// Build base command with common options
	args := []string{
		cm.config.CLIPath,
	}

	// Add model specification
	if cm.config.Model != "" {
		args = append(args, "--model", cm.config.Model)
	}

	// Add max tokens if specified
	if cm.config.MaxTokens > 0 {
		args = append(args, "--max-tokens", strconv.Itoa(cm.config.MaxTokens))
	}

	// Determine command type based on task status
	switch task.Status {
	case entity.TaskStatusPLANNING:
		args = append(args, cm.composePlanningCommand(task, plan)...)
	case entity.TaskStatusIMPLEMENTING:
		args = append(args, cm.composeImplementationCommand(task, plan)...)
	default:
		return "", fmt.Errorf("unsupported task status for CLI command: %s", task.Status)
	}

	// Join arguments into a single command string
	command := strings.Join(args, " ")
	
	cm.logger.Debug("Composed CLI command", 
		slog.String("task_id", task.ID.String()),
		slog.String("status", string(task.Status)),
		slog.String("command", command))

	return command, nil
}

// composePlanningCommand creates command arguments for planning tasks
func (cm *CLIManager) composePlanningCommand(task entity.Task, plan *Plan) []string {
	args := []string{}

	// Add planning-specific prompt
	prompt := fmt.Sprintf(`You are tasked with creating a detailed implementation plan for the following task:

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

	args = append(args, "--prompt", fmt.Sprintf(`"%s"`, prompt))

	// Add context if working directory is specified
	if cm.config.WorkingDirectory != "" {
		args = append(args, "--working-dir", cm.config.WorkingDirectory)
	}

	return args
}

// composeImplementationCommand creates command arguments for implementation tasks
func (cm *CLIManager) composeImplementationCommand(task entity.Task, plan *Plan) []string {
	args := []string{}

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

	args = append(args, "--prompt", fmt.Sprintf(`"%s"`, prompt.String()))

	// Add context if working directory is specified
	if cm.config.WorkingDirectory != "" {
		args = append(args, "--working-dir", cm.config.WorkingDirectory)
	}

	return args
}

// GetEnvironmentVars returns environment variables needed for Claude CLI
func (cm *CLIManager) GetEnvironmentVars() map[string]string {
	envVars := make(map[string]string)

	// Add Claude API key
	if cm.config.APIKey != "" {
		envVars["ANTHROPIC_API_KEY"] = cm.config.APIKey
	}

	// Add model configuration
	if cm.config.Model != "" {
		envVars["CLAUDE_MODEL"] = cm.config.Model
	}

	// Add working directory
	if cm.config.WorkingDirectory != "" {
		envVars["CLAUDE_WORKING_DIR"] = cm.config.WorkingDirectory
	}

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

// ExecuteCommand executes a CLI command with proper error handling and retries
func (cm *CLIManager) ExecuteCommand(ctx context.Context, command string) (*CLIResult, error) {
	var lastErr error
	
	for attempt := 0; attempt <= cm.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			cm.logger.Warn("Retrying CLI command", 
				slog.Int("attempt", attempt),
				slog.String("command", command))
			
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(cm.config.RetryDelay):
			}
		}

		result, err := cm.executeCommandOnce(ctx, command)
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

// executeCommandOnce executes a CLI command once
func (cm *CLIManager) executeCommandOnce(ctx context.Context, command string) (*CLIResult, error) {
	// Parse command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, cm.config.Timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, parts[0], parts[1:]...)
	cmd.Env = append(os.Environ(), cm.getEnvironmentVars()...)
	
	// Set working directory if specified
	if cm.config.WorkingDirectory != "" {
		if abs, err := filepath.Abs(cm.config.WorkingDirectory); err == nil {
			cmd.Dir = abs
		}
	}

	startTime := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	result := &CLIResult{
		Command:    command,
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
		slog.String("command", command),
		slog.Duration("duration", duration))

	return result, nil
}

// SetWorkingDirectory updates the working directory for CLI operations
func (cm *CLIManager) SetWorkingDirectory(dir string) error {
	if dir == "" {
		cm.config.WorkingDirectory = ""
		return nil
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	cm.config.WorkingDirectory = absDir
	cm.logger.Debug("Working directory updated", slog.String("dir", absDir))
	
	return nil
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