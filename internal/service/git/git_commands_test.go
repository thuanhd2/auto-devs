package git

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCommandExecutor is a mock implementation of CommandExecutor
type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) Execute(ctx context.Context, workingDir string, args ...string) (*CommandResult, error) {
	arguments := m.Called(ctx, workingDir, args)
	return arguments.Get(0).(*CommandResult), arguments.Error(1)
}

func (m *MockCommandExecutor) ExecuteWithTimeout(ctx context.Context, workingDir string, timeout time.Duration, args ...string) (*CommandResult, error) {
	arguments := m.Called(ctx, workingDir, timeout, args)
	return arguments.Get(0).(*CommandResult), arguments.Error(1)
}

func TestNewDefaultCommandExecutor(t *testing.T) {
	executor, err := NewDefaultCommandExecutor()

	// This test will pass if git is installed, fail if not
	if err == ErrGitNotInstalled {
		t.Skip("Git not installed, skipping test")
	}

	assert.NoError(t, err)
	assert.NotNil(t, executor)
	assert.NotEmpty(t, executor.gitPath)
	assert.Equal(t, 30*time.Second, executor.defaultTimeout)
}

func TestDefaultCommandExecutor_Execute(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git not available, skipping integration test")
	}

	executor, err := NewDefaultCommandExecutor()
	assert.NoError(t, err)

	ctx := context.Background()
	result, err := executor.Execute(ctx, "", "--version")

	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "git version")
	assert.Equal(t, "git --version", result.Command)
}

func TestGitCommands_Version(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)

	tests := []struct {
		name            string
		result          *CommandResult
		executeError    error
		expectedError   bool
		expectedVersion string
	}{
		{
			name: "successful version check",
			result: &CommandResult{
				ExitCode: 0,
				Stdout:   "git version 2.34.1\n",
				Command:  "git --version",
			},
			executeError:    nil,
			expectedError:   false,
			expectedVersion: "2.34.1",
		},
		{
			name:          "executor error",
			result:        &CommandResult{},
			executeError:  errors.New("execution failed"),
			expectedError: true,
		},
		{
			name: "git command error",
			result: &CommandResult{
				ExitCode: 1,
				Stderr:   "command not found",
				Command:  "git --version",
			},
			executeError:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor.On("Execute", mock.Anything, "", []string{"--version"}).
				Return(tt.result, tt.executeError).Once()

			version, err := commands.Version(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVersion, version)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestGitCommands_Init(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)

	tests := []struct {
		name          string
		workingDir    string
		bare          bool
		result        *CommandResult
		executeError  error
		expectedError bool
		expectedArgs  []string
	}{
		{
			name:       "successful init",
			workingDir: "/tmp/repo",
			bare:       false,
			result: &CommandResult{
				ExitCode: 0,
				Stdout:   "Initialized empty Git repository",
			},
			executeError:  nil,
			expectedError: false,
			expectedArgs:  []string{"init"},
		},
		{
			name:       "successful bare init",
			workingDir: "/tmp/repo",
			bare:       true,
			result: &CommandResult{
				ExitCode: 0,
				Stdout:   "Initialized bare Git repository",
			},
			executeError:  nil,
			expectedError: false,
			expectedArgs:  []string{"init", "--bare"},
		},
		{
			name:       "init failure",
			workingDir: "/tmp/repo",
			bare:       false,
			result: &CommandResult{
				ExitCode: 1,
				Stderr:   "permission denied",
			},
			executeError:  nil,
			expectedError: true,
			expectedArgs:  []string{"init"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor.On("Execute", mock.Anything, tt.workingDir, tt.expectedArgs).
				Return(tt.result, tt.executeError).Once()

			err := commands.Init(context.Background(), tt.workingDir, tt.bare)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestGitCommands_Clone(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)

	tests := []struct {
		name          string
		url           string
		destination   string
		options       *CloneOptions
		result        *CommandResult
		executeError  error
		expectedError bool
		expectedArgs  []string
	}{
		{
			name:        "basic clone",
			url:         "https://github.com/user/repo.git",
			destination: "/tmp/repo",
			options:     nil,
			result: &CommandResult{
				ExitCode: 0,
				Stdout:   "Cloning into '/tmp/repo'...",
			},
			executeError:  nil,
			expectedError: false,
			expectedArgs:  []string{"clone", "https://github.com/user/repo.git", "/tmp/repo"},
		},
		{
			name:        "clone with options",
			url:         "https://github.com/user/repo.git",
			destination: "/tmp/repo",
			options: &CloneOptions{
				Branch:       "develop",
				Depth:        1,
				SingleBranch: true,
			},
			result: &CommandResult{
				ExitCode: 0,
				Stdout:   "Cloning into '/tmp/repo'...",
			},
			executeError:  nil,
			expectedError: false,
			expectedArgs:  []string{"clone", "--branch", "develop", "--depth", "1", "--single-branch", "https://github.com/user/repo.git", "/tmp/repo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor.On("ExecuteWithTimeout", mock.Anything, "", 5*time.Minute, tt.expectedArgs).
				Return(tt.result, tt.executeError).Once()

			err := commands.Clone(context.Background(), tt.url, tt.destination, tt.options)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestGitCommands_Status(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)

	tests := []struct {
		name           string
		workingDir     string
		porcelain      bool
		result         *CommandResult
		executeError   error
		expectedError  bool
		expectedOutput string
		expectedArgs   []string
	}{
		{
			name:       "normal status",
			workingDir: "/tmp/repo",
			porcelain:  false,
			result: &CommandResult{
				ExitCode: 0,
				Stdout:   "On branch main\nnothing to commit, working tree clean",
			},
			executeError:   nil,
			expectedError:  false,
			expectedOutput: "On branch main\nnothing to commit, working tree clean",
			expectedArgs:   []string{"status"},
		},
		{
			name:       "porcelain status",
			workingDir: "/tmp/repo",
			porcelain:  true,
			result: &CommandResult{
				ExitCode: 0,
				Stdout:   " M file.txt\n?? newfile.txt",
			},
			executeError:   nil,
			expectedError:  false,
			expectedOutput: " M file.txt\n?? newfile.txt",
			expectedArgs:   []string{"status", "--porcelain"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor.On("Execute", mock.Anything, tt.workingDir, tt.expectedArgs).
				Return(tt.result, tt.executeError).Once()

			output, err := commands.Status(context.Background(), tt.workingDir, tt.porcelain)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestGitCommands_CurrentBranch(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)

	tests := []struct {
		name           string
		result         *CommandResult
		executeError   error
		expectedError  bool
		expectedBranch string
	}{
		{
			name: "current branch",
			result: &CommandResult{
				ExitCode: 0,
				Stdout:   "main\n",
			},
			executeError:   nil,
			expectedError:  false,
			expectedBranch: "main",
		},
		{
			name: "detached HEAD",
			result: &CommandResult{
				ExitCode: 0,
				Stdout:   "HEAD\n",
			},
			executeError:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"rev-parse", "--abbrev-ref", "HEAD"}).
				Return(tt.result, tt.executeError).Once()

			branch, err := commands.CurrentBranch(context.Background(), "/tmp/repo")

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBranch, branch)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestGitCommands_ListBranches(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)

	tests := []struct {
		name             string
		options          *ListBranchesOptions
		result           *CommandResult
		executeError     error
		expectedError    bool
		expectedBranches []string
		expectedArgs     []string
	}{
		{
			name:    "local branches",
			options: nil,
			result: &CommandResult{
				ExitCode: 0,
				Stdout:   "  develop\n* main\n  feature-branch\n",
			},
			executeError:     nil,
			expectedError:    false,
			expectedBranches: []string{"develop", "* main", "feature-branch"},
			expectedArgs:     []string{"branch"},
		},
		{
			name: "remote branches",
			options: &ListBranchesOptions{
				Remote: true,
			},
			result: &CommandResult{
				ExitCode: 0,
				Stdout:   "  origin/main\n  origin/develop\n",
			},
			executeError:     nil,
			expectedError:    false,
			expectedBranches: []string{"origin/main", "origin/develop"},
			expectedArgs:     []string{"branch", "--remote"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor.On("Execute", mock.Anything, mock.Anything, tt.expectedArgs).
				Return(tt.result, tt.executeError).Once()

			branches, err := commands.ListBranches(context.Background(), "/tmp/repo", tt.options)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBranches, branches)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestGitCommands_IsRepository(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)

	tests := []struct {
		name           string
		path           string
		result         *CommandResult
		executeError   error
		expectedResult bool
		expectedError  bool
	}{
		{
			name: "valid repository",
			path: "/tmp/repo",
			result: &CommandResult{
				ExitCode: 0,
				Stdout:   ".git\n",
			},
			executeError:   nil,
			expectedResult: true,
			expectedError:  false,
		},
		{
			name: "not a repository",
			path: "/tmp/notrepo",
			result: &CommandResult{
				ExitCode: 128,
				Stderr:   "fatal: not a git repository",
			},
			executeError:   nil,
			expectedResult: false,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"rev-parse", "--git-dir"}).
				Return(tt.result, tt.executeError).Once()

			isRepo, err := commands.IsRepository(context.Background(), tt.path)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, isRepo)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}
