package handler

import (
	"net/http"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/handler/dto"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	taskUsecase usecase.TaskUsecase
}

func NewTaskHandler(taskUsecase usecase.TaskUsecase) *TaskHandler {
	return &TaskHandler{
		taskUsecase: taskUsecase,
	}
}

// CreateTask godoc
// @Summary Create a new task
// @Description Create a new task with the provided details
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body dto.TaskCreateRequest true "Task creation data"
// @Success 201 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
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
	c.JSON(http.StatusCreated, response)
}

// GetTask godoc
// @Summary Get a task by ID
// @Description Get a single task by its ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	task, err := h.taskUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Task not found"))
		return
	}

	response := dto.TaskResponseFromEntity(task)
	c.JSON(http.StatusOK, response)
}

// GetTaskWithProject godoc
// @Summary Get a task with its project
// @Description Get a single task by its ID including the associated project
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} dto.TaskWithProjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id}/project [get]
func (h *TaskHandler) GetTaskWithProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	task, err := h.taskUsecase.GetWithProject(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Task not found"))
		return
	}

	var response dto.TaskWithProjectResponse
	response.FromEntity(task)
	c.JSON(http.StatusOK, response)
}

// ListTasks godoc
// @Summary List tasks with filtering
// @Description Get a list of tasks with optional filtering by status, project, or search term
// @Tags tasks
// @Accept json
// @Produce json
// @Param status query string false "Filter by status" Enums(TODO, PLANNING, PLAN_REVIEWING, IMPLEMENTING, CODE_REVIEWING, DONE, CANCELLED)
// @Param project_id query string false "Filter by project ID"
// @Param search query string false "Search in title and description"
// @Success 200 {object} dto.TaskListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	var query dto.TaskFilterQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid query parameters"))
		return
	}

	// For now, we'll implement basic filtering. A more complete implementation
	// would require repository method updates to handle all filters
	var tasks []*entity.Task
	var err error

	if query.Status != nil {
		status := entity.TaskStatus(*query.Status)
		tasks, err = h.taskUsecase.GetByStatus(c.Request.Context(), status)
	} else if query.ProjectID != nil {
		projectID, parseErr := uuid.Parse(*query.ProjectID)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(parseErr, http.StatusBadRequest, "Invalid project ID"))
			return
		}
		tasks, err = h.taskUsecase.GetByProjectID(c.Request.Context(), projectID)
	} else {
		// For now, we'll return all tasks. In a real implementation,
		// we'd implement a GetAll method or handle pagination properly
		c.JSON(http.StatusNotImplemented, dto.NewErrorResponse(nil, http.StatusNotImplemented, "General task listing not yet implemented"))
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to fetch tasks"))
		return
	}

	response := dto.TaskListResponseFromEntities(tasks)
	c.JSON(http.StatusOK, response)
}

// ListTasksByProject godoc
// @Summary List tasks by project
// @Description Get all tasks for a specific project
// @Tags tasks
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Success 200 {object} dto.TaskListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{project_id}/tasks [get]
func (h *TaskHandler) ListTasksByProject(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	tasks, err := h.taskUsecase.GetByProjectID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to fetch tasks"))
		return
	}

	response := dto.TaskListResponseFromEntities(tasks)
	c.JSON(http.StatusOK, response)
}

// UpdateTask godoc
// @Summary Update a task
// @Description Update a task with the provided details
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param task body dto.TaskUpdateRequest true "Task update data"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
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

	usecaseReq := usecase.UpdateTaskRequest{}
	if req.Title != nil {
		usecaseReq.Title = *req.Title
	}
	if req.Description != nil {
		usecaseReq.Description = *req.Description
	}
	if req.BranchName != nil {
		usecaseReq.BranchName = req.BranchName
	}
	if req.PullRequest != nil {
		usecaseReq.PullRequest = req.PullRequest
	}

	task, err := h.taskUsecase.Update(c.Request.Context(), id, usecaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to update task"))
		return
	}

	response := dto.TaskResponseFromEntity(task)
	c.JSON(http.StatusOK, response)
}

// UpdateTaskStatus godoc
// @Summary Update a task status
// @Description Update the status of a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param status body dto.TaskStatusUpdateRequest true "Task status update data"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id}/status [patch]
func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
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

	task, err := h.taskUsecase.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to update task status"))
		return
	}

	response := dto.TaskResponseFromEntity(task)
	c.JSON(http.StatusOK, response)
}

// DeleteTask godoc
// @Summary Delete a task
// @Description Delete a task by its ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	err = h.taskUsecase.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to delete task"))
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateTaskStatusWithHistory godoc
// @Summary Update task status with history tracking
// @Description Update the status of a task with validation and history tracking
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param status body dto.TaskStatusUpdateWithHistoryRequest true "Task status update with history data"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id}/status-with-history [patch]
func (h *TaskHandler) UpdateTaskStatusWithHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	var req dto.TaskStatusUpdateWithHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	usecaseReq := usecase.UpdateStatusRequest{
		TaskID:    id,
		Status:    req.Status,
		ChangedBy: req.ChangedBy,
		Reason:    req.Reason,
	}

	task, err := h.taskUsecase.UpdateStatusWithHistory(c.Request.Context(), usecaseReq)
	if err != nil {
		// Check if it's a validation error
		if _, ok := err.(*entity.TaskStatusValidationError); ok {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to update task status"))
		return
	}

	response := dto.TaskResponseFromEntity(task)
	c.JSON(http.StatusOK, response)
}

// BulkUpdateTaskStatus godoc
// @Summary Bulk update task statuses
// @Description Update status for multiple tasks at once
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body dto.BulkStatusUpdateRequest true "Bulk status update data"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/bulk-status [patch]
func (h *TaskHandler) BulkUpdateTaskStatus(c *gin.Context) {
	var req dto.BulkStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	usecaseReq := usecase.BulkUpdateStatusRequest{
		TaskIDs:   req.TaskIDs,
		Status:    req.Status,
		ChangedBy: req.ChangedBy,
	}

	err := h.taskUsecase.BulkUpdateStatus(c.Request.Context(), usecaseReq)
	if err != nil {
		// Check if it's a validation error
		if _, ok := err.(*entity.TaskStatusValidationError); ok {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to bulk update task statuses"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Task statuses updated successfully",
	})
}

// GetTaskStatusHistory godoc
// @Summary Get task status history
// @Description Get the status change history for a specific task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {array} dto.TaskStatusHistoryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id}/status-history [get]
func (h *TaskHandler) GetTaskStatusHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	history, err := h.taskUsecase.GetStatusHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Task not found or no history available"))
		return
	}

	response := dto.TaskStatusHistoryListFromEntities(history)
	c.JSON(http.StatusOK, response)
}

// GetProjectStatusAnalytics godoc
// @Summary Get project status analytics
// @Description Get comprehensive status analytics for a project
// @Tags tasks,analytics
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.TaskStatusAnalyticsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/status-analytics [get]
func (h *TaskHandler) GetProjectStatusAnalytics(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
		return
	}

	analytics, err := h.taskUsecase.GetStatusAnalytics(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to generate status analytics"))
		return
	}

	response := dto.TaskStatusAnalyticsResponseFromEntity(analytics)
	c.JSON(http.StatusOK, response)
}

// GetTasksWithFilters godoc
// @Summary Get tasks with advanced filtering
// @Description Get tasks with comprehensive filtering and sorting options
// @Tags tasks
// @Accept json
// @Produce json
// @Param project_id query string false "Filter by project ID"
// @Param status query string false "Filter by single status"
// @Param statuses query array false "Filter by multiple statuses"
// @Param created_after query string false "Filter by creation date (after)"
// @Param created_before query string false "Filter by creation date (before)"
// @Param search query string false "Search in title and description"
// @Param limit query int false "Limit number of results"
// @Param offset query int false "Offset for pagination"
// @Param order_by query string false "Field to order by"
// @Param order_dir query string false "Order direction (asc/desc)"
// @Success 200 {object} dto.TaskListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/filter [get]
func (h *TaskHandler) GetTasksWithFilters(c *gin.Context) {
	var query dto.TaskAdvancedFilterQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid query parameters"))
		return
	}

	// Convert DTO to usecase request
	usecaseReq := usecase.GetTasksFilterRequest{
		SearchTerm:    query.SearchTerm,
		CreatedAfter:  query.CreatedAfter,
		CreatedBefore: query.CreatedBefore,
		Limit:         query.Limit,
		Offset:        query.Offset,
		OrderBy:       query.OrderBy,
		OrderDir:      query.OrderDir,
	}

	// Parse project ID if provided
	if query.ProjectID != nil {
		projectID, err := uuid.Parse(*query.ProjectID)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid project ID"))
			return
		}
		usecaseReq.ProjectID = &projectID
	}

	// Parse statuses
	if len(query.Statuses) > 0 {
		statuses := make([]entity.TaskStatus, len(query.Statuses))
		for i, s := range query.Statuses {
			statuses[i] = entity.TaskStatus(s)
		}
		usecaseReq.Statuses = statuses
	} else if query.Status != nil {
		status := entity.TaskStatus(*query.Status)
		usecaseReq.Statuses = []entity.TaskStatus{status}
	}

	tasks, err := h.taskUsecase.GetTasksWithFilters(c.Request.Context(), usecaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to fetch tasks"))
		return
	}

	response := dto.TaskListResponseFromEntities(tasks)
	c.JSON(http.StatusOK, response)
}

// ValidateTaskStatusTransition godoc
// @Summary Validate task status transition
// @Description Check if a status transition is valid for a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param target_status query string true "Target status to validate"
// @Success 200 {object} dto.TaskStatusValidationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id}/validate-transition [get]
func (h *TaskHandler) ValidateTaskStatusTransition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	targetStatusStr := c.Query("target_status")
	if targetStatusStr == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(nil, http.StatusBadRequest, "target_status query parameter is required"))
		return
	}

	targetStatus := entity.TaskStatus(targetStatusStr)
	if !targetStatus.IsValid() {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(nil, http.StatusBadRequest, "Invalid target status"))
		return
	}

	// Get current task to show current status
	task, err := h.taskUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Task not found"))
		return
	}

	// Validate transition
	err = h.taskUsecase.ValidateStatusTransition(c.Request.Context(), id, targetStatus)

	response := dto.TaskStatusValidationResponse{
		Valid:         err == nil,
		CurrentStatus: task.Status,
		TargetStatus:  targetStatus,
	}

	if err != nil {
		response.Message = err.Error()
	} else {
		response.Message = "Transition is valid"
	}

	c.JSON(http.StatusOK, response)
}

// UpdateTaskGitStatus godoc
// @Summary Update task Git status
// @Description Update the Git status of a task with validation
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param git_status body dto.TaskGitStatusUpdateRequest true "Git status update data"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id}/git-status [patch]
func (h *TaskHandler) UpdateTaskGitStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	var req dto.TaskGitStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	task, err := h.taskUsecase.UpdateGitStatus(c.Request.Context(), id, req.GitStatus)
	if err != nil {
		// Check if it's a validation error
		if _, ok := err.(*entity.TaskGitStatusValidationError); ok {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to update task Git status"))
		return
	}

	response := dto.TaskResponseFromEntity(task)
	c.JSON(http.StatusOK, response)
}

// ValidateTaskGitStatusTransition godoc
// @Summary Validate task Git status transition
// @Description Check if a Git status transition is valid for a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param target_git_status query string true "Target Git status to validate"
// @Success 200 {object} dto.TaskGitStatusValidationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id}/validate-git-transition [get]
func (h *TaskHandler) ValidateTaskGitStatusTransition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	targetGitStatusStr := c.Query("target_git_status")
	if targetGitStatusStr == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(nil, http.StatusBadRequest, "target_git_status query parameter is required"))
		return
	}

	targetGitStatus := entity.TaskGitStatus(targetGitStatusStr)
	if !targetGitStatus.IsValid() {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(nil, http.StatusBadRequest, "Invalid target Git status"))
		return
	}

	// Get current task to show current Git status
	task, err := h.taskUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Task not found"))
		return
	}

	// Validate transition
	err = h.taskUsecase.ValidateGitStatusTransition(c.Request.Context(), id, targetGitStatus)

	response := dto.TaskGitStatusValidationResponse{
		Valid:            err == nil,
		CurrentGitStatus: task.GitStatus,
		TargetGitStatus:  targetGitStatus,
	}

	if err != nil {
		response.Message = err.Error()
	} else {
		response.Message = "Git status transition is valid"
	}

	c.JSON(http.StatusOK, response)
}

// StartPlanning godoc
// @Summary Start planning for a task
// @Description Start the planning phase for a task by selecting a branch and initiating background processing
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body dto.StartPlanningRequest true "Start planning request"
// @Success 200 {object} dto.StartPlanningResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id}/start-planning [post]
func (h *TaskHandler) StartPlanning(c *gin.Context) {
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

	// Validate that task exists and is in TODO status
	task, err := h.taskUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Task not found"))
		return
	}

	if task.Status != entity.TaskStatusTODO {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(nil, http.StatusBadRequest, "Task must be in TODO status to start planning"))
		return
	}

	// Start planning (this will enqueue a background job)
	jobID, err := h.taskUsecase.StartPlanning(c.Request.Context(), id, req.BranchName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to start planning"))
		return
	}

	response := dto.StartPlanningResponse{
		Message: "Planning started successfully",
		JobID:   jobID,
	}
	c.JSON(http.StatusOK, response)
}
