package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingMiddleware is a middleware for logging HTTP requests
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		
		// Process request
		c.Next()
		
		// End timer
		end := time.Now()
		latency := end.Sub(start)
		
		// Get status code and size
		status := c.Writer.Status()
		size := c.Writer.Size()
		
		// Log the request
		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("status", status),
			zap.String("latency", latency.String()),
			zap.Int("size", size),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}