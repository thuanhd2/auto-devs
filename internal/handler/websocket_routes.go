package handler

import (
	"github.com/auto-devs/auto-devs/internal/websocket"
	"github.com/gin-gonic/gin"
)

// SetupWebSocketRoutes configures WebSocket routes
func SetupWebSocketRoutes(router *gin.Engine, wsHandler *websocket.Handler, wsService *websocket.Service) {
	// WebSocket upgrade endpoint
	ws := router.Group("/ws")
	{
		// WebSocket connection endpoint
		ws.GET("/connect", wsHandler.GetWebSocketHandler())
	}
}
