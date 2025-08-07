package controllers

import (
	"net/http"
	"pnas/api/v1"
	"pnas/internal/models"
	"pnas/internal/repositories"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HealthController struct {
	logger              *zap.Logger
	healthCheckRepo     repositories.HealthCheckRepository
}

func NewHealthController(logger *zap.Logger, healthCheckRepo repositories.HealthCheckRepository) *HealthController {
	return &HealthController{
		logger:          logger,
		healthCheckRepo: healthCheckRepo,
	}
}

// Ping 健康检查接口
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} v1.PingResponse
// @Router /ping [get]
func (h *HealthController) Ping(c *gin.Context) {
	// 演示不同级别的彩色日志
	h.logger.Debug("调试信息: 开始处理健康检查请求", 
		zap.String("endpoint", "/ping"),
		zap.String("method", c.Request.Method),
	)
	
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	
	h.logger.Info("健康检查请求", 
		zap.String("client_ip", clientIP),
		zap.String("user_agent", userAgent),
	)
	
	// 模拟一个警告日志
	if userAgent == "" {
		h.logger.Warn("客户端未提供 User-Agent 头", 
			zap.String("client_ip", clientIP),
		)
	}
	
	// 记录健康检查到数据库
	healthCheck := &models.HealthCheck{
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Status:    "success",
	}
	
	if err := h.healthCheckRepo.Create(healthCheck); err != nil {
		h.logger.Error("保存健康检查记录失败", zap.Error(err))
		// 即使数据库记录失败，也不影响健康检查响应
	}
	
	response := v1.PingResponse{
		Message: "pong",
	}
	
	h.logger.Debug("健康检查响应", zap.Any("response", response))
	h.logger.Info("健康检查完成", zap.String("status", "success"))
	
	c.JSON(http.StatusOK, response)
}
