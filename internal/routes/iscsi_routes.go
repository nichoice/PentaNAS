package routes

import (
	"pnas/internal/app/handlers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterISCSIRoutes registers iSCSI-related routes
func RegisterISCSIRoutes(router *gin.Engine, db *gorm.DB) {
	handler := handlers.NewISCSIHandler()

	// iSCSI routes group
	iscsi := router.Group("/api/v1/iscsi")
	{
		// Target management routes
		targets := iscsi.Group("/targets")
		{
			targets.POST("", handler.CreateTarget)
			targets.GET("", handler.GetTargets)
			targets.GET("/:id", handler.GetTarget)
			targets.PUT("/:id", handler.UpdateTarget)
			targets.DELETE("/:id", handler.DeleteTarget)
			targets.POST("/:id/start", handler.StartTarget)
			targets.POST("/:id/stop", handler.StopTarget)
			targets.GET("/:id/status", handler.GetTargetStatus)

			// Target-specific LUN routes
			targets.GET("/:id/luns", handler.GetTargetLUNs)
		}

		// LUN management routes
		luns := iscsi.Group("/luns")
		{
			luns.POST("", handler.CreateLUN)
			luns.GET("", handler.GetLUNs)
			luns.GET("/:id", handler.GetLUN)
			luns.PUT("/:id", handler.UpdateLUN)
			luns.DELETE("/:id", handler.DeleteLUN)
			luns.POST("/:id/map", handler.MapLUNToTarget)
			luns.DELETE("/:id/map/:target_id", handler.UnmapLUNFromTarget)
		}

		// ACL management routes
		acls := iscsi.Group("/acls")
		{
			acls.POST("", handler.CreateACL)
			acls.GET("", handler.GetACLs)
			acls.GET("/:id", handler.GetACL)
			acls.PUT("/:id", handler.UpdateACL)
			acls.DELETE("/:id", handler.DeleteACL)
		}

		// Global configuration routes
		config := iscsi.Group("/config")
		{
			config.GET("", handler.GetGlobalConfig)
			config.PUT("", handler.UpdateGlobalConfig)
			config.POST("/reset", handler.ResetGlobalConfig)
		}

		// Service management routes
		service := iscsi.Group("/service")
		{
			service.GET("/status", handler.GetServiceStatus)
			service.POST("/start", handler.StartService)
			service.POST("/stop", handler.StopService)
			service.POST("/restart", handler.RestartService)
		}

		// Session management routes
		sessions := iscsi.Group("/sessions")
		{
			sessions.GET("", handler.GetSessions)
			sessions.GET("/:id", handler.GetSession)
			sessions.DELETE("/:id", handler.TerminateSession)
		}

		// Connection monitoring routes
		connections := iscsi.Group("/connections")
		{
			connections.GET("", handler.GetConnections)
			connections.GET("/history", handler.GetConnectionHistory)
		}

		// Storage pool management routes
		pools := iscsi.Group("/pools")
		{
			pools.POST("", handler.CreateStoragePool)
			pools.GET("", handler.GetStoragePools)
			pools.GET("/:id", handler.GetStoragePool)
			pools.PUT("/:id", handler.UpdateStoragePool)
			pools.DELETE("/:id", handler.DeleteStoragePool)
		}

		// Performance monitoring routes
		monitoring := iscsi.Group("/monitoring")
		{
			monitoring.GET("/performance", handler.GetPerformanceStats)
			monitoring.GET("/targets/:target_id/stats", handler.GetTargetStats)
			monitoring.GET("/luns/:lun_id/stats", handler.GetLUNStats)
		}

		// Audit logging routes
		audit := iscsi.Group("/audit")
		{
			audit.GET("/logs", handler.GetAuditLogs)
			audit.GET("/stats", handler.GetAuditStats)
		}
	}
}