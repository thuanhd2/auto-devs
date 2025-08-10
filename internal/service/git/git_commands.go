package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CommandExecutor defines the interface for executing Git commands
type CommandExecutor interface {
	Execute(ctx context.Context, workingDir string, args ...string) (*CommandResult, error)
	ExecuteWithTimeout(ctx context.Context, workingDir string, timeout time.Duration, args ...string) (*CommandResult, error)
}

// CommandResult represents the result of a Git command execution
type CommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Command  string
}

// DefaultCommandExecutor implements CommandExecutor using os/exec
type DefaultCommandExecutor struct {
	gitPath        string
	defaultTimeout time.Duration
}

// NewDefaultCommandExecutor creates a new DefaultCommandExecutor
func NewDefaultCommandExecutor() (*DefaultCommandExecutor, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, ErrGitNotInstalled
	}

	return &DefaultCommandExecutor{
		gitPath:        gitPath,
		defaultTimeout: 30 * time.Second,
	}, nil
}

// Execute runs a Git command in the specified working directory
func (e *DefaultCommandExecutor) Execute(ctx context.Context, workingDir string, args ...string) (*CommandResult, error) {
	return e.ExecuteWithTimeout(ctx, workingDir, e.defaultTimeout, args...)
}

// ExecuteWithTimeout runs a Git command with a specific timeout
func (e *DefaultCommandExecutor) ExecuteWithTimeout(ctx context.Context, workingDir string, timeout time.Duration, args ...string) (*CommandResult, error) {
	// Create context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build the full command
	cmd := exec.CommandContext(cmdCtx, e.gitPath, args...)
	cmd.Dir = workingDir

	// Set environment variables for Git
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0", // Disable interactive prompts
		"GIT_ASKPASS=echo",      // Prevent password prompts
	)

	// Execute the command
	stdout, stderr, err := e.executeCommand(cmd)

	result := &CommandResult{
		Command:  fmt.Sprintf("git %s", strings.Join(args, " ")),
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: 0,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			// Handle context cancellation or timeout
			if cmdCtx.Err() == context.DeadlineExceeded {
				return result, ErrCommandTimeout
			}
			if cmdCtx.Err() == context.Canceled {
				return result, ErrCommandCancelled
			}
			// Other execution errors
			return result, fmt.Errorf("failed to execute git command: %w", err)
		}
	}

	return result, nil
}

// executeCommand executes the command and captures stdout/stderr
func (e *DefaultCommandExecutor) executeCommand(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// GitCommands provides high-level Git command wrappers
type GitCommands struct {
	executor CommandExecutor
}

// NewGitCommands creates a new GitCommands instance
func NewGitCommands(executor CommandExecutor) *GitCommands {
	return &GitCommands{executor: executor}
}

// Version returns the Git version
func (g *GitCommands) Version(ctx context.Context) (string, error) {
	result, err := g.executor.Execute(ctx, "", "--version")
	if err != nil {
		return "", WrapWithOperation("version", err)
	}

	if result.ExitCode != 0 {
		return "", NewGitError("version", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return strings.TrimPrefix(strings.TrimSpace(result.Stdout), "git version "), nil
}

// Init initializes a new Git repository
func (g *GitCommands) Init(ctx context.Context, workingDir string, bare bool) error {
	args := []string{"init"}
	if bare {
		args = append(args, "--bare")
	}

	result, err := g.executor.Execute(ctx, workingDir, args...)
	if err != nil {
		return WrapWithOperation("init", err)
	}

	if result.ExitCode != 0 {
		return NewGitError("init", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return nil
}

// Clone clones a repository
func (g *GitCommands) Clone(ctx context.Context, url, destination string, options *CloneOptions) error {
	args := []string{"clone"}

	if options != nil {
		if options.Branch != "" {
			args = append(args, "--branch", options.Branch)
		}
		if options.Depth > 0 {
			args = append(args, "--depth", fmt.Sprintf("%d", options.Depth))
		}
		if options.SingleBranch {
			args = append(args, "--single-branch")
		}
		if options.NoCheckout {
			args = append(args, "--no-checkout")
		}
	}

	args = append(args, url, destination)

	// Use longer timeout for clone operations
	result, err := g.executor.ExecuteWithTimeout(ctx, "", 5*time.Minute, args...)
	if err != nil {
		return WrapWithOperation("clone", err)
	}

	if result.ExitCode != 0 {
		return NewGitError("clone", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return nil
}

// Status returns the repository status
func (g *GitCommands) Status(ctx context.Context, workingDir string, porcelain bool) (string, error) {
	args := []string{"status"}
	if porcelain {
		args = append(args, "--porcelain")
	}

	result, err := g.executor.Execute(ctx, workingDir, args...)
	if err != nil {
		return "", WrapWithOperation("status", err)
	}

	if result.ExitCode != 0 {
		return "", NewGitError("status", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return result.Stdout, nil
}

// CurrentBranch returns the current branch name
func (g *GitCommands) CurrentBranch(ctx context.Context, workingDir string) (string, error) {
	result, err := g.executor.Execute(ctx, workingDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", WrapWithOperation("current-branch", err)
	}

	if result.ExitCode != 0 {
		return "", NewGitError("current-branch", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	branch := strings.TrimSpace(result.Stdout)
	if branch == "HEAD" {
		// We're in a detached HEAD state
		return "", fmt.Errorf("repository is in detached HEAD state")
	}

	return branch, nil
}

// ListBranches returns a list of branches
func (g *GitCommands) ListBranches(ctx context.Context, workingDir string, options *ListBranchesOptions) ([]string, error) {
	args := []string{"branch"}

	if options != nil {
		if options.Remote {
			args = append(args, "--remote")
		}
		if options.All {
			args = append(args, "--all")
		}
		if options.Merged != "" {
			args = append(args, "--merged", options.Merged)
		}
		if options.NoMerged != "" {
			args = append(args, "--no-merged", options.NoMerged)
		}
	}

	result, err := g.executor.Execute(ctx, workingDir, args...)
	if err != nil {
		return nil, WrapWithOperation("list-branches", err)
	}

	if result.ExitCode != 0 {
		return nil, NewGitError("list-branches", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	// Parse branch output
	branches := []string{}
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remove the current branch marker (*) and remote prefix
		branch := line
		branch = strings.TrimSpace(branch)
		branches = append(branches, branch)
	}

	return branches, nil
}

// CreateBranch creates a new branch
func (g *GitCommands) CreateBranch(ctx context.Context, workingDir, branchName, startPoint string) error {
	args := []string{"branch", branchName}
	if startPoint != "" {
		args = append(args, startPoint)
	}

	result, err := g.executor.Execute(ctx, workingDir, args...)
	if err != nil {
		return WrapWithOperation("create-branch", err)
	}

	if result.ExitCode != 0 {
		return NewGitError("create-branch", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return nil
}

// Checkout switches to a branch or commit
func (g *GitCommands) Checkout(ctx context.Context, workingDir, target string, createBranch bool) error {
	args := []string{"checkout"}
	if createBranch {
		args = append(args, "-b")
	}
	args = append(args, target)

	result, err := g.executor.Execute(ctx, workingDir, args...)
	if err != nil {
		return WrapWithOperation("checkout", err)
	}

	if result.ExitCode != 0 {
		return NewGitError("checkout", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return nil
}

// IsRepository checks if a directory is a Git repository
func (g *GitCommands) IsRepository(ctx context.Context, path string) (bool, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, fmt.Errorf("invalid path: %w", err)
	}

	result, err := g.executor.Execute(ctx, absPath, "rev-parse", "--git-dir")
	if err != nil {
		return false, nil // Not a git repository
	}

	return result.ExitCode == 0, nil
}

// GetRemoteURL returns the URL of the specified remote
func (g *GitCommands) GetRemoteURL(ctx context.Context, workingDir, remoteName string) (string, error) {
	result, err := g.executor.Execute(ctx, workingDir, "remote", "get-url", remoteName)
	if err != nil {
		return "", WrapWithOperation("get-remote-url", err)
	}

	if result.ExitCode != 0 {
		return "", NewGitError("get-remote-url", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return strings.TrimSpace(result.Stdout), nil
}

// GetCommitInfo returns information about a commit
func (g *GitCommands) GetCommitInfo(ctx context.Context, workingDir, commitish string) (*CommitInfo, error) {
	// Format: hash|author|date|subject
	format := "--pretty=format:%H|%an|%ai|%s"
	args := []string{"show", format, "--no-patch"}
	if commitish != "" {
		args = append(args, commitish)
	}

	result, err := g.executor.Execute(ctx, workingDir, args...)
	if err != nil {
		return nil, WrapWithOperation("get-commit-info", err)
	}

	if result.ExitCode != 0 {
		return nil, NewGitError("get-commit-info", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	// Parse the output
	parts := strings.Split(strings.TrimSpace(result.Stdout), "|")
	if len(parts) < 4 {
		return nil, fmt.Errorf("unexpected git show output format")
	}

	commitTime, err := time.Parse("2006-01-02 15:04:05 -0700", parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse commit date: %w", err)
	}

	return &CommitInfo{
		Hash:    parts[0],
		Author:  parts[1],
		Date:    commitTime,
		Subject: parts[3],
	}, nil
}

// Option types for Git commands

// CloneOptions represents options for git clone
type CloneOptions struct {
	Branch       string
	Depth        int
	SingleBranch bool
	NoCheckout   bool
}

// ListBranchesOptions represents options for listing branches
type ListBranchesOptions struct {
	Remote   bool
	All      bool
	Merged   string
	NoMerged string
}

// CommitInfo represents information about a Git commit
type CommitInfo struct {
	Hash    string
	Author  string
	Date    time.Time
	Subject string
}

// CreateWorktree creates a new worktree
// run command git worktree add -b <worktree-branch-name> <worktree-path> <base-branch-name>

func (g *GitCommands) CreateWorktree(ctx context.Context, workingDir, baseBranchName, worktreeBranchName, worktreePath string) error {
	args := []string{"worktree", "add", "-b", worktreeBranchName, worktreePath, baseBranchName}
	result, err := g.executor.Execute(ctx, workingDir, args...)
	if err != nil {
		return WrapWithOperation("create-worktree", err)
	}

	if result.ExitCode != 0 {
		return NewGitError("create-worktree", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return nil
}

// DeleteWorktree deletes a worktree
// run command git worktree remove --force <worktree-path>
func (g *GitCommands) DeleteWorktree(ctx context.Context, workingDir, worktreePath string) error {
	args := []string{"worktree", "remove", "--force", worktreePath}
	result, err := g.executor.Execute(ctx, workingDir, args...)
	if err != nil {
		return WrapWithOperation("delete-worktree", err)
	}

	if result.ExitCode != 0 {
		return NewGitError("delete-worktree", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return nil
}

// AddAllChanges stages all changes in the working directory
func (g *GitCommands) AddAllChanges(ctx context.Context, workingDir string) error {
	result, err := g.executor.Execute(ctx, workingDir, "add", ".")
	if err != nil {
		return WrapWithOperation("add-all", err)
	}

	if result.ExitCode != 0 {
		return NewGitError("add-all", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return nil
}

// Commit creates a commit with the given message
func (g *GitCommands) Commit(ctx context.Context, workingDir, message string) error {
	result, err := g.executor.Execute(ctx, workingDir, "commit", "-m", message)
	if err != nil {
		return WrapWithOperation("commit", err)
	}

	if result.ExitCode != 0 {
		return NewGitError("commit", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return nil
}

// Push pushes commits to remote repository
func (g *GitCommands) Push(ctx context.Context, workingDir, remote, branch string) error {
	args := []string{"push"}
	if remote != "" && branch != "" {
		args = append(args, remote, branch)
	}

	result, err := g.executor.Execute(ctx, workingDir, args...)
	if err != nil {
		return WrapWithOperation("push", err)
	}

	if result.ExitCode != 0 {
		return NewGitError("push", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return nil
}

// PushWithUpstream pushes commits and sets upstream tracking
func (g *GitCommands) PushWithUpstream(ctx context.Context, workingDir, remote, branch string) error {
	args := []string{"push", "--set-upstream", remote, branch}
	
	result, err := g.executor.Execute(ctx, workingDir, args...)
	if err != nil {
		return WrapWithOperation("push-upstream", err)
	}

	if result.ExitCode != 0 {
		return NewGitError("push-upstream", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	return nil
}

// GetPendingChanges checks if there are uncommitted changes
func (g *GitCommands) GetPendingChanges(ctx context.Context, workingDir string) (bool, error) {
	result, err := g.executor.Execute(ctx, workingDir, "status", "--porcelain")
	if err != nil {
		return false, WrapWithOperation("get-pending-changes", err)
	}

	if result.ExitCode != 0 {
		return false, NewGitError("get-pending-changes", result.ExitCode, result.Command, result.Stdout, result.Stderr, nil)
	}

	// If there's any output, there are pending changes
	return strings.TrimSpace(result.Stdout) != "", nil
}
