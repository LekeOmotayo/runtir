package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
	Data       any    `json:"data"`
	Pagination struct {
		Page       int64 `json:"page"`
		PageSize   int64 `json:"page_size"`
		TotalCount int64 `json:"total_count"`
	} `json:"pagination"`
}

// SuccessResponse sends a standard JSON success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends a standard JSON error response
func ErrorResponse(c *gin.Context, statusCode int, message string, err any) {
	errStr := ""
	if e, ok := err.(error); ok && e != nil {
		errStr = e.Error()
	} else if s, ok := err.(string); ok {
		errStr = s
	}

	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Error:   errStr,
	})
}

// PaginatedSuccessResponse sends a standard paginated JSON success response
func PaginatedSuccessResponse(c *gin.Context, message string, data any, page, pageSize, total int64) {
	resp := PaginatedResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	resp.Pagination.Page = page
	resp.Pagination.PageSize = pageSize
	resp.Pagination.TotalCount = total

	c.JSON(http.StatusOK, resp)
}
