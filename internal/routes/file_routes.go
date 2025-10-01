package routes

import (
	"github.com/gin-gonic/gin"

	"pnas/internal/app/handlers"
	"pnas/internal/middleware"
)

// RegisterFileRoutes registers file management routes
func RegisterFileRoutes(router *gin.Engine) {
	handler := handlers.NewFileHandler()

	// File routes under protected group
	protected := router.Group("/api/v1/files")
	protected.Use(middleware.AuthMiddleware())
	{
		// Directory operations
		protected.POST("/directories", handler.CreateDirectory)

		// File upload/download
		protected.POST("/upload", handler.UploadFile)
		protected.GET("/:id/download", handler.DownloadFile)

		// File listing and search
		protected.GET("", handler.ListFiles)
		protected.GET("/search", handler.SearchFiles)

		// File operations
		protected.PUT("/move", handler.MoveFile)
		protected.PUT("/:id/rename", handler.RenameFile)
		protected.DELETE("/:id", handler.DeleteFile)

		// Permission management
		permissions := protected.Group("/permissions")
		{
			permissions.POST("/grant", handler.GrantPermission)
			permissions.POST("/revoke", handler.RevokePermission)
		}
		protected.GET("/:id/permissions", handler.GetFilePermissions)

		// Recycle bin
		recycleBin := protected.Group("/recycle-bin")
		{
			recycleBin.GET("", handler.ListRecycleBin)
			recycleBin.POST("/restore", handler.RestoreFile)
			recycleBin.POST("/empty", handler.EmptyRecycleBin)
			recycleBin.DELETE("/:id", handler.PermanentlyDeleteFile)
		}
	}
}
