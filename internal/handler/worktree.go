package handler

import (
	"net/http"

	"github.com/auto-devs/auto-devs/internal/handler/dto"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WorktreeHandler struct {
	worktreeUsecase usecase.WorktreeUsecase
}

func NewWorktreeHandler(worktreeUsecase usecase.WorktreeUsecase) *WorktreeHandler {
	return &WorktreeHandler{
		worktreeUsecase: worktreeUsecase,
	}
}

// CreateWorktreeForTask creates a worktree for a task
// @Summary Create worktree for task
// @Description Create a new Git worktree for a specific task
// @Tags worktrees
// @Accept json
// @Produce json
// @Param request body dto.CreateWorktreeRequest true "Create worktree request"
// @Success 201 {object} dto.WorktreeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /worktrees [post]
func (h *WorktreeHandler) CreateWorktreeForTask(c *gin.Context) {
	var req dto.CreateWorktreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Convert DTO to usecase request
	usecaseReq := usecase.CreateWorktreeRequest{
		TaskID:     req.TaskID,
		ProjectID:  req.ProjectID,
		TaskTitle:  req.TaskTitle,
		Repository: req.Repository,
	}

	worktree, err := h.worktreeUsecase.CreateWorktreeForTask(c.Request.Context(), usecaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to create worktree",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.WorktreeResponse{
		Worktree: worktree,
		Message:  "Worktree created successfully",
	})
}

// CleanupWorktreeForTask cleans up a worktree for a task
// @Summary Cleanup worktree for task
// @Description Clean up a Git worktree for a specific task
// @Tags worktrees
// @Accept json
// @Produce json
// @Param request body dto.CleanupWorktreeRequest true "Cleanup worktree request"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /worktrees/cleanup [post]
func (h *WorktreeHandler) CleanupWorktreeForTask(c *gin.Context) {
	var req dto.CleanupWorktreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Convert DTO to usecase request
	usecaseReq := usecase.CleanupWorktreeRequest{
		TaskID:     req.TaskID,
		ProjectID:  req.ProjectID,
		BranchName: req.BranchName,
		Force:      req.Force,
	}

	err := h.worktreeUsecase.CleanupWorktreeForTask(c.Request.Context(), usecaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to cleanup worktree",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Worktree cleaned up successfully",
	})
}

// GetWorktreeByTaskID gets worktree information for a task
// @Summary Get worktree by task ID
// @Description Get worktree information for a specific task
// @Tags worktrees
// @Produce json
// @Param taskId path string true "Task ID"
// @Success 200 {object} dto.WorktreeResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /worktrees/task/{taskId} [get]
func (h *WorktreeHandler) GetWorktreeByTaskID(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid task ID",
			Message: err.Error(),
		})
		return
	}

	worktree, err := h.worktreeUsecase.GetWorktreeByTaskID(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Worktree not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.WorktreeResponse{
		Worktree: worktree,
	})
}

// GetWorktreesByProjectID gets all worktrees for a project
// @Summary Get worktrees by project ID
// @Description Get all worktrees for a specific project
// @Tags worktrees
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {object} dto.WorktreesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /worktrees/project/{projectId} [get]
func (h *WorktreeHandler) GetWorktreesByProjectID(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid project ID",
			Message: err.Error(),
		})
		return
	}

	worktrees, err := h.worktreeUsecase.GetWorktreesByProjectID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get worktrees",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.WorktreesResponse{
		Worktrees: worktrees,
		Count:     len(worktrees),
	})
}

// UpdateWorktreeStatus updates the status of a worktree
// @Summary Update worktree status
// @Description Update the status of a worktree
// @Tags worktrees
// @Accept json
// @Produce json
// @Param worktreeId path string true "Worktree ID"
// @Param status body dto.UpdateWorktreeStatusRequest true "Update status request"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /worktrees/{worktreeId}/status [put]
func (h *WorktreeHandler) UpdateWorktreeStatus(c *gin.Context) {
	worktreeIDStr := c.Param("worktreeId")
	worktreeID, err := uuid.Parse(worktreeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid worktree ID",
			Message: err.Error(),
		})
		return
	}

	var req dto.UpdateWorktreeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	err = h.worktreeUsecase.UpdateWorktreeStatus(c.Request.Context(), worktreeID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to update worktree status",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Worktree status updated successfully",
	})
}

// ValidateWorktree validates a worktree
// @Summary Validate worktree
// @Description Validate the health and integrity of a worktree
// @Tags worktrees
// @Produce json
// @Param worktreeId path string true "Worktree ID"
// @Success 200 {object} dto.WorktreeValidationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /worktrees/{worktreeId}/validate [get]
func (h *WorktreeHandler) ValidateWorktree(c *gin.Context) {
	worktreeIDStr := c.Param("worktreeId")
	worktreeID, err := uuid.Parse(worktreeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid worktree ID",
			Message: err.Error(),
		})
		return
	}

	result, err := h.worktreeUsecase.ValidateWorktree(c.Request.Context(), worktreeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to validate worktree",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.WorktreeValidationResponse{
		ValidationResult: result,
	})
}

// GetWorktreeHealth gets health information for a worktree
// @Summary Get worktree health
// @Description Get health information for a worktree
// @Tags worktrees
// @Produce json
// @Param worktreeId path string true "Worktree ID"
// @Success 200 {object} dto.WorktreeHealthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /worktrees/{worktreeId}/health [get]
func (h *WorktreeHandler) GetWorktreeHealth(c *gin.Context) {
	worktreeIDStr := c.Param("worktreeId")
	worktreeID, err := uuid.Parse(worktreeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid worktree ID",
			Message: err.Error(),
		})
		return
	}

	health, err := h.worktreeUsecase.GetWorktreeHealth(c.Request.Context(), worktreeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get worktree health",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.WorktreeHealthResponse{
		Health: health,
	})
}

// GetBranchInfo gets branch information for a worktree
// @Summary Get branch info
// @Description Get branch information for a worktree
// @Tags worktrees
// @Produce json
// @Param worktreeId path string true "Worktree ID"
// @Success 200 {object} dto.BranchInfoResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /worktrees/{worktreeId}/branch [get]
func (h *WorktreeHandler) GetBranchInfo(c *gin.Context) {
	worktreeIDStr := c.Param("worktreeId")
	worktreeID, err := uuid.Parse(worktreeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid worktree ID",
			Message: err.Error(),
		})
		return
	}

	branchInfo, err := h.worktreeUsecase.GetBranchInfo(c.Request.Context(), worktreeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get branch info",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.BranchInfoResponse{
		BranchInfo: branchInfo,
	})
}

// InitializeWorktree initializes a worktree
// @Summary Initialize worktree
// @Description Initialize a worktree with basic configuration
// @Tags worktrees
// @Produce json
// @Param worktreeId path string true "Worktree ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /worktrees/{worktreeId}/initialize [post]
func (h *WorktreeHandler) InitializeWorktree(c *gin.Context) {
	worktreeIDStr := c.Param("worktreeId")
	worktreeID, err := uuid.Parse(worktreeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid worktree ID",
			Message: err.Error(),
		})
		return
	}

	err = h.worktreeUsecase.InitializeWorktree(c.Request.Context(), worktreeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to initialize worktree",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Worktree initialized successfully",
	})
}

// RecoverFailedWorktree recovers a failed worktree
// @Summary Recover failed worktree
// @Description Attempt to recover a worktree that is in error status
// @Tags worktrees
// @Produce json
// @Param worktreeId path string true "Worktree ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /worktrees/{worktreeId}/recover [post]
func (h *WorktreeHandler) RecoverFailedWorktree(c *gin.Context) {
	worktreeIDStr := c.Param("worktreeId")
	worktreeID, err := uuid.Parse(worktreeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid worktree ID",
			Message: err.Error(),
		})
		return
	}

	err = h.worktreeUsecase.RecoverFailedWorktree(c.Request.Context(), worktreeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to recover worktree",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Worktree recovered successfully",
	})
}

// GetWorktreeStatistics gets worktree statistics for a project
// @Summary Get worktree statistics
// @Description Get worktree statistics for a project
// @Tags worktrees
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {object} dto.WorktreeStatisticsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /worktrees/project/{projectId}/statistics [get]
func (h *WorktreeHandler) GetWorktreeStatistics(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid project ID",
			Message: err.Error(),
		})
		return
	}

	statistics, err := h.worktreeUsecase.GetWorktreeStatistics(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get worktree statistics",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.WorktreeStatisticsResponse{
		Statistics: statistics,
	})
}

// GetActiveWorktreesCount gets the count of active worktrees for a project
// @Summary Get active worktrees count
// @Description Get the count of active worktrees for a project
// @Tags worktrees
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {object} dto.WorktreeCountResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /worktrees/project/{projectId}/active-count [get]
func (h *WorktreeHandler) GetActiveWorktreesCount(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid project ID",
			Message: err.Error(),
		})
		return
	}

	count, err := h.worktreeUsecase.GetActiveWorktreesCount(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get active worktrees count",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.WorktreeCountResponse{
		Count: count,
	})
}
