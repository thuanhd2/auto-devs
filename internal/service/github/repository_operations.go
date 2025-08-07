package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GitHubCommit represents a GitHub commit
type GitHubCommit struct {
	SHA     string     `json:"sha"`
	Message string     `json:"message"`
	Author  GitHubUser `json:"author"`
	Committer GitHubUser `json:"committer"`
	Tree    struct {
		SHA string `json:"sha"`
	} `json:"tree"`
	Parents []struct {
		SHA string `json:"sha"`
	} `json:"parents"`
}

// GitHubTree represents a GitHub tree
type GitHubTree struct {
	SHA  string `json:"sha"`
	Tree []GitHubTreeEntry `json:"tree"`
}

// GitHubTreeEntry represents a GitHub tree entry
type GitHubTreeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
	Size *int   `json:"size,omitempty"`
}

// GitHubBlob represents a GitHub blob
type GitHubBlob struct {
	SHA      string `json:"sha"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	Size     int    `json:"size"`
}

// GitHubRef represents a GitHub reference
type GitHubRef struct {
	Ref    string `json:"ref"`
	NodeID string `json:"node_id"`
	URL    string `json:"url"`
	Object struct {
		SHA  string `json:"sha"`
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"object"`
}

// CreateFileRequest represents a request to create a file
type CreateFileRequest struct {
	Message   string `json:"message"`
	Content   string `json:"content"`
	Branch    string `json:"branch,omitempty"`
	Committer *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"committer,omitempty"`
}

// UpdateFileRequest represents a request to update a file
type UpdateFileRequest struct {
	Message   string `json:"message"`
	Content   string `json:"content"`
	SHA       string `json:"sha"`
	Branch    string `json:"branch,omitempty"`
	Committer *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"committer,omitempty"`
}

// DeleteFileRequest represents a request to delete a file
type DeleteFileRequest struct {
	Message   string `json:"message"`
	SHA       string `json:"sha"`
	Branch    string `json:"branch,omitempty"`
	Committer *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"committer,omitempty"`
}

// CreateBranchRequest represents a request to create a branch
type CreateBranchRequest struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// ValidateRepository checks if a repository exists and is accessible
func (gs *GitHubService) ValidateRepository(ctx context.Context, repo string) error {
	if err := gs.validateRepository(repo); err != nil {
		return err
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s", gs.config.BaseURL, repo)
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

// GetRepository retrieves repository information
func (gs *GitHubService) GetRepository(ctx context.Context, repo string) (*GitHubRepository, error) {
	if err := gs.validateRepository(repo); err != nil {
		return nil, err
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s", gs.config.BaseURL, repo)
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

	var repository GitHubRepository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &repository, nil
}

// CreateBranch creates a new branch
func (gs *GitHubService) CreateBranch(ctx context.Context, repo string, branchName string, baseSHA string) error {
	if err := gs.validateRepository(repo); err != nil {
		return err
	}

	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	if baseSHA == "" {
		return fmt.Errorf("base SHA cannot be empty")
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	reqBody := CreateBranchRequest{
		Ref: fmt.Sprintf("refs/heads/%s", branchName),
		SHA: baseSHA,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/git/refs", gs.config.BaseURL, repo)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
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

	if resp.StatusCode != http.StatusCreated {
		return gs.handleErrorResponse(resp)
	}

	return nil
}

// GetBranch retrieves branch information
func (gs *GitHubService) GetBranch(ctx context.Context, repo string, branchName string) (*GitHubRef, error) {
	if err := gs.validateRepository(repo); err != nil {
		return nil, err
	}

	if branchName == "" {
		return nil, fmt.Errorf("branch name cannot be empty")
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/git/refs/heads/%s", gs.config.BaseURL, repo, url.PathEscape(branchName))
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

	var ref GitHubRef
	if err := json.NewDecoder(resp.Body).Decode(&ref); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ref, nil
}

// DeleteBranch deletes a branch
func (gs *GitHubService) DeleteBranch(ctx context.Context, repo string, branchName string) error {
	if err := gs.validateRepository(repo); err != nil {
		return err
	}

	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/git/refs/heads/%s", gs.config.BaseURL, repo, url.PathEscape(branchName))
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
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

	if resp.StatusCode != http.StatusNoContent {
		return gs.handleErrorResponse(resp)
	}

	return nil
}

// GetFile retrieves a file from the repository
func (gs *GitHubService) GetFile(ctx context.Context, repo string, filePath string, ref string) (*GitHubBlob, error) {
	if err := gs.validateRepository(repo); err != nil {
		return nil, err
	}

	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/contents/%s", gs.config.BaseURL, repo, url.PathEscape(filePath))
	if ref != "" {
		url += "?ref=" + url.QueryEscape(ref)
	}

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

	var blob GitHubBlob
	if err := json.NewDecoder(resp.Body).Decode(&blob); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &blob, nil
}

// CreateFile creates a new file in the repository
func (gs *GitHubService) CreateFile(ctx context.Context, repo string, filePath string, content string, message string, branch string) error {
	if err := gs.validateRepository(repo); err != nil {
		return err
	}

	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	if message == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	reqBody := CreateFileRequest{
		Message: message,
		Content: content,
		Branch:  branch,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/contents/%s", gs.config.BaseURL, repo, url.PathEscape(filePath))
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

	if resp.StatusCode != http.StatusCreated {
		return gs.handleErrorResponse(resp)
	}

	return nil
}

// UpdateFile updates an existing file in the repository
func (gs *GitHubService) UpdateFile(ctx context.Context, repo string, filePath string, content string, message string, sha string, branch string) error {
	if err := gs.validateRepository(repo); err != nil {
		return err
	}

	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	if message == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	if sha == "" {
		return fmt.Errorf("file SHA cannot be empty")
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	reqBody := UpdateFileRequest{
		Message: message,
		Content: content,
		SHA:     sha,
		Branch:  branch,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/contents/%s", gs.config.BaseURL, repo, url.PathEscape(filePath))
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

// DeleteFile deletes a file from the repository
func (gs *GitHubService) DeleteFile(ctx context.Context, repo string, filePath string, message string, sha string, branch string) error {
	if err := gs.validateRepository(repo); err != nil {
		return err
	}

	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	if message == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	if sha == "" {
		return fmt.Errorf("file SHA cannot be empty")
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	reqBody := DeleteFileRequest{
		Message: message,
		SHA:     sha,
		Branch:  branch,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/contents/%s", gs.config.BaseURL, repo, url.PathEscape(filePath))
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, bytes.NewReader(bodyBytes))
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

// GetCommit retrieves a commit by SHA
func (gs *GitHubService) GetCommit(ctx context.Context, repo string, sha string) (*GitHubCommit, error) {
	if err := gs.validateRepository(repo); err != nil {
		return nil, err
	}

	if sha == "" {
		return nil, fmt.Errorf("commit SHA cannot be empty")
	}

	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/git/commits/%s", gs.config.BaseURL, repo, sha)
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

	var commit GitHubCommit
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &commit, nil
}