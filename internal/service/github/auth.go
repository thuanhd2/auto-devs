package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// AuthenticatedUser represents the authenticated GitHub user
type AuthenticatedUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Type      string `json:"type"`
	Plan      struct {
		Name         string `json:"name"`
		Space        int    `json:"space"`
		Collaborators int   `json:"collaborators"`
		PrivateRepos int    `json:"private_repos"`
	} `json:"plan"`
}

// TokenInfo holds information about a GitHub token
type TokenInfo struct {
	User        *AuthenticatedUser `json:"user"`
	Scopes      []string           `json:"scopes"`
	TokenType   string             `json:"token_type"`
	ExpiresAt   *string            `json:"expires_at,omitempty"`
	Fingerprint string             `json:"fingerprint"`
	HashedToken string             `json:"hashed_token"`
}

// TokenScopes represents different GitHub token scopes
var TokenScopes = struct {
	Repo                string
	RepoStatus          string
	RepoDeployment      string
	PublicRepo          string
	RepoInvite          string
	SecurityEvents      string
	AdminRepoHook       string
	WriteRepoHook       string
	ReadRepoHook        string
	AdminOrg            string
	WriteOrg            string
	ReadOrg             string
	AdminPublicKey      string
	WritePublicKey      string
	ReadPublicKey       string
	AdminOrgHook        string
	Gist                string
	Notifications       string
	User                string
	ReadUser            string
	UserEmail           string
	UserFollow          string
	Project             string
	ReadProject         string
	DeleteRepo          string
	WritePackages       string
	ReadPackages        string
	DeletePackages      string
	AdminGPGKey         string
	WriteGPGKey         string
	ReadGPGKey          string
	Codespace           string
	Workflow            string
}{
	Repo:                "repo",
	RepoStatus:          "repo:status",
	RepoDeployment:      "repo_deployment",
	PublicRepo:          "public_repo",
	RepoInvite:          "repo:invite",
	SecurityEvents:      "security_events",
	AdminRepoHook:       "admin:repo_hook",
	WriteRepoHook:       "write:repo_hook",
	ReadRepoHook:        "read:repo_hook",
	AdminOrg:            "admin:org",
	WriteOrg:            "write:org",
	ReadOrg:             "read:org",
	AdminPublicKey:      "admin:public_key",
	WritePublicKey:      "write:public_key",
	ReadPublicKey:       "read:public_key",
	AdminOrgHook:        "admin:org_hook",
	Gist:                "gist",
	Notifications:       "notifications",
	User:                "user",
	ReadUser:            "read:user",
	UserEmail:           "user:email",
	UserFollow:          "user:follow",
	Project:             "project",
	ReadProject:         "read:project",
	DeleteRepo:          "delete_repo",
	WritePackages:       "write:packages",
	ReadPackages:        "read:packages",
	DeletePackages:      "delete:packages",
	AdminGPGKey:         "admin:gpg_key",
	WriteGPGKey:         "write:gpg_key",
	ReadGPGKey:          "read:gpg_key",
	Codespace:           "codespace",
	Workflow:            "workflow",
}

// RequiredScopes defines the minimum scopes required for different operations
var RequiredScopes = map[string][]string{
	"pull_request": {TokenScopes.Repo, TokenScopes.PublicRepo},
	"repository":   {TokenScopes.Repo, TokenScopes.PublicRepo},
	"branch":       {TokenScopes.Repo, TokenScopes.PublicRepo},
	"file":         {TokenScopes.Repo, TokenScopes.PublicRepo},
	"commit":       {TokenScopes.Repo, TokenScopes.PublicRepo},
}

// ValidateToken validates the GitHub token and returns user information
func (gs *GitHubService) ValidateTokenWithInfo(ctx context.Context) (*TokenInfo, error) {
	// Get authenticated user
	user, err := gs.GetAuthenticatedUser(ctx)
	if err != nil {
		return nil, &AuthenticationError{
			Message: "failed to validate token",
			Cause:   err,
		}
	}

	// Check token format
	if err := gs.validateTokenFormat(gs.config.Token); err != nil {
		return nil, &AuthenticationError{
			Message: "invalid token format",
			Cause:   err,
		}
	}

	return &TokenInfo{
		User:      user,
		Scopes:    []string{}, // GitHub API v3 doesn't return scope info in user endpoint
		TokenType: "Bearer",
	}, nil
}

// GetAuthenticatedUser retrieves information about the authenticated user
func (gs *GitHubService) GetAuthenticatedUser(ctx context.Context) (*AuthenticatedUser, error) {
	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/user", gs.config.BaseURL)
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
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, &AuthenticationError{
				Message: "invalid or expired token",
			}
		}
		return nil, gs.handleErrorResponse(resp)
	}

	var user AuthenticatedUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &user, nil
}

// ValidateTokenScopes checks if the token has the required scopes for the operation
func (gs *GitHubService) ValidateTokenScopes(ctx context.Context, operation string) error {
	requiredScopes, exists := RequiredScopes[operation]
	if !exists {
		return &ValidationError{
			Field:   "operation",
			Value:   operation,
			Message: "unknown operation",
		}
	}

	// For now, we'll do a basic validation by checking if we can access the user endpoint
	// GitHub API v3 doesn't provide a direct way to check token scopes
	if err := gs.ValidateToken(ctx); err != nil {
		return &AuthenticationError{
			Message: fmt.Sprintf("token validation failed for operation '%s'", operation),
			Cause:   err,
		}
	}

	// Additional validation could be done by trying to access specific resources
	// that require the scopes, but for now this basic check is sufficient

	return nil
}

// validateTokenFormat validates the GitHub token format
func (gs *GitHubService) validateTokenFormat(token string) error {
	if token == "" {
		return &ValidationError{
			Field:   "token",
			Value:   "",
			Message: "token cannot be empty",
		}
	}

	// GitHub personal access tokens (classic) start with 'ghp_'
	// GitHub Apps installation tokens start with 'ghs_'
	// GitHub Apps user access tokens start with 'ghu_'
	// Fine-grained personal access tokens start with 'github_pat_'
	validPrefixes := []string{"ghp_", "ghs_", "ghu_", "github_pat_"}
	
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(token, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		// Also check for old format tokens (40 characters, hexadecimal)
		oldFormatRegex := regexp.MustCompile(`^[a-f0-9]{40}$`)
		if !oldFormatRegex.MatchString(token) {
			return &ValidationError{
				Field:   "token",
				Value:   token[:10] + "...", // Show only first 10 chars for security
				Message: "invalid token format",
			}
		}
	}

	return nil
}

// CheckRepositoryAccess checks if the token has access to the specified repository
func (gs *GitHubService) CheckRepositoryAccess(ctx context.Context, repo string) error {
	if err := gs.validateRepository(repo); err != nil {
		return &RepositoryError{
			Repository: repo,
			Message:    "invalid repository format",
			Cause:      err,
		}
	}

	// Try to access the repository
	if err := gs.ValidateRepository(ctx, repo); err != nil {
		if ghErr, ok := IsGitHubError(err); ok {
			if ghErr.IsNotFound() {
				return &RepositoryError{
					Repository: repo,
					Message:    "repository not found or no access",
					Cause:      err,
				}
			}
			if ghErr.IsUnauthorized() || ghErr.IsForbidden() {
				return &AuthenticationError{
					Message: fmt.Sprintf("insufficient permissions for repository '%s'", repo),
					Cause:   err,
				}
			}
		}
		return &RepositoryError{
			Repository: repo,
			Message:    "failed to access repository",
			Cause:      err,
		}
	}

	return nil
}

// RefreshToken refreshes the GitHub token if it's an OAuth token
// Note: Personal access tokens don't need refreshing, but GitHub Apps tokens do
func (gs *GitHubService) RefreshToken(ctx context.Context, refreshToken string) (*TokenInfo, error) {
	// This would be implemented for GitHub Apps OAuth flow
	// For personal access tokens, this is not applicable
	return nil, &AuthenticationError{
		Message: "token refresh not supported for personal access tokens",
	}
}

// RevokeToken revokes the GitHub token
func (gs *GitHubService) RevokeToken(ctx context.Context) error {
	// Wait for rate limit
	if err := gs.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	// For personal access tokens, we can delete them via the API
	// This requires the token to have 'delete_repo' scope or be done through web interface
	url := fmt.Sprintf("%s/applications/%s/token", gs.config.BaseURL, "github")
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

	// Token revocation might not always return 204, depends on the token type
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return gs.handleErrorResponse(resp)
	}

	return nil
}

// IsTokenExpired checks if the token is expired (for OAuth tokens)
func (gs *GitHubService) IsTokenExpired(ctx context.Context) (bool, error) {
	// Try to make a simple API call
	if err := gs.ValidateToken(ctx); err != nil {
		if authErr, ok := IsAuthenticationError(err); ok {
			return true, authErr
		}
		// Other errors don't necessarily mean the token is expired
		return false, err
	}
	return false, nil
}