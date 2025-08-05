package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlanStatus represents the status of a plan
type PlanStatus string

const (
	PlanStatusDraft      PlanStatus = "DRAFT"
	PlanStatusReviewing  PlanStatus = "REVIEWING"
	PlanStatusApproved   PlanStatus = "APPROVED"
	PlanStatusRejected   PlanStatus = "REJECTED"  
	PlanStatusExecuting  PlanStatus = "EXECUTING"
	PlanStatusCompleted  PlanStatus = "COMPLETED"
	PlanStatusCancelled  PlanStatus = "CANCELLED"
)

// IsValid checks if the plan status is valid
func (ps PlanStatus) IsValid() bool {
	switch ps {
	case PlanStatusDraft, PlanStatusReviewing, PlanStatusApproved, 
		 PlanStatusRejected, PlanStatusExecuting, PlanStatusCompleted, PlanStatusCancelled:
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
	case PlanStatusDraft:
		return "Draft"
	case PlanStatusReviewing:
		return "Reviewing"
	case PlanStatusApproved:
		return "Approved"
	case PlanStatusRejected:
		return "Rejected"
	case PlanStatusExecuting:
		return "Executing"
	case PlanStatusCompleted:
		return "Completed"
	case PlanStatusCancelled:
		return "Cancelled"
	default:
		return string(ps)
	}
}

// PlanStep represents a single step in a plan
type PlanStep struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Action      string            `json:"action"`
	Parameters  map[string]string `json:"parameters"`
	Order       int               `json:"order"`
	Completed   bool              `json:"completed"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
}

// Plan represents an implementation plan for a task
type Plan struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID      uuid.UUID      `json:"task_id" gorm:"type:uuid;not null" validate:"required"`
	Title       string         `json:"title" gorm:"size:255;not null" validate:"required,min=1,max=255"`
	Description string         `json:"description" gorm:"type:text" validate:"max=5000"`
	Status      PlanStatus     `json:"status" gorm:"size:50;not null;default:'DRAFT'" validate:"required,oneof=DRAFT REVIEWING APPROVED REJECTED EXECUTING COMPLETED CANCELLED"`
	Version     int            `json:"version" gorm:"not null;default:1"`
	
	// JSONB fields for complex data
	Steps       []PlanStep        `json:"steps,omitempty" gorm:"-"`
	StepsJSON   string            `json:"-" gorm:"column:steps;type:jsonb"`
	Context     map[string]string `json:"context,omitempty" gorm:"-"`
	ContextJSON string            `json:"-" gorm:"column:context;type:jsonb"`
	
	// Metadata
	CreatedBy   *string        `json:"created_by,omitempty" gorm:"size:255"`
	ApprovedBy  *string        `json:"approved_by,omitempty" gorm:"size:255"`
	ApprovedAt  *time.Time     `json:"approved_at,omitempty"`
	RejectedBy  *string        `json:"rejected_by,omitempty" gorm:"size:255"`
	RejectedAt  *time.Time     `json:"rejected_at,omitempty"`
	
	// Timestamps
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	
	// Relationships
	Task        Task           `json:"task,omitempty" gorm:"foreignKey:TaskID"`
	Versions    []PlanVersion  `json:"versions,omitempty" gorm:"foreignKey:PlanID"`
}

// PlanVersion represents a version history of a plan
type PlanVersion struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PlanID      uuid.UUID      `json:"plan_id" gorm:"type:uuid;not null" validate:"required"`
	Version     int            `json:"version" gorm:"not null"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Status      PlanStatus     `json:"status" gorm:"size:50;not null"`
	
	// JSONB fields for versioned data
	Steps       []PlanStep        `json:"steps,omitempty" gorm:"-"`
	StepsJSON   string            `json:"-" gorm:"column:steps;type:jsonb"`
	Context     map[string]string `json:"context,omitempty" gorm:"-"`
	ContextJSON string            `json:"-" gorm:"column:context;type:jsonb"`
	
	// Versioning metadata
	CreatedBy   *string        `json:"created_by,omitempty" gorm:"size:255"`
	ChangeLog   string         `json:"change_log,omitempty" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	
	// Relationships
	Plan        Plan           `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
}

// GORM hooks for JSON marshaling/unmarshaling

// BeforeSave marshals Steps and Context to JSON before saving
func (p *Plan) BeforeSave(tx *gorm.DB) error {
	if p.Steps != nil {
		stepsJSON, err := json.Marshal(p.Steps)
		if err != nil {
			return err
		}
		p.StepsJSON = string(stepsJSON)
	}
	
	if p.Context != nil {
		contextJSON, err := json.Marshal(p.Context)
		if err != nil {
			return err
		}
		p.ContextJSON = string(contextJSON)
	}
	
	return nil
}

// AfterFind unmarshals JSON to Steps and Context after loading
func (p *Plan) AfterFind(tx *gorm.DB) error {
	if p.StepsJSON != "" {
		if err := json.Unmarshal([]byte(p.StepsJSON), &p.Steps); err != nil {
			return err
		}
	}
	
	if p.ContextJSON != "" {
		if err := json.Unmarshal([]byte(p.ContextJSON), &p.Context); err != nil {
			return err
		}
	}
	
	return nil
}

// BeforeSave marshals Steps and Context to JSON before saving (PlanVersion)
func (pv *PlanVersion) BeforeSave(tx *gorm.DB) error {
	if pv.Steps != nil {
		stepsJSON, err := json.Marshal(pv.Steps)
		if err != nil {
			return err
		}
		pv.StepsJSON = string(stepsJSON)
	}
	
	if pv.Context != nil {
		contextJSON, err := json.Marshal(pv.Context)
		if err != nil {
			return err
		}
		pv.ContextJSON = string(contextJSON)
	}
	
	return nil
}

// AfterFind unmarshals JSON to Steps and Context after loading (PlanVersion)
func (pv *PlanVersion) AfterFind(tx *gorm.DB) error {
	if pv.StepsJSON != "" {
		if err := json.Unmarshal([]byte(pv.StepsJSON), &pv.Steps); err != nil {
			return err
		}
	}
	
	if pv.ContextJSON != "" {
		if err := json.Unmarshal([]byte(pv.ContextJSON), &pv.Context); err != nil {
			return err
		}
	}
	
	return nil
}

// CreateVersion creates a new version from the current plan
func (p *Plan) CreateVersion(changeLog string, createdBy *string) *PlanVersion {
	return &PlanVersion{
		ID:          uuid.New(),
		PlanID:      p.ID,
		Version:     p.Version,
		Title:       p.Title,
		Description: p.Description,
		Status:      p.Status,
		Steps:       p.Steps,
		Context:     p.Context,
		CreatedBy:   createdBy,
		ChangeLog:   changeLog,
		CreatedAt:   time.Now(),
	}
}

// GetStepByID returns a step by its ID
func (p *Plan) GetStepByID(stepID string) *PlanStep {
	for i := range p.Steps {
		if p.Steps[i].ID == stepID {
			return &p.Steps[i]
		}
	}
	return nil
}

// MarkStepCompleted marks a step as completed
func (p *Plan) MarkStepCompleted(stepID string) bool {
	step := p.GetStepByID(stepID)
	if step != nil {
		step.Completed = true
		now := time.Now()
		step.CompletedAt = &now
		return true
	}
	return false
}

// GetCompletionPercentage returns the completion percentage (0.0 to 1.0)
func (p *Plan) GetCompletionPercentage() float64 {
	if len(p.Steps) == 0 {
		return 0.0
	}
	
	completed := 0
	for _, step := range p.Steps {
		if step.Completed {
			completed++
		}
	}
	
	return float64(completed) / float64(len(p.Steps))
}

// IsFullyCompleted returns true if all steps are completed
func (p *Plan) IsFullyCompleted() bool {
	if len(p.Steps) == 0 {
		return false
	}
	
	for _, step := range p.Steps {
		if !step.Completed {
			return false
		}
	}
	
	return true
}

// Clone creates a deep copy of the plan
func (p *Plan) Clone() *Plan {
	clone := *p
	
	// Deep copy Steps
	if p.Steps != nil {
		clone.Steps = make([]PlanStep, len(p.Steps))
		copy(clone.Steps, p.Steps)
		
		// Deep copy Parameters in each step
		for i := range clone.Steps {
			if p.Steps[i].Parameters != nil {
				clone.Steps[i].Parameters = make(map[string]string)
				for k, v := range p.Steps[i].Parameters {
					clone.Steps[i].Parameters[k] = v
				}
			}
		}
	}
	
	// Deep copy Context
	if p.Context != nil {
		clone.Context = make(map[string]string)
		for k, v := range p.Context {
			clone.Context[k] = v
		}
	}
	
	return &clone
}

// PlanFilters represents filtering options for plans
type PlanFilters struct {
	TaskID       *uuid.UUID
	Statuses     []PlanStatus
	CreatedBy    *string
	CreatedAfter *time.Time
	CreatedBefore *time.Time
	SearchTerm   *string
	Limit        *int
	Offset       *int
	OrderBy      *string // "created_at", "updated_at", "title", "status", "version"
	OrderDir     *string // "asc", "desc"
}