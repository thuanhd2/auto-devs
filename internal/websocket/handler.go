package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler manages WebSocket connections and routing
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler() *Handler {
	hub := NewHub()

	handler := &Handler{
		hub: hub,
	}

	// Start the hub
	go hub.Run()

	return handler
}

// GetHub returns the hub instance
func (h *Handler) GetHub() *Hub {
	return h.hub
}

// Shutdown gracefully shuts down the WebSocket handler
func (h *Handler) Shutdown() {
	log.Printf("Shutting down WebSocket handler")
	h.hub.Shutdown()
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}

	// Create new connection
	wsConn := NewConnection(conn, h.hub)
	log.Printf("NewConnection ++++++++++++++++++++++: %v", wsConn.ID)

	// Register with hub
	h.hub.Register(wsConn)

	// Start connection pumps
	wsConn.Start()

	log.Printf("WebSocket connection established: %s", wsConn.ID)
}

// GetConnections returns information about all connections
func (h *Handler) GetConnections(c *gin.Context) {
	connections := h.hub.GetConnectionsInfo()
	c.JSON(http.StatusOK, gin.H{
		"connections": connections,
		"total":       len(connections),
	})
}

// GetMetrics returns hub metrics
func (h *Handler) GetMetrics(c *gin.Context) {
	metrics := h.hub.GetMetrics()
	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// BroadcastMessage broadcasts a message to connections (for testing/admin purposes)
func (h *Handler) BroadcastMessage(c *gin.Context) {
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

	message, err := NewMessage(request.Type, request.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	// Parse project ID if provided
	var projectID *string
	if request.ProjectID != nil && *request.ProjectID != "" {
		projectID = request.ProjectID
	}

	// Broadcast message
	h.hub.Broadcast(message, nil, projectID, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Message broadcasted successfully",
		"type":    request.Type,
	})
}
