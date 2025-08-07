package repository

import (
	"context"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// PullRequestRepository defines the interface for pull request data operations
type PullRequestRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, pr *entity.PullRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PullRequest, error)
	Update(ctx context.Context, pr *entity.PullRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Query operations
	GetByTaskID(ctx context.Context, taskID uuid.UUID) (*entity.PullRequest, error)
	GetByGitHubPRNumber(ctx context.Context, repo string, prNumber int) (*entity.PullRequest, error)
	GetByRepository(ctx context.Context, repo string) ([]*entity.PullRequest, error)
	GetByStatus(ctx context.Context, status entity.PullRequestStatus) ([]*entity.PullRequest, error)
	
	// Monitoring operations
	GetActiveMonitoringPRs(ctx context.Context) ([]*entity.PullRequest, error)
	GetOpenPRs(ctx context.Context) ([]*entity.PullRequest, error)
	
	// List operations with pagination
	List(ctx context.Context, offset, limit int) ([]*entity.PullRequest, error)
	ListByProjectID(ctx context.Context, projectID uuid.UUID, offset, limit int) ([]*entity.PullRequest, error)
}