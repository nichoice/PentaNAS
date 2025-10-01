package routes

import (
	"github.com/gin-gonic/gin"

	"pnas/internal/app/handlers"
	"pnas/internal/middleware"
)

// RegisterZFSRoutes registers ZFS-related routes
func RegisterZFSRoutes(router *gin.Engine) {
	handler := handlers.NewZFSHandler()

	// ZFS routes under protected group
	protected := router.Group("/api/v1/zfs")
	protected.Use(middleware.AuthMiddleware())
	{
		// Pool routes
		pools := protected.Group("/pools")
		{
			pools.POST("", handler.CreatePool)
			pools.GET("", handler.ListPools)
			pools.GET("/:name/status", handler.GetPoolStatus)
			pools.DELETE("/:name", handler.DestroyPool)
			pools.POST("/:name/scrub", handler.ScrubPool)
		}

		// Dataset routes
		datasets := protected.Group("/datasets")
		{
			datasets.POST("", handler.CreateDataset)
			datasets.GET("", handler.ListDatasets)
			datasets.PUT("/:name", handler.UpdateDataset)
			datasets.DELETE("/:name", handler.DestroyDataset)
		}

		// Snapshot routes
		snapshots := protected.Group("/snapshots")
		{
			snapshots.POST("", handler.CreateSnapshot)
			snapshots.GET("", handler.ListSnapshots)
			snapshots.POST("/rollback", handler.RollbackSnapshot)
			snapshots.POST("/clone", handler.CloneSnapshot)
			snapshots.DELETE("/:name", handler.DestroySnapshot)
		}

		// Volume routes
		volumes := protected.Group("/volumes")
		{
			volumes.POST("", handler.CreateVolume)
			volumes.GET("", handler.ListVolumes)
			volumes.PUT("/:name/resize", handler.ResizeVolume)
			volumes.DELETE("/:name", handler.DestroyVolume)
		}
	}
}
