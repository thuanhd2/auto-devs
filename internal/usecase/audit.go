package usecase

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuditService interface {
	LogProjectOperation(ctx context.Context, action entity.AuditAction, projectID uuid.UUID, oldProject, newProject *entity.Project, description string) error
	LogTaskOperation(ctx context.Context, action entity.AuditAction, taskID uuid.UUID, oldTask, newTask *entity.Task, description string) error
	GetAuditLogs(ctx context.Context, entityType string, entityID *uuid.UUID, limit int) ([]*entity.AuditLog, error)
}

type auditService struct {
	auditRepo repository.AuditRepository
}

func NewAuditService(auditRepo repository.AuditRepository) AuditService {
	return &auditService{
		auditRepo: auditRepo,
	}
}

func (s *auditService) LogProjectOperation(ctx context.Context, action entity.AuditAction, projectID uuid.UUID, oldProject, newProject *entity.Project, description string) error {
	return s.logOperation(ctx, "project", projectID, action, oldProject, newProject, description)
}

func (s *auditService) LogTaskOperation(ctx context.Context, action entity.AuditAction, taskID uuid.UUID, oldTask, newTask *entity.Task, description string) error {
	return s.logOperation(ctx, "task", taskID, action, oldTask, newTask, description)
}

func (s *auditService) logOperation(ctx context.Context, entityType string, entityID uuid.UUID, action entity.AuditAction, oldEntity, newEntity interface{}, description string) error {
	auditLog := &entity.AuditLog{
		ID:          uuid.New(),
		EntityType:  entityType,
		EntityID:    entityID,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
	}

	// Extract user info from Gin context if available
	if ginCtx, ok := ctx.(*gin.Context); ok {
		auditLog.IPAddress = getClientIP(ginCtx)
		auditLog.UserAgent = ginCtx.GetHeader("User-Agent")

		// Extract user info from context (you'll need to implement authentication first)
		if userID, exists := ginCtx.Get("user_id"); exists {
			if id, ok := userID.(uuid.UUID); ok {
				auditLog.UserID = &id
			}
		}
		if username, exists := ginCtx.Get("username"); exists {
			if name, ok := username.(string); ok {
				auditLog.Username = name
			}
		}
	}

	// Serialize old and new values
	if oldEntity != nil {
		if oldJSON, err := json.Marshal(oldEntity); err == nil {
			auditLog.OldValues = string(oldJSON)
		}
	}
	if newEntity != nil {
		if newJSON, err := json.Marshal(newEntity); err == nil {
			auditLog.NewValues = string(newJSON)
		}
	}

	return s.auditRepo.Create(ctx, auditLog)
}

func (s *auditService) GetAuditLogs(ctx context.Context, entityType string, entityID *uuid.UUID, limit int) ([]*entity.AuditLog, error) {
	return s.auditRepo.GetByEntity(ctx, entityType, entityID, limit)
}

// getClientIP extracts the real client IP from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first IP if multiple are present
		if ip := net.ParseIP(xff); ip != nil {
			return xff
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	if ip, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
		return ip
	}

	return c.Request.RemoteAddr
}
