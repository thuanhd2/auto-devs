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
func SetupRoutes(router *gin.Engine, projectUsecase usecase.ProjectUsecase, taskUsecase usecase.TaskUsecase, executionUsecase usecase.ExecutionUsecase, db *database.GormDB, wsService *websocket.Service) {
	// Initialize handlers
	projectHandler := NewProjectHandlerWithWebSocket(projectUsecase, wsService)
	taskHandler := NewTaskHandlerWithWebSocket(taskUsecase, wsService)
	executionHandler := NewExecutionHandler(executionUsecase)
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
	// router.GET("/ws", WebSocketMiddleware(), wsHandler.GetWebSocketHandler())

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
			projects.GET("/:id/statistics", projectHandler.GetProjectStatistics)
			projects.POST("/:id/archive", projectHandler.ArchiveProject)
			projects.POST("/:id/restore", projectHandler.RestoreProject)

			// Git repository management endpoints
			projects.POST("/:id/git/reinit", projectHandler.ReinitGitRepository)
			// Git branches endpoint
			projects.GET("/:id/branches", projectHandler.ListBranches)
		}

		// Task routes
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("", taskHandler.ListTasks)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)

			// Planning workflow endpoints
			tasks.POST("/:id/start-planning", taskHandler.StartPlanning)
			tasks.POST("/:id/approve-plan", taskHandler.ApprovePlan)

			// Execution endpoints for tasks
			tasks.GET("/:id/executions", executionHandler.GetTaskExecutions)

			// Pull request endpoints
			tasks.GET("/:id/pull-request", taskHandler.GetPullRequest)
			tasks.POST("/:id/pull-request", taskHandler.CreatePullRequest)

			// Plan endpoints
			tasks.GET("/:id/plans", taskHandler.GetTaskPlans)
			tasks.PUT("/:id/plans/:planId", taskHandler.UpdateTaskPlan)

			// Open with Cursor endpoint
			tasks.POST("/:id/open-with-cursor", taskHandler.OpenWithCursor)

			// Git diff endpoint
			tasks.GET("/:id/diff", taskHandler.GetTaskDiff)
		}

		// Project-scoped task routes
		v1.GET("/projects/:project_id/tasks", taskHandler.ListTasksByProject)
		v1.GET("/projects/:project_id/tasks/done", taskHandler.ListDoneTasksByProject)

		// Execution routes
		executions := v1.Group("/executions")
		{
			executions.POST("", executionHandler.CreateExecution)
			executions.GET("/stats", executionHandler.GetExecutionStats)
			executions.GET("/:id", executionHandler.GetExecutionByID)
			executions.PUT("/:id", executionHandler.UpdateExecution)
			executions.DELETE("/:id", executionHandler.DeleteExecution)
			executions.GET("/:id/logs", executionHandler.GetExecutionLogs)
		}
	}
}
