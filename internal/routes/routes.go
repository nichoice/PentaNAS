package routes

import (
	"pnas/internal/controllers"
	"pnas/internal/middlewares"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func SetupRoutes(r *gin.Engine, logger *zap.Logger) {
	// 添加日志中间件
	r.Use(middlewares.ZapLogger(logger))
	r.Use(middlewares.Recovery(logger))
	
	// 初始化控制器
	healthController := controllers.NewHealthController(logger)
	
	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	// 健康检查路由
	r.GET("/ping", healthController.Ping)
}
