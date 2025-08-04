package ai

import (
	"fmt"
	"time"
)

// CLIConfig holds configuration for Claude Code CLI
type CLIConfig struct {
	// CLIPath is the path to the Claude CLI binary
	CLIPath string `json:"cli_path" validate:"required"`

	// APIKey is the Claude API key for authentication
	APIKey string `json:"api_key" validate:"required"`

	// Model specifies the Claude model to use
	Model string `json:"model" validate:"required"`

	// MaxTokens sets the maximum tokens for AI responses
	MaxTokens int `json:"max_tokens" validate:"min=1,max=100000"`

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
		CLIPath:          "claude", // Assumes 'claude' is in PATH
		APIKey:           "", // Must be set before use
		Model:            "claude-3.5-sonnet",
		MaxTokens:        4000,
		Timeout:          30 * time.Minute,
		WorkingDirectory: "",
		EnableLogging:    true,
		RetryAttempts:    3,
		RetryDelay:       5 * time.Second,
	}
}

// Validate validates the CLI configuration
func (c *CLIConfig) Validate() error {
	if c.CLIPath == "" {
		return fmt.Errorf("CLI path is required")
	}

	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if c.Model == "" {
		return fmt.Errorf("model is required")
	}

	if c.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be greater than 0")
	}

	if c.MaxTokens > 100000 {
		return fmt.Errorf("max tokens cannot exceed 100000")
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
		CLIPath:          c.CLIPath,
		APIKey:           c.APIKey,
		Model:            c.Model,
		MaxTokens:        c.MaxTokens,
		Timeout:          c.Timeout,
		WorkingDirectory: c.WorkingDirectory,
		EnableLogging:    c.EnableLogging,
		RetryAttempts:    c.RetryAttempts,
		RetryDelay:       c.RetryDelay,
	}
}

// String returns a string representation of the config (without sensitive data)
func (c *CLIConfig) String() string {
	return fmt.Sprintf("CLIConfig{CLIPath: %s, Model: %s, MaxTokens: %d, Timeout: %v, RetryAttempts: %d}",
		c.CLIPath, c.Model, c.MaxTokens, c.Timeout, c.RetryAttempts)
}