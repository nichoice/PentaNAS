package main

import (
	"log"
	"os"
	"pnas/internal/config"
	"pnas/internal/routes"
	_ "pnas/cmd/docs" // 导入生成的 docs
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title PNAS API
// @version 1.0
// @description PNAS 项目 API 文档
// @host localhost:8080
// @BasePath /
func main() {
	// 根据环境变量选择日志配置文件
	env := os.Getenv("GO_ENV")
	var loggerConfigPath string
	
	switch env {
	case "production":
		loggerConfigPath = "config/logger-prod.yml"
		gin.SetMode(gin.ReleaseMode)
	case "development":
		loggerConfigPath = "config/logger-dev.yml"
		gin.SetMode(gin.DebugMode)
	default:
		loggerConfigPath = "config/logger-dev.yml" // 默认使用开发环境配置
		gin.SetMode(gin.DebugMode)
	}

	// 初始化日志器
	logger, err := config.InitLogger(loggerConfigPath)
	if err != nil {
		log.Fatalf("初始化日志器失败: %v", err)
	}
	defer logger.Sync()

	// 演示不同级别的彩色日志
	logger.Debug("🔧 调试信息: 应用正在启动...")
	logger.Info("🚀 应用启动中...", 
		zap.String("environment", env),
		zap.String("config", loggerConfigPath),
	)
	logger.Warn("⚠️  这是一个警告日志示例")
	
	// 创建 Gin 引擎（不使用默认中间件）
	r := gin.New()
	
	// 设置路由（包含日志中间件）
	routes.SetupRoutes(r, logger)

	logger.Info("✅ 服务启动成功", 
		zap.String("address", ":8080"),
		zap.String("swagger", "http://localhost:8080/swagger/index.html"),
	)
	
	// 启动服务
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("❌ 服务启动失败", zap.Error(err))
	}
}
