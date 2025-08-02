package handler

import (
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes and middleware
func SetupRoutes(router *gin.Engine, projectUsecase usecase.ProjectUsecase, taskUsecase usecase.TaskUsecase, db *database.GormDB) {
	// Initialize handlers
	projectHandler := NewProjectHandler(projectUsecase)
	taskHandler := NewTaskHandler(taskUsecase)

	// Global middleware
	router.Use(SecurityHeadersMiddleware())
	router.Use(CORSMiddleware())
	router.Use(RequestLoggingMiddleware())
	router.Use(ErrorHandlingMiddleware())
	router.Use(RateLimitMiddleware())
	router.Use(ValidationErrorMiddleware())

	// Health check endpoint (no versioning for health)
	SetupHealthRoutes(router, db)

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
		}

		// Task routes
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("", taskHandler.ListTasks)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
			tasks.PATCH("/:id/status", taskHandler.UpdateTaskStatus)
			tasks.GET("/:id/project", taskHandler.GetTaskWithProject)
		}
	}
}