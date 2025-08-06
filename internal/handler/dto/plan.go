package dto

import (
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

type PlanResponse struct {
	ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	TaskID    uuid.UUID `json:"task_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Content   string    `json:"content" example:"# Plan\n\nThis is a plan for a task"`
	Status    string    `json:"status" example:"DRAFT"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

func (p *PlanResponse) FromEntity(plan *entity.Plan) {
	p.ID = plan.ID
	p.TaskID = plan.TaskID
	p.Content = plan.Content
	p.Status = string(plan.Status)
	p.CreatedAt = plan.CreatedAt
	p.UpdatedAt = plan.UpdatedAt
}
