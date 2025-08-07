package websocket

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/auto-devs/auto-devs/config"
	"github.com/google/uuid"
)

// Service provides WebSocket functionality integration
type Service struct {
	handler           *Handler
	hub               *Hub
	middlewareManager *MiddlewareManager

	taskProcessor     *TaskEventProcessor
	projectProcessor  *ProjectEventProcessor
	statusProcessor   *StatusEventProcessor
	presenceProcessor *UserPresenceProcessor
	redisBroker       *RedisBroker // Redis broker for cross-process messaging
	logger            *slog.Logger
}

// NewService creates a new WebSocket service
func NewService(appConfig *config.CentrifugeRedisBrokerConfig) *Service {
	server, err := NewServer(appConfig)
	if err != nil {
		log.Fatalf("Failed to create WebSocket server: %v", err)
	}
	// Create core components
	handler := NewHandler(server)
	hub := handler.GetHub()
	middlewareManager := NewMiddlewareManager()

	// Create processors
	taskProcessor := NewTaskEventProcessor(hub)
	projectProcessor := NewProjectEventProcessor(hub)
	statusProcessor := NewStatusEventProcessor(hub)
	presenceProcessor := NewUserPresenceProcessor(hub)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := &Service{
		handler:           handler,
		hub:               hub,
		middlewareManager: middlewareManager,
		taskProcessor:     taskProcessor,
		projectProcessor:  projectProcessor,
		statusProcessor:   statusProcessor,
		presenceProcessor: presenceProcessor,
		logger:            logger,
	}

	return service
}

// GetHandler returns the WebSocket handler
func (s *Service) GetHandler() *Handler {
	return s.handler
}

// GetHub returns the WebSocket hub
func (s *Service) GetHub() *Hub {
	return s.hub
}

// Task event methods

// NotifyTaskCreated notifies about a task creation
func (s *Service) NotifyTaskCreated(task interface{}, projectID uuid.UUID) error {
	return s.taskProcessor.BroadcastTaskCreated(task, projectID, nil)
}

// NotifyTaskUpdated notifies about a task update
func (s *Service) NotifyTaskUpdated(taskID, projectID uuid.UUID, changes map[string]interface{}, task interface{}) error {
	// Try Redis broker first if available
	if s.redisBroker != nil && s.redisBroker.IsRunning() {
		if err := s.redisBroker.PublishTaskUpdated(taskID, projectID, changes, task); err != nil {
			s.logger.Warn("Failed to publish via Redis broker, falling back to direct broadcast", "error", err)
		} else {
			s.logger.Debug("Published task update via Redis broker", "task_id", taskID)
			return nil
		}
	}

	// Fallback to direct broadcast
	return s.taskProcessor.BroadcastTaskUpdated(taskID, projectID, changes, task, nil)
}

// NotifyTaskDeleted notifies about a task deletion
func (s *Service) NotifyTaskDeleted(taskID, projectID uuid.UUID) error {
	s.logger.Info("NotifyTaskDeleted", "taskID", taskID, "projectID", projectID)
	return s.taskProcessor.BroadcastTaskDeleted(taskID, projectID, nil)
}

// Project event methods

// NotifyProjectUpdated notifies about a project update
func (s *Service) NotifyProjectUpdated(projectID uuid.UUID, changes map[string]interface{}, project interface{}) error {
	return s.projectProcessor.BroadcastProjectUpdated(projectID, changes, project, nil)
}

// Status event methods

// NotifyStatusChanged notifies about a status change
func (s *Service) NotifyStatusChanged(entityID, projectID uuid.UUID, entityType, oldStatus, newStatus string) error {
	// Try Redis broker first if available
	if s.redisBroker != nil && s.redisBroker.IsRunning() {
		if err := s.redisBroker.PublishStatusChanged(entityID, projectID, entityType, oldStatus, newStatus); err != nil {
			s.logger.Warn("Failed to publish via Redis broker, falling back to direct broadcast", "error", err)
		} else {
			s.logger.Debug("Published status change via Redis broker", "entity_id", entityID)
			return nil
		}
	}

	// Fallback to direct broadcast
	return s.statusProcessor.BroadcastStatusChanged(entityID, projectID, entityType, oldStatus, newStatus, nil)
}

// User presence methods

// NotifyUserJoined notifies about a user joining a project
func (s *Service) NotifyUserJoined(userID string, projectID uuid.UUID) error {
	return s.presenceProcessor.BroadcastUserJoined(userID, projectID, nil)
}

// NotifyUserLeft notifies about a user leaving a project
func (s *Service) NotifyUserLeft(userID string, projectID uuid.UUID) error {
	return s.presenceProcessor.BroadcastUserLeft(userID, projectID, nil)
}

// Connection management methods

// GetConnectionCount returns the total number of active connections
func (s *Service) GetConnectionCount() int64 {
	metrics := s.hub.GetMetrics()
	return metrics.ActiveConnections
}

// Metrics and monitoring

// GetMetrics returns comprehensive WebSocket metrics
func (s *Service) GetMetrics() map[string]interface{} {
	hubMetrics := s.hub.GetMetrics()
	middlewareStats := s.middlewareManager.GetMiddlewareStats()

	return map[string]interface{}{
		"hub":        hubMetrics,
		"middleware": middlewareStats,
		"timestamp":  time.Now(),
	}
}

// Health check

// IsHealthy checks if the WebSocket service is healthy
func (s *Service) IsHealthy() bool {
	// Check if hub is responsive by getting metrics
	_ = s.hub.GetMetrics()
	return true
}

// GetHealthStatus returns detailed health status
func (s *Service) GetHealthStatus() map[string]interface{} {
	metrics := s.hub.GetMetrics()

	return map[string]interface{}{
		"status":             "healthy",
		"active_connections": metrics.ActiveConnections,
		"total_connections":  metrics.TotalConnections,
		"messages_sent":      metrics.MessagesSent,
		"messages_received":  metrics.MessagesReceived,
		"uptime":             time.Since(time.Now()).String(), // This would need to track actual start time
		"timestamp":          time.Now(),
	}
}

// Administrative methods

// BroadcastMessage broadcasts a custom message (for admin/testing purposes)
func (s *Service) BroadcastMessage(msgType MessageType, data interface{}, projectID *uuid.UUID, userID *string) error {
	message, err := NewMessage(msgType, data)
	if err != nil {
		return err
	}

	s.hub.Broadcast(message, projectID, userID, nil)
	return nil
}

// DisconnectUser disconnects all connections for a specific user
func (s *Service) DisconnectUser(userID string) {
	// This would require additional methods in the hub to get user connections
	// and disconnect them. For now, this is a placeholder.
	log.Printf("Disconnecting all connections for user: %s", userID)
}

// DisconnectProject disconnects all connections from a specific project
func (s *Service) DisconnectProject(projectID uuid.UUID) {
	// This would require additional methods in the hub to get project connections
	// and disconnect them. For now, this is a placeholder.
	log.Printf("Disconnecting all connections from project: %s", projectID)
}

// Configuration

// ServiceConfig holds WebSocket service configuration
type ServiceConfig struct {
	// Rate limiting
	RequestsPerSecond float64
	BurstSize         int

	// Error handling
	MaxErrors          int
	ErrorResetInterval time.Duration

	// Message persistence
	MaxStoredMessages int
	MessageTTL        time.Duration

	// Connection health
	PingInterval   time.Duration
	PongTimeout    time.Duration
	MaxMessageSize int64
}

// DefaultServiceConfig returns default service configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		RequestsPerSecond:  10.0,
		BurstSize:          20,
		MaxErrors:          10,
		ErrorResetInterval: 5 * time.Minute,
		MaxStoredMessages:  1000,
		MessageTTL:         24 * time.Hour,
		PingInterval:       54 * time.Second,
		PongTimeout:        60 * time.Second,
		MaxMessageSize:     512,
	}
}

// Utility methods for integration

// ValidateProjectAccess validates if a user has access to a project
func (s *Service) ValidateProjectAccess(userID string, projectID uuid.UUID) bool {
	// This should integrate with your authorization system
	// For now, we'll allow all access
	return true
}

// GetActiveProjectUsers returns active users for a project
func (s *Service) GetActiveProjectUsers(projectID uuid.UUID) []string {
	// This would return users currently connected to a project
	// Implementation would require tracking user IDs by project
	return []string{}
}

// SendDirectMessage sends a message directly to a specific user
func (s *Service) SendDirectMessage(userID string, msgType MessageType, data interface{}) error {
	message, err := NewMessage(msgType, data)
	if err != nil {
		return err
	}

	s.hub.BroadcastToUser(message, userID, nil)
	return nil
}

// SendProjectMessage sends a message to all users in a project
func (s *Service) SendProjectMessage(projectID uuid.UUID, msgType MessageType, data interface{}) error {
	message, err := NewMessage(msgType, data)
	if err != nil {
		return err
	}

	s.hub.BroadcastToProject(message, projectID, nil)
	return nil
}
