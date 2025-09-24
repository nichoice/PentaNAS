package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"pnas/internal/models"
	"pnas/internal/services"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditConfig 审计配置
type AuditConfig struct {
	EnableFileOperations bool     `json:"enable_file_operations"`
	EnableUserOperations bool     `json:"enable_user_operations"`
	ExcludePaths         []string `json:"exclude_paths"`
	MaxBodySize          int64    `json:"max_body_size"`
}

// FileOperationInfo 文件操作信息
type FileOperationInfo struct {
	Operation models.AuditOperation
	FilePath  string
	FileSize  int64
	Metadata  map[string]interface{}
}

var (
	defaultAuditConfig = AuditConfig{
		EnableFileOperations: true,
		EnableUserOperations: true,
		ExcludePaths: []string{
			"/health",
			"/metrics",
			"/swagger",
			"/docs",
			"/api/v1/auth/refresh",
		},
		MaxBodySize: 1024 * 1024, // 1MB
	}

	// 文件操作路径模式
	fileOperationPatterns = map[*regexp.Regexp]models.AuditOperation{
		regexp.MustCompile(`/api/v1/files/upload`):     models.AuditOperationUpload,
		regexp.MustCompile(`/api/v1/files/download`):   models.AuditOperationDownload,
		regexp.MustCompile(`/api/v1/files/delete`):     models.AuditOperationDelete,
		regexp.MustCompile(`/api/v1/files/rename`):     models.AuditOperationRename,
		regexp.MustCompile(`/api/v1/files/move`):       models.AuditOperationMove,
		regexp.MustCompile(`/api/v1/files/copy`):       models.AuditOperationCopy,
		regexp.MustCompile(`/api/v1/files/compress`):   models.AuditOperationCompress,
		regexp.MustCompile(`/api/v1/files/extract`):    models.AuditOperationExtract,
		regexp.MustCompile(`/api/v1/files`):            models.AuditOperationRead, // GET请求默认为读操作
		regexp.MustCompile(`/api/v1/storage/create.*`): models.AuditOperationCreate,
	}
)

// AuditMiddleware 审计中间件
func AuditMiddleware(config ...AuditConfig) gin.HandlerFunc {
	cfg := defaultAuditConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// 检查是否需要审计此路径
		if !shouldAudit(c.Request.URL.Path, cfg.ExcludePaths) {
			c.Next()
			return
		}

		// 记录开始时间
		startTime := time.Now()

		// 获取用户信息
		userID, username := getUserInfo(c)

		// 解析文件操作信息
		operation, filePath, fileSize, metadata := parseFileOperation(c, cfg.MaxBodySize)

		// 如果不是文件操作且未启用用户操作审计，跳过
		if operation == "" && !cfg.EnableUserOperations {
			c.Next()
			return
		}

		// 如果是文件操作但未启用文件操作审计，跳过
		if operation != "" && !cfg.EnableFileOperations {
			c.Next()
			return
		}

		// 创建响应写入器包装器来捕获响应
		responseWriter := &auditResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(startTime)

		// 确定操作状态
		status := models.AuditStatusSuccess
		errorMsg := ""
		if c.Writer.Status() >= 400 {
			status = models.AuditStatusFailed
			errorMsg = extractErrorMessage(responseWriter.body.String())
		}

		// 如果没有明确的操作类型，根据HTTP方法和状态推断
		if operation == "" {
			operation = inferOperationFromMethod(c.Request.Method, c.Writer.Status())
		}

		// 记录审计日志
		if operation != "" {
			auditEvent := &services.AuditEvent{
				UserID:    userID,
				Username:  username,
				Operation: operation,
				FilePath:  filePath,
				FileSize:  fileSize,
				ClientIP:  c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Status:    status,
				ErrorMsg:  errorMsg,
				Duration:  duration,
				Source:    models.AuditSourceAPI,
				Metadata:  metadata,
			}

			services.GetAuditService().LogEvent(auditEvent)
		}
	}
}

// auditResponseWriter 响应写入器包装器
type auditResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *auditResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// shouldAudit 检查是否应该审计此路径
func shouldAudit(path string, excludePaths []string) bool {
	for _, excludePath := range excludePaths {
		if strings.HasPrefix(path, excludePath) {
			return false
		}
	}
	return true
}

// getUserInfo 获取用户信息
func getUserInfo(c *gin.Context) (userID, username string) {
	// 尝试从JWT token获取用户信息
	if userIDInterface, exists := c.Get("userID"); exists {
		userID = userIDInterface.(string)
	}
	if usernameInterface, exists := c.Get("username"); exists {
		username = usernameInterface.(string)
	}

	// 如果没有用户信息，标记为匿名用户
	if userID == "" {
		userID = "anonymous"
	}
	if username == "" {
		username = "anonymous"
	}

	return userID, username
}

// parseFileOperation 解析文件操作信息
func parseFileOperation(c *gin.Context, maxBodySize int64) (models.AuditOperation, string, int64, map[string]interface{}) {
	var operation models.AuditOperation
	var filePath string
	var fileSize int64
	metadata := make(map[string]interface{})

	// 根据路径模式匹配操作类型
	for pattern, op := range fileOperationPatterns {
		if pattern.MatchString(c.Request.URL.Path) {
			operation = op
			break
		}
	}

	// 从URL参数获取文件路径
	if path := c.Query("path"); path != "" {
		filePath = path
	} else if path := c.Param("path"); path != "" {
		filePath = path
	}

	// 从请求体获取文件信息
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		contentType := c.Request.Header.Get("Content-Type")

		if strings.Contains(contentType, "multipart/form-data") {
			// 处理文件上传
			if form, err := c.MultipartForm(); err == nil {
				if files := form.File["file"]; len(files) > 0 {
					file := files[0]
					filePath = file.Filename
					fileSize = file.Size
					metadata["file_count"] = len(files)
					metadata["content_type"] = file.Header.Get("Content-Type")
				}
			}
		} else if strings.Contains(contentType, "application/json") && c.Request.ContentLength < maxBodySize {
			// 处理JSON请求体
			body, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// 恢复请求体供后续处理
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

				var jsonData map[string]interface{}
				if json.Unmarshal(body, &jsonData) == nil {
					if path, ok := jsonData["path"].(string); ok {
						filePath = path
					}
					if size, ok := jsonData["size"].(float64); ok {
						fileSize = int64(size)
					}
					metadata["request_body"] = jsonData
				}
			}
		}
	}

	// 从响应头获取文件大小（下载操作）
	if operation == models.AuditOperationDownload {
		if contentLength := c.Writer.Header().Get("Content-Length"); contentLength != "" {
			if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
				fileSize = size
			}
		}
	}

	// 添加请求信息到元数据
	metadata["method"] = c.Request.Method
	metadata["url"] = c.Request.URL.String()
	metadata["content_length"] = c.Request.ContentLength

	return operation, filePath, fileSize, metadata
}

// inferOperationFromMethod 根据HTTP方法推断操作类型
func inferOperationFromMethod(method string, statusCode int) models.AuditOperation {
	if statusCode >= 400 {
		// 失败的请求，根据方法推断意图
		switch method {
		case "GET":
			return models.AuditOperationRead
		case "POST":
			return models.AuditOperationCreate
		case "PUT", "PATCH":
			return models.AuditOperationUpdate
		case "DELETE":
			return models.AuditOperationDelete
		}
	}

	// 成功的请求，根据方法确定操作
	switch method {
	case "GET":
		return models.AuditOperationRead
	case "POST":
		return models.AuditOperationCreate
	case "PUT", "PATCH":
		return models.AuditOperationUpdate
	case "DELETE":
		return models.AuditOperationDelete
	default:
		return ""
	}
}

// extractErrorMessage 从响应中提取错误信息
func extractErrorMessage(responseBody string) string {
	var response map[string]interface{}
	if json.Unmarshal([]byte(responseBody), &response) == nil {
		if error, ok := response["error"].(string); ok {
			return error
		}
		if message, ok := response["message"].(string); ok {
			return message
		}
	}
	return "Unknown error"
}

// FileAuditMiddleware 专门的文件操作审计中间件
func FileAuditMiddleware() gin.HandlerFunc {
	return AuditMiddleware(AuditConfig{
		EnableFileOperations: true,
		EnableUserOperations: false,
		ExcludePaths:         defaultAuditConfig.ExcludePaths,
		MaxBodySize:          defaultAuditConfig.MaxBodySize,
	})
}

// UserAuditMiddleware 专门的用户操作审计中间件
func UserAuditMiddleware() gin.HandlerFunc {
	return AuditMiddleware(AuditConfig{
		EnableFileOperations: false,
		EnableUserOperations: true,
		ExcludePaths:         defaultAuditConfig.ExcludePaths,
		MaxBodySize:          defaultAuditConfig.MaxBodySize,
	})
}

// SetAuditConfig 更新审计配置
func SetAuditConfig(config AuditConfig) {
	defaultAuditConfig = config
}