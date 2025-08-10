package github

import (
	"github.com/google/go-github/v74/github"
)

// GitHubServiceV2 provides GitHub API integration capabilities using go-github library
type GitHubServiceV2 struct {
	config      *GitHubConfig
	client      *github.Client
	rateLimiter *RateLimiter
}

// implement github service by using go-github library
func NewGitHubServiceV2(config *GitHubConfig) *GitHubServiceV2 {
	return &GitHubServiceV2{
		config: config,
	}
}

/* the implementation of the github service interface
type GitHubServiceInterface interface {
	CreatePullRequest(ctx context.Context, repo, base, head, title, body string) (*entity.PullRequest, error)
	UpdatePullRequest(ctx context.Context, repo string, prNumber int, updates map[string]interface{}) error
	GetPullRequest(ctx context.Context, repo string, prNumber int) (*entity.PullRequest, error)
}
*/
