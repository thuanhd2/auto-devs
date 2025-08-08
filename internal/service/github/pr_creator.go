package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// GitHubServiceInterface defines the interface for GitHub operations needed by PRCreator and PRMonitor
type GitHubServiceInterface interface {
	CreatePullRequest(ctx context.Context, repo, base, head, title, body string) (*entity.PullRequest, error)
	UpdatePullRequest(ctx context.Context, repo string, prNumber int, updates map[string]interface{}) error
	GetPullRequest(ctx context.Context, repo string, prNumber int) (*entity.PullRequest, error)
}

// PRCreator handles automatic pull request creation from completed implementations
type PRCreator struct {
	githubService GitHubServiceInterface
	baseURL       string // Base URL for task links (e.g., "https://auto-devs.example.com")
}

// NewPRCreator creates a new PR creator instance
func NewPRCreator(githubService GitHubServiceInterface, baseURL string) *PRCreator {
	return &PRCreator{
		githubService: githubService,
		baseURL:       strings.TrimSuffix(baseURL, "/"),
	}
}

// CreatePRFromImplementation automatically creates a pull request when implementation is complete
func (prc *PRCreator) CreatePRFromImplementation(ctx context.Context, task entity.Task, execution entity.Execution, plan *entity.Plan) (*entity.PullRequest, error) {
	// Validate inputs using comprehensive validation
	if err := prc.ValidateTaskForPRCreation(task, execution); err != nil {
		return nil, err
	}

	// Generate PR title
	title, err := prc.GeneratePRTitle(task)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PR title: %w", err)
	}

	// Generate PR description
	description, err := prc.GeneratePRDescription(task, plan, execution)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PR description: %w", err)
	}

	// Extract repository from task's project (this would need to be available via Task.Project)
	// For now, assume repository is stored in project or can be derived
	repository := prc.getRepositoryFromTask(task)
	if repository == "" {
		return nil, fmt.Errorf("unable to determine repository from task")
	}

	// Create the pull request via GitHub API
	githubPR, err := prc.githubService.CreatePullRequest(
		ctx,
		repository,
		"main",           // base branch - could be configurable
		*task.BranchName, // head branch
		title,
		description,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub pull request: %w", err)
	}

	// Add task links to the created PR
	if err := prc.AddTaskLinks(ctx, githubPR, task); err != nil {
		// Log the error but don't fail the PR creation
		// This could be handled with a logger in the future
		_ = fmt.Errorf("failed to add task links to PR: %w", err)
	}

	return githubPR, nil
}

// GeneratePRTitle creates an informative and unique title for the pull request
func (prc *PRCreator) GeneratePRTitle(task entity.Task) (string, error) {
	if task.Title == "" {
		return "", fmt.Errorf("task title cannot be empty")
	}

	// Determine type prefix based on task characteristics
	typePrefix := prc.determineTypePrefix(task)

	// Create title with format: "[TYPE] Task Title (Task-ID)"
	// Truncate title if too long to fit within GitHub's PR title limits
	maxTitleLength := 255 - len(typePrefix) - len(task.ID.String()) - 5 // Account for brackets and spaces

	title := task.Title
	if len(title) > maxTitleLength {
		title = title[:maxTitleLength-3] + "..."
	}

	return fmt.Sprintf("%s %s (%s)", typePrefix, title, task.ID.String()[:8]), nil
}

// GeneratePRDescription creates a comprehensive description for the pull request
func (prc *PRCreator) GeneratePRDescription(task entity.Task, plan *entity.Plan, execution entity.Execution) (string, error) {
	var description strings.Builder

	// Add task information
	description.WriteString("## Task Information\n\n")
	description.WriteString(fmt.Sprintf("**Task ID:** %s\n", task.ID.String()))
	description.WriteString(fmt.Sprintf("**Title:** %s\n", task.Title))

	if task.Description != "" {
		description.WriteString(fmt.Sprintf("**Description:**\n%s\n\n", task.Description))
	}

	description.WriteString(fmt.Sprintf("**Priority:** %s\n", task.Priority.GetDisplayName()))
	description.WriteString(fmt.Sprintf("**Status:** %s\n\n", task.Status.GetDisplayName()))

	// Add task link
	if prc.baseURL != "" {
		taskURL := fmt.Sprintf("%s/projects/%s/tasks/%s", prc.baseURL, task.ProjectID.String(), task.ID.String())
		description.WriteString(fmt.Sprintf("**Task URL:** %s\n\n", taskURL))
	}

	// Add plan reference if available
	if plan != nil {
		description.WriteString("## Implementation Plan\n\n")
		description.WriteString(fmt.Sprintf("**Plan Status:** %s\n", plan.Status.GetDisplayName()))
		description.WriteString(fmt.Sprintf("**Plan ID:** %s\n\n", plan.ID.String()))

		// Add truncated plan content for context
		planContent := plan.Content
		if len(planContent) > 500 {
			planContent = planContent[:500] + "...\n\n[See full plan in task details]"
		}
		description.WriteString(fmt.Sprintf("**Plan Summary:**\n```\n%s\n```\n\n", planContent))
	}

	// Add implementation summary
	description.WriteString("## Implementation Summary\n\n")
	description.WriteString(fmt.Sprintf("**Execution ID:** %s\n", execution.ID.String()))
	description.WriteString(fmt.Sprintf("**Execution Status:** %s\n", execution.Status))
	description.WriteString(fmt.Sprintf("**Started At:** %s\n", execution.StartedAt.Format(time.RFC3339)))

	if execution.CompletedAt != nil {
		description.WriteString(fmt.Sprintf("**Completed At:** %s\n", execution.CompletedAt.Format(time.RFC3339)))
		duration := execution.GetDuration()
		description.WriteString(fmt.Sprintf("**Duration:** %v\n", duration.Round(time.Second)))
	}

	if execution.Result != "" {
		description.WriteString(fmt.Sprintf("**Implementation Result:**\n```json\n%s\n```\n\n", execution.Result))
	}

	// Add testing instructions
	description.WriteString("## Testing Instructions\n\n")
	description.WriteString("1. Check out this branch locally\n")
	description.WriteString("2. Run the application and verify the implemented functionality\n")
	description.WriteString("3. Run tests to ensure no regressions:\n")
	description.WriteString("   ```bash\n")
	description.WriteString("   make test\n")
	description.WriteString("   ```\n")
	description.WriteString("4. Verify the changes meet the requirements outlined in the task description\n\n")

	// Add checklist
	description.WriteString("## Review Checklist\n\n")
	description.WriteString("- [ ] Code follows project conventions and style guidelines\n")
	description.WriteString("- [ ] All tests pass\n")
	description.WriteString("- [ ] No breaking changes introduced\n")
	description.WriteString("- [ ] Documentation updated if needed\n")
	description.WriteString("- [ ] Security considerations addressed\n")
	description.WriteString("- [ ] Performance impact assessed\n\n")

	// Add metadata
	description.WriteString("---\n")
	description.WriteString("*This pull request was automatically generated by Auto-Devs AI system*\n")

	// Sanitize the description before returning
	return prc.SanitizeForGitHub(description.String()), nil
}

// AddTaskLinks creates bidirectional links between the PR and the task
func (prc *PRCreator) AddTaskLinks(ctx context.Context, pr *entity.PullRequest, task entity.Task) error {
	if pr == nil {
		return fmt.Errorf("pull request cannot be nil")
	}

	// Update PR description to include task reference if not already present
	taskRef := fmt.Sprintf("Task-%s", task.ID.String()[:8])
	if !strings.Contains(pr.Body, taskRef) {
		updatedBody := pr.Body + fmt.Sprintf("\n\n**Related Task:** %s", taskRef)

		// Update the PR via GitHub API
		updates := map[string]interface{}{
			"body": updatedBody,
		}

		err := prc.githubService.UpdatePullRequest(ctx, pr.Repository, pr.GitHubPRNumber, updates)
		if err != nil {
			return fmt.Errorf("failed to update PR with task link: %w", err)
		}

		// Update local entity
		pr.Body = updatedBody
	}

	return nil
}

// determineTypePrefix determines the appropriate type prefix for the PR title
func (prc *PRCreator) determineTypePrefix(task entity.Task) string {
	title := strings.ToLower(task.Title)
	description := strings.ToLower(task.Description)
	combined := title + " " + description

	// Check for common patterns
	if strings.Contains(combined, "fix") || strings.Contains(combined, "bug") || strings.Contains(combined, "error") {
		return "[fix]"
	}
	if strings.Contains(combined, "feature") || strings.Contains(combined, "add") || strings.Contains(combined, "implement") {
		return "[feat]"
	}
	if strings.Contains(combined, "refactor") || strings.Contains(combined, "improve") || strings.Contains(combined, "optimize") {
		return "[refactor]"
	}
	if strings.Contains(combined, "docs") || strings.Contains(combined, "documentation") {
		return "[docs]"
	}
	if strings.Contains(combined, "test") {
		return "[test]"
	}
	if strings.Contains(combined, "style") || strings.Contains(combined, "format") {
		return "[style]"
	}
	if strings.Contains(combined, "chore") || strings.Contains(combined, "maintenance") {
		return "[chore]"
	}

	// Default to feature for new functionality
	return "[feat]"
}

// getRepositoryFromTask extracts the repository information from a task
// Expected format: "https://github.com/owner/repo" -> "owner/repo"
func (prc *PRCreator) getRepositoryFromTask(task entity.Task) string {
	if task.Project.RepositoryURL == "" {
		return ""
	}

	// Parse GitHub URL to extract owner/repo format
	repoURL := task.Project.RepositoryURL

	// Remove common prefixes
	prefixes := []string{
		"https://github.com/",
		"http://github.com/",
		"git@github.com:",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(repoURL, prefix) {
			repoURL = strings.TrimPrefix(repoURL, prefix)
			break
		}
	}

	// Remove .git suffix if present
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Validate format (should be owner/repo)
	parts := strings.Split(repoURL, "/")
	if len(parts) >= 2 && parts[0] != "" && parts[1] != "" {
		return fmt.Sprintf("%s/%s", parts[0], parts[1])
	}

	return ""
}

// PRCreationError represents errors that occur during PR creation
type PRCreationError struct {
	TaskID     string
	Step       string
	Underlying error
}

func (e *PRCreationError) Error() string {
	return fmt.Sprintf("PR creation failed at step '%s' for task %s: %v", e.Step, e.TaskID, e.Underlying)
}

func (e *PRCreationError) Unwrap() error {
	return e.Underlying
}

// CreatePRCreationError creates a new PR creation error
func CreatePRCreationError(taskID, step string, err error) *PRCreationError {
	return &PRCreationError{
		TaskID:     taskID,
		Step:       step,
		Underlying: err,
	}
}

// ValidateTaskForPRCreation validates that a task is ready for PR creation
func (prc *PRCreator) ValidateTaskForPRCreation(task entity.Task, execution entity.Execution) error {
	// Check task has required fields
	if task.ID == (uuid.UUID{}) {
		return CreatePRCreationError(task.ID.String(), "validation", fmt.Errorf("task ID cannot be nil"))
	}

	if task.Title == "" {
		return CreatePRCreationError(task.ID.String(), "validation", fmt.Errorf("task title cannot be empty"))
	}

	if task.BranchName == nil || *task.BranchName == "" {
		return CreatePRCreationError(task.ID.String(), "validation", fmt.Errorf("task must have a branch name"))
	}

	// Check execution is complete
	if execution.Status != entity.ExecutionStatusCompleted {
		return CreatePRCreationError(task.ID.String(), "validation",
			fmt.Errorf("execution must be completed, current status: %s", execution.Status))
	}

	// Check repository is available
	repository := prc.getRepositoryFromTask(task)
	if repository == "" {
		return CreatePRCreationError(task.ID.String(), "validation",
			fmt.Errorf("unable to determine repository from task project"))
	}

	return nil
}

// SanitizeForGitHub sanitizes text for safe use in GitHub API calls
func (prc *PRCreator) SanitizeForGitHub(text string) string {
	// Remove null bytes and other problematic characters
	text = strings.ReplaceAll(text, string(rune(0)), "")
	text = strings.TrimSpace(text)

	// Limit length to prevent API errors
	if len(text) > 65535 { // GitHub's limit for PR descriptions
		text = text[:65535-3] + "..."
	}

	return text
}
