package git

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// GitValidator provides validation functionality for Git operations
type GitValidator struct {
	commands *GitCommands
}

// NewGitValidator creates a new GitValidator instance
func NewGitValidator(commands *GitCommands) *GitValidator {
	return &GitValidator{commands: commands}
}

// ValidateGitInstallation checks if Git is properly installed and configured
func (v *GitValidator) ValidateGitInstallation(ctx context.Context) error {
	// Check Git version
	version, err := v.commands.Version(ctx)
	if err != nil {
		return fmt.Errorf("git installation check failed: %w", err)
	}

	// Validate minimum version (2.20.0)
	if !v.isVersionSupported(version) {
		return fmt.Errorf("%w: found %s, minimum required 2.20.0", ErrGitVersionUnsupported, version)
	}

	return nil
}

// ValidateRepository checks if a directory is a valid Git repository
func (v *GitValidator) ValidateRepository(ctx context.Context, path string) (*RepositoryInfo, error) {
	// Check if path exists and is accessible

	log.Println("path!!!!", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: directory does not exist", ErrRepositoryNotFound)
	}

	// Check if it's a Git repository
	isRepo, err := v.commands.IsRepository(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("repository validation failed: %w", err)
	}

	if !isRepo {
		return nil, ErrNotGitRepository
	}

	// Get repository information
	info, err := v.getRepositoryInfo(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository info: %w", err)
	}

	return info, nil
}

// ValidateRepositoryURL checks if a repository URL is valid and accessible
func (v *GitValidator) ValidateRepositoryURL(ctx context.Context, repoURL string) error {
	if repoURL == "" {
		return ErrInvalidRepositoryURL
	}

	// Parse URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Errorf("%w: invalid URL format", ErrInvalidRepositoryURL)
	}

	// Validate URL scheme
	if !v.isSupportedScheme(parsedURL.Scheme) {
		return fmt.Errorf("%w: unsupported scheme '%s'", ErrInvalidRepositoryURL, parsedURL.Scheme)
	}

	// Validate URL format for different schemes
	switch parsedURL.Scheme {
	case "https", "http":
		if err := v.validateHTTPRepoURL(parsedURL); err != nil {
			return err
		}
	case "ssh":
		if err := v.validateSSHRepoURL(parsedURL); err != nil {
			return err
		}
	case "git":
		if err := v.validateGitRepoURL(parsedURL); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBranchName checks if a branch name is valid according to Git rules
func (v *GitValidator) ValidateBranchName(branchName string) error {
	if branchName == "" {
		return fmt.Errorf("%w: branch name cannot be empty", ErrInvalidBranchName)
	}

	// Git branch name rules
	invalidPatterns := []string{
		`^\.`,       // Cannot start with dot
		`\.\.$`,     // Cannot end with two dots
		`^/`,        // Cannot start with slash
		`/$`,        // Cannot end with slash
		`//`,        // Cannot contain double slash
		`\.lock$`,   // Cannot end with .lock
		`@{`,        // Cannot contain @{
		`\^`,        // Cannot contain ^
		`~`,         // Cannot contain ~
		`:`,         // Cannot contain :
		`\?`,        // Cannot contain ?
		`\*`,        // Cannot contain *
		`\[`,        // Cannot contain [
		`\\`,        // Cannot contain backslash
		`\s`,        // Cannot contain whitespace
		`\x00-\x1f`, // Cannot contain control characters
		`\x7f`,      // Cannot contain DEL character
	}

	for _, pattern := range invalidPatterns {
		matched, err := regexp.MatchString(pattern, branchName)
		if err != nil {
			continue // Skip invalid regex patterns
		}
		if matched {
			return fmt.Errorf("%w: contains invalid characters or format", ErrInvalidBranchName)
		}
	}

	// Additional Git branch name rules
	if strings.Contains(branchName, "..") {
		return fmt.Errorf("%w: cannot contain consecutive dots", ErrInvalidBranchName)
	}

	if len(branchName) > 255 {
		return fmt.Errorf("%w: branch name too long (max 255 characters)", ErrInvalidBranchName)
	}

	return nil
}

// ValidateGitConfig checks if Git user configuration is set
func (v *GitValidator) ValidateGitConfig(ctx context.Context, workingDir string) (*GitConfig, error) {
	config := &GitConfig{}

	// Check user.name
	result, err := v.commands.executor.Execute(ctx, workingDir, "config", "user.name")
	if err != nil || result.ExitCode != 0 {
		return nil, fmt.Errorf("%w: user.name not configured", ErrGitConfigNotSet)
	}
	config.UserName = strings.TrimSpace(result.Stdout)

	// Check user.email
	result, err = v.commands.executor.Execute(ctx, workingDir, "config", "user.email")
	if err != nil || result.ExitCode != 0 {
		return nil, fmt.Errorf("%w: user.email not configured", ErrGitConfigNotSet)
	}
	config.UserEmail = strings.TrimSpace(result.Stdout)

	// Validate email format
	if !v.isValidEmail(config.UserEmail) {
		return nil, fmt.Errorf("%w: invalid email format", ErrInvalidGitConfig)
	}

	// Get additional config if available
	if result, err := v.commands.executor.Execute(ctx, workingDir, "config", "core.editor"); err == nil && result.ExitCode == 0 {
		config.CoreEditor = strings.TrimSpace(result.Stdout)
	}

	if result, err := v.commands.executor.Execute(ctx, workingDir, "config", "init.defaultBranch"); err == nil && result.ExitCode == 0 {
		config.DefaultBranch = strings.TrimSpace(result.Stdout)
	}

	return config, nil
}

// ValidateWorkingDirectory checks the state of the working directory
func (v *GitValidator) ValidateWorkingDirectory(ctx context.Context, workingDir string) (*WorkingDirStatus, error) {
	status := &WorkingDirStatus{}

	// Get status
	statusOutput, err := v.commands.Status(ctx, workingDir, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory status: %w", err)
	}

	// Parse porcelain status
	lines := strings.Split(strings.TrimSpace(statusOutput), "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}

		indexStatus := line[0]
		workingStatus := line[1]

		// Check for staged changes (index status)
		if indexStatus != ' ' && indexStatus != '?' {
			status.HasStagedChanges = true
		}

		// Check for unstaged changes (working tree status)
		if workingStatus != ' ' && workingStatus != '?' {
			status.HasUnstagedChanges = true
		}

		// Check for untracked files
		if indexStatus == '?' && workingStatus == '?' {
			status.HasUntrackedFiles = true
		}
	}

	status.IsClean = !status.HasStagedChanges && !status.HasUnstagedChanges && !status.HasUntrackedFiles

	return status, nil
}

// CheckBranchExists verifies if a branch exists in the repository
func (v *GitValidator) CheckBranchExists(ctx context.Context, workingDir, branchName string) (bool, error) {
	branches, err := v.commands.ListBranches(ctx, workingDir, &ListBranchesOptions{All: true})
	if err != nil {
		return false, fmt.Errorf("failed to list branches: %w", err)
	}

	for _, branch := range branches {
		if branch == branchName {
			return true, nil
		}
	}

	return false, nil
}

// Helper methods

// isVersionSupported checks if the Git version meets minimum requirements
func (v *GitValidator) isVersionSupported(version string) bool {
	// Extract version numbers (e.g., "2.34.1" from "git version 2.34.1")
	versionRegex := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	matches := versionRegex.FindStringSubmatch(version)

	if len(matches) < 4 {
		return false
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	// Minimum version: 2.20.0
	if major > 2 {
		return true
	}
	if major == 2 && minor > 20 {
		return true
	}
	if major == 2 && minor == 20 && patch >= 0 {
		return true
	}

	return false
}

// isSupportedScheme checks if URL scheme is supported
func (v *GitValidator) isSupportedScheme(scheme string) bool {
	supportedSchemes := []string{"https", "http", "ssh", "git"}
	for _, supported := range supportedSchemes {
		if scheme == supported {
			return true
		}
	}
	return false
}

// validateHTTPRepoURL validates HTTP/HTTPS repository URLs
func (v *GitValidator) validateHTTPRepoURL(parsedURL *url.URL) error {
	if parsedURL.Host == "" {
		return fmt.Errorf("%w: missing hostname", ErrInvalidRepositoryURL)
	}

	if !strings.HasSuffix(parsedURL.Path, ".git") && !v.isKnownGitHost(parsedURL.Host) {
		return fmt.Errorf("%w: URL should end with .git or be from a known Git hosting service", ErrInvalidRepositoryURL)
	}

	return nil
}

// validateSSHRepoURL validates SSH repository URLs
func (v *GitValidator) validateSSHRepoURL(parsedURL *url.URL) error {
	if parsedURL.Host == "" {
		return fmt.Errorf("%w: missing hostname", ErrInvalidRepositoryURL)
	}

	if parsedURL.User == nil {
		return fmt.Errorf("%w: SSH URL must include username", ErrInvalidRepositoryURL)
	}

	return nil
}

// validateGitRepoURL validates git:// repository URLs
func (v *GitValidator) validateGitRepoURL(parsedURL *url.URL) error {
	if parsedURL.Host == "" {
		return fmt.Errorf("%w: missing hostname", ErrInvalidRepositoryURL)
	}

	return nil
}

// isKnownGitHost checks if the host is a known Git hosting service
func (v *GitValidator) isKnownGitHost(host string) bool {
	knownHosts := []string{
		"github.com",
		"gitlab.com",
		"bitbucket.org",
		"codeberg.org",
		"git.sr.ht",
	}

	for _, knownHost := range knownHosts {
		if host == knownHost {
			return true
		}
		// Check for enterprise/subdomain versions (e.g., github.company.com)
		if strings.Contains(host, knownHost) {
			return true
		}
	}

	return false
}

// isValidEmail validates email format
func (v *GitValidator) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// getRepositoryInfo retrieves detailed information about a repository
func (v *GitValidator) getRepositoryInfo(ctx context.Context, workingDir string) (*RepositoryInfo, error) {
	info := &RepositoryInfo{Path: workingDir}

	// Get current branch
	branch, err := v.commands.CurrentBranch(ctx, workingDir)
	if err != nil {
		// Repository might be in detached HEAD state or have no commits
		info.CurrentBranch = ""
	} else {
		info.CurrentBranch = branch
	}

	// Get remote URL
	remoteURL, err := v.commands.GetRemoteURL(ctx, workingDir, "origin")
	if err != nil {
		// Repository might not have a remote
		info.RemoteURL = ""
	} else {
		info.RemoteURL = remoteURL
	}

	// Get working directory status
	workingDirStatus, err := v.ValidateWorkingDirectory(ctx, workingDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory status: %w", err)
	}
	info.WorkingDirStatus = *workingDirStatus

	// Get last commit info if available
	commitInfo, err := v.commands.GetCommitInfo(ctx, workingDir, "HEAD")
	if err != nil {
		// Repository might have no commits
		info.LastCommit = nil
	} else {
		info.LastCommit = commitInfo
	}

	return info, nil
}

// Data structures for validation results

// RepositoryInfo contains information about a Git repository
type RepositoryInfo struct {
	Path             string
	CurrentBranch    string
	RemoteURL        string
	WorkingDirStatus WorkingDirStatus
	LastCommit       *CommitInfo
}

// WorkingDirStatus represents the status of the working directory
type WorkingDirStatus struct {
	IsClean            bool
	HasStagedChanges   bool
	HasUnstagedChanges bool
	HasUntrackedFiles  bool
}

// GitConfig represents Git configuration settings
type GitConfig struct {
	UserName      string
	UserEmail     string
	CoreEditor    string
	DefaultBranch string
}
