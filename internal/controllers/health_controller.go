package controllers

import (
	"net/http"
	"pnas/api/v1"
	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
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
	response := v1.PingResponse{
		Message: "pong",
	}
	c.JSON(http.StatusOK, response)
}
