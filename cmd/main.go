package main

import (
	"log"
	"os"
	"pnas/internal/config"
	"pnas/internal/database"
	"pnas/internal/repositories"
	"pnas/internal/routes"
	"pnas/internal/services"
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
	// 根据环境变量选择配置
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development" // 默认开发环境
	}
	
	var loggerConfigPath string
	switch env {
	case "production":
		loggerConfigPath = "config/logger-prod.yml"
		gin.SetMode(gin.ReleaseMode)
	case "test":
		loggerConfigPath = "config/logger-dev.yml"
		gin.SetMode(gin.TestMode)
	default: // development
		loggerConfigPath = "config/logger-dev.yml"
		gin.SetMode(gin.DebugMode)
	}

	// 初始化日志器
	logger, err := config.InitLogger(loggerConfigPath)
	if err != nil {
		log.Fatalf("初始化日志器失败: %v", err)
	}
	defer logger.Sync()

	logger.Debug("🔧 调试信息: 应用正在启动...")
	logger.Info("🚀 应用启动中...", 
		zap.String("environment", env),
		zap.String("config", loggerConfigPath),
	)

	// 初始化数据库
	db, err := config.InitDatabase("config/database.yml", env, logger)
	if err != nil {
		logger.Fatal("❌ 数据库初始化失败", zap.Error(err))
	}

	// 执行数据库迁移
	if err := database.AutoMigrate(db, logger); err != nil {
		logger.Fatal("❌ 数据库迁移失败", zap.Error(err))
	}

	// 初始化种子数据（仅在开发环境）
	if env == "development" {
		if err := database.SeedData(db, logger); err != nil {
			logger.Warn("⚠️ 种子数据初始化失败", zap.Error(err))
		}
	}

	// 加载JWT配置
	jwtConfig, err := config.LoadJWTConfig("config/jwt.yml")
	if err != nil {
		logger.Fatal("❌ 加载JWT配置失败", zap.Error(err))
	}

	// 初始化仓库层
	healthCheckRepo := repositories.NewHealthCheckRepository(db, logger)
	userRepo := repositories.NewUserRepository(db, logger)
	userGroupRepo := repositories.NewUserGroupRepository(db, logger)

	logger.Debug("📦 仓库层初始化完成")

	// 初始化服务层
	jwtService := services.NewJWTService(jwtConfig, logger)
	authService := services.NewAuthService(userRepo, jwtService, logger)

	logger.Debug("🔐 认证服务初始化完成")
	
	// 创建 Gin 引擎（不使用默认中间件）
	r := gin.New()
	
	// 设置路由（包含日志中间件、数据库依赖和认证服务）
	routes.SetupRoutes(r, logger, healthCheckRepo, userRepo, userGroupRepo, authService)

	logger.Info("✅ 服务启动成功", 
		zap.String("address", ":8080"),
		zap.String("swagger", "http://localhost:8080/swagger/index.html"),
		zap.String("database", "已连接"),
	)
	
	// 启动服务
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("❌ 服务启动失败", zap.Error(err))
	}
}
