package handler

import (
	"net/http"

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

	// WebSocket management API (for monitoring/admin)
	wsAPI := router.Group("/api/v1/websocket")
	{

		// Connection management endpoints
		wsAPI.GET("/connections", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"connections": 100,
			})
		})
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

			count := 100
			c.JSON(http.StatusOK, gin.H{
				"project_id": projectID,
				"count":      count,
			})
		})

		wsAPI.GET("/users/:userId/connections/count", func(c *gin.Context) {
			userID := c.Param("userId")
			count := 100
			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"count":   count,
			})
		})

		// Administrative endpoints
		wsAPI.POST("/users/:userId/disconnect", func(c *gin.Context) {
			userID := c.Param("userId")
			wsService.DisconnectUser(userID)
			c.JSON(http.StatusOK, gin.H{
				"message": "User disconnected",
				"user_id": userID,
			})
		})

		wsAPI.POST("/projects/:projectId/disconnect", func(c *gin.Context) {
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
