package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultCLIConfig(t *testing.T) {
	config := DefaultCLIConfig()

	assert.Equal(t, "npx -y @anthropic-ai/claude-code@latest -p --dangerously-skip-permissions --verbose --output-format=stream-json", config.CLICommand)
	assert.Equal(t, 30*time.Minute, config.Timeout)
	assert.Equal(t, "", config.WorkingDirectory)
	assert.True(t, config.EnableLogging)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, 5*time.Second, config.RetryDelay)
}

func TestCLIConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *CLIConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: func() *CLIConfig {
				config := DefaultCLIConfig()
				return config
			}(),
			expectError: false,
		},
		{
			name: "empty CLI command",
			config: &CLIConfig{
				CLICommand:    "",
				Timeout:       30 * time.Minute,
				RetryAttempts: 3,
			},
			expectError: true,
			errorMsg:    "CLI command is required",
		},
		{
			name: "empty CLI command",
			config: &CLIConfig{
				CLICommand:    "",
				Timeout:       30 * time.Minute,
				RetryAttempts: 3,
			},
			expectError: true,
			errorMsg:    "CLI command is required",
		},
		{
			name: "zero timeout",
			config: &CLIConfig{
				CLICommand:    "npx claude-code",
				Timeout:       0,
				RetryAttempts: 3,
			},
			expectError: true,
			errorMsg:    "timeout must be greater than 0",
		},
		{
			name: "negative retry attempts",
			config: &CLIConfig{
				CLICommand:    "npx claude-code",
				Timeout:       30 * time.Minute,
				RetryAttempts: -1,
			},
			expectError: true,
			errorMsg:    "retry attempts cannot be negative",
		},
		{
			name: "too many retry attempts",
			config: &CLIConfig{
				CLICommand:    "npx claude-code",
				Timeout:       30 * time.Minute,
				RetryAttempts: 11,
			},
			expectError: true,
			errorMsg:    "retry attempts cannot exceed 10",
		},
		{
			name: "negative retry delay",
			config: &CLIConfig{
				CLICommand:    "npx claude-code",
				Timeout:       30 * time.Minute,
				RetryAttempts: 3,
				RetryDelay:    -1 * time.Second,
			},
			expectError: true,
			errorMsg:    "retry delay cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCLIConfig_Clone(t *testing.T) {
	original := &CLIConfig{
		CLICommand:       "npx claude-code",
		Timeout:          30 * time.Minute,
		WorkingDirectory: "/tmp",
		EnableLogging:    true,
		RetryAttempts:    3,
		RetryDelay:       5 * time.Second,
	}

	cloned := original.Clone()

	// Verify all fields are copied
	assert.Equal(t, original.CLICommand, cloned.CLICommand)
	assert.Equal(t, original.Timeout, cloned.Timeout)
	assert.Equal(t, original.EnableLogging, cloned.EnableLogging)
	assert.Equal(t, original.RetryAttempts, cloned.RetryAttempts)
	assert.Equal(t, original.RetryDelay, cloned.RetryDelay)

	// Verify it's a different instance
	assert.NotSame(t, original, cloned)

	// Verify changing one doesn't affect the other
	cloned.CLICommand = "different-command"
	assert.NotEqual(t, original.CLICommand, cloned.CLICommand)
}

func TestCLIConfig_String(t *testing.T) {
	config := &CLIConfig{
		CLICommand:    "npx claude-code",
		Timeout:       30 * time.Minute,
		RetryAttempts: 3,
	}

	str := config.String()

	// Should include non-sensitive information
	assert.Contains(t, str, "npx claude-code")
	assert.Contains(t, str, "30m0s")
	assert.Contains(t, str, "3")

	// Should not include sensitive information
	assert.NotContains(t, str, "secret-key")
}

func TestCLIConfig_ValidateEdgeCases(t *testing.T) {
	t.Run("boundary values", func(t *testing.T) {
		config := DefaultCLIConfig()

		// Test minimum valid values
		config.Timeout = 1 * time.Nanosecond
		config.RetryAttempts = 0
		config.RetryDelay = 0

		err := config.Validate()
		assert.NoError(t, err)

		// Test maximum valid values
		config.RetryAttempts = 10

		err = config.Validate()
		assert.NoError(t, err)
	})
}
