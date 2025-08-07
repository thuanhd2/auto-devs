package websocket

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/auto-devs/auto-devs/config"
)

// EnhancedService provides WebSocket functionality with both legacy and Centrifuge support
type EnhancedService struct {
	*Service // Embed the original service
	
	// Centrifuge components
	centrifugeHandler *CentrifugeHandler
	useCentrifuge     bool
	
	// Configuration
	config *config.CentrifugeConfig
}

// NewEnhancedService creates a new enhanced WebSocket service that supports both backends
func NewEnhancedService(cfg *config.CentrifugeConfig) (*EnhancedService, error) {
	// Create the original service
	originalService := NewService()
	
	// Determine which backend to use
	useCentrifuge := false
	if legacyStr := os.Getenv("WEBSOCKET_USE_LEGACY"); legacyStr != "" {
		if parsed, err := strconv.ParseBool(legacyStr); err == nil {
			useCentrifuge = !parsed // If legacy is false, use Centrifuge
		}
	}
	
	enhancedService := &EnhancedService{
		Service:       originalService,
		useCentrifuge: useCentrifuge,
		config:        cfg,
	}
	
	// Initialize Centrifuge handler if needed
	if useCentrifuge {
		centrifugeHandler, err := NewCentrifugeHandler(cfg, false)
		if err != nil {
			log.Printf("Failed to create Centrifuge handler, falling back to legacy: %v", err)
			enhancedService.useCentrifuge = false
		} else {
			enhancedService.centrifugeHandler = centrifugeHandler
			log.Printf("Enhanced service initialized with Centrifuge backend")
		}
	} else {
		log.Printf("Enhanced service initialized with legacy backend")
	}
	
	return enhancedService, nil
}

// GetCentrifugeHandler returns the Centrifuge handler if available
func (s *EnhancedService) GetCentrifugeHandler() *CentrifugeHandler {
	return s.centrifugeHandler
}

// IsUsingCentrifuge returns true if the service is using Centrifuge backend
func (s *EnhancedService) IsUsingCentrifuge() bool {
	return s.useCentrifuge && s.centrifugeHandler != nil
}

// SwitchToCentrifuge switches to Centrifuge backend
func (s *EnhancedService) SwitchToCentrifuge() error {
	if s.centrifugeHandler == nil {
		centrifugeHandler, err := NewCentrifugeHandler(s.config, false)
		if err != nil {
			return fmt.Errorf("failed to create Centrifuge handler: %w", err)
		}
		s.centrifugeHandler = centrifugeHandler
	}
	
	s.useCentrifuge = true
	log.Printf("Switched to Centrifuge backend")
	return nil
}

// SwitchToLegacy switches to legacy backend
func (s *EnhancedService) SwitchToLegacy() {
	s.useCentrifuge = false
	log.Printf("Switched to legacy backend")
}

// Enhanced notification methods that use the appropriate backend

// NotifyTaskCreated notifies about a task creation using the appropriate backend
func (s *EnhancedService) NotifyTaskCreated(task interface{}, projectID uuid.UUID) error {
	if s.IsUsingCentrifuge() {
		return s.centrifugeHandler.BroadcastToProject(projectID, TaskCreated, task)
	}
	return s.Service.NotifyTaskCreated(task, projectID)
}

// NotifyTaskUpdated notifies about a task update using the appropriate backend
func (s *EnhancedService) NotifyTaskUpdated(taskID, projectID uuid.UUID, changes map[string]interface{}, task interface{}) error {
	if s.IsUsingCentrifuge() {
		data := map[string]interface{}{
			"task":    task,
			"changes": changes,
		}
		return s.centrifugeHandler.BroadcastToProject(projectID, TaskUpdated, data)
	}
	return s.Service.NotifyTaskUpdated(taskID, projectID, changes, task)
}

// NotifyTaskDeleted notifies about a task deletion using the appropriate backend
func (s *EnhancedService) NotifyTaskDeleted(taskID, projectID uuid.UUID) error {
	if s.IsUsingCentrifuge() {
		data := map[string]interface{}{
			"task_id":    taskID,
			"project_id": projectID,
		}
		return s.centrifugeHandler.BroadcastToProject(projectID, TaskDeleted, data)
	}
	return s.Service.NotifyTaskDeleted(taskID, projectID)
}

// NotifyProjectUpdated notifies about a project update using the appropriate backend
func (s *EnhancedService) NotifyProjectUpdated(projectID uuid.UUID, changes map[string]interface{}, project interface{}) error {
	if s.IsUsingCentrifuge() {
		data := map[string]interface{}{
			"project": project,
			"changes": changes,
		}
		return s.centrifugeHandler.BroadcastToProject(projectID, ProjectUpdated, data)
	}
	return s.Service.NotifyProjectUpdated(projectID, changes, project)
}

// NotifyStatusChanged notifies about a status change using the appropriate backend
func (s *EnhancedService) NotifyStatusChanged(entityID, projectID uuid.UUID, entityType, oldStatus, newStatus string) error {
	if s.IsUsingCentrifuge() {
		data := map[string]interface{}{
			"entity_id":   entityID,
			"project_id":  projectID,
			"entity_type": entityType,
			"old_status":  oldStatus,
			"new_status":  newStatus,
		}
		return s.centrifugeHandler.BroadcastToProject(projectID, StatusChanged, data)
	}
	return s.Service.NotifyStatusChanged(entityID, projectID, entityType, oldStatus, newStatus)
}

// NotifyUserJoined notifies about a user joining using the appropriate backend
func (s *EnhancedService) NotifyUserJoined(userID string, projectID uuid.UUID) error {
	if s.IsUsingCentrifuge() {
		data := map[string]interface{}{
			"user_id":    userID,
			"project_id": projectID,
			"action":     "joined",
		}
		return s.centrifugeHandler.BroadcastToProject(projectID, UserJoined, data)
	}
	return s.Service.NotifyUserJoined(userID, projectID)
}

// NotifyUserLeft notifies about a user leaving using the appropriate backend
func (s *EnhancedService) NotifyUserLeft(userID string, projectID uuid.UUID) error {
	if s.IsUsingCentrifuge() {
		data := map[string]interface{}{
			"user_id":    userID,
			"project_id": projectID,
			"action":     "left",
		}
		return s.centrifugeHandler.BroadcastToProject(projectID, UserLeft, data)
	}
	return s.Service.NotifyUserLeft(userID, projectID)
}

// Enhanced direct messaging methods

// SendDirectMessage sends a message directly to a user using the appropriate backend
func (s *EnhancedService) SendDirectMessage(userID string, msgType MessageType, data interface{}) error {
	if s.IsUsingCentrifuge() {
		return s.centrifugeHandler.BroadcastToUser(userID, msgType, data)
	}
	return s.Service.SendDirectMessage(userID, msgType, data)
}

// SendProjectMessage sends a message to all users in a project using the appropriate backend
func (s *EnhancedService) SendProjectMessage(projectID uuid.UUID, msgType MessageType, data interface{}) error {
	if s.IsUsingCentrifuge() {
		return s.centrifugeHandler.BroadcastToProject(projectID, msgType, data)
	}
	return s.Service.SendProjectMessage(projectID, msgType, data)
}

// BroadcastMessage broadcasts a custom message using the appropriate backend
func (s *EnhancedService) BroadcastMessage(msgType MessageType, data interface{}, projectID *uuid.UUID, userID *string) error {
	if s.IsUsingCentrifuge() {
		switch {
		case projectID != nil:
			return s.centrifugeHandler.BroadcastToProject(*projectID, msgType, data)
		case userID != nil:
			return s.centrifugeHandler.BroadcastToUser(*userID, msgType, data)
		default:
			return s.centrifugeHandler.BroadcastToAll(msgType, data)
		}
	}
	return s.Service.BroadcastMessage(msgType, data, projectID, userID)
}

// Enhanced metrics and monitoring

// GetEnhancedMetrics returns metrics from both backends
func (s *EnhancedService) GetEnhancedMetrics() map[string]interface{} {
	metrics := s.Service.GetMetrics()
	
	// Add backend information
	metrics["backend"] = "legacy"
	if s.IsUsingCentrifuge() {
		metrics["backend"] = "centrifuge"
		// Add Centrifuge-specific metrics if available
		metrics["centrifuge_available"] = true
	} else {
		metrics["centrifuge_available"] = s.centrifugeHandler != nil
	}
	
	metrics["use_centrifuge"] = s.useCentrifuge
	return metrics
}

// GetBackendInfo returns information about the current backend
func (s *EnhancedService) GetBackendInfo() map[string]interface{} {
	return map[string]interface{}{
		"current_backend":      s.getCurrentBackendName(),
		"centrifuge_available": s.centrifugeHandler != nil,
		"legacy_available":     true, // Always available
		"can_switch":           true,
		"use_centrifuge":       s.useCentrifuge,
		"timestamp":            time.Now(),
	}
}

// getCurrentBackendName returns the name of the current backend
func (s *EnhancedService) getCurrentBackendName() string {
	if s.IsUsingCentrifuge() {
		return "centrifuge"
	}
	return "legacy"
}

// Enhanced health checks

// IsHealthy checks if the current backend is healthy
func (s *EnhancedService) IsHealthy() bool {
	if s.IsUsingCentrifuge() {
		// For Centrifuge, we could check node health
		return s.centrifugeHandler != nil
	}
	return s.Service.IsHealthy()
}

// GetEnhancedHealthStatus returns detailed health status for both backends
func (s *EnhancedService) GetEnhancedHealthStatus() map[string]interface{} {
	status := s.Service.GetHealthStatus()
	
	// Add backend-specific information
	status["backend"] = s.getCurrentBackendName()
	status["centrifuge_healthy"] = s.centrifugeHandler != nil
	status["legacy_healthy"] = s.Service.IsHealthy()
	
	return status
}

// Channel management methods (Centrifuge-specific)

// GetChannelInfo returns information about a Centrifuge channel
func (s *EnhancedService) GetChannelInfo(channel string) (interface{}, error) {
	if !s.IsUsingCentrifuge() {
		return nil, fmt.Errorf("channel info only available with Centrifuge backend")
	}
	
	server := s.centrifugeHandler.server
	if server == nil {
		return nil, fmt.Errorf("Centrifuge server not available")
	}
	
	return server.GetChannelInfo(channel)
}

// GetPresence returns presence information for a Centrifuge channel
func (s *EnhancedService) GetPresence(channel string) (interface{}, error) {
	if !s.IsUsingCentrifuge() {
		return nil, fmt.Errorf("presence info only available with Centrifuge backend")
	}
	
	server := s.centrifugeHandler.server
	if server == nil {
		return nil, fmt.Errorf("Centrifuge server not available")
	}
	
	return server.GetPresence(channel)
}

// GetHistory returns message history for a Centrifuge channel
func (s *EnhancedService) GetHistory(channel string, limit int) (interface{}, error) {
	if !s.IsUsingCentrifuge() {
		return nil, fmt.Errorf("history only available with Centrifuge backend")
	}
	
	server := s.centrifugeHandler.server
	if server == nil {
		return nil, fmt.Errorf("Centrifuge server not available")
	}
	
	return server.GetHistory(channel, limit, nil)
}