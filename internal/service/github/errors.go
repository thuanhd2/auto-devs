package github

import (
	"errors"
	"fmt"
)

// Error types for GitHub service
var (
	ErrInvalidToken        = errors.New("invalid GitHub token")
	ErrRepositoryNotFound  = errors.New("repository not found")
	ErrBranchNotFound      = errors.New("branch not found")
	ErrFileNotFound        = errors.New("file not found")
	ErrPullRequestNotFound = errors.New("pull request not found")
	ErrRateLimitExceeded   = errors.New("GitHub API rate limit exceeded")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrForbidden           = errors.New("access forbidden")
	ErrConflict            = errors.New("resource conflict")
	ErrValidationFailed    = errors.New("validation failed")
)

// GitHubError represents a GitHub API error
type GitHubError struct {
	StatusCode       int
	Message          string
	DocumentationURL string
	Errors           []GitHubFieldError
}

// GitHubFieldError represents a field-specific error
type GitHubFieldError struct {
	Resource string `json:"resource"`
	Field    string `json:"field"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

// Error implements the error interface
func (ge *GitHubError) Error() string {
	if len(ge.Errors) > 0 {
		return fmt.Sprintf("GitHub API error (HTTP %d): %s - %v", ge.StatusCode, ge.Message, ge.Errors)
	}
	return fmt.Sprintf("GitHub API error (HTTP %d): %s", ge.StatusCode, ge.Message)
}

// IsNotFound checks if the error is a 404 Not Found error
func (ge *GitHubError) IsNotFound() bool {
	return ge.StatusCode == 404
}

// IsUnauthorized checks if the error is a 401 Unauthorized error
func (ge *GitHubError) IsUnauthorized() bool {
	return ge.StatusCode == 401
}

// IsForbidden checks if the error is a 403 Forbidden error
func (ge *GitHubError) IsForbidden() bool {
	return ge.StatusCode == 403
}

// IsRateLimit checks if the error is a rate limit error
func (ge *GitHubError) IsRateLimit() bool {
	return ge.StatusCode == 403 && (ge.Message == "API rate limit exceeded" || 
		ge.Message == "You have exceeded a secondary rate limit")
}

// IsConflict checks if the error is a 409 Conflict error
func (ge *GitHubError) IsConflict() bool {
	return ge.StatusCode == 409
}

// IsValidationError checks if the error is a 422 Validation error
func (ge *GitHubError) IsValidationError() bool {
	return ge.StatusCode == 422
}

// AuthenticationError represents authentication-related errors
type AuthenticationError struct {
	Message string
	Cause   error
}

// Error implements the error interface
func (ae *AuthenticationError) Error() string {
	if ae.Cause != nil {
		return fmt.Sprintf("authentication error: %s - %v", ae.Message, ae.Cause)
	}
	return fmt.Sprintf("authentication error: %s", ae.Message)
}

// Unwrap returns the underlying error
func (ae *AuthenticationError) Unwrap() error {
	return ae.Cause
}

// RepositoryError represents repository-related errors
type RepositoryError struct {
	Repository string
	Message    string
	Cause      error
}

// Error implements the error interface
func (re *RepositoryError) Error() string {
	if re.Cause != nil {
		return fmt.Sprintf("repository error [%s]: %s - %v", re.Repository, re.Message, re.Cause)
	}
	return fmt.Sprintf("repository error [%s]: %s", re.Repository, re.Message)
}

// Unwrap returns the underlying error
func (re *RepositoryError) Unwrap() error {
	return re.Cause
}

// BranchError represents branch-related errors
type BranchError struct {
	Repository string
	Branch     string
	Message    string
	Cause      error
}

// Error implements the error interface
func (be *BranchError) Error() string {
	if be.Cause != nil {
		return fmt.Sprintf("branch error [%s:%s]: %s - %v", be.Repository, be.Branch, be.Message, be.Cause)
	}
	return fmt.Sprintf("branch error [%s:%s]: %s", be.Repository, be.Branch, be.Message)
}

// Unwrap returns the underlying error
func (be *BranchError) Unwrap() error {
	return be.Cause
}

// PullRequestError represents pull request-related errors
type PullRequestError struct {
	Repository string
	PRNumber   int
	Message    string
	Cause      error
}

// Error implements the error interface
func (pre *PullRequestError) Error() string {
	if pre.Cause != nil {
		return fmt.Sprintf("pull request error [%s#%d]: %s - %v", pre.Repository, pre.PRNumber, pre.Message, pre.Cause)
	}
	return fmt.Sprintf("pull request error [%s#%d]: %s", pre.Repository, pre.PRNumber, pre.Message)
}

// Unwrap returns the underlying error
func (pre *PullRequestError) Unwrap() error {
	return pre.Cause
}

// RateLimitError represents rate limit errors
type RateLimitError struct {
	Limit     int
	Remaining int
	ResetAt   string
	Message   string
}

// Error implements the error interface
func (rle *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit error: %s (limit: %d, remaining: %d, resets at: %s)", 
		rle.Message, rle.Limit, rle.Remaining, rle.ResetAt)
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s' with value '%v': %s", ve.Field, ve.Value, ve.Message)
}

// IsGitHubError checks if an error is a GitHubError
func IsGitHubError(err error) (*GitHubError, bool) {
	var ghErr *GitHubError
	if errors.As(err, &ghErr) {
		return ghErr, true
	}
	return nil, false
}

// IsAuthenticationError checks if an error is an AuthenticationError
func IsAuthenticationError(err error) (*AuthenticationError, bool) {
	var authErr *AuthenticationError
	if errors.As(err, &authErr) {
		return authErr, true
	}
	return nil, false
}

// IsRepositoryError checks if an error is a RepositoryError
func IsRepositoryError(err error) (*RepositoryError, bool) {
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return repoErr, true
	}
	return nil, false
}

// IsBranchError checks if an error is a BranchError
func IsBranchError(err error) (*BranchError, bool) {
	var branchErr *BranchError
	if errors.As(err, &branchErr) {
		return branchErr, true
	}
	return nil, false
}

// IsPullRequestError checks if an error is a PullRequestError
func IsPullRequestError(err error) (*PullRequestError, bool) {
	var prErr *PullRequestError
	if errors.As(err, &prErr) {
		return prErr, true
	}
	return nil, false
}

// IsRateLimitError checks if an error is a RateLimitError
func IsRateLimitError(err error) (*RateLimitError, bool) {
	var rlErr *RateLimitError
	if errors.As(err, &rlErr) {
		return rlErr, true
	}
	return nil, false
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) (*ValidationError, bool) {
	var valErr *ValidationError
	if errors.As(err, &valErr) {
		return valErr, true
	}
	return nil, false
}