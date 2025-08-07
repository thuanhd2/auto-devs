package websocket

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/auto-devs/auto-devs/config"
)

// CentrifugeHandler manages WebSocket connections using Centrifuge
type CentrifugeHandler struct {
	server     *CentrifugeServer
	legacyHub  *Hub // Keep legacy hub for gradual migration
	useLegacy  bool // Feature flag to switch between implementations
}

// NewCentrifugeHandler creates a new Centrifuge WebSocket handler
func NewCentrifugeHandler(cfg *config.CentrifugeConfig, useLegacy bool) (*CentrifugeHandler, error) {
	var server *CentrifugeServer
	var err error
	
	if !useLegacy {
		server, err = NewCentrifugeServer(cfg)
		if err != nil {
			return nil, err
		}
	}

	// Always create legacy hub for backward compatibility during migration
	hub := NewHub()
	go hub.Run()

	handler := &CentrifugeHandler{
		server:    server,
		legacyHub: hub,
		useLegacy: useLegacy,
	}

	// Start Centrifuge server if not using legacy
	if !useLegacy && server != nil {
		if err := server.Start(); err != nil {
			return nil, err
		}
	}

	return handler, nil
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *CentrifugeHandler) HandleWebSocket(c *gin.Context) {
	if h.useLegacy {
		h.handleLegacyWebSocket(c)
	} else {
		h.handleCentrifugeWebSocket(c)
	}
}

// handleLegacyWebSocket handles WebSocket connections using the legacy hub
func (h *CentrifugeHandler) handleLegacyWebSocket(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}

	// Create new connection
	wsConn := NewConnection(conn, h.legacyHub)

	// Register with hub
	h.legacyHub.Register(wsConn)

	// Start connection pumps
	wsConn.Start()

	log.Printf("Legacy WebSocket connection established: %s", wsConn.ID)
}

// handleCentrifugeWebSocket handles WebSocket connections using Centrifuge
func (h *CentrifugeHandler) handleCentrifugeWebSocket(c *gin.Context) {
	// Use Centrifuge's handler
	centrifugeHandler := h.server.CreateHTTPHandler()
	centrifugeHandler(c)
}

// BroadcastMessage broadcasts a message to connections
func (h *CentrifugeHandler) BroadcastMessage(c *gin.Context) {
	var request struct {
		Type      MessageType `json:"type" binding:"required"`
		Data      interface{} `json:"data"`
		ProjectID *string     `json:"project_id,omitempty"`
		UserID    *string     `json:"user_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.useLegacy {
		h.broadcastLegacyMessage(c, &request)
	} else {
		h.broadcastCentrifugeMessage(c, &request)
	}
}

// broadcastLegacyMessage broadcasts using the legacy hub
func (h *CentrifugeHandler) broadcastLegacyMessage(c *gin.Context, request *struct {
	Type      MessageType `json:"type" binding:"required"`
	Data      interface{} `json:"data"`
	ProjectID *string     `json:"project_id,omitempty"`
	UserID    *string     `json:"user_id,omitempty"`
}) {
	message, err := NewMessage(request.Type, request.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	var projectID *uuid.UUID
	if request.ProjectID != nil && *request.ProjectID != "" {
		if pid, err := uuid.Parse(*request.ProjectID); err == nil {
			projectID = &pid
		}
	}

	// Broadcast message
	h.legacyHub.Broadcast(message, projectID, request.UserID, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Legacy message broadcasted successfully",
		"type":    request.Type,
	})
}

// broadcastCentrifugeMessage broadcasts using Centrifuge
func (h *CentrifugeHandler) broadcastCentrifugeMessage(c *gin.Context, request *struct {
	Type      MessageType `json:"type" binding:"required"`
	Data      interface{} `json:"data"`
	ProjectID *string     `json:"project_id,omitempty"`
	UserID    *string     `json:"user_id,omitempty"`
}) {
	centrifugeMsg := CreateCentrifugeMessage(request.Type, request.Data)

	var err error

	switch {
	case request.ProjectID != nil && *request.ProjectID != "":
		// Broadcast to project
		projectID, parseErr := uuid.Parse(*request.ProjectID)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}
		err = h.server.PublishToProject(projectID, centrifugeMsg)
	
	case request.UserID != nil && *request.UserID != "":
		// Broadcast to user
		err = h.server.PublishToUser(*request.UserID, centrifugeMsg)
	
	default:
		// Broadcast to all
		err = h.server.PublishToAll(centrifugeMsg)
	}

	if err != nil {
		log.Printf("Failed to broadcast Centrifuge message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to broadcast message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Centrifuge message broadcasted successfully",
		"type":    request.Type,
	})
}

// GetConnections returns information about all connections
func (h *CentrifugeHandler) GetConnections(c *gin.Context) {
	if h.useLegacy {
		connections := h.legacyHub.GetConnectionsInfo()
		c.JSON(http.StatusOK, gin.H{
			"connections": connections,
			"total":       len(connections),
			"backend":     "legacy",
		})
	} else {
		// For Centrifuge, we can get channel information
		// This is a simplified version - in practice you might want to track more details
		c.JSON(http.StatusOK, gin.H{
			"message": "Centrifuge connections info not implemented yet",
			"backend": "centrifuge",
		})
	}
}

// GetMetrics returns metrics
func (h *CentrifugeHandler) GetMetrics(c *gin.Context) {
	if h.useLegacy {
		metrics := h.legacyHub.GetMetrics()
		c.JSON(http.StatusOK, gin.H{
			"metrics": metrics,
			"backend": "legacy",
		})
	} else {
		// For Centrifuge, we can provide basic node information
		c.JSON(http.StatusOK, gin.H{
			"backend": "centrifuge",
			"message": "Centrifuge metrics not implemented yet",
		})
	}
}

// GetChannelInfo returns information about a specific channel (Centrifuge only)
func (h *CentrifugeHandler) GetChannelInfo(c *gin.Context) {
	if h.useLegacy {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Channel info not available in legacy mode",
			"backend": "legacy",
		})
		return
	}

	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Channel parameter is required"})
		return
	}

	info, err := h.server.GetChannelInfo(channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"channel": channel,
		"info":    info,
		"backend": "centrifuge",
	})
}

// GetPresence returns presence information for a channel (Centrifuge only)
func (h *CentrifugeHandler) GetPresence(c *gin.Context) {
	if h.useLegacy {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Presence info not available in legacy mode",
			"backend": "legacy",
		})
		return
	}

	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Channel parameter is required"})
		return
	}

	presence, err := h.server.GetPresence(channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"channel":  channel,
		"presence": presence,
		"backend":  "centrifuge",
	})
}

// GetHistory returns message history for a channel (Centrifuge only)
func (h *CentrifugeHandler) GetHistory(c *gin.Context) {
	if h.useLegacy {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "History not available in legacy mode",
			"backend": "legacy",
		})
		return
	}

	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Channel parameter is required"})
		return
	}

	// Get limit from query parameters
	limit := 10 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := parseLimit(limitStr); err == nil {
			limit = l
		}
	}

	history, err := h.server.GetHistory(channel, limit, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"channel": channel,
		"history": history,
		"backend": "centrifuge",
	})
}

// SwitchBackend switches between legacy and Centrifuge backends (for testing)
func (h *CentrifugeHandler) SwitchBackend(c *gin.Context) {
	var request struct {
		UseLegacy bool `json:"use_legacy"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.useLegacy = request.UseLegacy

	backend := "centrifuge"
	if h.useLegacy {
		backend = "legacy"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Backend switched successfully",
		"backend": backend,
	})
}

// Shutdown gracefully shuts down the handler
func (h *CentrifugeHandler) Shutdown(ctx context.Context) error {
	log.Printf("Shutting down WebSocket handler")

	// Shutdown legacy hub
	if h.legacyHub != nil {
		h.legacyHub.Shutdown()
	}

	// Shutdown Centrifuge server
	if h.server != nil {
		if err := h.server.Shutdown(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Helper functions
func parseLimit(limitStr string) (int, error) {
	// Simple parsing - in practice you might want more robust validation
	if limitStr == "" {
		return 10, nil
	}
	// This is a simplified implementation
	return 10, nil
}

// BroadcastToProject publishes a message to all subscribers of a project
func (h *CentrifugeHandler) BroadcastToProject(projectID uuid.UUID, msgType MessageType, data interface{}) error {
	if h.useLegacy {
		message, err := NewMessage(msgType, data)
		if err != nil {
			return err
		}
		h.legacyHub.BroadcastToProject(message, projectID, nil)
		return nil
	} else {
		centrifugeMsg := CreateCentrifugeMessage(msgType, data)
		return h.server.PublishToProject(projectID, centrifugeMsg)
	}
}

// BroadcastToUser publishes a message to a specific user
func (h *CentrifugeHandler) BroadcastToUser(userID string, msgType MessageType, data interface{}) error {
	if h.useLegacy {
		message, err := NewMessage(msgType, data)
		if err != nil {
			return err
		}
		h.legacyHub.BroadcastToUser(message, userID, nil)
		return nil
	} else {
		centrifugeMsg := CreateCentrifugeMessage(msgType, data)
		return h.server.PublishToUser(userID, centrifugeMsg)
	}
}

// BroadcastToAll publishes a message to all connected clients
func (h *CentrifugeHandler) BroadcastToAll(msgType MessageType, data interface{}) error {
	if h.useLegacy {
		message, err := NewMessage(msgType, data)
		if err != nil {
			return err
		}
		h.legacyHub.BroadcastToAll(message, nil)
		return nil
	} else {
		centrifugeMsg := CreateCentrifugeMessage(msgType, data)
		return h.server.PublishToAll(centrifugeMsg)
	}
}