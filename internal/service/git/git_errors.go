package git

import (
	"errors"
	"fmt"
	"strings"
)

// Git operation errors
var (
	// General Git errors
	ErrGitNotInstalled       = errors.New("git is not installed or not available in PATH")
	ErrGitVersionUnsupported = errors.New("git version is not supported (minimum required: 2.20.0)")
	ErrInvalidWorkingDir     = errors.New("invalid working directory")
	ErrCommandTimeout        = errors.New("git command timed out")
	ErrCommandCancelled      = errors.New("git command was cancelled")

	// Repository errors
	ErrNotGitRepository     = errors.New("not a git repository")
	ErrRepositoryNotFound   = errors.New("repository not found")
	ErrRepositoryCorrupted  = errors.New("repository appears to be corrupted")
	ErrInvalidRepositoryURL = errors.New("invalid repository URL")
	ErrRemoteNotAccessible  = errors.New("remote repository is not accessible")

	// Authentication errors
	ErrAuthenticationFailed = errors.New("git authentication failed")
	ErrNoSSHKey             = errors.New("SSH key not found or invalid")
	ErrInvalidCredentials   = errors.New("invalid git credentials")

	// Branch and reference errors
	ErrBranchNotFound       = errors.New("branch not found")
	ErrBranchAlreadyExists  = errors.New("branch already exists")
	ErrInvalidBranchName    = errors.New("invalid branch name")
	ErrCannotDeleteBranch   = errors.New("cannot delete branch")
	ErrInvalidReference     = errors.New("invalid git reference")

	// Working directory errors
	ErrWorkingDirDirty      = errors.New("working directory has uncommitted changes")
	ErrMergeConflicts       = errors.New("merge conflicts detected")
	ErrUnstagedChanges      = errors.New("unstaged changes detected")
		
	// Configuration errors
	ErrGitConfigNotSet      = errors.New("git configuration not set")
	ErrInvalidGitConfig     = errors.New("invalid git configuration")
)

// GitError represents a structured Git operation error
type GitError struct {
	Operation   string // The Git operation that failed (e.g., "clone", "checkout", "push")
	ExitCode    int    // Git command exit code
	Command     string // The Git command that was executed
	Stderr      string // Standard error output from Git
	Stdout      string // Standard output from Git
	Underlying  error  // The underlying error
	Suggestions []string // Suggested solutions for the error
}

func (e *GitError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("git %s failed: %v", e.Operation, e.Underlying)
	}
	return fmt.Sprintf("git %s failed with exit code %d", e.Operation, e.ExitCode)
}

func (e *GitError) Unwrap() error {
	return e.Underlying
}

// NewGitError creates a new GitError with the provided details
func NewGitError(operation string, exitCode int, command, stdout, stderr string, underlying error) *GitError {
	gitErr := &GitError{
		Operation:  operation,
		ExitCode:   exitCode,
		Command:    command,
		Stderr:     stderr,
		Stdout:     stdout,
		Underlying: underlying,
	}

	// Add contextual suggestions based on the error
	gitErr.Suggestions = getSuggestionsForError(gitErr)
	
	return gitErr
}

// getSuggestionsForError provides helpful suggestions based on the error type
func getSuggestionsForError(err *GitError) []string {
	// Check authentication errors first
	if containsAny(err.Stderr, []string{"permission denied", "authentication failed"}) {
		return []string{
			"Check your Git credentials",
			"Ensure SSH keys are properly configured",
			"Verify repository access permissions",
		}
	}
	
	switch {
	case err.ExitCode == 128:
		if containsAny(err.Stderr, []string{"not a git repository", "not found"}) {
			return []string{
				"Ensure you are in a valid Git repository directory",
				"Run 'git init' to initialize a new repository",
				"Check if the repository path is correct",
			}
		}
		if containsAny(err.Stderr, []string{"remote", "origin"}) {
			return []string{
				"Check if the remote repository URL is correct",
				"Verify network connectivity to the remote repository",
				"Ensure you have proper access permissions to the repository",
			}
		}
	case err.ExitCode == 1:
		if containsAny(err.Stderr, []string{"merge conflict", "conflict"}) {
			return []string{
				"Resolve merge conflicts manually",
				"Use 'git status' to see conflicted files",
				"After resolving conflicts, run 'git add' and 'git commit'",
			}
		}
		if containsAny(err.Stderr, []string{"branch", "already exists"}) {
			return []string{
				"Use a different branch name",
				"Delete the existing branch first if appropriate",
				"Switch to the existing branch instead",
			}
		}
	}

	return []string{
		"Check Git command syntax and parameters",
		"Ensure Git is properly installed and configured",
		"Review Git status and repository state",
	}
}

// containsAny checks if the text contains any of the provided substrings (case-insensitive)
func containsAny(text string, substrings []string) bool {
	textLower := strings.ToLower(text)
	for _, substr := range substrings {
		if strings.Contains(textLower, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// Common error classification functions

// IsAuthenticationError checks if the error is related to authentication
func IsAuthenticationError(err error) bool {
	var gitErr *GitError
	if errors.As(err, &gitErr) {
		return containsAny(gitErr.Stderr, []string{
			"authentication failed",
			"permission denied",
			"access denied",
			"unauthorized",
			"could not read username",
			"could not read password",
		})
	}
	return errors.Is(err, ErrAuthenticationFailed) ||
		errors.Is(err, ErrNoSSHKey) ||
		errors.Is(err, ErrInvalidCredentials)
}

// IsRepositoryError checks if the error is related to repository state
func IsRepositoryError(err error) bool {
	return errors.Is(err, ErrNotGitRepository) ||
		errors.Is(err, ErrRepositoryNotFound) ||
		errors.Is(err, ErrRepositoryCorrupted) ||
		errors.Is(err, ErrInvalidRepositoryURL)
}

// IsBranchError checks if the error is related to branch operations
func IsBranchError(err error) bool {
	return errors.Is(err, ErrBranchNotFound) ||
		errors.Is(err, ErrBranchAlreadyExists) ||
		errors.Is(err, ErrInvalidBranchName) ||
		errors.Is(err, ErrCannotDeleteBranch)
}

// IsWorkingDirError checks if the error is related to working directory state
func IsWorkingDirError(err error) bool {
	return errors.Is(err, ErrWorkingDirDirty) ||
		errors.Is(err, ErrMergeConflicts) ||
		errors.Is(err, ErrUnstagedChanges)
}

// WrapWithOperation wraps an error with Git operation context
func WrapWithOperation(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("git %s: %w", operation, err)
}