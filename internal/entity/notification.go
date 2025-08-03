package entity

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents different types of notifications
type NotificationType string

const (
	NotificationTypeTaskStatusChanged NotificationType = "TASK_STATUS_CHANGED"
	NotificationTypeTaskCreated      NotificationType = "TASK_CREATED"
	NotificationTypeTaskUpdated      NotificationType = "TASK_UPDATED"
	NotificationTypeTaskDeleted      NotificationType = "TASK_DELETED"
)

// NotificationEvent represents a notification event
type NotificationEvent struct {
	ID        uuid.UUID        `json:"id"`
	Type      NotificationType `json:"type"`
	ProjectID uuid.UUID        `json:"project_id"`
	TaskID    *uuid.UUID       `json:"task_id,omitempty"`
	UserID    *string          `json:"user_id,omitempty"`
	Message   string           `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

// TaskStatusChangeNotificationData represents specific data for task status change notifications
type TaskStatusChangeNotificationData struct {
	TaskID       uuid.UUID   `json:"task_id"`
	TaskTitle    string      `json:"task_title"`
	FromStatus   *TaskStatus `json:"from_status,omitempty"`
	ToStatus     TaskStatus  `json:"to_status"`
	ChangedBy    *string     `json:"changed_by,omitempty"`
	Reason       *string     `json:"reason,omitempty"`
	ProjectID    uuid.UUID   `json:"project_id"`
	ProjectName  string      `json:"project_name"`
}

// NotificationHandler defines the interface for handling notifications
type NotificationHandler interface {
	HandleNotification(event NotificationEvent) error
}

// NotificationService defines the interface for the notification service
type NotificationService interface {
	SendNotification(event NotificationEvent) error
	RegisterHandler(notificationType NotificationType, handler NotificationHandler) error
	UnregisterHandler(notificationType NotificationType) error
}