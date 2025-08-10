package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PullRequestStatus represents the status of a pull request
type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
	PullRequestStatusClosed PullRequestStatus = "CLOSED"
)

// IsValid checks if the pull request status is valid
func (prs PullRequestStatus) IsValid() bool {
	switch prs {
	case PullRequestStatusOpen, PullRequestStatusMerged, PullRequestStatusClosed:
		return true
	default:
		return false
	}
}

// String returns the string representation of PullRequestStatus
func (prs PullRequestStatus) String() string {
	return string(prs)
}

// GetDisplayName returns a user-friendly display name for the status
func (prs PullRequestStatus) GetDisplayName() string {
	switch prs {
	case PullRequestStatusOpen:
		return "Open"
	case PullRequestStatusMerged:
		return "Merged"
	case PullRequestStatusClosed:
		return "Closed"
	default:
		return string(prs)
	}
}

// GetAllPullRequestStatuses returns all valid pull request statuses
func GetAllPullRequestStatuses() []PullRequestStatus {
	return []PullRequestStatus{
		PullRequestStatusOpen,
		PullRequestStatusMerged,
		PullRequestStatusClosed,
	}
}

// PullRequest represents a GitHub pull request associated with a task
type PullRequest struct {
	ID             uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID         uuid.UUID         `json:"task_id" gorm:"type:uuid;not null" validate:"required"`
	GitHubPRNumber int               `json:"github_pr_number" gorm:"column:github_pr_number;not null" validate:"required,min=1"`
	Repository     string            `json:"repository" gorm:"size:255;not null" validate:"required"`
	Title          string            `json:"title" gorm:"size:255;not null" validate:"required,min=1,max=255"`
	Body           string            `json:"body" gorm:"type:text"`
	Status         PullRequestStatus `json:"status" gorm:"size:20;not null;default:'OPEN'" validate:"required,oneof=OPEN MERGED CLOSED"`
	HeadBranch     string            `json:"head_branch" gorm:"size:255;not null" validate:"required"`
	BaseBranch     string            `json:"base_branch" gorm:"size:255;not null;default:'main'" validate:"required"`
	GitHubURL      string            `json:"github_url" gorm:"column:github_url;size:500"`
	MergeCommitSHA *string           `json:"merge_commit_sha,omitempty" gorm:"size:40"`
	MergedAt       *time.Time        `json:"merged_at,omitempty"`
	ClosedAt       *time.Time        `json:"closed_at,omitempty"`
	CreatedBy      *string           `json:"created_by,omitempty" gorm:"size:255"`
	MergedBy       *string           `json:"merged_by,omitempty" gorm:"size:255"`
	Reviewers      []string          `json:"reviewers,omitempty" gorm:"-"` // Will be stored as JSON
	ReviewersJSON  string            `json:"-" gorm:"column:reviewers;type:jsonb"`
	Labels         []string          `json:"labels,omitempty" gorm:"-"` // Will be stored as JSON
	LabelsJSON     string            `json:"-" gorm:"column:labels;type:jsonb"`
	Assignees      []string          `json:"assignees,omitempty" gorm:"-"` // Will be stored as JSON
	AssigneesJSON  string            `json:"-" gorm:"column:assignees;type:jsonb"`
	IsDraft        bool              `json:"is_draft" gorm:"default:false"`
	Mergeable      *bool             `json:"mergeable,omitempty"`
	MergeableState *string           `json:"mergeable_state,omitempty" gorm:"size:50"`
	Additions      *int              `json:"additions,omitempty"`
	Deletions      *int              `json:"deletions,omitempty"`
	ChangedFiles   *int              `json:"changed_files,omitempty"`
	CreatedAt      time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt    `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Task *Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

// PullRequestComment represents comments on a pull request
type PullRequestComment struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PullRequestID uuid.UUID      `json:"pull_request_id" gorm:"type:uuid;not null"`
	GitHubID      *int64         `json:"github_id,omitempty" gorm:"column:github_id;unique"`
	Author        string         `json:"author" gorm:"size:255;not null"`
	Body          string         `json:"body" gorm:"type:text;not null"`
	FilePath      *string        `json:"file_path,omitempty" gorm:"size:500"`
	Line          *int           `json:"line,omitempty"`
	IsResolved    bool           `json:"is_resolved" gorm:"default:false"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	PullRequest *PullRequest `json:"pull_request,omitempty" gorm:"foreignKey:PullRequestID"`
}

// PullRequestReview represents code reviews on a pull request
type PullRequestReview struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PullRequestID uuid.UUID      `json:"pull_request_id" gorm:"type:uuid;not null"`
	GitHubID      *int64         `json:"github_id,omitempty" gorm:"column:github_id;unique"`
	Reviewer      string         `json:"reviewer" gorm:"size:255;not null"`
	State         string         `json:"state" gorm:"size:50;not null"` // APPROVED, CHANGES_REQUESTED, COMMENTED
	Body          *string        `json:"body,omitempty" gorm:"type:text"`
	SubmittedAt   *time.Time     `json:"submitted_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	PullRequest *PullRequest `json:"pull_request,omitempty" gorm:"foreignKey:PullRequestID"`
}

// PullRequestCheck represents CI/CD checks on a pull request
type PullRequestCheck struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PullRequestID uuid.UUID      `json:"pull_request_id" gorm:"type:uuid;not null"`
	CheckName     string         `json:"check_name" gorm:"size:255;not null"`
	Status        string         `json:"status" gorm:"size:50;not null"` // PENDING, SUCCESS, FAILURE, ERROR
	Conclusion    *string        `json:"conclusion,omitempty" gorm:"size:50"`
	DetailsURL    *string        `json:"details_url,omitempty" gorm:"size:500"`
	StartedAt     *time.Time     `json:"started_at,omitempty"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	PullRequest *PullRequest `json:"pull_request,omitempty" gorm:"foreignKey:PullRequestID"`
}

// BeforeCreate GORM hook to convert slices to JSON before saving
func (pr *PullRequest) BeforeCreate(tx *gorm.DB) error {
	return pr.convertSlicesToJSON()
}

// BeforeUpdate GORM hook to convert slices to JSON before updating
func (pr *PullRequest) BeforeUpdate(tx *gorm.DB) error {
	return pr.convertSlicesToJSON()
}

// AfterFind GORM hook to convert JSON to slices after loading
func (pr *PullRequest) AfterFind(tx *gorm.DB) error {
	return pr.convertJSONToSlices()
}

// convertSlicesToJSON converts slice fields to JSON strings
func (pr *PullRequest) convertSlicesToJSON() error {
	if len(pr.Reviewers) > 0 {
		reviewersJSON, err := json.Marshal(pr.Reviewers)
		if err != nil {
			return err
		}
		pr.ReviewersJSON = string(reviewersJSON)
	} else {
		pr.ReviewersJSON = "[]"
	}

	if len(pr.Labels) > 0 {
		labelsJSON, err := json.Marshal(pr.Labels)
		if err != nil {
			return err
		}
		pr.LabelsJSON = string(labelsJSON)
	} else {
		pr.LabelsJSON = "[]"
	}

	if len(pr.Assignees) > 0 {
		assigneesJSON, err := json.Marshal(pr.Assignees)
		if err != nil {
			return err
		}
		pr.AssigneesJSON = string(assigneesJSON)
	} else {
		pr.AssigneesJSON = "[]"
	}

	return nil
}

// convertJSONToSlices converts JSON strings to slice fields
func (pr *PullRequest) convertJSONToSlices() error {
	if pr.ReviewersJSON != "" {
		if err := json.Unmarshal([]byte(pr.ReviewersJSON), &pr.Reviewers); err != nil {
			return err
		}
	}

	if pr.LabelsJSON != "" {
		if err := json.Unmarshal([]byte(pr.LabelsJSON), &pr.Labels); err != nil {
			return err
		}
	}

	if pr.AssigneesJSON != "" {
		if err := json.Unmarshal([]byte(pr.AssigneesJSON), &pr.Assignees); err != nil {
			return err
		}
	}

	return nil
}
