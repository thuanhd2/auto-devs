package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/auto-devs/auto-devs/internal/websocket"
	"github.com/gin-gonic/gin"
)

// SetupWebSocketRoutes configures WebSocket routes with support for both legacy and Centrifuge backends
func SetupWebSocketRoutes(router *gin.Engine, wsHandler *websocket.Handler, wsService *websocket.EnhancedService) {
	// Create debug handler for testing
	debugHandler := NewWebSocketDebugHandler(wsService)
	// Check if we should use legacy WebSocket (via environment variable)
	useLegacy := true // Default to legacy for safety
	if legacyStr := os.Getenv("WEBSOCKET_USE_LEGACY"); legacyStr != "" {
		if parsed, err := strconv.ParseBool(legacyStr); err == nil {
			useLegacy = parsed
		}
	}
	// WebSocket upgrade endpoint
	ws := router.Group("/ws")
	{
		if useLegacy {
			// Legacy WebSocket implementation
			ws.Use(websocket.WebSocketMiddleware())
			ws.Use(websocket.WebSocketAuthMiddleware(wsService.GetAuthService()))
			ws.GET("/connect", wsHandler.HandleWebSocket)
		} else {
			// Centrifuge WebSocket implementation
			// Create Centrifuge handler (we'll implement this integration)
			centrifugeHandler, err := createCentrifugeHandler(wsService)
			if err != nil {
				// Fallback to legacy if Centrifuge setup fails
				ws.Use(websocket.WebSocketMiddleware())
				ws.Use(websocket.WebSocketAuthMiddleware(wsService.GetAuthService()))
				ws.GET("/connect", wsHandler.HandleWebSocket)
			} else {
				ws.GET("/connect", centrifugeHandler.HandleWebSocket)
			}
		}
	}
	
	// WebSocket management API (for monitoring/admin)
	wsAPI := router.Group("/api/v1/websocket")
	{
		// Apply authentication middleware for admin endpoints
		wsAPI.Use(websocket.AuthMiddleware(wsService.GetAuthService()))
		
		// Connection management endpoints
		wsAPI.GET("/connections", wsHandler.GetConnections)
		wsAPI.GET("/metrics", wsHandler.GetMetrics)
		wsAPI.POST("/broadcast", wsHandler.BroadcastMessage)
		
		// Health and status endpoints
		wsAPI.GET("/health", func(c *gin.Context) {
			if wsService.IsHealthy() {
				c.JSON(http.StatusOK, gin.H{
					"status": "healthy",
					"data":   wsService.GetHealthStatus(),
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": "unhealthy",
				})
			}
		})
		
		// Statistics endpoint
		wsAPI.GET("/stats", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"data": wsService.GetMetrics(),
			})
		})
		
		// Connection count endpoints
		wsAPI.GET("/connections/count", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"total": wsService.GetConnectionCount(),
			})
		})
		
		wsAPI.GET("/projects/:projectId/connections/count", func(c *gin.Context) {
			projectIDStr := c.Param("projectId")
			projectID, err := parseUUID(projectIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
				return
			}
			
			count := wsService.GetProjectConnectionCount(projectID)
			c.JSON(http.StatusOK, gin.H{
				"project_id": projectID,
				"count":      count,
			})
		})
		
		wsAPI.GET("/users/:userId/connections/count", func(c *gin.Context) {
			userID := c.Param("userId")
			count := wsService.GetUserConnectionCount(userID)
			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"count":   count,
			})
		})
		
		// Backend information endpoint
		wsAPI.GET("/backend", func(c *gin.Context) {
			backend := "legacy"
			if !useLegacy {
				backend = "centrifuge"
			}
			c.JSON(http.StatusOK, gin.H{
				"backend":    backend,
				"use_legacy": useLegacy,
			})
		})
		
		// Debug endpoints for testing enhanced service
		wsAPI.GET("/debug/backend-info", debugHandler.GetBackendInfo)
		wsAPI.POST("/debug/test-switch", debugHandler.TestBackendSwitch)

		// Administrative endpoints
		wsAPI.POST("/users/:userId/disconnect", websocket.RequireRole("admin"), func(c *gin.Context) {
			userID := c.Param("userId")
			wsService.DisconnectUser(userID)
			c.JSON(http.StatusOK, gin.H{
				"message": "User disconnected",
				"user_id": userID,
			})
		})
		
		wsAPI.POST("/projects/:projectId/disconnect", websocket.RequireRole("admin"), func(c *gin.Context) {
			projectIDStr := c.Param("projectId")
			projectID, err := parseUUID(projectIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
				return
			}
			
			wsService.DisconnectProject(projectID)
			c.JSON(http.StatusOK, gin.H{
				"message":    "Project connections disconnected",
				"project_id": projectID,
			})
		})
		
		// Message sending endpoints
		wsAPI.POST("/users/:userId/message", func(c *gin.Context) {
			userID := c.Param("userId")
			
			var request struct {
				Type websocket.MessageType `json:"type" binding:"required"`
				Data interface{}           `json:"data"`
			}
			
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			
			err := wsService.SendDirectMessage(userID, request.Type, request.Data)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			
			c.JSON(http.StatusOK, gin.H{
				"message": "Direct message sent",
				"user_id": userID,
				"type":    request.Type,
			})
		})
		
		wsAPI.POST("/projects/:projectId/message", func(c *gin.Context) {
			projectIDStr := c.Param("projectId")
			projectID, err := parseUUID(projectIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
				return
			}
			
			var request struct {
				Type websocket.MessageType `json:"type" binding:"required"`
				Data interface{}           `json:"data"`
			}
			
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			
			err = wsService.SendProjectMessage(projectID, request.Type, request.Data)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			
			c.JSON(http.StatusOK, gin.H{
				"message":    "Project message sent",
				"project_id": projectID,
				"type":       request.Type,
			})
		})
	}
}

// createCentrifugeHandler creates a Centrifuge WebSocket handler
func createCentrifugeHandler(wsService *websocket.EnhancedService) (*websocket.CentrifugeHandler, error) {
	// Use the Centrifuge handler from the enhanced service if available
	if centrifugeHandler := wsService.GetCentrifugeHandler(); centrifugeHandler != nil {
		return centrifugeHandler, nil
	}
	
	// If not available, return an error to fall back to legacy
	return nil, fmt.Errorf("Centrifuge handler not available in enhanced service")
}