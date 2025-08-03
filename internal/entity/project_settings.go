package entity

import (
	"time"

	"github.com/google/uuid"
)

type ProjectSettings struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID        uuid.UUID `json:"project_id" gorm:"type:uuid;not null;uniqueIndex"`
	AutoArchiveDays  *int      `json:"auto_archive_days,omitempty"`
	NotificationsEnabled bool  `json:"notifications_enabled" gorm:"default:true"`
	EmailNotifications   bool  `json:"email_notifications" gorm:"default:false"`
	SlackWebhookURL      string `json:"slack_webhook_url,omitempty" gorm:"size:500"`
	GitBranch            string `json:"git_branch" gorm:"size:255;default:'main'"`
	GitAutoSync          bool   `json:"git_auto_sync" gorm:"default:false"`
	TaskPrefix           string `json:"task_prefix" gorm:"size:10"`
	CreatedAt            time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}