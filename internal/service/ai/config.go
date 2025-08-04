package ai

import (
	"fmt"
	"time"
)

// CLIConfig holds configuration for Claude Code CLI
type CLIConfig struct {
	// CLICommand is the complete CLI command to execute
	// Example: "npx -y @anthropic-ai/claude-code@latest -p --dangerously-skip-permissions --verbose --output-format=stream-json"
	CLICommand string `json:"cli_command" validate:"required"`

	// Timeout sets the maximum duration for CLI operations
	Timeout time.Duration `json:"timeout"`

	// WorkingDirectory is the base directory for CLI operations
	WorkingDirectory string `json:"working_directory"`

	// EnableLogging controls whether to enable detailed logging
	EnableLogging bool `json:"enable_logging"`

	// RetryAttempts sets the number of retry attempts on failure
	RetryAttempts int `json:"retry_attempts" validate:"min=0,max=10"`

	// RetryDelay sets the delay between retry attempts
	RetryDelay time.Duration `json:"retry_delay"`
}

// DefaultCLIConfig returns a default configuration for Claude CLI
func DefaultCLIConfig() *CLIConfig {
	return &CLIConfig{
		CLICommand:       "npx -y @anthropic-ai/claude-code@latest -p --dangerously-skip-permissions --verbose --output-format=stream-json",
		Timeout:          30 * time.Minute,
		WorkingDirectory: "",
		EnableLogging:    true,
		RetryAttempts:    3,
		RetryDelay:       5 * time.Second,
	}
}

// Validate validates the CLI configuration
func (c *CLIConfig) Validate() error {
	if c.CLICommand == "" {
		return fmt.Errorf("CLI command is required")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	if c.RetryAttempts < 0 {
		return fmt.Errorf("retry attempts cannot be negative")
	}

	if c.RetryAttempts > 10 {
		return fmt.Errorf("retry attempts cannot exceed 10")
	}

	if c.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}

	return nil
}

// Clone creates a deep copy of the configuration
func (c *CLIConfig) Clone() *CLIConfig {
	return &CLIConfig{
		CLICommand:       c.CLICommand,
		Timeout:          c.Timeout,
		WorkingDirectory: c.WorkingDirectory,
		EnableLogging:    c.EnableLogging,
		RetryAttempts:    c.RetryAttempts,
		RetryDelay:       c.RetryDelay,
	}
}

// String returns a string representation of the config (without sensitive data)
func (c *CLIConfig) String() string {
	return fmt.Sprintf("CLIConfig{CLICommand: %s, Timeout: %v, RetryAttempts: %d}",
		c.CLICommand, c.Timeout, c.RetryAttempts)
}
