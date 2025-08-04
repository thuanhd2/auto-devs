package git

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitError(t *testing.T) {
	tests := []struct {
		name       string
		gitError   *GitError
		expectMsg  string
		underlying error
	}{
		{
			name: "with underlying error",
			gitError: &GitError{
				Operation:  "clone",
				ExitCode:   1,
				Command:    "git clone",
				Underlying: errors.New("network error"),
			},
			expectMsg:  "git clone failed: network error",
			underlying: errors.New("network error"),
		},
		{
			name: "without underlying error",
			gitError: &GitError{
				Operation: "push",
				ExitCode:  128,
				Command:   "git push",
			},
			expectMsg:  "git push failed with exit code 128",
			underlying: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectMsg, tt.gitError.Error())
			assert.Equal(t, tt.underlying, tt.gitError.Unwrap())
		})
	}
}

func TestNewGitError(t *testing.T) {
	underlying := errors.New("test error")
	gitErr := NewGitError("test", 1, "git test", "stdout", "stderr", underlying)

	assert.Equal(t, "test", gitErr.Operation)
	assert.Equal(t, 1, gitErr.ExitCode)
	assert.Equal(t, "git test", gitErr.Command)
	assert.Equal(t, "stdout", gitErr.Stdout)
	assert.Equal(t, "stderr", gitErr.Stderr)
	assert.Equal(t, underlying, gitErr.Underlying)
	assert.NotEmpty(t, gitErr.Suggestions)
}

func TestGetSuggestionsForError(t *testing.T) {
	tests := []struct {
		name     string
		gitError *GitError
		contains string
	}{
		{
			name: "not a git repository error",
			gitError: &GitError{
				ExitCode: 128,
				Stderr:   "fatal: not a git repository",
			},
			contains: "repository directory",
		},
		{
			name: "remote repository error",
			gitError: &GitError{
				ExitCode: 128,
				Stderr:   "fatal: remote origin does not exist",
			},
			contains: "remote repository URL",
		},
		{
			name: "merge conflict error",
			gitError: &GitError{
				ExitCode: 1,
				Stderr:   "error: merge conflict in file.txt",
			},
			contains: "merge conflicts",
		},
		{
			name: "branch exists error",
			gitError: &GitError{
				ExitCode: 1,
				Stderr:   "fatal: branch 'main' already exists",
			},
			contains: "different branch name",
		},
		{
			name: "authentication error",
			gitError: &GitError{
				ExitCode: 128,
				Stderr:   "fatal: permission denied",
			},
			contains: "credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := getSuggestionsForError(tt.gitError)
			assert.NotEmpty(t, suggestions)

			found := false
			for _, suggestion := range suggestions {
				if assert.Contains(t, suggestion, tt.contains) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected suggestions to contain '%s', got %v", tt.contains, suggestions)
			}
		})
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		substrings  []string
		expected    bool
	}{
		{
			name:       "contains match",
			text:       "This is a test error message",
			substrings: []string{"test", "example"},
			expected:   true,
		},
		{
			name:       "case insensitive match",
			text:       "This is a TEST error message",
			substrings: []string{"test", "example"},
			expected:   true,
		},
		{
			name:       "no match",
			text:       "This is an error message",
			substrings: []string{"success", "complete"},
			expected:   false,
		},
		{
			name:       "empty substrings",
			text:       "This is a test",
			substrings: []string{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsAny(tt.text, tt.substrings)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAuthenticationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "direct authentication error",
			err:      ErrAuthenticationFailed,
			expected: true,
		},
		{
			name:     "SSH key error",
			err:      ErrNoSSHKey,
			expected: true,
		},
		{
			name: "git error with auth failure",
			err: &GitError{
				Stderr: "fatal: authentication failed for repository",
			},
			expected: true,
		},
		{
			name: "git error with permission denied",
			err: &GitError{
				Stderr: "fatal: permission denied",
			},
			expected: true,
		},
		{
			name:     "non-authentication error",
			err:      ErrBranchNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAuthenticationError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsRepositoryError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "not git repository",
			err:      ErrNotGitRepository,
			expected: true,
		},
		{
			name:     "repository not found",
			err:      ErrRepositoryNotFound,
			expected: true,
		},
		{
			name:     "invalid repository URL",
			err:      ErrInvalidRepositoryURL,
			expected: true,
		},
		{
			name:     "non-repository error",
			err:      ErrBranchNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRepositoryError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsBranchError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "branch not found",
			err:      ErrBranchNotFound,
			expected: true,
		},
		{
			name:     "branch already exists",
			err:      ErrBranchAlreadyExists,
			expected: true,
		},
		{
			name:     "invalid branch name",
			err:      ErrInvalidBranchName,
			expected: true,
		},
		{
			name:     "non-branch error",
			err:      ErrNotGitRepository,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBranchError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsWorkingDirError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "working dir dirty",
			err:      ErrWorkingDirDirty,
			expected: true,
		},
		{
			name:     "merge conflicts",
			err:      ErrMergeConflicts,
			expected: true,
		},
		{
			name:     "unstaged changes",
			err:      ErrUnstagedChanges,
			expected: true,
		},
		{
			name:     "non-working dir error",
			err:      ErrBranchNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWorkingDirError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrapWithOperation(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		err       error
		expected  string
	}{
		{
			name:      "with error",
			operation: "clone",
			err:       errors.New("failed"),
			expected:  "git clone: failed",
		},
		{
			name:      "with nil error",
			operation: "push",
			err:       nil,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapWithOperation(tt.operation, tt.err)
			if tt.err == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result.Error())
			}
		})
	}
}