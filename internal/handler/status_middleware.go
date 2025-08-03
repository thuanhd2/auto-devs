package handler

import (
	"net/http"
	"strings"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

// TaskStatusValidationMiddleware validates task status values in requests
func TaskStatusValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to specific endpoints that handle status updates
		path := c.Request.URL.Path
		method := c.Request.Method
		
		// Check if this is a status update endpoint
		isStatusUpdate := (method == "PATCH" && strings.Contains(path, "/status")) ||
						 (method == "PATCH" && strings.Contains(path, "/bulk-status")) ||
						 (method == "POST" && strings.Contains(path, "/tasks"))

		if !isStatusUpdate {
			c.Next()
			return
		}

		// For POST /tasks, validate the default status if provided
		if method == "POST" && strings.Contains(path, "/tasks") {
			var taskReq dto.TaskCreateRequest
			if err := c.ShouldBindJSON(&taskReq); err != nil {
				c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
				c.Abort()
				return
			}
			// Rebind the request for the actual handler
			c.Set("validated_task_request", taskReq)
			c.Next()
			return
		}

		// For status update endpoints, validate the status value
		if strings.Contains(path, "/status") {
			// Handle different request types
			if strings.Contains(path, "/bulk-status") {
				var bulkReq dto.BulkStatusUpdateRequest
				if err := c.ShouldBindJSON(&bulkReq); err != nil {
					c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
					c.Abort()
					return
				}

				// Validate status
				if !bulkReq.Status.IsValid() {
					c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
						nil, 
						http.StatusBadRequest, 
						"Invalid status value: "+string(bulkReq.Status)))
					c.Abort()
					return
				}

				c.Set("validated_bulk_request", bulkReq)
			} else if strings.Contains(path, "/status-with-history") {
				var statusReq dto.TaskStatusUpdateWithHistoryRequest
				if err := c.ShouldBindJSON(&statusReq); err != nil {
					c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
					c.Abort()
					return
				}

				// Validate status
				if !statusReq.Status.IsValid() {
					c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
						nil, 
						http.StatusBadRequest, 
						"Invalid status value: "+string(statusReq.Status)))
					c.Abort()
					return
				}

				c.Set("validated_status_history_request", statusReq)
			} else {
				var statusReq dto.TaskStatusUpdateRequest
				if err := c.ShouldBindJSON(&statusReq); err != nil {
					c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err, http.StatusBadRequest, "Invalid request data"))
					c.Abort()
					return
				}

				// Validate status
				if !statusReq.Status.IsValid() {
					c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
						nil, 
						http.StatusBadRequest, 
						"Invalid status value: "+string(statusReq.Status)))
					c.Abort()
					return
				}

				c.Set("validated_status_request", statusReq)
			}
		}

		c.Next()
	}
}

// QueryParameterStatusValidationMiddleware validates status values in query parameters
func QueryParameterStatusValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate status query parameter if present
		status := c.Query("status")
		if status != "" {
			taskStatus := entity.TaskStatus(status)
			if !taskStatus.IsValid() {
				c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
					nil, 
					http.StatusBadRequest, 
					"Invalid status query parameter: "+status))
				c.Abort()
				return
			}
		}

		// Validate statuses array parameter if present
		statuses := c.QueryArray("statuses")
		for _, statusStr := range statuses {
			taskStatus := entity.TaskStatus(statusStr)
			if !taskStatus.IsValid() {
				c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
					nil, 
					http.StatusBadRequest, 
					"Invalid status in statuses parameter: "+statusStr))
				c.Abort()
				return
			}
		}

		// Validate target_status parameter for transition validation
		targetStatus := c.Query("target_status")
		if targetStatus != "" {
			taskStatus := entity.TaskStatus(targetStatus)
			if !taskStatus.IsValid() {
				c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
					nil, 
					http.StatusBadRequest, 
					"Invalid target_status query parameter: "+targetStatus))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// GetValidatedRequest retrieves a validated request from the context
func GetValidatedTaskRequest(c *gin.Context) (*dto.TaskCreateRequest, bool) {
	if req, exists := c.Get("validated_task_request"); exists {
		if taskReq, ok := req.(dto.TaskCreateRequest); ok {
			return &taskReq, true
		}
	}
	return nil, false
}

func GetValidatedStatusRequest(c *gin.Context) (*dto.TaskStatusUpdateRequest, bool) {
	if req, exists := c.Get("validated_status_request"); exists {
		if statusReq, ok := req.(dto.TaskStatusUpdateRequest); ok {
			return &statusReq, true
		}
	}
	return nil, false
}

func GetValidatedStatusHistoryRequest(c *gin.Context) (*dto.TaskStatusUpdateWithHistoryRequest, bool) {
	if req, exists := c.Get("validated_status_history_request"); exists {
		if statusReq, ok := req.(dto.TaskStatusUpdateWithHistoryRequest); ok {
			return &statusReq, true
		}
	}
	return nil, false
}

func GetValidatedBulkRequest(c *gin.Context) (*dto.BulkStatusUpdateRequest, bool) {
	if req, exists := c.Get("validated_bulk_request"); exists {
		if bulkReq, ok := req.(dto.BulkStatusUpdateRequest); ok {
			return &bulkReq, true
		}
	}
	return nil, false
}