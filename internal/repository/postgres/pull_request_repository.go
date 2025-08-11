package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PullRequestRepository implements the pull request repository interface using PostgreSQL
type pullRequestRepository struct {
	db *database.GormDB
}

// NewPullRequestRepository creates a new pull request repository
func NewPullRequestRepository(db *database.GormDB) repository.PullRequestRepository {
	return &pullRequestRepository{
		db: db,
	}
}

// Create creates a new pull request
func (r *pullRequestRepository) Create(ctx context.Context, pr *entity.PullRequest) error {
	if pr == nil {
		return fmt.Errorf("pull request cannot be nil")
	}

	result := r.db.WithContext(ctx).Create(pr)
	if result.Error != nil {
		return fmt.Errorf("failed to create pull request: %w", result.Error)
	}

	return nil
}

// GetByID retrieves a pull request by ID
func (r *pullRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.PullRequest, error) {
	var pr entity.PullRequest
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&pr)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("pull request not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get pull request: %w", result.Error)
	}

	return &pr, nil
}

// Update updates an existing pull request
func (r *pullRequestRepository) Update(ctx context.Context, pr *entity.PullRequest) error {
	if pr == nil {
		return fmt.Errorf("pull request cannot be nil")
	}

	result := r.db.WithContext(ctx).Save(pr)
	if result.Error != nil {
		return fmt.Errorf("failed to update pull request: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("pull request not found: %s", pr.ID)
	}

	return nil
}

// Delete deletes a pull request
func (r *pullRequestRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.PullRequest{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete pull request: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("pull request not found: %s", id)
	}

	return nil
}

// GetByTaskID retrieves a pull request by task ID
func (r *pullRequestRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.PullRequest, error) {
	var pr entity.PullRequest
	result := r.db.WithContext(ctx).Where("task_id = ?", taskID).First(&pr)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No PR found for task (which is valid)
		}
		return nil, fmt.Errorf("failed to get pull request by task ID: %w", result.Error)
	}

	return &pr, nil
}

// GetByGitHubPRNumber retrieves a pull request by GitHub PR number and repository
func (r *pullRequestRepository) GetByGitHubPRNumber(ctx context.Context, repo string, prNumber int) (*entity.PullRequest, error) {
	var pr entity.PullRequest
	result := r.db.WithContext(ctx).Where("repository = ? AND github_pr_number = ?", repo, prNumber).First(&pr)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("pull request not found: repo=%s, pr=%d", repo, prNumber)
		}
		return nil, fmt.Errorf("failed to get pull request by GitHub PR number: %w", result.Error)
	}

	return &pr, nil
}

// GetByRepository retrieves all pull requests for a repository
func (r *pullRequestRepository) GetByRepository(ctx context.Context, repo string) ([]*entity.PullRequest, error) {
	var prs []*entity.PullRequest
	result := r.db.WithContext(ctx).Where("repository = ?", repo).Order("created_at DESC").Find(&prs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get pull requests by repository: %w", result.Error)
	}

	return prs, nil
}

// GetByStatus retrieves pull requests by status
func (r *pullRequestRepository) GetByStatus(ctx context.Context, status entity.PullRequestStatus) ([]*entity.PullRequest, error) {
	var prs []*entity.PullRequest
	result := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC").Find(&prs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get pull requests by status: %w", result.Error)
	}

	return prs, nil
}

// GetActiveMonitoringPRs retrieves pull requests that should be actively monitored
func (r *pullRequestRepository) GetActiveMonitoringPRs(ctx context.Context) ([]*entity.PullRequest, error) {
	var prs []*entity.PullRequest

	// Get all open PRs that should be monitored
	result := r.db.WithContext(ctx).
		Where("status = ?", entity.PullRequestStatusOpen).
		Order("created_at DESC").
		Find(&prs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get active monitoring PRs: %w", result.Error)
	}

	return prs, nil
}

// GetOpenPRs retrieves all open pull requests
func (r *pullRequestRepository) GetOpenPRs(ctx context.Context) ([]*entity.PullRequest, error) {
	var prs []*entity.PullRequest
	result := r.db.WithContext(ctx).Where("status = ?", entity.PullRequestStatusOpen).Order("created_at DESC").Find(&prs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get open pull requests: %w", result.Error)
	}

	return prs, nil
}

// List retrieves pull requests with pagination
func (r *pullRequestRepository) List(ctx context.Context, offset, limit int) ([]*entity.PullRequest, error) {
	var prs []*entity.PullRequest
	result := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&prs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", result.Error)
	}

	return prs, nil
}

// ListByProjectID retrieves pull requests by project ID with pagination
func (r *pullRequestRepository) ListByProjectID(ctx context.Context, projectID uuid.UUID, offset, limit int) ([]*entity.PullRequest, error) {
	var prs []*entity.PullRequest

	// Join with tasks to filter by project ID
	result := r.db.WithContext(ctx).
		Joins("JOIN tasks ON tasks.id = pull_requests.task_id").
		Where("tasks.project_id = ?", projectID).
		Order("pull_requests.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&prs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to list pull requests by project ID: %w", result.Error)
	}

	return prs, nil
}
