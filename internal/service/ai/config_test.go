package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultCLIConfig(t *testing.T) {
	config := DefaultCLIConfig()
	
	assert.Equal(t, "claude", config.CLIPath)
	assert.Equal(t, "", config.APIKey) // Empty by default, must be set
	assert.Equal(t, "claude-3.5-sonnet", config.Model)
	assert.Equal(t, 4000, config.MaxTokens)
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
				config.APIKey = "test-api-key"
				return config
			}(),
			expectError: false,
		},
		{
			name: "empty CLI path",
			config: &CLIConfig{
				CLIPath:       "",
				APIKey:        "test-key",
				Model:         "claude-3.5-sonnet",
				MaxTokens:     4000,
				Timeout:       30 * time.Minute,
				RetryAttempts: 3,
			},
			expectError: true,
			errorMsg:    "CLI path is required",
		},
		{
			name: "empty API key",
			config: &CLIConfig{
				CLIPath:       "claude",
				APIKey:        "",
				Model:         "claude-3.5-sonnet",
				MaxTokens:     4000,
				Timeout:       30 * time.Minute,
				RetryAttempts: 3,
			},
			expectError: true,
			errorMsg:    "API key is required",
		},
		{
			name: "empty model",
			config: &CLIConfig{
				CLIPath:       "claude",
				APIKey:        "test-key",
				Model:         "",
				MaxTokens:     4000,
				Timeout:       30 * time.Minute,
				RetryAttempts: 3,
			},
			expectError: true,
			errorMsg:    "model is required",
		},
		{
			name: "zero max tokens",
			config: &CLIConfig{
				CLIPath:       "claude",
				APIKey:        "test-key",
				Model:         "claude-3.5-sonnet",
				MaxTokens:     0,
				Timeout:       30 * time.Minute,
				RetryAttempts: 3,
			},
			expectError: true,
			errorMsg:    "max tokens must be greater than 0",
		},
		{
			name: "max tokens too high",
			config: &CLIConfig{
				CLIPath:       "claude",
				APIKey:        "test-key",
				Model:         "claude-3.5-sonnet",
				MaxTokens:     100001,
				Timeout:       30 * time.Minute,
				RetryAttempts: 3,
			},
			expectError: true,
			errorMsg:    "max tokens cannot exceed 100000",
		},
		{
			name: "zero timeout",
			config: &CLIConfig{
				CLIPath:       "claude",
				APIKey:        "test-key",
				Model:         "claude-3.5-sonnet",
				MaxTokens:     4000,
				Timeout:       0,
				RetryAttempts: 3,
			},
			expectError: true,
			errorMsg:    "timeout must be greater than 0",
		},
		{
			name: "negative retry attempts",
			config: &CLIConfig{
				CLIPath:       "claude",
				APIKey:        "test-key",
				Model:         "claude-3.5-sonnet",
				MaxTokens:     4000,
				Timeout:       30 * time.Minute,
				RetryAttempts: -1,
			},
			expectError: true,
			errorMsg:    "retry attempts cannot be negative",
		},
		{
			name: "too many retry attempts",
			config: &CLIConfig{
				CLIPath:       "claude",
				APIKey:        "test-key",
				Model:         "claude-3.5-sonnet",
				MaxTokens:     4000,
				Timeout:       30 * time.Minute,
				RetryAttempts: 11,
			},
			expectError: true,
			errorMsg:    "retry attempts cannot exceed 10",
		},
		{
			name: "negative retry delay",
			config: &CLIConfig{
				CLIPath:       "claude",
				APIKey:        "test-key",
				Model:         "claude-3.5-sonnet",
				MaxTokens:     4000,
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
		CLIPath:          "claude",
		APIKey:           "test-key",
		Model:            "claude-3.5-sonnet",
		MaxTokens:        4000,
		Timeout:          30 * time.Minute,
		WorkingDirectory: "/tmp",
		EnableLogging:    true,
		RetryAttempts:    3,
		RetryDelay:       5 * time.Second,
	}

	cloned := original.Clone()

	// Verify all fields are copied
	assert.Equal(t, original.CLIPath, cloned.CLIPath)
	assert.Equal(t, original.APIKey, cloned.APIKey)
	assert.Equal(t, original.Model, cloned.Model)
	assert.Equal(t, original.MaxTokens, cloned.MaxTokens)
	assert.Equal(t, original.Timeout, cloned.Timeout)
	assert.Equal(t, original.WorkingDirectory, cloned.WorkingDirectory)
	assert.Equal(t, original.EnableLogging, cloned.EnableLogging)
	assert.Equal(t, original.RetryAttempts, cloned.RetryAttempts)
	assert.Equal(t, original.RetryDelay, cloned.RetryDelay)

	// Verify it's a different instance
	assert.NotSame(t, original, cloned)

	// Verify changing one doesn't affect the other
	cloned.CLIPath = "different-path"
	assert.NotEqual(t, original.CLIPath, cloned.CLIPath)
}

func TestCLIConfig_String(t *testing.T) {
	config := &CLIConfig{
		CLIPath:       "claude",
		APIKey:        "secret-key",
		Model:         "claude-3.5-sonnet",
		MaxTokens:     4000,
		Timeout:       30 * time.Minute,
		RetryAttempts: 3,
	}

	str := config.String()
	
	// Should include non-sensitive information
	assert.Contains(t, str, "claude")
	assert.Contains(t, str, "claude-3.5-sonnet")
	assert.Contains(t, str, "4000")
	assert.Contains(t, str, "30m0s")
	assert.Contains(t, str, "3")
	
	// Should not include sensitive information
	assert.NotContains(t, str, "secret-key")
}

func TestCLIConfig_ValidateEdgeCases(t *testing.T) {
	t.Run("boundary values", func(t *testing.T) {
		config := DefaultCLIConfig()
		config.APIKey = "test-api-key" // Set API key for validation to pass
		
		// Test minimum valid values
		config.MaxTokens = 1
		config.Timeout = 1 * time.Nanosecond
		config.RetryAttempts = 0
		config.RetryDelay = 0
		
		err := config.Validate()
		assert.NoError(t, err)
		
		// Test maximum valid values
		config.MaxTokens = 100000
		config.RetryAttempts = 10
		
		err = config.Validate()
		assert.NoError(t, err)
	})
}