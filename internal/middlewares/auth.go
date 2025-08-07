package middlewares

import (
	"net/http"
	"pnas/api/v1"
	"pnas/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// JWTAuth JWT认证中间件
func JWTAuth(authService *services.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("缺少Authorization头", 
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, v1.ErrorResponse{
				Message: "缺少认证信息",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			logger.Warn("无效的Authorization格式", 
				zap.String("auth_header", authHeader),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, v1.ErrorResponse{
				Message: "无效的认证格式",
			})
			c.Abort()
			return
		}

		// 提取Token
		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			logger.Warn("空的Token", zap.String("client_ip", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, v1.ErrorResponse{
				Message: "缺少Token",
			})
			c.Abort()
			return
		}

		// 验证Token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			logger.Warn("Token验证失败", 
				zap.Error(err),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, v1.ErrorResponse{
				Message: "无效的Token",
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", claims)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_type", claims.UserType)
		c.Set("group_id", claims.GroupID)

		logger.Debug("JWT认证成功", 
			zap.String("username", claims.Username),
			zap.Uint("user_id", claims.UserID),
			zap.String("path", c.Request.URL.Path),
		)

		c.Next()
	}
}

// OptionalJWTAuth 可选的JWT认证中间件（用于某些可选认证的接口）
func OptionalJWTAuth(authService *services.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有认证信息，继续执行
			c.Next()
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.Next()
			return
		}

		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			c.Next()
			return
		}

		// 尝试验证Token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			logger.Debug("可选JWT认证失败", zap.Error(err))
			c.Next()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", claims)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_type", claims.UserType)
		c.Set("group_id", claims.GroupID)

		logger.Debug("可选JWT认证成功", 
			zap.String("username", claims.Username),
			zap.Uint("user_id", claims.UserID),
		)

		c.Next()
	}
}