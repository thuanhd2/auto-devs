package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGitValidator_ValidateGitInstallation(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)
	validator := NewGitValidator(commands)

	tests := []struct {
		name         string
		version      string
		executeError error
		expectedError bool
	}{
		{
			name:          "supported version",
			version:       "2.34.1",
			executeError:  nil,
			expectedError: false,
		},
		{
			name:          "minimum supported version",
			version:       "2.20.0",
			executeError:  nil,
			expectedError: false,
		},
		{
			name:          "unsupported version",
			version:       "2.19.0",
			executeError:  nil,
			expectedError: true,
		},
		{
			name:          "git not available",
			version:       "",
			executeError:  ErrGitNotInstalled,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &CommandResult{
				ExitCode: 0,
				Stdout:   "git version " + tt.version + "\n",
			}
			if tt.executeError != nil {
				result.ExitCode = 1
			}

			mockExecutor.On("Execute", mock.Anything, "", []string{"--version"}).
				Return(result, tt.executeError).Once()

			err := validator.ValidateGitInstallation(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestGitValidator_ValidateBranchName(t *testing.T) {
	validator := &GitValidator{}

	tests := []struct {
		name       string
		branchName string
		expectError bool
	}{
		{
			name:        "valid branch name",
			branchName:  "feature-branch",
			expectError: false,
		},
		{
			name:        "valid branch with slashes",
			branchName:  "feature/user-management",
			expectError: false,
		},
		{
			name:        "empty branch name",
			branchName:  "",
			expectError: true,
		},
		{
			name:        "starts with dot",
			branchName:  ".invalid",
			expectError: true,
		},
		{
			name:        "ends with .lock",
			branchName:  "branch.lock",
			expectError: true,
		},
		{
			name:        "contains @{",
			branchName:  "branch@{upstream}",
			expectError: true,
		},
		{
			name:        "contains colon",
			branchName:  "branch:invalid",
			expectError: true,
		},
		{
			name:        "contains space",
			branchName:  "branch name",
			expectError: true,
		},
		{
			name:        "contains consecutive dots",
			branchName:  "branch..name",
			expectError: true,
		},
		{
			name:        "too long",
			branchName:  string(make([]byte, 256)),
			expectError: true,
		},
		{
			name:        "starts with slash",
			branchName:  "/invalid",
			expectError: true,
		},
		{
			name:        "ends with slash",
			branchName:  "invalid/",
			expectError: true,
		},
		{
			name:        "double slash",
			branchName:  "branch//invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateBranchName(tt.branchName)
			if tt.expectError {
				assert.Error(t, err)
				assert.True(t, IsBranchError(err) || err.Error() != "")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitValidator_ValidateRepositoryURL(t *testing.T) {
	validator := &GitValidator{}

	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "valid HTTPS URL",
			url:         "https://github.com/user/repo.git",
			expectError: false,
		},
		{
			name:        "valid SSH URL",
			url:         "ssh://git@github.com/user/repo.git",
			expectError: false,
		},
		{
			name:        "valid git protocol URL",
			url:         "git://github.com/user/repo.git",
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
		},
		{
			name:        "invalid URL format",
			url:         "not-a-url",
			expectError: true,
		},
		{
			name:        "unsupported scheme",
			url:         "ftp://example.com/repo.git",
			expectError: true,
		},
		{
			name:        "HTTPS without hostname",
			url:         "https:///repo.git",
			expectError: true,
		},
		{
			name:        "SSH without username",
			url:         "ssh://github.com/user/repo.git",
			expectError: true,
		},
		{
			name:        "HTTP URL from known host without .git",
			url:         "https://github.com/user/repo",
			expectError: false,
		},
		{
			name:        "HTTP URL from unknown host without .git",
			url:         "https://unknown.com/user/repo",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRepositoryURL(context.Background(), tt.url)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitValidator_isVersionSupported(t *testing.T) {
	validator := &GitValidator{}

	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "supported version 2.34.1",
			version:  "2.34.1",
			expected: true,
		},
		{
			name:     "minimum version 2.20.0",
			version:  "2.20.0",
			expected: true,
		},
		{
			name:     "unsupported version 2.19.9",
			version:  "2.19.9",
			expected: false,
		},
		{
			name:     "major version 3.0.0",
			version:  "3.0.0",
			expected: true,
		},
		{
			name:     "version 2.21.0",
			version:  "2.21.0",
			expected: true,
		},
		{
			name:     "invalid version format",
			version:  "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isVersionSupported(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitValidator_isSupportedScheme(t *testing.T) {
	validator := &GitValidator{}

	tests := []struct {
		name     string
		scheme   string
		expected bool
	}{
		{
			name:     "https scheme",
			scheme:   "https",
			expected: true,
		},
		{
			name:     "http scheme",
			scheme:   "http",
			expected: true,
		},
		{
			name:     "ssh scheme",
			scheme:   "ssh",
			expected: true,
		},
		{
			name:     "git scheme",
			scheme:   "git",
			expected: true,
		},
		{
			name:     "unsupported ftp scheme",
			scheme:   "ftp",
			expected: false,
		},
		{
			name:     "unsupported file scheme",
			scheme:   "file",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isSupportedScheme(tt.scheme)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitValidator_isKnownGitHost(t *testing.T) {
	validator := &GitValidator{}

	tests := []struct {
		name     string
		host     string
		expected bool
	}{
		{
			name:     "github.com",
			host:     "github.com",
			expected: true,
		},
		{
			name:     "gitlab.com",
			host:     "gitlab.com",
			expected: true,
		},
		{
			name:     "bitbucket.org",
			host:     "bitbucket.org",
			expected: true,
		},
		{
			name:     "codeberg.org",
			host:     "codeberg.org",
			expected: true,
		},
		{
			name:     "git.sr.ht",
			host:     "git.sr.ht",
			expected: true,
		},
		{
			name:     "enterprise github",
			host:     "github.company.com",
			expected: true,
		},
		{
			name:     "unknown host",
			host:     "unknown.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isKnownGitHost(tt.host)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitValidator_isValidEmail(t *testing.T) {
	validator := &GitValidator{}

	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "valid email",
			email:    "user@example.com",
			expected: true,
		},
		{
			name:     "valid email with subdomain",
			email:    "user@mail.example.com",
			expected: true,
		},
		{
			name:     "valid email with plus",
			email:    "user+tag@example.com",
			expected: true,
		},
		{
			name:     "invalid email without @",
			email:    "userexample.com",
			expected: false,
		},
		{
			name:     "invalid email without domain",
			email:    "user@",
			expected: false,
		},
		{
			name:     "invalid email without tld",
			email:    "user@example",
			expected: false,
		},
		{
			name:     "empty email",
			email:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitValidator_ValidateWorkingDirectory(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)
	validator := NewGitValidator(commands)

	tests := []struct {
		name                   string
		statusOutput           string
		expectedStatus         *WorkingDirStatus
		expectedError          bool
	}{
		{
			name:         "clean working directory",
			statusOutput: "",
			expectedStatus: &WorkingDirStatus{
				IsClean:            true,
				HasStagedChanges:   false,
				HasUnstagedChanges: false,
				HasUntrackedFiles:  false,
			},
			expectedError: false,
		},
		{
			name:         "working directory with changes",
			statusOutput: "AM file1.txt\nA  file2.txt\n?? file3.txt",
			expectedStatus: &WorkingDirStatus{
				IsClean:            false,
				HasStagedChanges:   true,
				HasUnstagedChanges: true,
				HasUntrackedFiles:  true,
			},
			expectedError: false,
		},
		{
			name:         "only untracked files",
			statusOutput: "?? newfile.txt\n?? another.txt",
			expectedStatus: &WorkingDirStatus{
				IsClean:            false,
				HasStagedChanges:   false,
				HasUnstagedChanges: false,
				HasUntrackedFiles:  true,
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"status", "--porcelain"}).
				Return(&CommandResult{
					ExitCode: 0,
					Stdout:   tt.statusOutput,
				}, nil).Once()

			status, err := validator.ValidateWorkingDirectory(context.Background(), "/tmp/repo")

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestGitValidator_CheckBranchExists(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	commands := NewGitCommands(mockExecutor)
	validator := NewGitValidator(commands)

	tests := []struct {
		name         string
		branchName   string
		branchOutput string
		expectedResult bool
		expectedError  bool
	}{
		{
			name:         "branch exists",
			branchName:   "feature-branch",
			branchOutput: "  develop\n* main\n  feature-branch\n  origin/main\n  origin/develop",
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:         "branch does not exist",
			branchName:   "nonexistent",
			branchOutput: "  develop\n* main\n  feature-branch",
			expectedResult: false,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor.On("Execute", mock.Anything, mock.Anything, []string{"branch", "--all"}).
				Return(&CommandResult{
					ExitCode: 0,
					Stdout:   tt.branchOutput,
				}, nil).Once()

			exists, err := validator.CheckBranchExists(context.Background(), "/tmp/repo", tt.branchName)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, exists)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}