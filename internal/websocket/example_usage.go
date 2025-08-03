package websocket

import (
	"log"
	"time"

	"github.com/google/uuid"
)

// This file contains examples of how to use the WebSocket infrastructure

// ExampleBasicUsage demonstrates basic WebSocket service usage
func ExampleBasicUsage() {
	// Initialize the WebSocket service
	wsService := NewService()

	// Example 1: Send a task creation notification
	taskID := uuid.New()
	projectID := uuid.New()

	task := map[string]interface{}{
		"id":          taskID,
		"project_id":  projectID,
		"title":       "Implement new feature",
		"description": "Add real-time notifications",
		"status":      "TODO",
		"created_at":  time.Now(),
	}

	if err := wsService.NotifyTaskCreated(task, projectID); err != nil {
		log.Printf("Failed to notify task creation: %v", err)
	}

	// Example 2: Send a task update notification
	changes := map[string]interface{}{
		"status": map[string]interface{}{
			"old": "TODO",
			"new": "IN_PROGRESS",
		},
		"updated_at": time.Now(),
	}

	updatedTask := map[string]interface{}{
		"id":          taskID,
		"project_id":  projectID,
		"title":       "Implement new feature",
		"description": "Add real-time notifications",
		"status":      "IN_PROGRESS",
		"updated_at":  time.Now(),
	}

	if err := wsService.NotifyTaskUpdated(taskID, projectID, changes, updatedTask); err != nil {
		log.Printf("Failed to notify task update: %v", err)
	}

	// Example 3: Send a status change notification
	if err := wsService.NotifyStatusChanged(taskID, projectID, "task", "TODO", "IN_PROGRESS"); err != nil {
		log.Printf("Failed to notify status change: %v", err)
	}

	// Example 4: Send a project update notification
	projectChanges := map[string]interface{}{
		"name": map[string]interface{}{
			"old": "Old Project Name",
			"new": "New Project Name",
		},
	}

	project := map[string]interface{}{
		"id":          projectID,
		"name":        "New Project Name",
		"description": "Updated project description",
		"updated_at":  time.Now(),
	}

	if err := wsService.NotifyProjectUpdated(projectID, projectChanges, project); err != nil {
		log.Printf("Failed to notify project update: %v", err)
	}

	// Example 5: Send user presence notifications
	userID := "user-123"
	if err := wsService.NotifyUserJoined(userID, projectID); err != nil {
		log.Printf("Failed to notify user joined: %v", err)
	}

	// Example 6: Send direct message to user
	if err := wsService.SendDirectMessage(userID, TaskCreated, task); err != nil {
		log.Printf("Failed to send direct message: %v", err)
	}

	// Example 7: Send message to all project members
	announcement := map[string]interface{}{
		"message": "Project deployment scheduled for tonight",
		"type":    "announcement",
		"time":    time.Now(),
	}

	if err := wsService.SendProjectMessage(projectID, ProjectUpdated, announcement); err != nil {
		log.Printf("Failed to send project message: %v", err)
	}

	// Example 8: Get connection metrics
	metrics := wsService.GetMetrics()
	log.Printf("WebSocket metrics: %+v", metrics)

	// Example 9: Check connection counts
	totalConnections := wsService.GetConnectionCount()
	projectConnections := wsService.GetProjectConnectionCount(projectID)
	userConnections := wsService.GetUserConnectionCount(userID)

	log.Printf("Total connections: %d", totalConnections)
	log.Printf("Project %s connections: %d", projectID, projectConnections)
	log.Printf("User %s connections: %d", userID, userConnections)
}

// ExampleCustomMessage demonstrates sending custom messages
func ExampleCustomMessage() {
	wsService := NewService()

	// Create a custom message type (you would define this in message.go)
	customData := map[string]interface{}{
		"action":    "system_maintenance",
		"message":   "System will be down for maintenance in 5 minutes",
		"scheduled": time.Now().Add(5 * time.Minute),
		"duration":  "30 minutes",
	}

	// Broadcast to all users
	if err := wsService.BroadcastMessage("system_announcement", customData, nil, nil); err != nil {
		log.Printf("Failed to broadcast system announcement: %v", err)
	}

	// Broadcast to specific project
	projectID := uuid.New()
	if err := wsService.BroadcastMessage("project_announcement", customData, &projectID, nil); err != nil {
		log.Printf("Failed to broadcast project announcement: %v", err)
	}

	// Send to specific user
	userID := "admin-user"
	if err := wsService.BroadcastMessage("admin_alert", customData, nil, &userID); err != nil {
		log.Printf("Failed to send admin alert: %v", err)
	}
}

// CustomProcessor is an example custom message processor
type CustomProcessor struct{}

func (p *CustomProcessor) ProcessMessage(conn *Connection, message *Message) error {
	log.Printf("Processing custom message: %s from connection %s", message.Type, conn.ID)

	// Handle custom logic here
	switch message.Type {
	case "custom_action":
		// Process custom action
		var data map[string]interface{}
		if err := message.ParseData(&data); err != nil {
			return err
		}

		log.Printf("Custom action data: %+v", data)

		// Send response back to client
		response, _ := NewMessage("custom_response", map[string]string{
			"status": "processed",
			"action": "custom_action",
		})
		return conn.SendMessage(response)
	}

	return nil
}

// ExampleMessageHandling demonstrates how to handle different message types
func ExampleMessageHandling() {
	// This would typically be done in your application startup
	wsService := NewService()
	hub := wsService.GetHub()

	// Register custom processor
	customProcessor := &CustomProcessor{}
	hub.RegisterProcessor("custom_action", customProcessor)

	log.Printf("Custom message processor registered")
}

// ExampleConnectionManagement demonstrates connection lifecycle management
func ExampleConnectionManagement() {
	wsService := NewService()

	// Monitor connection health
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if !wsService.IsHealthy() {
				log.Printf("WebSocket service is unhealthy!")
				// Handle unhealthy state
			}

			status := wsService.GetHealthStatus()
			log.Printf("WebSocket health status: %+v", status)
		}
	}()

	// Monitor metrics
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			metrics := wsService.GetMetrics()
			log.Printf("WebSocket metrics: %+v", metrics)

			// Check for any issues
			if hubMetrics, ok := metrics["hub"].(HubMetrics); ok {
				if hubMetrics.ActiveConnections > 1000 {
					log.Printf("Warning: High number of active connections: %d", hubMetrics.ActiveConnections)
				}
			}
		}
	}()
}

// ExampleErrorHandling demonstrates error handling patterns
func ExampleErrorHandling() {
	wsService := NewService()

	// Example of handling various error scenarios
	taskID := uuid.New()
	projectID := uuid.New()

	// 1. Handle task creation with error recovery
	task := map[string]interface{}{
		"id":         taskID,
		"project_id": projectID,
		"title":      "Test task",
	}

	if err := wsService.NotifyTaskCreated(task, projectID); err != nil {
		log.Printf("Task creation notification failed: %v", err)

		// Retry or fallback logic
		time.AfterFunc(5*time.Second, func() {
			if retryErr := wsService.NotifyTaskCreated(task, projectID); retryErr != nil {
				log.Printf("Retry also failed: %v", retryErr)
			} else {
				log.Printf("Retry successful for task creation notification")
			}
		})
	}

	// 2. Handle invalid data gracefully
	invalidData := "invalid-data-structure"
	if err := wsService.BroadcastMessage(TaskCreated, invalidData, &projectID, nil); err != nil {
		log.Printf("Expected error for invalid data: %v", err)
	}

	// 3. Handle network issues
	connectionCount := wsService.GetConnectionCount()
	if connectionCount == 0 {
		log.Printf("No active connections - messages will be queued for offline delivery")
	}
}

// ExampleIntegrationWithHTTPHandlers shows how to integrate with HTTP handlers
func ExampleIntegrationWithHTTPHandlers() {
	// This example shows how the WebSocket notifications would be integrated
	// into your existing HTTP handlers (as implemented in task_websocket.go)

	wsService := NewService()

	// Simulate HTTP handler logic
	simulateTaskCreation := func() {
		// 1. Create task in database (your existing logic)
		taskID := uuid.New()
		projectID := uuid.New()

		task := map[string]interface{}{
			"id":          taskID,
			"project_id":  projectID,
			"title":       "New task from API",
			"description": "Created via HTTP API",
			"status":      "TODO",
			"created_at":  time.Now(),
		}

		// 2. Send WebSocket notification
		if err := wsService.NotifyTaskCreated(task, projectID); err != nil {
			log.Printf("Failed to send WebSocket notification: %v", err)
			// Don't fail the HTTP request, just log the error
		}

		log.Printf("Task created and notification sent: %s", taskID)
	}

	simulateTaskUpdate := func() {
		taskID := uuid.New()
		projectID := uuid.New()

		// 1. Update task in database
		changes := map[string]interface{}{
			"status": map[string]interface{}{
				"old": "TODO",
				"new": "IN_PROGRESS",
			},
		}

		updatedTask := map[string]interface{}{
			"id":         taskID,
			"project_id": projectID,
			"title":      "Updated task",
			"status":     "IN_PROGRESS",
			"updated_at": time.Now(),
		}

		// 2. Send WebSocket notifications
		if err := wsService.NotifyTaskUpdated(taskID, projectID, changes, updatedTask); err != nil {
			log.Printf("Failed to send task update notification: %v", err)
		}

		if err := wsService.NotifyStatusChanged(taskID, projectID, "task", "TODO", "IN_PROGRESS"); err != nil {
			log.Printf("Failed to send status change notification: %v", err)
		}

		log.Printf("Task updated and notifications sent: %s", taskID)
	}

	// Simulate the handlers
	simulateTaskCreation()
	simulateTaskUpdate()
}
