package handler

import (
	"github.com/gin-gonic/gin"
)

// RegisterWorktreeRoutes registers all worktree-related routes
func RegisterWorktreeRoutes(router *gin.RouterGroup, worktreeHandler *WorktreeHandler) {
	worktrees := router.Group("/worktrees")
	{
		// Basic worktree operations
		worktrees.POST("", worktreeHandler.CreateWorktreeForTask)
		worktrees.POST("/cleanup", worktreeHandler.CleanupWorktreeForTask)

		// Worktree retrieval
		worktrees.GET("/task/:taskId", worktreeHandler.GetWorktreeByTaskID)
		worktrees.GET("/project/:projectId", worktreeHandler.GetWorktreesByProjectID)

		// Worktree management
		worktrees.PUT("/:worktreeId/status", worktreeHandler.UpdateWorktreeStatus)
		worktrees.POST("/:worktreeId/initialize", worktreeHandler.InitializeWorktree)
		worktrees.POST("/:worktreeId/recover", worktreeHandler.RecoverFailedWorktree)

		// Worktree validation and health
		worktrees.GET("/:worktreeId/validate", worktreeHandler.ValidateWorktree)
		worktrees.GET("/:worktreeId/health", worktreeHandler.GetWorktreeHealth)

		// Branch management
		worktrees.GET("/:worktreeId/branch", worktreeHandler.GetBranchInfo)

		// Statistics and monitoring
		worktrees.GET("/project/:projectId/statistics", worktreeHandler.GetWorktreeStatistics)
		worktrees.GET("/project/:projectId/active-count", worktreeHandler.GetActiveWorktreesCount)
	}
}
