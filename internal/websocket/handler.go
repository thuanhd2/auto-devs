package websocket

import (
	"log"
	"net/http"

	"github.com/centrifugal/centrifuge"
	"github.com/gin-gonic/gin"
)

// Handler manages WebSocket connections and routing
type Handler struct {
	hub    *Hub
	server *Server
}

// NewHandler creates a new WebSocket handler
func NewHandler(server *Server) *Handler {
	hub := NewHub(server.node)
	handler := &Handler{
		hub:    hub,
		server: server,
	}

	log.Printf("WebSocket handler created successfully")
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
func (h *Handler) GetWebSocketHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("WebSocket connection request from %s", c.ClientIP())

		// Check if server is ready
		if h.server == nil || h.server.node == nil {
			log.Printf("WebSocket server not ready")
			c.JSON(503, gin.H{"error": "WebSocket server not ready"})
			return
		}

		// Create Centrifuge WebSocket handler
		Handler := centrifuge.NewWebsocketHandler(h.server.node, centrifuge.WebsocketConfig{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for now - in production, implement proper origin checking
				log.Printf("Checking origin: %s", r.Header.Get("Origin"))
				return true
			},
		})

		// Serve the WebSocket request
		log.Printf("Serving WebSocket request for %s", c.ClientIP())
		Handler.ServeHTTP(c.Writer, c.Request)
	}
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
