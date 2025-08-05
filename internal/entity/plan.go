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

// PlanVersion represents a version of a plan for tracking changes
type PlanVersion struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PlanID    uuid.UUID      `json:"plan_id" gorm:"type:uuid;not null" validate:"required"`
	Version   int            `json:"version" gorm:"not null" validate:"required,min=1"`
	Content   string         `json:"content" gorm:"type:text;not null" validate:"required"`
	CreatedBy string         `json:"created_by" gorm:"size:255;not null" validate:"required"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Plan Plan `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
}

// PlanVersionComparison represents a comparison between two plan versions
type PlanVersionComparison struct {
	PlanID      uuid.UUID `json:"plan_id"`
	FromVersion int       `json:"from_version"`
	ToVersion   int       `json:"to_version"`
	Differences []string  `json:"differences"`
	ChangedAt   time.Time `json:"changed_at"`
}

// PlanStatistics represents statistics for plans in a project
type PlanStatistics struct {
	ProjectID            uuid.UUID                   `json:"project_id"`
	TotalPlans           int                         `json:"total_plans"`
	StatusDistribution   map[PlanStatus]int          `json:"status_distribution"`
	AverageContentLength float64                     `json:"average_content_length"`
	PlansWithVersions    int                         `json:"plans_with_versions"`
	MostActiveTask       *uuid.UUID                  `json:"most_active_task,omitempty"`
	GeneratedAt          time.Time                   `json:"generated_at"`
}