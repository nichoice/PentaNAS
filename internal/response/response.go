package response

import (
	"net/http"
	"pnas/internal/i18n"
	"pnas/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// BaseResponse 基础响应结构
type BaseResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

// PaginationMeta 分页元数据
type PaginationMeta struct {
	Total       int64 `json:"total"`
	PerPage     int   `json:"per_page"`
	CurrentPage int   `json:"current_page"`
	LastPage    int   `json:"last_page"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	t := middlewares.GetT(c)
	c.JSON(http.StatusOK, BaseResponse{
		Code:    http.StatusOK,
		Message: t("success"),
		Data:    data,
	})
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, messageKey string, data interface{}, args ...interface{}) {
	t := middlewares.GetT(c)
	c.JSON(http.StatusOK, BaseResponse{
		Code:    http.StatusOK,
		Message: t(messageKey, args...),
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, messageKey string, err error, args ...interface{}) {
	t := middlewares.GetT(c)
	response := BaseResponse{
		Code:    code,
		Message: t(messageKey, args...),
	}
	
	if err != nil {
		response.Error = err.Error()
	}
	
	c.JSON(code, response)
}

// BadRequest 400错误响应
func BadRequest(c *gin.Context, messageKey string, err error, args ...interface{}) {
	Error(c, http.StatusBadRequest, messageKey, err, args...)
}

// Unauthorized 401错误响应
func Unauthorized(c *gin.Context, messageKey string, err error, args ...interface{}) {
	Error(c, http.StatusUnauthorized, messageKey, err, args...)
}

// Forbidden 403错误响应
func Forbidden(c *gin.Context, messageKey string, err error, args ...interface{}) {
	Error(c, http.StatusForbidden, messageKey, err, args...)
}

// NotFound 404错误响应
func NotFound(c *gin.Context, messageKey string, err error, args ...interface{}) {
	Error(c, http.StatusNotFound, messageKey, err, args...)
}

// InternalServerError 500错误响应
func InternalServerError(c *gin.Context, messageKey string, err error, args ...interface{}) {
	Error(c, http.StatusInternalServerError, messageKey, err, args...)
}

// Pagination 分页响应
func Pagination(c *gin.Context, messageKey string, data interface{}, total int64, page, pageSize int, args ...interface{}) {
	t := middlewares.GetT(c)
	
	lastPage := int(total) / pageSize
	if int(total)%pageSize > 0 {
		lastPage++
	}
	
	c.JSON(http.StatusOK, PaginationResponse{
		Code:    http.StatusOK,
		Message: t(messageKey, args...),
		Data:    data,
		Meta: PaginationMeta{
			Total:       total,
			PerPage:     pageSize,
			CurrentPage: page,
			LastPage:    lastPage,
		},
	})
}

// ValidationError 验证错误响应
func ValidationError(c *gin.Context, err error) {
	t := middlewares.GetT(c)
	c.JSON(http.StatusBadRequest, BaseResponse{
		Code:    http.StatusBadRequest,
		Message: t("request.invalid_params"),
		Error:   err.Error(),
	})
}

// GetLanguageInfo 获取当前语言信息
func GetLanguageInfo(c *gin.Context) map[string]interface{} {
	locale := middlewares.GetLocale(c)
	return map[string]interface{}{
		"locale":    string(locale),
		"language":  getLanguageName(locale),
		"supported": []string{"zh-CN", "en-US"},
	}
}

// getLanguageName 获取语言名称
func getLanguageName(locale i18n.Locale) string {
	switch locale {
	case i18n.LocaleZhCN:
		return "简体中文"
	case i18n.LocaleEnUS:
		return "English"
	default:
		return "简体中文"
	}
}