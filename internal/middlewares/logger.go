package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 使用 zap 记录请求日志
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.Int("body_size", param.BodySize),
		)
		return ""
	})
}

// ZapLogger 返回 zap 日志中间件
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 记录日志
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()
		
		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("body_size", bodySize),
		)
	}
}

// Recovery 恢复中间件
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, recovered interface{}) {
		logger.Error("Panic recovered",
			zap.Any("error", recovered),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
		)
		c.AbortWithStatus(500)
	})
}