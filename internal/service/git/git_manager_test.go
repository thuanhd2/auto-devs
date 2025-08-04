package git

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test helper to create a test logger that doesn't output during tests
func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestNewGitManager(t *testing.T) {
	tests := []struct {
		name     string
		config   *ManagerConfig
		expectError bool
	}{
		{
			name:        "with nil config",
			config:      nil,
			expectError: false,
		},
		{
			name: "with custom config",
			config: &ManagerConfig{
				DefaultTimeout: 60 * time.Second,
				MaxRetries:     5,
				EnableLogging:  false,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewGitManager(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, manager)
			} else {
				// This might fail if git is not installed
				if err == ErrGitNotInstalled {
					t.Skip("Git not installed, skipping test")
				}
				assert.NoError(t, err)
				assert.NotNil(t, manager)
				assert.NotNil(t, manager.commands)
				assert.NotNil(t, manager.validator)
				assert.NotNil(t, manager.config)
			}
		})
	}
}

func TestGitManager_ValidateRepository(t *testing.T) {
	// Create a manager with mock components for testing
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)
	validator := NewGitValidator(commands)
	
	manager := &GitManager{
		commands:  commands,
		validator: validator,
		logger:    createTestLogger(),
		config: &ManagerConfig{
			DefaultTimeout: 30 * time.Second,
			MaxRetries:     3,
		},
	}

	tests := []struct {
		name           string
		repoPath       string
		setupMocks     func()
		expectedError  bool
		expectedInfo   *RepositoryInfo
	}{
		{
			name:     "valid repository",
			repoPath: ".",
			setupMocks: func() {
				// Mock IsRepository check
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"rev-parse", "--git-dir"}).
					Return(&CommandResult{ExitCode: 0, Stdout: ".git\n"}, nil).Once()
				
				// Mock CurrentBranch
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"rev-parse", "--abbrev-ref", "HEAD"}).
					Return(&CommandResult{ExitCode: 0, Stdout: "main\n"}, nil).Once()
				
				// Mock GetRemoteURL
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"remote", "get-url", "origin"}).
					Return(&CommandResult{ExitCode: 0, Stdout: "https://github.com/user/repo.git\n"}, nil).Once()
				
				// Mock Status
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"status", "--porcelain"}).
					Return(&CommandResult{ExitCode: 0, Stdout: ""}, nil).Once()
				
				// Mock GetCommitInfo
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"show", "--pretty=format:%H|%an|%ai|%s", "--no-patch", "HEAD"}).
					Return(&CommandResult{ExitCode: 0, Stdout: "abc123|John Doe|2023-01-01 12:00:00 +0000|Initial commit"}, nil).Once()
			},
			expectedError: false,
			expectedInfo: &RepositoryInfo{
				CurrentBranch: "main",
				RemoteURL:     "https://github.com/user/repo.git",
				WorkingDirStatus: WorkingDirStatus{
					IsClean:            true,
					HasStagedChanges:   false,
					HasUnstagedChanges: false,
					HasUntrackedFiles:  false,
				},
			},
		},
		{
			name:     "not a git repository",
			repoPath: "/tmp",
			setupMocks: func() {
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"rev-parse", "--git-dir"}).
					Return(&CommandResult{ExitCode: 128, Stderr: "fatal: not a git repository"}, nil).Once()
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockExecutor.ExpectedCalls = nil
			mockExecutor.Calls = nil
			
			tt.setupMocks()

			info, err := manager.ValidateRepository(context.Background(), tt.repoPath)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, info)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				if tt.expectedInfo != nil {
					assert.Equal(t, tt.expectedInfo.CurrentBranch, info.CurrentBranch)
					assert.Equal(t, tt.expectedInfo.RemoteURL, info.RemoteURL)
					assert.Equal(t, tt.expectedInfo.WorkingDirStatus.IsClean, info.WorkingDirStatus.IsClean)
				}
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestGitManager_CreateBranch(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)
	validator := NewGitValidator(commands)
	
	manager := &GitManager{
		commands:  commands,
		validator: validator,
		logger:    createTestLogger(),
		config: &ManagerConfig{
			DefaultTimeout: 30 * time.Second,
			MaxRetries:     3,
			WorkingDir:     "/tmp/repo",
		},
	}

	tests := []struct {
		name         string
		request      *CreateBranchRequest
		setupMocks   func()
		expectedError bool
	}{
		{
			name: "successful branch creation",
			request: &CreateBranchRequest{
				BranchName: "feature-branch",
				StartPoint: "main",
				WorkingDir: "/tmp/repo",
			},
			setupMocks: func() {
				// Mock CheckBranchExists
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"branch", "--all"}).
					Return(&CommandResult{ExitCode: 0, Stdout: "  main\n  develop\n"}, nil).Once()
				
				// Mock CreateBranch
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"branch", "feature-branch", "main"}).
					Return(&CommandResult{ExitCode: 0}, nil).Once()
			},
			expectedError: false,
		},
		{
			name: "invalid branch name",
			request: &CreateBranchRequest{
				BranchName: ".invalid-branch",
				WorkingDir: "/tmp/repo",
			},
			setupMocks:   func() {},
			expectedError: true,
		},
		{
			name: "branch already exists",
			request: &CreateBranchRequest{
				BranchName: "existing-branch",
				WorkingDir: "/tmp/repo",
			},
			setupMocks: func() {
				// Mock CheckBranchExists - return true
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"branch", "--all"}).
					Return(&CommandResult{ExitCode: 0, Stdout: "  main\n  existing-branch\n"}, nil).Once()
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockExecutor.ExpectedCalls = nil
			mockExecutor.Calls = nil
			
			tt.setupMocks()

			err := manager.CreateBranch(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestGitManager_SwitchBranch(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)
	validator := NewGitValidator(commands)
	
	manager := &GitManager{
		commands:  commands,
		validator: validator,
		logger:    createTestLogger(),
		config: &ManagerConfig{
			DefaultTimeout: 30 * time.Second,
			MaxRetries:     3,
		},
	}

	tests := []struct {
		name         string
		request      *SwitchBranchRequest
		setupMocks   func()
		expectedError bool
	}{
		{
			name: "successful branch switch",
			request: &SwitchBranchRequest{
				BranchName: "develop",
				WorkingDir: "/tmp/repo",
			},
			setupMocks: func() {
				// Mock ValidateWorkingDirectory
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"status", "--porcelain"}).
					Return(&CommandResult{ExitCode: 0, Stdout: ""}, nil).Once()
				
				// Mock Checkout
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"checkout", "develop"}).
					Return(&CommandResult{ExitCode: 0}, nil).Once()
			},
			expectedError: false,
		},
		{
			name: "working directory dirty",
			request: &SwitchBranchRequest{
				BranchName: "develop",
				WorkingDir: "/tmp/repo",
			},
			setupMocks: func() {
				// Mock ValidateWorkingDirectory - return dirty state
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"status", "--porcelain"}).
					Return(&CommandResult{ExitCode: 0, Stdout: " M file.txt\n"}, nil).Once()
			},
			expectedError: true,
		},
		{
			name: "force switch with dirty working directory",
			request: &SwitchBranchRequest{
				BranchName: "develop",
				WorkingDir: "/tmp/repo",
				Force:      true,
			},
			setupMocks: func() {
				// Mock Checkout (skip working directory validation due to Force flag)
				mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"checkout", "develop"}).
					Return(&CommandResult{ExitCode: 0}, nil).Once()
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockExecutor.ExpectedCalls = nil
			mockExecutor.Calls = nil
			
			tt.setupMocks()

			err := manager.SwitchBranch(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestGitManager_executeWithRetry(t *testing.T) {
	manager := &GitManager{
		logger: createTestLogger(),
		config: &ManagerConfig{
			MaxRetries: 2,
		},
	}

	tests := []struct {
		name         string
		operation    func() error
		expectedError bool
		expectedCalls int
	}{
		{
			name: "successful on first try",
			operation: func() error {
				return nil
			},
			expectedError: false,
			expectedCalls: 1,
		},
		{
			name: "non-retryable error",
			operation: func() error {
				return ErrAuthenticationFailed
			},
			expectedError: true,
			expectedCalls: 1,
		},
		{
			name: "timeout error",
			operation: func() error {
				return ErrCommandTimeout
			},
			expectedError: true,
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			operation := func() error {
				callCount++
				return tt.operation()
			}

			err := manager.executeWithRetry(context.Background(), operation)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			assert.Equal(t, tt.expectedCalls, callCount)
		})
	}
}

func TestGitManager_shouldRetry(t *testing.T) {
	manager := &GitManager{}

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "authentication error - no retry",
			err:      ErrAuthenticationFailed,
			expected: false,
		},
		{
			name:     "branch error - no retry",
			err:      ErrBranchNotFound,
			expected: false,
		},
		{
			name:     "repository error - no retry",
			err:      ErrNotGitRepository,
			expected: false,
		},
		{
			name:     "timeout error - no retry",
			err:      ErrCommandTimeout,
			expected: false,
		},
		{
			name:     "cancelled error - no retry",
			err:      ErrCommandCancelled,
			expected: false,
		},
		{
			name:     "generic error - retry",
			err:      errors.New("network error"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.shouldRetry(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitManager_getWorkingDir(t *testing.T) {
	tests := []struct {
		name       string
		manager    *GitManager
		workingDir string
		expected   string
	}{
		{
			name: "provided working dir",
			manager: &GitManager{
				config: &ManagerConfig{WorkingDir: "/default"},
			},
			workingDir: "/custom",
			expected:   "/custom",
		},
		{
			name: "config working dir",
			manager: &GitManager{
				config: &ManagerConfig{WorkingDir: "/default"},
			},
			workingDir: "",
			expected:   "/default",
		},
		{
			name: "default working dir",
			manager: &GitManager{
				config: &ManagerConfig{},
			},
			workingDir: "",
			expected:   ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.manager.getWorkingDir(tt.workingDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}