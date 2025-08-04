package git

import (
	"testing"
)

func TestBranchNamingConfig(t *testing.T) {
	// Test default configuration
	config := DefaultBranchNamingConfig()
	if config.Prefix != "task" {
		t.Errorf("Expected prefix 'task', got '%s'", config.Prefix)
	}
	if !config.IncludeID {
		t.Error("Expected IncludeID to be true")
	}
	if config.Separator != "-" {
		t.Errorf("Expected separator '-', got '%s'", config.Separator)
	}
	if config.MaxLength != 255 {
		t.Errorf("Expected max length 255, got %d", config.MaxLength)
	}
	if !config.UseSlug {
		t.Error("Expected UseSlug to be true")
	}
}

func TestBranchManager_GenerateBranchName(t *testing.T) {
	// Create mock dependencies
	executor := &MockCommandExecutor{}
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)

	// Test cases
	testCases := []struct {
		name     string
		taskID   string
		title    string
		config   *BranchNamingConfig
		expected string
		hasError bool
	}{
		{
			name:     "Basic branch name generation",
			taskID:   "123",
			title:    "Implement user authentication",
			config:   DefaultBranchNamingConfig(),
			expected: "task-123-implement-user-authentication",
			hasError: false,
		},
		{
			name:     "Branch name without ID",
			taskID:   "123",
			title:    "Fix bug",
			config:   &BranchNamingConfig{Prefix: "feature", IncludeID: false, Separator: "-", UseSlug: true, MaxLength: 255},
			expected: "feature-fix-bug",
			hasError: false,
		},
		{
			name:     "Branch name with special characters",
			taskID:   "456",
			title:    "Add API endpoint for user@example.com",
			config:   DefaultBranchNamingConfig(),
			expected: "task-456-add-api-endpoint-for-user-example-com",
			hasError: false,
		},
		{
			name:     "Branch name with underscores",
			taskID:   "789",
			title:    "Update_database_schema",
			config:   DefaultBranchNamingConfig(),
			expected: "task-789-update-database-schema",
			hasError: false,
		},
		{
			name:     "Empty title",
			taskID:   "999",
			title:    "",
			config:   DefaultBranchNamingConfig(),
			expected: "task-999",
			hasError: false,
		},
		{
			name:     "Very long title",
			taskID:   "100",
			title:    "This is a very long title that should be truncated to fit within the maximum branch name length limit",
			config:   DefaultBranchNamingConfig(),
			expected: "task-100-this-is-a-very-long-title-that-should-be",
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bm := NewBranchManager(commands, validator, tc.config)

			result, err := bm.GenerateBranchName(tc.taskID, tc.title)

			if tc.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}

			// Validate the generated name
			if err := bm.validator.ValidateBranchName(result); err != nil {
				t.Errorf("Generated branch name is invalid: %v", err)
			}
		})
	}
}

func TestBranchManager_slugifyTitle(t *testing.T) {
	executor := &MockCommandExecutor{}
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	bm := NewBranchManager(commands, validator, DefaultBranchNamingConfig())

	testCases := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"API Endpoint", "api-endpoint"},
		{"User@example.com", "user-example-com"},
		{"Fix bug #123", "fix-bug-123"},
		{"Update_database", "update-database"},
		{"Special!@#$%^&*()", "special"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"", ""},
		{"A", "a"},
		{"Very long title that should be truncated to fit within reasonable limits", "very-long-title-that-should-be-truncated-to-fit"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := bm.slugifyTitle(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestBranchManager_simpleTitleProcess(t *testing.T) {
	executor := &MockCommandExecutor{}
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	bm := NewBranchManager(commands, validator, DefaultBranchNamingConfig())

	testCases := []struct {
		input    string
		expected string
	}{
		{"Hello World", "Hello-World"},
		{"API Endpoint", "API-Endpoint"},
		{"User@example.com", "Userexamplecom"},
		{"Fix bug #123", "Fix-bug-123"},
		{"Update_database", "Updatedatabase"},
		{"Special!@#$%^&*()", "Special"},
		{"Multiple   Spaces", "Multiple-Spaces"},
		{"", ""},
		{"A", "A"},
		{"Very long title that should be truncated", "Very-long-title-that-should"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := bm.simpleTitleProcess(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestBranchManager_cleanBranchName(t *testing.T) {
	executor := &MockCommandExecutor{}
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	bm := NewBranchManager(commands, validator, DefaultBranchNamingConfig())

	testCases := []struct {
		input    string
		expected string
	}{
		{"task-123-title", "task-123-title"},
		{"-task-123-title-", "task-123-title"},
		{"task--123--title", "task-123-title"},
		{".hidden", "branch-.hidden"},
		{"branch.lock", "branch"},
		{"normal-branch", "normal-branch"},
		{"---multiple---separators---", "multiple-separators"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := bm.cleanBranchName(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestBranchManager_ValidateBranchNameFormat(t *testing.T) {
	executor := &MockCommandExecutor{}
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	bm := NewBranchManager(commands, validator, DefaultBranchNamingConfig())

	testCases := []struct {
		name           string
		branchName     string
		shouldBeValid  bool
		expectedIssues []string
	}{
		{
			name:           "Valid branch name",
			branchName:     "task-123-title",
			shouldBeValid:  true,
			expectedIssues: []string{},
		},
		{
			name:           "Empty branch name",
			branchName:     "",
			shouldBeValid:  false,
			expectedIssues: []string{"Branch name cannot be empty"},
		},
		{
			name:           "Branch name too long",
			branchName:     string(make([]byte, 256)),
			shouldBeValid:  false,
			expectedIssues: []string{"Branch name too long (max 255 characters)"},
		},
		{
			name:           "Branch name starting with dot",
			branchName:     ".hidden",
			shouldBeValid:  false,
			expectedIssues: []string{"Cannot start with dot"},
		},
		{
			name:           "Branch name ending with .lock",
			branchName:     "branch.lock",
			shouldBeValid:  false,
			expectedIssues: []string{"Cannot end with .lock"},
		},
		{
			name:           "Branch name with spaces",
			branchName:     "branch name",
			shouldBeValid:  false,
			expectedIssues: []string{"Cannot contain whitespace"},
		},
		{
			name:           "Branch name with consecutive dots",
			branchName:     "branch..name",
			shouldBeValid:  false,
			expectedIssues: []string{"Cannot contain consecutive dots"},
		},
		{
			name:           "Reserved name HEAD",
			branchName:     "HEAD",
			shouldBeValid:  false,
			expectedIssues: []string{"Cannot use reserved name 'HEAD'"},
		},
		{
			name:           "Branch name with special characters",
			branchName:     "branch@{name",
			shouldBeValid:  false,
			expectedIssues: []string{"Cannot contain @{"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := bm.ValidateBranchNameFormat(tc.branchName)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.IsValid != tc.shouldBeValid {
				t.Errorf("Expected validity %v, got %v", tc.shouldBeValid, result.IsValid)
			}

			if len(result.Issues) != len(tc.expectedIssues) {
				t.Errorf("Expected %d issues, got %d: %v", len(tc.expectedIssues), len(result.Issues), result.Issues)
			}
		})
	}
}

func TestBranchManager_calculateSimilarity(t *testing.T) {
	executor := &MockCommandExecutor{}
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	bm := NewBranchManager(commands, validator, DefaultBranchNamingConfig())

	testCases := []struct {
		s1       string
		s2       string
		expected float64
	}{
		{"hello", "hello", 1.0},
		{"hello", "world", 0.6},   // 1 common character 'o' out of 5 max length
		{"task", "task-123", 0.5}, // 4 common characters out of 8 max length
		{"", "hello", 0.0},
		{"hello", "", 0.0},
		{"", "", 1.0},
		{"abc", "def", 0.0},
		{"abc", "abc", 1.0},
	}

	for _, tc := range testCases {
		t.Run(tc.s1+"_"+tc.s2, func(t *testing.T) {
			result := bm.calculateSimilarity(tc.s1, tc.s2)
			if result != tc.expected {
				t.Errorf("Expected %f, got %f", tc.expected, result)
			}
		})
	}
}

// MockCommandExecutor is already defined in git_commands_test.go

// Integration test for branch lifecycle
func TestBranchManager_Integration(t *testing.T) {
	// This test would require a real Git repository
	// For now, we'll test the integration with mock data
	executor := &MockCommandExecutor{}
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	bm := NewBranchManager(commands, validator, DefaultBranchNamingConfig())

	// Test branch name generation
	branchName, err := bm.GenerateBranchName("123", "Test task")
	if err != nil {
		t.Errorf("Failed to generate branch name: %v", err)
	}

	if branchName != "task-123-test-task" {
		t.Errorf("Expected 'task-123-test-task', got '%s'", branchName)
	}

	// Test validation
	validationResult, err := bm.ValidateBranchNameFormat(branchName)
	if err != nil {
		t.Errorf("Failed to validate branch name: %v", err)
	}

	if !validationResult.IsValid {
		t.Errorf("Expected valid branch name, got invalid: %v", validationResult.Issues)
	}
}

// Benchmark tests
func BenchmarkBranchManager_GenerateBranchName(b *testing.B) {
	executor := &MockCommandExecutor{}
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	bm := NewBranchManager(commands, validator, DefaultBranchNamingConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bm.GenerateBranchName("123", "Implement user authentication feature")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBranchManager_slugifyTitle(b *testing.B) {
	executor := &MockCommandExecutor{}
	commands := NewGitCommands(executor)
	validator := NewGitValidator(commands)
	bm := NewBranchManager(commands, validator, DefaultBranchNamingConfig())

	title := "Implement user authentication feature with OAuth2 support"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.slugifyTitle(title)
	}
}
