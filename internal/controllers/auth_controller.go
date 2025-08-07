package controllers

import (
	"net/http"
	"pnas/api/v1"
	"pnas/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthController 认证控制器
type AuthController struct {
	logger      *zap.Logger
	authService *services.AuthService
}

// NewAuthController 创建认证控制器实例
func NewAuthController(logger *zap.Logger, authService *services.AuthService) *AuthController {
	return &AuthController{
		logger:      logger,
		authService: authService,
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取JWT Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body v1.LoginRequest true "登录请求"
// @Success 200 {object} v1.LoginResponse "登录成功"
// @Failure 400 {object} v1.ErrorResponse "请求参数错误"
// @Failure 401 {object} v1.ErrorResponse "用户名或密码错误"
// @Router /api/v1/auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var req v1.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.Warn("登录请求参数错误", 
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 验证用户登录
	token, user, err := ac.authService.Login(req.Username, req.Password)
	if err != nil {
		ac.logger.Warn("用户登录失败", 
			zap.String("username", req.Username),
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusUnauthorized, v1.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "用户名或密码错误",
			Error:   err.Error(),
		})
		return
	}

	ac.logger.Info("用户登录成功", 
		zap.String("username", user.Username),
		zap.Uint("user_id", user.ID),
		zap.Int("user_type", int(user.UserType)),
		zap.String("client_ip", c.ClientIP()),
	)

	// 构造响应
	response := v1.LoginResponse{
		Status:  "success",
		Message: "登录成功",
		Data: v1.LoginData{
			Token: token,
			User: v1.UserInfo{
				ID:       user.ID,
				Username: user.Username,
				UserType: user.UserType,
				Status:   user.Status,
				Group: v1.UserGroupInfo{
					ID:          user.Group.ID,
					Name:        user.Group.Name,
					Description: user.Group.Description,
				},
			},
			ExpiresAt: "2025-08-08T23:59:59Z", // 临时固定值，实际应该从JWT配置获取
		},
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken 刷新Token
// @Summary 刷新Token
// @Description 使用现有Token获取新的Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body v1.RefreshTokenRequest true "刷新Token请求"
// @Success 200 {object} v1.RefreshTokenResponse "Token刷新成功"
// @Failure 400 {object} v1.ErrorResponse "请求参数错误"
// @Failure 401 {object} v1.ErrorResponse "无效的Token"
// @Router /api/v1/auth/refresh [post]
func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req v1.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.Warn("刷新Token请求参数错误", 
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 刷新Token
	newToken, err := ac.authService.RefreshToken(req.Token)
	if err != nil {
		ac.logger.Warn("Token刷新失败", 
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusUnauthorized, v1.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "无效的Token",
			Error:   err.Error(),
		})
		return
	}

	ac.logger.Info("Token刷新成功", 
		zap.String("client_ip", c.ClientIP()),
	)

	// 构造响应
	response := v1.RefreshTokenResponse{
		Status:  "success",
		Message: "Token刷新成功",
		Data: v1.RefreshTokenData{
			Token:     newToken,
			ExpiresAt: "2025-08-08T23:59:59Z", // 临时固定值，实际应该从JWT配置获取
		},
	}

	c.JSON(http.StatusOK, response)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出（使Token失效）
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} v1.LogoutResponse "登出成功"
// @Failure 500 {object} v1.ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	// 从上下文中获取用户信息
	username, exists := c.Get("username")
	if !exists {
		username = "unknown"
	}

	// 这里可以实现Token黑名单机制
	// 目前简单返回成功响应
	ac.logger.Info("用户登出成功", 
		zap.String("username", username.(string)),
		zap.String("client_ip", c.ClientIP()),
	)

	response := v1.LogoutResponse{
		Status:  "success",
		Message: "登出成功",
	}

	c.JSON(http.StatusOK, response)
}