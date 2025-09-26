package routes

import (
	"pnas/internal/app/handlers"
	"pnas/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupIscsiRoutes(router *gin.Engine) {
	handler := handlers.NewiSCSIHandler()

	// Apply middleware
	api := router.Group("/api/v1/iscsi")
	api.Use(middleware.AuthMiddleware())
	api.Use(middleware.AuditMiddleware())

	// Global configuration
	api.GET("/config", handler.GetGlobalConfig)
	api.PUT("/config", handler.UpdateGlobalConfig)

	// Targets
	targets := api.Group("/targets")
	{
		targets.POST("", handler.CreateTarget)
		targets.GET("", handler.ListTargets)
		targets.GET("/stats", handler.GetTargetStats)
		targets.POST("/batch", handler.BatchOperationTargets)
		targets.GET("/:id", handler.GetTarget)
		targets.PUT("/:id", handler.UpdateTarget)
		targets.DELETE("/:id", handler.DeleteTarget)
	}

	// LUNs
	luns := api.Group("/luns")
	{
		luns.POST("", handler.CreateLUN)
		luns.GET("", handler.ListLUNs)
		luns.POST("/batch", handler.BatchOperationLUNs)
		luns.GET("/:id", handler.GetLUN)
		luns.PUT("/:id", handler.UpdateLUN)
		luns.DELETE("/:id", handler.DeleteLUN)
	}

	// ACLs
	acls := api.Group("/acls")
	{
		acls.POST("", handler.CreateACL)
		acls.GET("", handler.ListACLs)
		acls.POST("/batch", handler.BatchOperationACLs)
		acls.GET("/:id", handler.GetACL)
		acls.PUT("/:id", handler.UpdateACL)
		acls.DELETE("/:id", handler.DeleteACL)
	}

	// LUN Mappings
	lunMappings := api.Group("/lun-mappings")
	{
		lunMappings.POST("", handler.CreateLUNMapping)
		lunMappings.PUT("/:id", handler.UpdateLUNMapping)
		lunMappings.DELETE("/:id", handler.DeleteLUNMapping)
	}
}