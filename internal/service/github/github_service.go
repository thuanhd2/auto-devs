package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
)

// GitHubConfig holds the configuration for GitHub API integration
type GitHubConfig struct {
	Token     string
	BaseURL   string
	UserAgent string
	Timeout   time.Duration
}

// GitHubService provides GitHub API integration capabilities
type GitHubService struct {
	config     *GitHubConfig
	httpClient *http.Client
	rateLimiter *RateLimiter
}

// NewGitHubService creates a new GitHub service instance
func NewGitHubService(config *GitHubConfig) *GitHubService {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.github.com"
	}
	if config.UserAgent == "" {
		config.UserAgent = "auto-devs/1.0"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &GitHubService{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		rateLimiter: NewRateLimiter(),
	}
}

// GitHubPullRequest represents a GitHub pull request response
type GitHubPullRequest struct {
	Number      int                    `json:"number"`
	Title       string                 `json:"title"`
	Body        *string                `json:"body"`
	State       string                 `json:"state"`
	Head        GitHubBranch           `json:"head"`
	Base        GitHubBranch           `json:"base"`
	HTMLURL     string                 `json:"html_url"`
	MergeCommitSHA *string             `json:"merge_commit_sha"`
	MergedAt    *time.Time             `json:"merged_at"`
	ClosedAt    *time.Time             `json:"closed_at"`
	Draft       bool                   `json:"draft"`
	Mergeable   *bool                  `json:"mergeable"`
	MergeableState *string             `json:"mergeable_state"`
	Additions   *int                   `json:"additions"`
	Deletions   *int                   `json:"deletions"`
	ChangedFiles *int                  `json:"changed_files"`
	User        GitHubUser             `json:"user"`
	MergedBy    *GitHubUser            `json:"merged_by"`
	Assignees   []GitHubUser           `json:"assignees"`
	RequestedReviewers []GitHubUser    `json:"requested_reviewers"`
	Labels      []GitHubLabel          `json:"labels"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// GitHubBranch represents a GitHub branch
type GitHubBranch struct {
	Ref  string          `json:"ref"`
	SHA  string          `json:"sha"`
	Repo GitHubRepository `json:"repo"`
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

// GitHubLabel represents a GitHub label
type GitHubLabel struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    GitHubUser `json:"owner"`
}

// CreatePullRequestRequest represents the request body for creating a pull request
type CreatePullRequestRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
	Draft bool   `json:"draft,omitempty"`
}

// UpdatePullRequestRequest represents the request body for updating a pull request
type UpdatePullRequestRequest struct {
	Title *string `json:"title,omitempty"`
	Body  *string `json:"body,omitempty"`
	State *string `json:"state,omitempty"`
	Base  *string `json:"base,omitempty"`
}

// MergePullRequestRequest represents the request body for merging a pull request
type MergePullRequestRequest struct {
	CommitTitle   *string `json:"commit_title,omitempty"`
	CommitMessage *string `json:"commit_message,omitempty"`
	SHA           *string `json:"sha,omitempty"`
	MergeMethod   string  `json:"merge_method"`
}

// CreatePullRequest creates a new pull request on GitHub
func (gs *GitHubService) CreatePullRequest(ctx context.Context, repo string, base string, head string, title string, body string) (*entity.PullRequest, error) {
	if err := gs.validateRepository(repo); err != nil {
		return nil, fmt.Errorf("invalid repository: %w", err)
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	reqBody := CreatePullRequestRequest{
		Title: title,
		Body:  body,
		Head:  head,
		Base:  base,
		Draft: false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/pulls", gs.config.BaseURL, repo)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	gs.setHeaders(req)

	resp, err := gs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limiter
	gs.rateLimiter.UpdateFromResponse(resp)

	if resp.StatusCode != http.StatusCreated {
		return nil, gs.handleErrorResponse(resp)
	}

	var ghPR GitHubPullRequest
	if err := json.NewDecoder(resp.Body).Decode(&ghPR); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return gs.convertToEntityPR(&ghPR, repo), nil
}

// GetPullRequest retrieves a pull request from GitHub
func (gs *GitHubService) GetPullRequest(ctx context.Context, repo string, prNumber int) (*entity.PullRequest, error) {
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

	url := fmt.Sprintf("%s/repos/%s/pulls/%d", gs.config.BaseURL, repo, prNumber)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	gs.setHeaders(req)

	resp, err := gs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limiter
	gs.rateLimiter.UpdateFromResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, gs.handleErrorResponse(resp)
	}

	var ghPR GitHubPullRequest
	if err := json.NewDecoder(resp.Body).Decode(&ghPR); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return gs.convertToEntityPR(&ghPR, repo), nil
}

// UpdatePullRequest updates a pull request on GitHub
func (gs *GitHubService) UpdatePullRequest(ctx context.Context, repo string, prNumber int, updates map[string]interface{}) error {
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

	reqBody := UpdatePullRequestRequest{}
	
	if title, ok := updates["title"].(string); ok {
		reqBody.Title = &title
	}
	if body, ok := updates["body"].(string); ok {
		reqBody.Body = &body
	}
	if state, ok := updates["state"].(string); ok {
		reqBody.State = &state
	}
	if base, ok := updates["base"].(string); ok {
		reqBody.Base = &base
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/pulls/%d", gs.config.BaseURL, repo, prNumber)
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	gs.setHeaders(req)

	resp, err := gs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limiter
	gs.rateLimiter.UpdateFromResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return gs.handleErrorResponse(resp)
	}

	return nil
}

// MergePullRequest merges a pull request on GitHub
func (gs *GitHubService) MergePullRequest(ctx context.Context, repo string, prNumber int, mergeMethod string) error {
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

	reqBody := MergePullRequestRequest{
		MergeMethod: mergeMethod,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/pulls/%d/merge", gs.config.BaseURL, repo, prNumber)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	gs.setHeaders(req)

	resp, err := gs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limiter
	gs.rateLimiter.UpdateFromResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return gs.handleErrorResponse(resp)
	}

	return nil
}

// ValidateToken validates the GitHub token by making a test API call
func (gs *GitHubService) ValidateToken(ctx context.Context) error {
	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/user", gs.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	gs.setHeaders(req)

	resp, err := gs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limiter
	gs.rateLimiter.UpdateFromResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return gs.handleErrorResponse(resp)
	}

	return nil
}

// setHeaders sets the required headers for GitHub API requests
func (gs *GitHubService) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "token "+gs.config.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", gs.config.UserAgent)
}

// handleErrorResponse handles error responses from GitHub API
func (gs *GitHubService) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP %d: failed to read error response", resp.StatusCode)
	}

	var errorResp struct {
		Message          string `json:"message"`
		DocumentationURL string `json:"documentation_url"`
		Errors          []struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &errorResp); err != nil {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("GitHub API error: %s (HTTP %d)", errorResp.Message, resp.StatusCode)
}

// validateRepository validates the repository format (owner/repo)
func (gs *GitHubService) validateRepository(repo string) error {
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

// isValidMergeMethod checks if the merge method is valid
func (gs *GitHubService) isValidMergeMethod(method string) bool {
	validMethods := []string{"merge", "squash", "rebase"}
	for _, valid := range validMethods {
		if method == valid {
			return true
		}
	}
	return false
}

// convertToEntityPR converts GitHub PR response to entity PR
func (gs *GitHubService) convertToEntityPR(ghPR *GitHubPullRequest, repo string) *entity.PullRequest {
	var status entity.PullRequestStatus
	switch strings.ToLower(ghPR.State) {
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

	pr := &entity.PullRequest{
		GitHubPRNumber: ghPR.Number,
		Repository:     repo,
		Title:          ghPR.Title,
		Status:         status,
		HeadBranch:     ghPR.Head.Ref,
		BaseBranch:     ghPR.Base.Ref,
		GitHubURL:      ghPR.HTMLURL,
		MergeCommitSHA: ghPR.MergeCommitSHA,
		MergedAt:       ghPR.MergedAt,
		ClosedAt:       ghPR.ClosedAt,
		IsDraft:        ghPR.Draft,
		Mergeable:      ghPR.Mergeable,
		MergeableState: ghPR.MergeableState,
		Additions:      ghPR.Additions,
		Deletions:      ghPR.Deletions,
		ChangedFiles:   ghPR.ChangedFiles,
		CreatedBy:      &ghPR.User.Login,
	}

	if ghPR.Body != nil {
		pr.Body = *ghPR.Body
	}

	if ghPR.MergedBy != nil {
		pr.MergedBy = &ghPR.MergedBy.Login
	}

	// Convert assignees
	if len(ghPR.Assignees) > 0 {
		pr.Assignees = make([]string, len(ghPR.Assignees))
		for i, assignee := range ghPR.Assignees {
			pr.Assignees[i] = assignee.Login
		}
	}

	// Convert reviewers
	if len(ghPR.RequestedReviewers) > 0 {
		pr.Reviewers = make([]string, len(ghPR.RequestedReviewers))
		for i, reviewer := range ghPR.RequestedReviewers {
			pr.Reviewers[i] = reviewer.Login
		}
	}

	// Convert labels
	if len(ghPR.Labels) > 0 {
		pr.Labels = make([]string, len(ghPR.Labels))
		for i, label := range ghPR.Labels {
			pr.Labels[i] = label.Name
		}
	}

	return pr
}