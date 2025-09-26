package routes

import (
	"pnas/internal/app/handlers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterNFSRoutes registers NFS-related routes
func RegisterNFSRoutes(router *gin.Engine, db *gorm.DB) {
	handler := handlers.NewNFSHandler()

	// NFS routes group
	nfs := router.Group("/api/v1/nfs")
	{
		// Export management routes
		exports := nfs.Group("/exports")
		{
			exports.POST("", handler.CreateExport)
			exports.GET("", handler.GetExports)
			exports.GET("/:id", handler.GetExport)
			exports.PUT("/:id", handler.UpdateExport)
			exports.DELETE("/:id", handler.DeleteExport)

			// Export-specific multipath routes
			exports.GET("/:id/multipath", handler.GetMultipathConfs)
			exports.GET("/:id/multipath/load-balance", handler.GetPathLoadBalance)
			exports.POST("/:id/multipath/optimize", handler.AutoOptimizePaths)

			// Export-specific client access routes
			exports.GET("/:id/client-access", handler.GetClientAccess)

			// Export-specific quota routes
			exports.GET("/:id/quotas", handler.GetQuotas)
		}

		// Global configuration routes
		config := nfs.Group("/config")
		{
			config.GET("", handler.GetGlobalConfig)
			config.PUT("", handler.UpdateGlobalConfig)
			config.POST("/reset", handler.ResetGlobalConfig)
		}

		// Client access control routes
		clientAccess := nfs.Group("/client-access")
		{
			clientAccess.POST("", handler.CreateClientAccess)
			clientAccess.PUT("/:id", handler.UpdateClientAccess)
			clientAccess.DELETE("/:id", handler.DeleteClientAccess)
		}

		// Operation management routes
		operations := nfs.Group("/operations")
		{
			operations.GET("", handler.GetAllOperations)
			operations.GET("/:operation_id", handler.GetOperationStatus)
			operations.POST("/async", handler.ExecuteAsyncOperation)
			operations.POST("/sync", handler.ExecuteSyncOperation)
			operations.POST("/:operation_id/cancel", handler.CancelOperation)
		}

		// Multipath management routes
		multipath := nfs.Group("/multipath")
		{
			multipath.POST("", handler.CreateMultipathConf)
			multipath.PUT("/:id", handler.UpdateMultipathConf)
			multipath.DELETE("/:id", handler.DeleteMultipathConf)
			multipath.GET("/health", handler.CheckMultipathHealth)
		}

		// Quota management routes
		quotas := nfs.Group("/quotas")
		{
			quotas.POST("", handler.CreateQuota)
		}
	}
}