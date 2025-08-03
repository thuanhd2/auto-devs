package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/uuid"
)

type auditRepository struct {
	db *database.GormDB
}

func NewAuditRepository(db *database.GormDB) repository.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, auditLog *entity.AuditLog) error {
	if auditLog.ID == uuid.Nil {
		auditLog.ID = uuid.New()
	}

	result := r.db.WithContext(ctx).Create(auditLog)
	if result.Error != nil {
		return fmt.Errorf("failed to create audit log: %w", result.Error)
	}

	return nil
}

func (r *auditRepository) GetByEntity(ctx context.Context, entityType string, entityID *uuid.UUID, limit int) ([]*entity.AuditLog, error) {
	var auditLogs []entity.AuditLog

	query := r.db.WithContext(ctx).Where("entity_type = ?", entityType)
	
	if entityID != nil {
		query = query.Where("entity_id = ?", *entityID)
	}

	if limit <= 0 {
		limit = 100
	}

	result := query.Order("created_at DESC").Limit(limit).Find(&auditLogs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", result.Error)
	}

	// Convert to slice of pointers
	auditLogPtrs := make([]*entity.AuditLog, len(auditLogs))
	for i := range auditLogs {
		auditLogPtrs[i] = &auditLogs[i]
	}

	return auditLogPtrs, nil
}

func (r *auditRepository) GetByTimeRange(ctx context.Context, entityType string, startTime, endTime *time.Time, limit int) ([]*entity.AuditLog, error) {
	var auditLogs []entity.AuditLog

	query := r.db.WithContext(ctx).Where("entity_type = ?", entityType)
	
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	if limit <= 0 {
		limit = 100
	}

	result := query.Order("created_at DESC").Limit(limit).Find(&auditLogs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", result.Error)
	}

	// Convert to slice of pointers
	auditLogPtrs := make([]*entity.AuditLog, len(auditLogs))
	for i := range auditLogs {
		auditLogPtrs[i] = &auditLogs[i]
	}

	return auditLogPtrs, nil
}

func (r *auditRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entity.AuditLog, error) {
	var auditLogs []entity.AuditLog

	if limit <= 0 {
		limit = 100
	}

	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&auditLogs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", result.Error)
	}

	// Convert to slice of pointers
	auditLogPtrs := make([]*entity.AuditLog, len(auditLogs))
	for i := range auditLogs {
		auditLogPtrs[i] = &auditLogs[i]
	}

	return auditLogPtrs, nil
}