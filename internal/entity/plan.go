package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlanStatus represents the status of a plan
type PlanStatus string

const (
	PlanStatusDRAFT     PlanStatus = "DRAFT"
	PlanStatusREVIEWING PlanStatus = "REVIEWING"
	PlanStatusAPPROVED  PlanStatus = "APPROVED"
	PlanStatusREJECTED  PlanStatus = "REJECTED"
)

// IsValid checks if the plan status is valid
func (ps PlanStatus) IsValid() bool {
	switch ps {
	case PlanStatusDRAFT, PlanStatusREVIEWING, PlanStatusAPPROVED, PlanStatusREJECTED:
		return true
	default:
		return false
	}
}

// String returns the string representation of PlanStatus
func (ps PlanStatus) String() string {
	return string(ps)
}

// GetDisplayName returns a user-friendly display name for the status
func (ps PlanStatus) GetDisplayName() string {
	switch ps {
	case PlanStatusDRAFT:
		return "Draft"
	case PlanStatusREVIEWING:
		return "Reviewing"
	case PlanStatusAPPROVED:
		return "Approved"
	case PlanStatusREJECTED:
		return "Rejected"
	default:
		return string(ps)
	}
}

// GetAllPlanStatuses returns all valid plan statuses
func GetAllPlanStatuses() []PlanStatus {
	return []PlanStatus{
		PlanStatusDRAFT,
		PlanStatusREVIEWING,
		PlanStatusAPPROVED,
		PlanStatusREJECTED,
	}
}

// Plan represents a plan for a task stored as markdown content
type Plan struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID    uuid.UUID      `json:"task_id" gorm:"type:uuid;not null" validate:"required"`
	Status    PlanStatus     `json:"status" gorm:"size:50;not null;default:'DRAFT'" validate:"required,oneof=DRAFT REVIEWING APPROVED REJECTED"`
	Content   string         `json:"content" gorm:"type:text;not null" validate:"required"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Task Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}