package shared

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a standardized error code
type ErrorCode string

const (
	// Authentication & Authorization errors (1xxx)
	ErrCodeUnauthorized     ErrorCode = "AUTH_1001"
	ErrCodeForbidden        ErrorCode = "AUTH_1002"
	ErrCodeInvalidToken     ErrorCode = "AUTH_1003"
	ErrCodeTokenExpired     ErrorCode = "AUTH_1004"
	ErrCodeInvalidCredentials ErrorCode = "AUTH_1005"
	ErrCodeAccountDisabled  ErrorCode = "AUTH_1006"

	// Validation errors (2xxx)
	ErrCodeInvalidInput     ErrorCode = "VAL_2001"
	ErrCodeInvalidUsername  ErrorCode = "VAL_2002"
	ErrCodeInvalidPassword  ErrorCode = "VAL_2003"
	ErrCodeInvalidEmail     ErrorCode = "VAL_2004"
	ErrCodeInvalidPath      ErrorCode = "VAL_2005"
	ErrCodeInvalidIQN       ErrorCode = "VAL_2006"

	// Resource errors (3xxx)
	ErrCodeNotFound         ErrorCode = "RES_3001"
	ErrCodeAlreadyExists    ErrorCode = "RES_3002"
	ErrCodeConflict         ErrorCode = "RES_3003"

	// Database errors (4xxx)
	ErrCodeDatabaseError    ErrorCode = "DB_4001"
	ErrCodeTransactionFailed ErrorCode = "DB_4002"

	// System errors (5xxx)
	ErrCodeSystemCommand    ErrorCode = "SYS_5001"
	ErrCodeFileSystem       ErrorCode = "SYS_5002"
	ErrCodeNetwork          ErrorCode = "SYS_5003"
	ErrCodeISCSI            ErrorCode = "SYS_5004"
	ErrCodeSamba            ErrorCode = "SYS_5005"
	ErrCodeNFS              ErrorCode = "SYS_5006"

	// Internal errors (9xxx)
	ErrCodeInternal         ErrorCode = "INT_9001"
	ErrCodeNotImplemented   ErrorCode = "INT_9002"
)

// AppError represents a standardized application error
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, details string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: getHTTPStatus(code),
	}
}

// getHTTPStatus maps error codes to HTTP status codes
func getHTTPStatus(code ErrorCode) int {
	switch {
	case code[:4] == "AUTH":
		if code == ErrCodeUnauthorized || code == ErrCodeInvalidToken ||
		   code == ErrCodeTokenExpired || code == ErrCodeInvalidCredentials {
			return http.StatusUnauthorized
		}
		return http.StatusForbidden

	case code[:3] == "VAL":
		return http.StatusBadRequest

	case code[:3] == "RES":
		if code == ErrCodeNotFound {
			return http.StatusNotFound
		}
		if code == ErrCodeAlreadyExists || code == ErrCodeConflict {
			return http.StatusConflict
		}
		return http.StatusBadRequest

	case code[:2] == "DB" || code[:3] == "SYS":
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}

// Common error constructors
func ErrUnauthorized(details string) *AppError {
	return NewAppError(ErrCodeUnauthorized, "Unauthorized", details)
}

func ErrInvalidInput(details string) *AppError {
	return NewAppError(ErrCodeInvalidInput, "Invalid input", details)
}

func ErrNotFound(resource string) *AppError {
	return NewAppError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), "")
}

func ErrAlreadyExists(resource string) *AppError {
	return NewAppError(ErrCodeAlreadyExists, fmt.Sprintf("%s already exists", resource), "")
}

func ErrDatabaseError(details string) *AppError {
	return NewAppError(ErrCodeDatabaseError, "Database error", details)
}

func ErrSystemCommand(details string) *AppError {
	return NewAppError(ErrCodeSystemCommand, "System command failed", details)
}

func ErrInternal(details string) *AppError {
	return NewAppError(ErrCodeInternal, "Internal server error", details)
}

// ToGinH converts AppError to Gin response format
func (e *AppError) ToGinH() map[string]interface{} {
	result := map[string]interface{}{
		"code":    e.Code,
		"message": e.Message,
	}
	if e.Details != "" {
		result["details"] = e.Details
	}
	return result
}
