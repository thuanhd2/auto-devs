package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// NotificationUsecase defines the interface for notification operations
type NotificationUsecase interface {
	SendTaskStatusChangeNotification(ctx context.Context, data entity.TaskStatusChangeNotificationData) error
	SendTaskCreatedNotification(ctx context.Context, task *entity.Task, project *entity.Project) error
	RegisterHandler(notificationType entity.NotificationType, handler entity.NotificationHandler) error
	UnregisterHandler(notificationType entity.NotificationType) error
}

type notificationUsecase struct {
	handlers map[entity.NotificationType]entity.NotificationHandler
}

// NewNotificationUsecase creates a new notification usecase
func NewNotificationUsecase() NotificationUsecase {
	return &notificationUsecase{
		handlers: make(map[entity.NotificationType]entity.NotificationHandler),
	}
}

// SendTaskStatusChangeNotification sends a notification when a task status changes
func (n *notificationUsecase) SendTaskStatusChangeNotification(ctx context.Context, data entity.TaskStatusChangeNotificationData) error {
	// Create notification event
	event := entity.NotificationEvent{
		ID:        uuid.New(),
		Type:      entity.NotificationTypeTaskStatusChanged,
		ProjectID: data.ProjectID,
		TaskID:    &data.TaskID,
		UserID:    data.ChangedBy,
		CreatedAt: time.Now(),
	}

	// Create message
	fromStatusStr := "initial"
	if data.FromStatus != nil {
		fromStatusStr = data.FromStatus.GetDisplayName()
	}
	toStatusStr := data.ToStatus.GetDisplayName()
	
	event.Message = fmt.Sprintf("Task '%s' status changed from %s to %s", 
		data.TaskTitle, fromStatusStr, toStatusStr)

	// Add structured data
	dataMap := make(map[string]interface{})
	dataBytes, _ := json.Marshal(data)
	json.Unmarshal(dataBytes, &dataMap)
	event.Data = dataMap

	return n.sendNotification(event)
}

// SendTaskCreatedNotification sends a notification when a task is created
func (n *notificationUsecase) SendTaskCreatedNotification(ctx context.Context, task *entity.Task, project *entity.Project) error {
	event := entity.NotificationEvent{
		ID:        uuid.New(),
		Type:      entity.NotificationTypeTaskCreated,
		ProjectID: task.ProjectID,
		TaskID:    &task.ID,
		Message:   fmt.Sprintf("New task '%s' created in project '%s'", task.Title, project.Name),
		Data: map[string]interface{}{
			"task_id":      task.ID,
			"task_title":   task.Title,
			"project_id":   task.ProjectID,
			"project_name": project.Name,
			"status":       task.Status,
		},
		CreatedAt: time.Now(),
	}

	return n.sendNotification(event)
}

// RegisterHandler registers a handler for a specific notification type
func (n *notificationUsecase) RegisterHandler(notificationType entity.NotificationType, handler entity.NotificationHandler) error {
	n.handlers[notificationType] = handler
	return nil
}

// UnregisterHandler removes a handler for a specific notification type
func (n *notificationUsecase) UnregisterHandler(notificationType entity.NotificationType) error {
	delete(n.handlers, notificationType)
	return nil
}

// sendNotification sends a notification to the appropriate handler
func (n *notificationUsecase) sendNotification(event entity.NotificationEvent) error {
	handler, exists := n.handlers[event.Type]
	if !exists {
		// Log that no handler is registered, but don't return an error
		log.Printf("No handler registered for notification type: %s", event.Type)
		return nil
	}

	return handler.HandleNotification(event)
}

// WebSocketNotificationHandler implements NotificationHandler for WebSocket notifications
type WebSocketNotificationHandler struct {
	// This would integrate with your existing WebSocket service
	// For now, it's a placeholder implementation
}

func NewWebSocketNotificationHandler() *WebSocketNotificationHandler {
	return &WebSocketNotificationHandler{}
}

func (w *WebSocketNotificationHandler) HandleNotification(event entity.NotificationEvent) error {
	// This would send the notification via WebSocket to connected clients
	// For now, just log the notification
	log.Printf("WebSocket Notification: %s - %s", event.Type, event.Message)
	
	// TODO: Integrate with actual WebSocket service
	// wsService.BroadcastToProject(event.ProjectID, "notification", event)
	
	return nil
}

// LogNotificationHandler implements NotificationHandler for logging notifications
type LogNotificationHandler struct{}

func NewLogNotificationHandler() *LogNotificationHandler {
	return &LogNotificationHandler{}
}

func (l *LogNotificationHandler) HandleNotification(event entity.NotificationEvent) error {
	log.Printf("Notification [%s]: %s (Project: %s, Task: %v)", 
		event.Type, event.Message, event.ProjectID, event.TaskID)
	return nil
}