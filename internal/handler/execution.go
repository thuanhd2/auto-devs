package handler

import (
	"net/http"
	"strconv"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/handler/dto"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExecutionHandler struct {
	executionUsecase usecase.ExecutionUsecase
}

func NewExecutionHandler(executionUsecase usecase.ExecutionUsecase) *ExecutionHandler {
	return &ExecutionHandler{
		executionUsecase: executionUsecase,
	}
}

// GetTaskExecutions godoc
// @Summary Get all executions for a task
// @Description Get all executions for a specific task with optional filtering
// @Tags executions
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param status query string false "Filter by status" Enums(pending,running,paused,completed,failed,cancelled)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param order_by query string false "Order by field" default("started_at")
// @Param order_dir query string false "Order direction" default("desc") Enums(asc,desc)
// @Success 200 {object} dto.ExecutionListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id}/executions [get]
func (h *ExecutionHandler) GetTaskExecutions(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	var query dto.ExecutionFilterQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid query parameters"))
		return
	}

	// Build filter request
	filterReq := usecase.GetExecutionsFilterRequest{
		TaskID: &taskID,
		Limit:  query.PageSize,
		Offset: (query.Page - 1) * query.PageSize,
	}

	// Apply optional filters
	if query.Status != nil {
		status := entity.ExecutionStatus(*query.Status)
		filterReq.Statuses = []entity.ExecutionStatus{status}
	}
	if query.StartedAfter != nil {
		filterReq.StartedAfter = query.StartedAfter
	}
	if query.StartedBefore != nil {
		filterReq.StartedBefore = query.StartedBefore
	}
	if query.WithErrors != nil {
		filterReq.WithErrors = query.WithErrors
	}
	if query.OrderBy != nil {
		filterReq.OrderBy = *query.OrderBy
	} else {
		filterReq.OrderBy = "started_at"
	}
	if query.OrderDir != nil {
		filterReq.OrderDir = *query.OrderDir
	} else {
		filterReq.OrderDir = "desc"
	}

	executions, total, err := h.executionUsecase.GetByStatusFiltered(c.Request.Context(), filterReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to get executions"))
		return
	}

	// Calculate pagination metadata
	totalPages := int(total) / query.PageSize
	if int(total)%query.PageSize > 0 {
		totalPages++
	}

	meta := dto.PaginationMeta{
		Page:       query.Page,
		PageSize:   query.PageSize,
		Total:      int(total),
		TotalPages: totalPages,
	}

	response := dto.ToExecutionListResponse(executions, meta)
	c.JSON(http.StatusOK, response)
}

// GetExecutionByID godoc
// @Summary Get execution by ID
// @Description Get a single execution with detailed information
// @Tags executions
// @Accept json
// @Produce json
// @Param id path string true "Execution ID"
// @Param include_logs query bool false "Include execution logs" default(false)
// @Param log_limit query int false "Maximum number of logs to include" default(100)
// @Success 200 {object} dto.ExecutionResponse
// @Success 200 {object} dto.ExecutionWithLogsResponse "When include_logs=true"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/executions/{id} [get]
func (h *ExecutionHandler) GetExecutionByID(c *gin.Context) {
	executionIDStr := c.Param("id")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid execution ID"))
		return
	}

	includeLogs := c.Query("include_logs") == "true"
	logLimit := 100 // default

	if limitStr := c.Query("log_limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			logLimit = limit
		}
	}

	var execution *entity.Execution
	if includeLogs {
		execution, err = h.executionUsecase.GetWithLogs(c.Request.Context(), executionID, logLimit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to get execution with logs"))
			return
		}

		response := dto.ToExecutionWithLogsResponse(execution, execution.Logs)
		c.JSON(http.StatusOK, response)
	} else {
		execution, err = h.executionUsecase.GetByID(c.Request.Context(), executionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to get execution"))
			return
		}

		response := dto.ToExecutionResponse(execution)
		c.JSON(http.StatusOK, response)
	}
}

// GetExecutionLogs godoc
// @Summary Get execution logs
// @Description Get logs for a specific execution with pagination and filtering
// @Tags executions
// @Accept json
// @Produce json
// @Param id path string true "Execution ID"
// @Param level query string false "Filter by log level" Enums(debug,info,warn,error)
// @Param source query string false "Filter by log source"
// @Param search query string false "Search in log messages"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(50)
// @Param order_by query string false "Order by field" default("timestamp")
// @Param order_dir query string false "Order direction" default("desc") Enums(asc,desc)
// @Success 200 {object} dto.ExecutionLogListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/executions/{id}/logs [get]
func (h *ExecutionHandler) GetExecutionLogs(c *gin.Context) {
	executionIDStr := c.Param("id")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid execution ID"))
		return
	}

	var query dto.ExecutionLogFilterQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid query parameters"))
		return
	}

	// Build filter request
	filterReq := usecase.GetExecutionLogsRequest{
		Limit:  query.PageSize,
		Offset: (query.Page - 1) * query.PageSize,
	}

	// Apply optional filters
	if query.Level != nil {
		level := entity.LogLevel(*query.Level)
		filterReq.Levels = []entity.LogLevel{level}
	}
	if query.Source != nil {
		filterReq.Sources = []string{*query.Source}
	}
	if query.Search != nil {
		filterReq.SearchTerm = query.Search
	}
	if query.TimeAfter != nil {
		filterReq.TimeAfter = query.TimeAfter
	}
	if query.TimeBefore != nil {
		filterReq.TimeBefore = query.TimeBefore
	}
	if query.OrderBy != nil {
		filterReq.OrderBy = *query.OrderBy
	} else {
		filterReq.OrderBy = "timestamp"
	}
	if query.OrderDir != nil {
		filterReq.OrderDir = *query.OrderDir
	} else {
		filterReq.OrderDir = "desc"
	}

	logs, total, err := h.executionUsecase.GetExecutionLogs(c.Request.Context(), executionID, filterReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to get execution logs"))
		return
	}

	// Calculate pagination metadata
	totalPages := int(total) / query.PageSize
	if int(total)%query.PageSize > 0 {
		totalPages++
	}

	meta := dto.PaginationMeta{
		Page:       query.Page,
		PageSize:   query.PageSize,
		Total:      int(total),
		TotalPages: totalPages,
	}

	response := dto.ToExecutionLogListResponse(logs, meta)
	c.JSON(http.StatusOK, response)
}

// CreateExecution godoc
// @Summary Create a new execution
// @Description Create a new execution for a task
// @Tags executions
// @Accept json
// @Produce json
// @Param execution body dto.ExecutionCreateRequest true "Execution creation data"
// @Success 201 {object} dto.ExecutionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/executions [post]
func (h *ExecutionHandler) CreateExecution(c *gin.Context) {
	var req dto.ExecutionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	usecaseReq := usecase.CreateExecutionRequest{
		TaskID: req.TaskID,
	}

	execution, err := h.executionUsecase.Create(c.Request.Context(), usecaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to create execution"))
		return
	}

	response := dto.ToExecutionResponse(execution)
	c.JSON(http.StatusCreated, response)
}

// UpdateExecution godoc
// @Summary Update an execution
// @Description Update execution status, progress, or error information
// @Tags executions
// @Accept json
// @Produce json
// @Param id path string true "Execution ID"
// @Param execution body dto.ExecutionUpdateRequest true "Execution update data"
// @Success 200 {object} dto.ExecutionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/executions/{id} [put]
func (h *ExecutionHandler) UpdateExecution(c *gin.Context) {
	executionIDStr := c.Param("id")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid execution ID"))
		return
	}

	var req dto.ExecutionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
		return
	}

	usecaseReq := usecase.UpdateExecutionRequest{
		Status:   req.Status,
		Progress: req.Progress,
		Error:    req.Error,
	}

	execution, err := h.executionUsecase.Update(c.Request.Context(), executionID, usecaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to update execution"))
		return
	}

	response := dto.ToExecutionResponse(execution)
	c.JSON(http.StatusOK, response)
}

// DeleteExecution godoc
// @Summary Delete an execution
// @Description Delete an execution and all its associated logs
// @Tags executions
// @Accept json
// @Produce json
// @Param id path string true "Execution ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/executions/{id} [delete]
func (h *ExecutionHandler) DeleteExecution(c *gin.Context) {
	executionIDStr := c.Param("id")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid execution ID"))
		return
	}

	err = h.executionUsecase.Delete(c.Request.Context(), executionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to delete execution"))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetExecutionStats godoc
// @Summary Get execution statistics
// @Description Get execution statistics for a task or globally
// @Tags executions
// @Accept json
// @Produce json
// @Param task_id query string false "Filter by task ID"
// @Success 200 {object} repository.ExecutionStats
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/executions/stats [get]
func (h *ExecutionHandler) GetExecutionStats(c *gin.Context) {
	var taskID *uuid.UUID

	if taskIDStr := c.Query("task_id"); taskIDStr != "" {
		parsedTaskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid task ID"))
			return
		}
		taskID = &parsedTaskID
	}

	stats, err := h.executionUsecase.GetExecutionStats(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err, http.StatusInternalServerError, "Failed to get execution stats"))
		return
	}

	c.JSON(http.StatusOK, stats)
}