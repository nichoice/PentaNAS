package errors

import (
	"github.com/gin-gonic/gin"
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

// HandleError 统一错误处理
func HandleError(c *gin.Context, statusCode int, message string, err error) {
	response := ErrorResponse{
		Message: message,
		Code:    statusCode,
	}

	if err != nil {
		response.Error = err.Error()
	}

	c.JSON(statusCode, response)
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse 验证错误响应
type ValidationErrorResponse struct {
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors"`
}

// HandleValidationError 处理验证错误
func HandleValidationError(c *gin.Context, message string, validationErrors []ValidationError) {
	response := ValidationErrorResponse{
		Message: message,
		Errors:  validationErrors,
	}

	c.JSON(400, response)
}