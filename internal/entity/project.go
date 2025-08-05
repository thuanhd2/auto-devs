package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name             string         `json:"name" gorm:"size:255;not null" validate:"required,min=1,max=255"`
	Description      string         `json:"description" gorm:"size:1000" validate:"max=1000"`
	RepositoryURL    string         `json:"repository_url" gorm:"column:repository_url;size:500"`
	WorktreeBasePath string         `json:"worktree_base_path" gorm:"column:worktree_base_path;size:500"`
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Tasks []Task `json:"tasks,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}
