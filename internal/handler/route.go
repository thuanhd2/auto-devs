package handler

import (
	"github.com/auto-devs/auto-devs/docs"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/auto-devs/auto-devs/internal/websocket"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all API routes and middleware
func SetupRoutes(router *gin.Engine, projectUsecase usecase.ProjectUsecase, taskUsecase usecase.TaskUsecase, db *database.GormDB, wsService *websocket.Service) {
	// Initialize handlers
	projectHandler := NewProjectHandlerWithWebSocket(projectUsecase, wsService)
	taskHandler := NewTaskHandlerWithWebSocket(taskUsecase, wsService)
	wsHandler := wsService.GetHandler()

	// Global middleware
	router.Use(SecurityHeadersMiddleware())
	router.Use(CORSMiddleware())
	router.Use(RequestLoggingMiddleware())
	router.Use(ErrorHandlingMiddleware())
	router.Use(RateLimitMiddleware())
	router.Use(ValidationErrorMiddleware())

	docs.SwaggerInfo.BasePath = "/api/v1"
	// Swagger documentation endpoints (must be before other routes)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.PersistAuthorization(true)))
	// SetupSwaggerRoutes(router)

	// Health check endpoint (no versioning for health)
	SetupHealthRoutes(router, db)

	// WebSocket endpoints
	SetupWebSocketRoutes(router, wsHandler, wsService)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Project routes
		projects := v1.Group("/projects")
		{
			projects.POST("", projectHandler.CreateProject)
			projects.GET("", projectHandler.ListProjects)
			projects.GET("/:id", projectHandler.GetProject)
			projects.PUT("/:id", projectHandler.UpdateProject)
			projects.DELETE("/:id", projectHandler.DeleteProject)
			projects.GET("/:id/tasks", projectHandler.GetProjectWithTasks)
			projects.GET("/:id/statistics", projectHandler.GetProjectStatistics)
			projects.GET("/:id/status-analytics", taskHandler.GetProjectStatusAnalytics)
			projects.POST("/:id/archive", projectHandler.ArchiveProject)
			projects.POST("/:id/restore", projectHandler.RestoreProject)
			projects.GET("/:id/settings", projectHandler.GetProjectSettings)
			projects.PUT("/:id/settings", projectHandler.UpdateProjectSettings)
		}

		// Task routes
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("", taskHandler.ListTasks)
			tasks.GET("/filter", taskHandler.GetTasksWithFilters)
			tasks.PATCH("/bulk-status", taskHandler.BulkUpdateTaskStatus)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
			tasks.PATCH("/:id/status", taskHandler.UpdateTaskStatus)
			tasks.PATCH("/:id/status-with-history", taskHandler.UpdateTaskStatusWithHistory)
			tasks.GET("/:id/status-history", taskHandler.GetTaskStatusHistory)
			tasks.GET("/:id/validate-transition", taskHandler.ValidateTaskStatusTransition)
			tasks.GET("/:id/project", taskHandler.GetTaskWithProject)
		}
	}
}
