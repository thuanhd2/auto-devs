package dto

import (
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// Execution request DTOs
type ExecutionCreateRequest struct {
	TaskID uuid.UUID `json:"task_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type ExecutionUpdateRequest struct {
	Status   *entity.ExecutionStatus `json:"status,omitempty" binding:"omitempty,oneof=pending running paused completed failed cancelled" example:"running"`
	Progress *float64                `json:"progress,omitempty" binding:"omitempty,min=0,max=1" example:"0.5"`
	Error    *string                 `json:"error,omitempty" example:"Process failed"`
}

// Execution response DTOs
type ExecutionResponse struct {
	ID          uuid.UUID               `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	TaskID      uuid.UUID               `json:"task_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status      entity.ExecutionStatus  `json:"status" example:"running"`
	StartedAt   time.Time               `json:"started_at" example:"2024-01-01T00:00:00Z"`
	CompletedAt *time.Time              `json:"completed_at,omitempty" example:"2024-01-01T01:00:00Z"`
	Error       string                  `json:"error,omitempty" example:"Process failed"`
	Progress    float64                 `json:"progress" example:"0.75"`
	Result      *entity.ExecutionResult `json:"result,omitempty"`
	Duration    *time.Duration          `json:"duration,omitempty" example:"3600000000000"`
	CreatedAt   time.Time               `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time               `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

type ExecutionWithLogsResponse struct {
	ExecutionResponse
	Logs []ExecutionLogResponse `json:"logs"`
}

type ExecutionListResponse struct {
	Data []ExecutionResponse `json:"data"`
	Meta PaginationMeta      `json:"meta"`
}

// Execution log response DTOs
type ExecutionLogResponse struct {
	ID          uuid.UUID       `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ExecutionID uuid.UUID       `json:"execution_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ProcessID   *uuid.UUID      `json:"process_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Level       entity.LogLevel `json:"level" example:"info"`
	Message     string          `json:"message" example:"Process started successfully"`
	Timestamp   time.Time       `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	Source      string          `json:"source" example:"stdout"`
	Metadata    interface{}     `json:"metadata,omitempty"`
    // Structured fields
    LogType       string      `json:"log_type,omitempty" example:"assistant"`
    ToolName      string      `json:"tool_name,omitempty" example:"read_file"`
    ToolUseID     string      `json:"tool_use_id,omitempty" example:"toolu_01ABC..."`
    ParsedContent interface{} `json:"parsed_content,omitempty"`
    IsError       *bool       `json:"is_error,omitempty"`
    DurationMs    *int        `json:"duration_ms,omitempty" example:"1234"`
    NumTurns      *int        `json:"num_turns,omitempty" example:"5"`
	CreatedAt   time.Time       `json:"created_at" example:"2024-01-01T00:00:00Z"`
	Line        int             `json:"line" example:"1"`
}

type ExecutionLogListResponse struct {
	Data []ExecutionLogResponse `json:"data"`
	Meta PaginationMeta         `json:"meta"`
}

// Filter and query DTOs
type ExecutionFilterQuery struct {
	PaginationQuery
	Status        *string    `form:"status" binding:"omitempty,oneof=pending running paused completed failed cancelled" example:"running"`
	Statuses      []string   `form:"statuses" example:"running,completed"`
	StartedAfter  *time.Time `form:"started_after" example:"2024-01-01T00:00:00Z"`
	StartedBefore *time.Time `form:"started_before" example:"2024-12-31T23:59:59Z"`
	WithErrors    *bool      `form:"with_errors" example:"true"`
	OrderBy       *string    `form:"order_by" binding:"omitempty,oneof=started_at completed_at progress status" example:"started_at"`
	OrderDir      *string    `form:"order_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
}

type ExecutionLogFilterQuery struct {
	PaginationQuery
	Level      *string    `form:"level" binding:"omitempty,oneof=debug info warn error" example:"info"`
	Levels     []string   `form:"levels" example:"info,error"`
	Source     *string    `form:"source" example:"stdout"`
	Sources    []string   `form:"sources" example:"stdout,stderr"`
    LogType    *string    `form:"log_type" example:"assistant"`
    ToolName   *string    `form:"tool_name" example:"read_file"`
    ToolUseID  *string    `form:"tool_use_id" example:"toolu_01ABC..."`
	Search     *string    `form:"search" example:"error"`
	TimeAfter  *time.Time `form:"time_after" example:"2024-01-01T00:00:00Z"`
	TimeBefore *time.Time `form:"time_before" example:"2024-12-31T23:59:59Z"`
	OrderBy    *string    `form:"order_by" binding:"omitempty,oneof=timestamp level source" example:"timestamp"`
	OrderDir   *string    `form:"order_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
}

// Conversion functions
func ToExecutionResponse(execution *entity.Execution) ExecutionResponse {
	response := ExecutionResponse{
		ID:        execution.ID,
		TaskID:    execution.TaskID,
		Status:    execution.Status,
		StartedAt: execution.StartedAt,
		Error:     execution.ErrorMessage,
		Progress:  execution.Progress,
		CreatedAt: execution.CreatedAt,
		UpdatedAt: execution.UpdatedAt,
	}

	if execution.CompletedAt != nil {
		response.CompletedAt = execution.CompletedAt
	}

	if execution.Result != nil {
		// Parse result if needed
		response.Result = &entity.ExecutionResult{}
	}

	// Calculate duration
	if execution.CompletedAt != nil {
		duration := execution.CompletedAt.Sub(execution.StartedAt)
		response.Duration = &duration
	}

	return response
}

func ToExecutionWithLogsResponse(execution *entity.Execution, logs []entity.ExecutionLog) ExecutionWithLogsResponse {
	response := ExecutionWithLogsResponse{
		ExecutionResponse: ToExecutionResponse(execution),
		Logs:              make([]ExecutionLogResponse, len(logs)),
	}

	for i, log := range logs {
		response.Logs[i] = ToExecutionLogResponse(&log)
	}

	return response
}

func ToExecutionLogResponse(log *entity.ExecutionLog) ExecutionLogResponse {
	response := ExecutionLogResponse{
		ID:          log.ID,
		ExecutionID: log.ExecutionID,
		Level:       log.Level,
		Message:     log.Message,
		Timestamp:   log.Timestamp,
		Source:      log.Source,
        LogType:     log.LogType,
        ToolName:    log.ToolName,
        ToolUseID:   log.ToolUseID,
        IsError:     log.IsError,
        DurationMs:  log.DurationMs,
        NumTurns:    log.NumTurns,
		CreatedAt:   log.CreatedAt,
		Line:        log.Line,
	}

	if log.Metadata != nil {
		// Parse metadata if needed
		response.Metadata = log.Metadata
	}

    if log.ParsedContent != nil {
        response.ParsedContent = log.ParsedContent
    }

	return response
}

func ToExecutionListResponse(executions []*entity.Execution, meta PaginationMeta) ExecutionListResponse {
	responses := make([]ExecutionResponse, len(executions))
	for i, execution := range executions {
		responses[i] = ToExecutionResponse(execution)
	}

	return ExecutionListResponse{
		Data: responses,
		Meta: meta,
	}
}

func ToExecutionLogListResponse(logs []*entity.ExecutionLog, meta PaginationMeta) ExecutionLogListResponse {
	responses := make([]ExecutionLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = ToExecutionLogResponse(log)
	}

	return ExecutionLogListResponse{
		Data: responses,
		Meta: meta,
	}
}
