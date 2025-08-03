package repository

import (
	"context"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

type AuditRepository interface {
	Create(ctx context.Context, auditLog *entity.AuditLog) error
	GetByEntity(ctx context.Context, entityType string, entityID *uuid.UUID, limit int) ([]*entity.AuditLog, error)
	GetByTimeRange(ctx context.Context, entityType string, startTime, endTime *time.Time, limit int) ([]*entity.AuditLog, error)
	GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entity.AuditLog, error)
}