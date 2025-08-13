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

	// Notify task deleted

	c.Status(http.StatusNoContent)
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
	jobID, err := h.taskUsecase.StartPlanning(c.Request.Context(), id, req.BranchName, req.AIType)
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

// ApprovePlan godoc
// @Summary Approve plan and start implementation
// @Description Approve the plan for a task and enqueue implementation job
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body dto.ApprovePlanRequest true "Approve plan request"
// @Success 200 {object} dto.StartPlanningResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id}/approve-plan [post]
func (h *TaskHandler) ApprovePlan(c *gin.Context) {
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

	// Validate that task exists and is in PLAN_REVIEWING status
	task, err := h.taskUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Task not found"))
		return
	}

	if task.Status != entity.TaskStatusPLANREVIEWING {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(nil, http.StatusBadRequest, "Task must be in PLAN_REVIEWING status to approve plan"))
		return
	}

	// Approve plan and start implementation (this will enqueue a background job)
	jobID, err := h.taskUsecase.ApprovePlan(c.Request.Context(), id, req.AIType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to approve plan and start implementation"))
		return
	}

	response := dto.StartPlanningResponse{
		Message: "Plan approved and implementation started successfully",
		JobID:   jobID,
	}
	c.JSON(http.StatusOK, response)
}

func (h *TaskHandler) GetPullRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	pr, err := h.taskUsecase.GetPullRequest(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(err, http.StatusNotFound, "Pull request not found"))
		return
	}

	c.JSON(http.StatusOK, pr)
}
