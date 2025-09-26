package routes

import (
	"pnas/cmd/docs"
	"pnas/internal/config"
	"pnas/internal/controllers"
	"pnas/internal/database"
	"pnas/internal/logging"
	"pnas/internal/middleware"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes sets up the API routes
func SetupRoutes(router *gin.Engine) {
	// Add logging middleware to all routes
	router.Use(middleware.LoggingMiddleware(logging.Logger))

	// Add audit middleware to all routes if enabled
	if config.AppConfig.Audit.Enabled && config.AppConfig.Audit.EnableAPI {
		router.Use(middleware.AuditMiddleware())
	}

	// Swagger information
	docs.SwaggerInfo.Title = config.AppConfig.Server.Title
	docs.SwaggerInfo.Description = config.AppConfig.Server.Description
	docs.SwaggerInfo.Version = config.AppConfig.Server.Version
	docs.SwaggerInfo.Host = "localhost:" + config.AppConfig.Server.Port
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes (no authentication required)
	public := router.Group("/api/v1")
	{
		// Login endpoint
		public.POST("/login", controllers.Login)

	}

	// Protected routes (authentication required)
	protected := router.Group("/api/v1")
	// protected.Use(middleware.AuthMiddleware())
	{
		// User management routes
		users := protected.Group("/users")
		{
			users.POST("", controllers.CreateUser)
			users.GET("", controllers.GetUsers)
			users.GET("/:id", controllers.GetUser)
			users.PUT("/:id", controllers.UpdateUser)
			users.DELETE("/:id", controllers.DeleteUser)
			users.PUT("/:id/status", controllers.UpdateUserStatus)
		}

		// Role management routes
		roles := protected.Group("/roles")
		{
			roles.GET("", controllers.GetRoles)
			roles.GET("/:id", controllers.GetRole)
		}

		// Authorization management routes
		auth := protected.Group("/auth")
		{
			auth.POST("/assign", controllers.AssignRole)
			auth.POST("/revoke", controllers.RevokeRole)
			auth.GET("/users/:user_id/roles", controllers.GetUserRoles)
			auth.GET("/roles/:role_id/users", controllers.GetRoleUsers)
		}

		storage := protected.Group("/storage")
		storage.Use(middleware.RateLimitMiddleware(5, 10*time.Second))
		{
			storage.GET("/disks", controllers.GetDisks)
			storage.GET("/vgs", controllers.GetVG)
			storage.GET("/lvs", controllers.GetLV)
			storage.POST("create_vg", controllers.CreateVG)
		}

		// Monitoring routes
		monitoring := protected.Group("/monitoring")
		{
			monitoring.GET("/system", controllers.GetSystemInfo)
			monitoring.GET("/cpu", controllers.GetCPUInfo)
			monitoring.GET("/memory", controllers.GetMemoryInfo)
			monitoring.GET("/disk", controllers.GetDiskInfo)
			monitoring.GET("/network", controllers.GetNetworkInfo)
			monitoring.GET("/storage-protocols", controllers.GetStorageProtocolInfo)
			monitoring.GET("/complete", controllers.GetCompleteMonitoringData)
			monitoring.GET("/service/status", controllers.GetServiceStatus)
			monitoring.GET("/metrics", controllers.GetMetrics)

			// WebSocket routes
			monitoring.GET("/websocket", controllers.HandleWebSocket)
			monitoring.GET("/websocket/info", controllers.GetWebSocketInfo)
			monitoring.POST("/websocket/broadcast", controllers.BroadcastMessage)
			monitoring.POST("/websocket/client/:client_id", controllers.SendToClient)
		}

		// Audit routes
		audit := protected.Group("/audit")
		{
			audit.GET("/logs", controllers.GetAuditLogs)
			audit.GET("/stats", controllers.GetAuditStats)
			audit.GET("/heatmap", controllers.GetFileAccessHeatmap)
			audit.GET("/anomalies", controllers.DetectAnomalies)
			audit.GET("/users/:user_id/timeline", controllers.GetUserActivityTimeline)
		}
	}

	// Register Samba routes
	RegisterSambaRoutes(router, database.DB)

	// Register NFS routes
	RegisterNFSRoutes(router, database.DB)
}
