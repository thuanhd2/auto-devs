package entity

import (
	"time"

	"github.com/google/uuid"
)

type AuditAction string

const (
	AuditActionCreate  AuditAction = "CREATE"
	AuditActionUpdate  AuditAction = "UPDATE"
	AuditActionDelete  AuditAction = "DELETE"
	AuditActionArchive AuditAction = "ARCHIVE"
	AuditActionRestore AuditAction = "RESTORE"
)

type AuditLog struct {
	ID           uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EntityType   string      `json:"entity_type" gorm:"size:50;not null"`
	EntityID     uuid.UUID   `json:"entity_id" gorm:"type:uuid;not null"`
	Action       AuditAction `json:"action" gorm:"size:20;not null"`
	UserID       *uuid.UUID  `json:"user_id,omitempty" gorm:"type:uuid"`
	Username     string      `json:"username" gorm:"size:255"`
	IPAddress    string      `json:"ip_address" gorm:"size:45"`
	UserAgent    string      `json:"user_agent" gorm:"size:500"`
	OldValues    string      `json:"old_values,omitempty" gorm:"type:jsonb"`
	NewValues    string      `json:"new_values,omitempty" gorm:"type:jsonb"`
	Description  string      `json:"description" gorm:"size:500"`
	CreatedAt    time.Time   `json:"created_at" gorm:"autoCreateTime"`
}