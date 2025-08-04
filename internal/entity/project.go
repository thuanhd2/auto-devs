package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"size:255;not null" validate:"required,min=1,max=255"`
	Description string         `json:"description" gorm:"size:1000" validate:"max=1000"`
	RepoURL     string         `json:"repo_url" gorm:"size:500;not null" validate:"required,url,max=500"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Git-related fields
	RepositoryURL    string `json:"repository_url" gorm:"column:repository_url;size:500"`
	MainBranch       string `json:"main_branch" gorm:"column:main_branch;default:main;size:100"`
	WorktreeBasePath string `json:"worktree_base_path" gorm:"column:worktree_base_path;size:500"`
	GitAuthMethod    string `json:"git_auth_method" gorm:"column:git_auth_method;size:20"` // "ssh" or "https"
	GitEnabled       bool   `json:"git_enabled" gorm:"column:git_enabled;default:false"`

	// Relationships
	Tasks []Task `json:"tasks,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}
