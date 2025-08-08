package routes

import (
	"pnas/internal/controllers"
	"pnas/internal/i18n"
	"pnas/internal/middlewares"
	"pnas/internal/repositories"
	"pnas/internal/response"
	"pnas/internal/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func SetupRoutes(r *gin.Engine, logger *zap.Logger, healthCheckRepo repositories.HealthCheckRepository, userRepo repositories.UserRepository, userGroupRepo repositories.UserGroupRepository, authService *services.AuthService, i18nManager *i18n.I18n) {
	// 添加全局中间件
	r.Use(middlewares.ZapLogger(logger))
	r.Use(middlewares.Recovery(logger))
	r.Use(middlewares.I18nMiddleware(i18nManager))
	
	// 初始化控制器
	healthController := controllers.NewHealthController(logger, healthCheckRepo)
	userController := controllers.NewUserController(logger, userRepo, userGroupRepo)
	groupController := controllers.NewUserGroupController(logger, userGroupRepo)
	authController := controllers.NewAuthController(logger, authService)
	
	// Swagger 文档路由（不需要认证）
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	// 健康检查路由（不需要认证）
	r.GET("/ping", healthController.Ping)
	
	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证相关路由（不需要认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authController.Login)
			auth.POST("/refresh", authController.RefreshToken)
			auth.POST("/logout", middlewares.JWTAuth(authService, logger), authController.Logout)
		}

		// 健康检查相关路由（不需要认证）
		health := v1.Group("/health")
		{
			health.GET("/ping", healthController.Ping)
		}
		
		// 语言信息路由（不需要认证）
		v1.GET("/language", func(c *gin.Context) {
			response.Success(c, response.GetLanguageInfo(c))
		})
		
		// 需要认证的路由组
		authenticated := v1.Group("/")
		authenticated.Use(middlewares.JWTAuth(authService, logger))
		{
			// 用户相关路由
			users := authenticated.Group("/users")
			{
				users.POST("/", userController.CreateUser)
				users.GET("/", userController.ListUsers)
				users.GET("/:id", userController.GetUser)
				users.PUT("/:id", userController.UpdateUser)
				users.DELETE("/:id", userController.DeleteUser)
			}
			
			// 用户组相关路由
			groups := authenticated.Group("/user-groups")
			{
				groups.POST("/", groupController.CreateUserGroup)
				groups.GET("/", groupController.ListUserGroups)
				groups.GET("/:id", groupController.GetUserGroup)
				groups.GET("/:id/users", groupController.GetUserGroupWithUsers)
				groups.PUT("/:id", groupController.UpdateUserGroup)
				groups.DELETE("/:id", groupController.DeleteUserGroup)
			}
		}
	}
}
