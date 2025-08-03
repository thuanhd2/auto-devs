package dto

// Common response DTOs
type ErrorResponse struct {
	Error   string            `json:"error" example:"Invalid request"`
	Message string            `json:"message" example:"The provided data is invalid"`
	Code    int               `json:"code" example:"400"`
	Details map[string]string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// Pagination DTOs
type PaginationQuery struct {
	Page     int `form:"page,default=1" binding:"min=1" example:"1"`
	PageSize int `form:"page_size,default=10" binding:"min=1,max=100" example:"10"`
}

type PaginationMeta struct {
	Page       int `json:"page" example:"1"`
	PageSize   int `json:"page_size" example:"10"`
	Total      int `json:"total" example:"100"`
	TotalPages int `json:"total_pages" example:"10"`
}

type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// Filter DTOs for tasks
type TaskFilterQuery struct {
	PaginationQuery
	Status    *string    `form:"status" binding:"omitempty,oneof=TODO PLANNING PLAN_REVIEWING IMPLEMENTING CODE_REVIEWING DONE CANCELLED" example:"TODO"`
	ProjectID *string    `form:"project_id" binding:"omitempty,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	Search    *string    `form:"search" binding:"omitempty,max=255" example:"authentication"`
}

// Helper functions
func NewErrorResponse(err error, code int, message string) ErrorResponse {
	return ErrorResponse{
		Error:   err.Error(),
		Message: message,
		Code:    code,
	}
}

func NewValidationErrorResponse(details map[string]string) ErrorResponse {
	return ErrorResponse{
		Error:   "Validation failed",
		Message: "The provided data failed validation",
		Code:    400,
		Details: details,
	}
}

func NewSuccessResponse(message string, data interface{}) SuccessResponse {
	return SuccessResponse{
		Message: message,
		Data:    data,
	}
}