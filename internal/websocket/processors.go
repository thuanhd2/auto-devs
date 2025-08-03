package websocket

import (
	"log"

	"github.com/google/uuid"
)

// TaskEventProcessor handles task-related WebSocket messages
type TaskEventProcessor struct {
	hub *Hub
}

// NewTaskEventProcessor creates a new task event processor
func NewTaskEventProcessor(hub *Hub) *TaskEventProcessor {
	return &TaskEventProcessor{
		hub: hub,
	}
}

// ProcessMessage processes task-related messages
func (p *TaskEventProcessor) ProcessMessage(conn *Connection, message *Message) error {
	switch message.Type {
	case TaskCreated, TaskUpdated, TaskDeleted:
		return p.handleTaskEvent(conn, message)
	default:
		return ErrProcessingFailed
	}
}

// handleTaskEvent processes task events and broadcasts them
func (p *TaskEventProcessor) handleTaskEvent(conn *Connection, message *Message) error {
	var taskData TaskData
	if err := message.ParseData(&taskData); err != nil {
		return err
	}

	// Broadcast to all connections subscribed to the project
	p.hub.BroadcastToProject(message, taskData.ProjectID, conn)

	log.Printf("Task event broadcasted: %s for task %s in project %s",
		message.Type, taskData.TaskID, taskData.ProjectID)

	return nil
}

// BroadcastTaskCreated broadcasts a task created event
func (p *TaskEventProcessor) BroadcastTaskCreated(task interface{}, projectID uuid.UUID, excludeConn *Connection) error {
	data := TaskData{
		TaskID:    uuid.New(), // This should come from the actual task
		ProjectID: projectID,
		Task:      task,
	}

	message, err := NewMessage(TaskCreated, data)
	if err != nil {
		return err
	}

	p.hub.BroadcastToProject(message, projectID, excludeConn)
	return nil
}

// BroadcastTaskUpdated broadcasts a task updated event
func (p *TaskEventProcessor) BroadcastTaskUpdated(taskID, projectID uuid.UUID, changes map[string]interface{}, task interface{}, excludeConn *Connection) error {
	data := TaskData{
		TaskID:    taskID,
		ProjectID: projectID,
		Changes:   changes,
		Task:      task,
	}

	message, err := NewMessage(TaskUpdated, data)
	if err != nil {
		return err
	}

	p.hub.BroadcastToProject(message, projectID, excludeConn)
	return nil
}

// BroadcastTaskDeleted broadcasts a task deleted event
func (p *TaskEventProcessor) BroadcastTaskDeleted(taskID, projectID uuid.UUID, excludeConn *Connection) error {
	data := TaskData{
		TaskID:    taskID,
		ProjectID: projectID,
	}

	message, err := NewMessage(TaskDeleted, data)
	if err != nil {
		return err
	}

	p.hub.BroadcastToProject(message, projectID, excludeConn)
	return nil
}

// ProjectEventProcessor handles project-related WebSocket messages
type ProjectEventProcessor struct {
	hub *Hub
}

// NewProjectEventProcessor creates a new project event processor
func NewProjectEventProcessor(hub *Hub) *ProjectEventProcessor {
	return &ProjectEventProcessor{
		hub: hub,
	}
}

// ProcessMessage processes project-related messages
func (p *ProjectEventProcessor) ProcessMessage(conn *Connection, message *Message) error {
	switch message.Type {
	case ProjectUpdated:
		return p.handleProjectEvent(conn, message)
	default:
		return ErrProcessingFailed
	}
}

// handleProjectEvent processes project events and broadcasts them
func (p *ProjectEventProcessor) handleProjectEvent(conn *Connection, message *Message) error {
	var projectData ProjectData
	if err := message.ParseData(&projectData); err != nil {
		return err
	}

	// Broadcast to all connections subscribed to the project
	p.hub.BroadcastToProject(message, projectData.ProjectID, conn)

	log.Printf("Project event broadcasted: %s for project %s",
		message.Type, projectData.ProjectID)

	return nil
}

// BroadcastProjectUpdated broadcasts a project updated event
func (p *ProjectEventProcessor) BroadcastProjectUpdated(projectID uuid.UUID, changes map[string]interface{}, project interface{}, excludeConn *Connection) error {
	data := ProjectData{
		ProjectID: projectID,
		Changes:   changes,
		Project:   project,
	}

	message, err := NewMessage(ProjectUpdated, data)
	if err != nil {
		return err
	}

	p.hub.BroadcastToProject(message, projectID, excludeConn)
	return nil
}

// StatusEventProcessor handles status change events
type StatusEventProcessor struct {
	hub *Hub
}

// NewStatusEventProcessor creates a new status event processor
func NewStatusEventProcessor(hub *Hub) *StatusEventProcessor {
	return &StatusEventProcessor{
		hub: hub,
	}
}

// ProcessMessage processes status change messages
func (p *StatusEventProcessor) ProcessMessage(conn *Connection, message *Message) error {
	switch message.Type {
	case StatusChanged:
		return p.handleStatusEvent(conn, message)
	default:
		return ErrProcessingFailed
	}
}

// handleStatusEvent processes status change events and broadcasts them
func (p *StatusEventProcessor) handleStatusEvent(conn *Connection, message *Message) error {
	var statusData StatusData
	if err := message.ParseData(&statusData); err != nil {
		return err
	}

	// Broadcast to all connections subscribed to the project
	p.hub.BroadcastToProject(message, statusData.ProjectID, conn)

	log.Printf("Status change broadcasted: %s changed from %s to %s in project %s",
		statusData.EntityType, statusData.OldStatus, statusData.NewStatus, statusData.ProjectID)

	return nil
}

// BroadcastStatusChanged broadcasts a status change event
func (p *StatusEventProcessor) BroadcastStatusChanged(entityID, projectID uuid.UUID, entityType, oldStatus, newStatus string, excludeConn *Connection) error {
	data := StatusData{
		EntityID:   entityID,
		EntityType: entityType,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
		ProjectID:  projectID,
	}

	message, err := NewMessage(StatusChanged, data)
	if err != nil {
		return err
	}

	p.hub.BroadcastToProject(message, projectID, excludeConn)
	return nil
}

// UserPresenceProcessor handles user presence events
type UserPresenceProcessor struct {
	hub *Hub
}

// NewUserPresenceProcessor creates a new user presence processor
func NewUserPresenceProcessor(hub *Hub) *UserPresenceProcessor {
	return &UserPresenceProcessor{
		hub: hub,
	}
}

// ProcessMessage processes user presence messages
func (p *UserPresenceProcessor) ProcessMessage(conn *Connection, message *Message) error {
	switch message.Type {
	case UserJoined, UserLeft:
		return p.handlePresenceEvent(conn, message)
	default:
		return ErrProcessingFailed
	}
}

// handlePresenceEvent processes user presence events and broadcasts them
func (p *UserPresenceProcessor) handlePresenceEvent(conn *Connection, message *Message) error {
	var presenceData UserPresenceData
	if err := message.ParseData(&presenceData); err != nil {
		return err
	}

	// Broadcast to all connections subscribed to the project
	p.hub.BroadcastToProject(message, presenceData.ProjectID, conn)

	log.Printf("User presence broadcasted: %s %s project %s",
		presenceData.UserID, presenceData.Action, presenceData.ProjectID)

	return nil
}

// BroadcastUserJoined broadcasts a user joined event
func (p *UserPresenceProcessor) BroadcastUserJoined(userID string, projectID uuid.UUID, excludeConn *Connection) error {
	data := UserPresenceData{
		UserID:    userID,
		ProjectID: projectID,
		Action:    "joined",
	}

	message, err := NewMessage(UserJoined, data)
	if err != nil {
		return err
	}

	p.hub.BroadcastToProject(message, projectID, excludeConn)
	return nil
}

// BroadcastUserLeft broadcasts a user left event
func (p *UserPresenceProcessor) BroadcastUserLeft(userID string, projectID uuid.UUID, excludeConn *Connection) error {
	data := UserPresenceData{
		UserID:    userID,
		ProjectID: projectID,
		Action:    "left",
	}

	message, err := NewMessage(UserLeft, data)
	if err != nil {
		return err
	}

	p.hub.BroadcastToProject(message, projectID, excludeConn)
	return nil
}

// SubscriptionProcessor handles subscription management messages
type SubscriptionProcessor struct {
	hub *Hub
}

// NewSubscriptionProcessor creates a new subscription processor
func NewSubscriptionProcessor(hub *Hub) *SubscriptionProcessor {
	return &SubscriptionProcessor{
		hub: hub,
	}
}

// SubscriptionMessage represents a subscription request
type SubscriptionMessage struct {
	Action    string    `json:"action"` // "subscribe" or "unsubscribe"
	ProjectID uuid.UUID `json:"project_id"`
}

// ProcessMessage processes subscription messages
func (p *SubscriptionProcessor) ProcessMessage(conn *Connection, message *Message) error {
	var subMsg SubscriptionMessage
	if err := message.ParseData(&subMsg); err != nil {
		return err
	}

	switch subMsg.Action {
	case "subscribe":
		p.hub.SubscribeConnectionToProject(conn, subMsg.ProjectID)

		// Send user joined notification
		if userID := conn.GetUserID(); userID != "" {
			presenceProcessor := NewUserPresenceProcessor(p.hub)
			presenceProcessor.BroadcastUserJoined(userID, subMsg.ProjectID, conn)
		}

	case "unsubscribe":
		p.hub.UnsubscribeConnectionFromProject(conn, subMsg.ProjectID)

		// Send user left notification
		if userID := conn.GetUserID(); userID != "" {
			presenceProcessor := NewUserPresenceProcessor(p.hub)
			presenceProcessor.BroadcastUserLeft(userID, subMsg.ProjectID, conn)
		}

	default:
		return ErrProcessingFailed
	}

	return nil
}

// GetEventProcessors returns all configured event processors
func GetEventProcessors(hub *Hub) map[MessageType]MessageProcessor {
	processors := make(map[MessageType]MessageProcessor)

	// Task event processor
	taskProcessor := NewTaskEventProcessor(hub)
	processors[TaskCreated] = taskProcessor
	processors[TaskUpdated] = taskProcessor
	processors[TaskDeleted] = taskProcessor

	// Project event processor
	projectProcessor := NewProjectEventProcessor(hub)
	processors[ProjectUpdated] = projectProcessor

	// Status event processor
	statusProcessor := NewStatusEventProcessor(hub)
	processors[StatusChanged] = statusProcessor

	// User presence processor
	presenceProcessor := NewUserPresenceProcessor(hub)
	processors[UserJoined] = presenceProcessor
	processors[UserLeft] = presenceProcessor

	// Auth processor
	authService := NewMockAuthService()
	authProcessor := NewAuthProcessor(authService, hub)
	processors[AuthRequired] = authProcessor

	// Subscription processor (handled in connection's handleIncomingMessage method)
	// subscriptionProcessor := NewSubscriptionProcessor(hub)

	return processors
}
