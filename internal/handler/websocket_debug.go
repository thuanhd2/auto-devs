package handler

import (
	"net/http"
	"os"

	"github.com/auto-devs/auto-devs/internal/websocket"
	"github.com/gin-gonic/gin"
)

// WebSocketDebugHandler provides debug endpoints for WebSocket functionality
type WebSocketDebugHandler struct {
	enhancedService *websocket.EnhancedService
}

// NewWebSocketDebugHandler creates a new WebSocket debug handler
func NewWebSocketDebugHandler(enhancedService *websocket.EnhancedService) *WebSocketDebugHandler {
	return &WebSocketDebugHandler{
		enhancedService: enhancedService,
	}
}

// GetBackendInfo returns information about which WebSocket backend is being used
func (h *WebSocketDebugHandler) GetBackendInfo(c *gin.Context) {
	info := map[string]interface{}{
		"backend_type":         h.getBackendType(),
		"using_centrifuge":     h.enhancedService.IsCentrifugeEnabled(),
		"websocket_use_legacy": os.Getenv("WEBSOCKET_USE_LEGACY"),
		"use_centrifuge":       os.Getenv("USE_CENTRIFUGE"),
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   info,
	})
}

// TestBackendSwitch tests switching between backends
func (h *WebSocketDebugHandler) TestBackendSwitch(c *gin.Context) {
	// Test message broadcasting
	testMessage := map[string]interface{}{
		"type":    "debug_test",
		"message": "Backend switching test",
		"backend": h.getBackendType(),
	}

	err := h.enhancedService.BroadcastToProject(c.Request.Context(), 1, testMessage)
	
	response := gin.H{
		"status":       "success",
		"backend_type": h.getBackendType(),
		"test_result":  "broadcast_attempted",
	}
	
	if err != nil {
		response["error"] = err.Error()
		response["note"] = "Error expected if no active connections"
	}

	c.JSON(http.StatusOK, response)
}

func (h *WebSocketDebugHandler) getBackendType() string {
	if h.enhancedService.IsCentrifugeEnabled() {
		return "centrifuge"
	}
	return "legacy"
}