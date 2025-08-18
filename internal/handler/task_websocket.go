package handler

import (
	"log"
	"net/http"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/handler/dto"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/auto-devs/auto-devs/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TaskHandlerWithWebSocket extends the basic task handler with WebSocket notifications
type TaskHandlerWithWebSocket struct {
	*TaskHandler
	wsService *websocket.Service
}

// NewTaskHandlerWithWebSocket creates a new task handler with WebSocket support
func NewTaskHandlerWithWebSocket(taskUsecase usecase.TaskUsecase, wsService *websocket.Service) *TaskHandlerWithWebSocket {
	return &TaskHandlerWithWebSocket{
		TaskHandler: NewTaskHandler(taskUsecase),
		wsService:   wsService,
	}
}

// CreateTask creates a new task and sends WebSocket notification
func (h *TaskHandlerWithWebSocket) CreateTask(c *gin.Context) {
	var req dto.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	usecaseReq := usecase.CreateTaskRequest{
		ProjectID:   req.ProjectID,
		Title:       req.Title,
		Description: req.Description,
	}

	task, err := h.taskUsecase.Create(c.Request.Context(), usecaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to create task"))
		return
	}

	response := dto.TaskResponseFromEntity(task)

	// Send WebSocket notification
	if err := h.wsService.NotifyTaskCreated(response, task.ProjectID); err != nil {
		log.Printf("Failed to send WebSocket notification for task creation: %v", err)
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateTask updates a task and sends WebSocket notification
func (h *TaskHandlerWithWebSocket) UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	var req dto.TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	// Get the original task to track changes
	originalTask, err := h.taskUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Task not found"))
		return
	}

	usecaseReq := usecase.UpdateTaskRequest{}
	changes := make(map[string]interface{})

	if req.Title != nil && *req.Title != originalTask.Title {
		usecaseReq.Title = *req.Title
		changes["title"] = map[string]interface{}{
			"old": originalTask.Title,
			"new": *req.Title,
		}
	}
	if req.Description != nil && *req.Description != originalTask.Description {
		usecaseReq.Description = *req.Description
		changes["description"] = map[string]interface{}{
			"old": originalTask.Description,
			"new": *req.Description,
		}
	}
	if req.BranchName != nil && (originalTask.BranchName == nil || *req.BranchName != *originalTask.BranchName) {
		usecaseReq.BranchName = req.BranchName
		changes["branch_name"] = map[string]interface{}{
			"old": originalTask.BranchName,
			"new": req.BranchName,
		}
	}
	if req.PullRequest != nil && (originalTask.PullRequest == nil || *req.PullRequest != *originalTask.PullRequest) {
		usecaseReq.PullRequest = req.PullRequest
		changes["pull_request"] = map[string]interface{}{
			"old": originalTask.PullRequest,
			"new": req.PullRequest,
		}
	}
	if req.Status != nil && *req.Status != originalTask.Status {
		usecaseReq.Status = req.Status
		changes["status"] = map[string]interface{}{
			"old": originalTask.Status,
			"new": req.Status,
		}
	}

	task, err := h.taskUsecase.Update(c.Request.Context(), id, usecaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to update task"))
		return
	}

	response := dto.TaskResponseFromEntity(task)

	// Send WebSocket notification if there were changes
	if len(changes) > 0 {
		if err := h.wsService.NotifyTaskUpdated(task.ID, task.ProjectID, changes, response); err != nil {
			log.Printf("Failed to send WebSocket notification for task update: %v", err)
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateTaskStatus updates a task status and sends WebSocket notification
func (h *TaskHandlerWithWebSocket) UpdateTaskStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	var req dto.TaskStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	// Get the original task to track status change
	originalTask, err := h.taskUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Task not found"))
		return
	}

	task, err := h.taskUsecase.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to update task status"))
		return
	}

	response := dto.TaskResponseFromEntity(task)

	// Send WebSocket notifications for status change
	if originalTask.Status != task.Status {
		changes := map[string]interface{}{
			"status": map[string]interface{}{
				"old": originalTask.Status,
				"new": task.Status,
			},
		}

		// Send task updated notification
		if err := h.wsService.NotifyTaskUpdated(task.ID, task.ProjectID, changes, response); err != nil {
			log.Printf("Failed to send WebSocket notification for task update: %v", err)
		}

		// Send status changed notification
		if err := h.wsService.NotifyStatusChanged(task.ID, task.ProjectID, "task", string(originalTask.Status), string(task.Status)); err != nil {
			log.Printf("Failed to send WebSocket notification for status change: %v", err)
		}
	}

	c.JSON(http.StatusOK, response)
}

// DeleteTask deletes a task and sends WebSocket notification
func (h *TaskHandlerWithWebSocket) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	// Get the task before deleting to get the project ID
	task, err := h.taskUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Task not found"))
		return
	}

	err = h.taskUsecase.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to delete task"))
		return
	}

	// Send WebSocket notification
	if err := h.wsService.NotifyTaskDeleted(task.ID, task.ProjectID); err != nil {
		log.Printf("Failed to send WebSocket notification for task deletion: %v", err)
	}

	c.Status(http.StatusNoContent)
}

// StartPlanning starts planning for a task with immediate status update and WebSocket notification
func (h *TaskHandlerWithWebSocket) StartPlanning(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	var req dto.StartPlanningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	// Get the original task to track changes
	originalTask, err := h.taskUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Task not found"))
		return
	}

	if originalTask.Status != entity.TaskStatusTODO {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(nil, http.StatusBadRequest, "Task must be in TODO status to start planning"))
		return
	}

	// Immediately update task status to PLANNING to provide instant UI feedback
	updatedTask, err := h.taskUsecase.UpdateStatus(c.Request.Context(), id, entity.TaskStatusPLANNING)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to update task status"))
		return
	}

	response := dto.TaskResponseFromEntity(updatedTask)

	// Send WebSocket notifications for status change
	changes := map[string]interface{}{
		"status": map[string]interface{}{
			"old": originalTask.Status,
			"new": updatedTask.Status,
		},
	}

	// Send task updated notification
	if err := h.wsService.NotifyTaskUpdated(updatedTask.ID, updatedTask.ProjectID, changes, response); err != nil {
		log.Printf("Failed to send WebSocket notification for task update: %v", err)
	}

	// Send status changed notification
	if err := h.wsService.NotifyStatusChanged(updatedTask.ID, updatedTask.ProjectID, "task", string(originalTask.Status), string(updatedTask.Status)); err != nil {
		log.Printf("Failed to send WebSocket notification for status change: %v", err)
	}

	// Start the background planning job using the usecase
	jobID, err := h.TaskHandler.taskUsecase.StartPlanning(c.Request.Context(), id, req.BranchName, req.AIType)
	if err != nil {
		// Revert status if job enqueueing fails
		_, revertErr := h.taskUsecase.UpdateStatus(c.Request.Context(), id, entity.TaskStatusTODO)
		if revertErr != nil {
			log.Printf("Failed to revert task status after job enqueueing failed: %v", revertErr)
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to start planning"))
		return
	}

	planningResponse := dto.StartPlanningResponse{
		Message: "Planning started successfully",
		JobID:   jobID,
	}
	c.JSON(http.StatusOK, planningResponse)
}

// ApprovePlan approves a plan and starts implementation with WebSocket notification
func (h *TaskHandlerWithWebSocket) ApprovePlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	var req dto.ApprovePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	// Get the original task to track changes
	originalTask, err := h.taskUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Task not found"))
		return
	}

	if originalTask.Status != entity.TaskStatusPLANREVIEWING {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(nil, http.StatusBadRequest, "Task must be in PLAN_REVIEWING status to approve plan"))
		return
	}

	// Immediately update task status to IMPLEMENTING to provide instant UI feedback
	updatedTask, err := h.taskUsecase.UpdateStatus(c.Request.Context(), id, entity.TaskStatusIMPLEMENTING)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to update task status"))
		return
	}

	response := dto.TaskResponseFromEntity(updatedTask)

	// Send WebSocket notifications for status change
	changes := map[string]interface{}{
		"status": map[string]interface{}{
			"old": originalTask.Status,
			"new": updatedTask.Status,
		},
	}

	// Send task updated notification
	if err := h.wsService.NotifyTaskUpdated(updatedTask.ID, updatedTask.ProjectID, changes, response); err != nil {
		log.Printf("Failed to send WebSocket notification for task update: %v", err)
	}

	// Send status changed notification
	if err := h.wsService.NotifyStatusChanged(updatedTask.ID, updatedTask.ProjectID, "task", string(originalTask.Status), string(updatedTask.Status)); err != nil {
		log.Printf("Failed to send WebSocket notification for status change: %v", err)
	}

	// Start the background implementation job using the usecase
	jobID, err := h.TaskHandler.taskUsecase.ApprovePlan(c.Request.Context(), id, req.AIType)
	if err != nil {
		// Revert status if job enqueueing fails
		_, revertErr := h.taskUsecase.UpdateStatus(c.Request.Context(), id, entity.TaskStatusPLANREVIEWING)
		if revertErr != nil {
			log.Printf("Failed to revert task status after job enqueueing failed: %v", revertErr)
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to approve plan and start implementation"))
		return
	}

	planningResponse := dto.StartPlanningResponse{
		Message: "Plan approved and implementation started successfully",
		JobID:   jobID,
	}
	c.JSON(http.StatusOK, planningResponse)
}
