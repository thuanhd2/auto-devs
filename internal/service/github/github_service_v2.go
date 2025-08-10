package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

// GitHubServiceV2 provides GitHub API integration capabilities using go-github library
type GitHubServiceV2 struct {
	config      *GitHubConfig
	client      *github.Client
	rateLimiter *RateLimiter
}

// NewGitHubServiceV2 creates a new GitHub service instance using go-github library
func NewGitHubServiceV2(config *GitHubConfig) *GitHubServiceV2 {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.github.com"
	}
	if config.UserAgent == "" {
		config.UserAgent = "auto-devs/1.0"
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	// Create OAuth2 token source
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)

	// Create HTTP client with OAuth2 transport
	httpClient := oauth2.NewClient(context.Background(), ts)
	httpClient.Timeout = time.Duration(config.Timeout) * time.Second

	// Create GitHub client
	var client *github.Client
	if config.BaseURL == "https://api.github.com" {
		client = github.NewClient(httpClient)
	} else {
		// For GitHub Enterprise
		client, _ = github.NewEnterpriseClient(config.BaseURL, config.BaseURL, httpClient)
	}

	return &GitHubServiceV2{
		config:      config,
		client:      client,
		rateLimiter: NewRateLimiter(),
	}
}

// CreatePullRequest creates a new pull request on GitHub
func (gs *GitHubServiceV2) CreatePullRequest(ctx context.Context, repo, base, head, title, body string) (*entity.PullRequest, error) {
	if err := gs.validateRepository(repo); err != nil {
		return nil, fmt.Errorf("invalid repository: %w", err)
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	// Parse repository owner and name
	owner, name := gs.parseRepository(repo)

	// Create pull request request
	prRequest := &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		Head:  &head,
		Base:  &base,
		Draft: github.Bool(false),
	}

	// Create pull request
	ghPR, resp, err := gs.client.PullRequests.Create(ctx, owner, name, prRequest)
	if err != nil {
		// Update rate limiter from response
		if resp != nil {
			gs.rateLimiter.UpdateFromGitHubResponse(resp)
		}
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	// Update rate limiter
	gs.rateLimiter.UpdateFromGitHubResponse(resp)

	return gs.convertToEntityPR(ghPR, repo), nil
}

// GetPullRequest retrieves a pull request from GitHub
func (gs *GitHubServiceV2) GetPullRequest(ctx context.Context, repo string, prNumber int) (*entity.PullRequest, error) {
	if err := gs.validateRepository(repo); err != nil {
		return nil, fmt.Errorf("invalid repository: %w", err)
	}

	if prNumber <= 0 {
		return nil, fmt.Errorf("invalid pull request number: %d", prNumber)
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	// Parse repository owner and name
	owner, name := gs.parseRepository(repo)

	// Get pull request
	ghPR, resp, err := gs.client.PullRequests.Get(ctx, owner, name, prNumber)
	if err != nil {
		// Update rate limiter from response
		if resp != nil {
			gs.rateLimiter.UpdateFromGitHubResponse(resp)
		}
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	// Update rate limiter
	gs.rateLimiter.UpdateFromGitHubResponse(resp)

	return gs.convertToEntityPR(ghPR, repo), nil
}

// UpdatePullRequest updates a pull request on GitHub
func (gs *GitHubServiceV2) UpdatePullRequest(ctx context.Context, repo string, prNumber int, updates map[string]interface{}) error {
	if err := gs.validateRepository(repo); err != nil {
		return fmt.Errorf("invalid repository: %w", err)
	}

	if prNumber <= 0 {
		return fmt.Errorf("invalid pull request number: %d", prNumber)
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	// Parse repository owner and name
	owner, name := gs.parseRepository(repo)

	// Create update request
	updateRequest := &github.PullRequest{}

	if title, ok := updates["title"].(string); ok {
		updateRequest.Title = &title
	}
	if body, ok := updates["body"].(string); ok {
		updateRequest.Body = &body
	}
	if state, ok := updates["state"].(string); ok {
		updateRequest.State = &state
	}
	if base, ok := updates["base"].(string); ok {
		updateRequest.Base = &github.PullRequestBranch{Ref: &base}
	}

	// Update pull request
	_, resp, err := gs.client.PullRequests.Edit(ctx, owner, name, prNumber, updateRequest)
	if err != nil {
		// Update rate limiter from response
		if resp != nil {
			gs.rateLimiter.UpdateFromGitHubResponse(resp)
		}
		return fmt.Errorf("failed to update pull request: %w", err)
	}

	// Update rate limiter
	gs.rateLimiter.UpdateFromGitHubResponse(resp)

	return nil
}

// MergePullRequest merges a pull request on GitHub
func (gs *GitHubServiceV2) MergePullRequest(ctx context.Context, repo string, prNumber int, mergeMethod string) error {
	if err := gs.validateRepository(repo); err != nil {
		return fmt.Errorf("invalid repository: %w", err)
	}

	if prNumber <= 0 {
		return fmt.Errorf("invalid pull request number: %d", prNumber)
	}

	if !gs.isValidMergeMethod(mergeMethod) {
		return fmt.Errorf("invalid merge method: %s", mergeMethod)
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	// Parse repository owner and name
	owner, name := gs.parseRepository(repo)

	// Merge pull request with merge method
	result, resp, err := gs.client.PullRequests.Merge(ctx, owner, name, prNumber, "", &github.PullRequestOptions{
		MergeMethod: mergeMethod,
	})
	if err != nil {
		// Update rate limiter from response
		if resp != nil {
			gs.rateLimiter.UpdateFromGitHubResponse(resp)
		}
		return fmt.Errorf("failed to merge pull request: %w", err)
	}

	// Update rate limiter
	gs.rateLimiter.UpdateFromGitHubResponse(resp)

	// Check if merge was successful
	if result.Merged == nil || !*result.Merged {
		return fmt.Errorf("pull request was not merged: %s", result.GetMessage())
	}

	return nil
}

// ValidateToken validates the GitHub token by making a test API call
func (gs *GitHubServiceV2) ValidateToken(ctx context.Context) error {
	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	// Get authenticated user
	user, resp, err := gs.client.Users.Get(ctx, "")
	if err != nil {
		// Update rate limiter from response
		if resp != nil {
			gs.rateLimiter.UpdateFromGitHubResponse(resp)
		}
		return fmt.Errorf("failed to validate token: %w", err)
	}

	// Update rate limiter
	gs.rateLimiter.UpdateFromGitHubResponse(resp)

	// Check if user is authenticated
	if user == nil || user.Login == nil {
		return fmt.Errorf("invalid token: no user information returned")
	}

	return nil
}

// validateRepository validates the repository format (owner/repo)
func (gs *GitHubServiceV2) validateRepository(repo string) error {
	if repo == "" {
		return fmt.Errorf("repository cannot be empty")
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("repository must be in format 'owner/repo'")
	}

	if parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("repository owner and name cannot be empty")
	}

	return nil
}

// parseRepository parses repository string into owner and name
func (gs *GitHubServiceV2) parseRepository(repo string) (owner, name string) {
	parts := strings.Split(repo, "/")
	return parts[0], parts[1]
}

// isValidMergeMethod checks if the merge method is valid
func (gs *GitHubServiceV2) isValidMergeMethod(method string) bool {
	validMethods := []string{"merge", "squash", "rebase"}
	for _, valid := range validMethods {
		if method == valid {
			return true
		}
	}
	return false
}

// convertToEntityPR converts GitHub PR response to entity PR
func (gs *GitHubServiceV2) convertToEntityPR(ghPR *github.PullRequest, repo string) *entity.PullRequest {
	var status entity.PullRequestStatus
	if ghPR.State != nil {
		switch strings.ToLower(*ghPR.State) {
		case "open":
			status = entity.PullRequestStatusOpen
		case "closed":
			if ghPR.MergedAt != nil {
				status = entity.PullRequestStatusMerged
			} else {
				status = entity.PullRequestStatusClosed
			}
		default:
			status = entity.PullRequestStatusOpen
		}
	} else {
		status = entity.PullRequestStatusOpen
	}

	mergedAt := ghPR.MergedAt.GetTime()
	closedAt := ghPR.ClosedAt.GetTime()

	pr := &entity.PullRequest{
		GitHubPRNumber: ghPR.GetNumber(),
		Repository:     repo,
		Title:          ghPR.GetTitle(),
		Status:         status,
		HeadBranch:     ghPR.GetHead().GetRef(),
		BaseBranch:     ghPR.GetBase().GetRef(),
		GitHubURL:      ghPR.GetHTMLURL(),
		MergeCommitSHA: ghPR.MergeCommitSHA,
		MergedAt:       mergedAt,
		ClosedAt:       closedAt,
		IsDraft:        ghPR.GetDraft(),
		Mergeable:      ghPR.Mergeable,
		MergeableState: ghPR.MergeableState,
		Additions:      ghPR.Additions,
		Deletions:      ghPR.Deletions,
		ChangedFiles:   ghPR.ChangedFiles,
		CreatedBy:      ghPR.User.Login,
	}

	if ghPR.Body != nil {
		pr.Body = *ghPR.Body
	}

	if ghPR.MergedBy != nil {
		pr.MergedBy = ghPR.MergedBy.Login
	}

	// Convert assignees
	if ghPR.Assignees != nil {
		pr.Assignees = make([]string, len(ghPR.Assignees))
		for i, assignee := range ghPR.Assignees {
			pr.Assignees[i] = assignee.GetLogin()
		}
	}

	// Convert requested reviewers
	if ghPR.RequestedReviewers != nil {
		pr.Reviewers = make([]string, len(ghPR.RequestedReviewers))
		for i, reviewer := range ghPR.RequestedReviewers {
			pr.Reviewers[i] = reviewer.GetLogin()
		}
	}

	// Convert labels
	if ghPR.Labels != nil {
		pr.Labels = make([]string, len(ghPR.Labels))
		for i, label := range ghPR.Labels {
			pr.Labels[i] = label.GetName()
		}
	}

	return pr
}
