package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pnas/internal/app/handlers"
)

// SetupSambaRoutes 设置Samba相关路由
func SetupSambaRoutes(router *gin.RouterGroup, db *gorm.DB) {
	sambaHandler := handlers.NewSambaHandler(db)

	// Samba API路由组
	samba := router.Group("/samba")
	{
		// 账号管理
		accounts := samba.Group("/accounts")
		{
			accounts.POST("", sambaHandler.CreateAccount)                    // 创建Samba账号
			accounts.GET("", sambaHandler.ListAccounts)                     // 获取账号列表
			accounts.GET("/:id", sambaHandler.GetAccount)                   // 获取单个账号
			accounts.PUT("/:id", sambaHandler.UpdateAccount)                // 更新账号
			accounts.DELETE("/:id", sambaHandler.DeleteAccount)             // 删除账号
		}

		// 共享管理
		shares := samba.Group("/shares")
		{
			shares.POST("", sambaHandler.CreateShare)                       // 创建共享
			shares.GET("", sambaHandler.ListShares)                         // 获取共享列表
			shares.GET("/:id", sambaHandler.GetShare)                       // 获取单个共享
			shares.PUT("/:id", sambaHandler.UpdateShare)                    // 更新共享
			shares.DELETE("/:id", sambaHandler.DeleteShare)                 // 删除共享

			// 共享访问权限管理
			shares.POST("/access", sambaHandler.SetShareAccess)             // 设置访问权限
			shares.GET("/:id/access", sambaHandler.GetShareAccess)          // 获取访问权限列表
			shares.DELETE("/:id/access/:accountId", sambaHandler.RemoveShareAccess) // 移除访问权限

			// 时间机器功能
			shares.PUT("/:id/timemachine/enable", sambaHandler.EnableTimeMachine)    // 启用时间机器
			shares.PUT("/:id/timemachine/disable", sambaHandler.DisableTimeMachine)  // 禁用时间机器
			shares.GET("/:id/timemachine/status", sambaHandler.GetTimeMachineStatus) // 获取时间机器状态

			// 多通道功能
			shares.PUT("/:id/multichannel/enable", sambaHandler.EnableMultiChannel)   // 启用多通道
			shares.PUT("/:id/multichannel/disable", sambaHandler.DisableMultiChannel) // 禁用多通道
		}

		// 时间机器专用路由
		timemachine := samba.Group("/timemachine")
		{
			timemachine.GET("/shares", sambaHandler.GetTimeMachineShares)    // 获取时间机器共享列表
		}

		// 回收站管理
		recycle := samba.Group("/recycle")
		{
			recycle.GET("/items", sambaHandler.ListRecycleBinItems)          // 获取回收站条目列表
			recycle.PUT("/items/restore", sambaHandler.RestoreRecycleBinItem) // 恢复回收站条目
			recycle.DELETE("/items/:id", sambaHandler.PermanentDeleteRecycleBinItem) // 永久删除条目
			recycle.DELETE("/shares/:shareId/empty", sambaHandler.EmptyRecycleBin)   // 清空指定共享的回收站
			recycle.GET("/statistics", sambaHandler.GetRecycleBinStatistics)         // 获取回收站统计信息
		}

		// 配置管理
		config := samba.Group("/config")
		{
			config.GET("/global", sambaHandler.GetGlobalConfig)             // 获取全局配置
			config.PUT("/global", sambaHandler.UpdateGlobalConfig)          // 更新全局配置
			config.GET("/generate", sambaHandler.GenerateConfig)            // 生成配置文件
			config.POST("/write", sambaHandler.WriteConfig)                 // 写入配置文件
			config.POST("/reload", sambaHandler.ReloadConfig)               // 重载配置
		}

		// 监控和状态
		monitoring := samba.Group("/monitoring")
		{
			monitoring.GET("/status", sambaHandler.GetSambaStatus)          // 获取Samba整体状态
			monitoring.GET("/services", sambaHandler.GetServiceStatus)       // 获取服务状态
			monitoring.GET("/connections", sambaHandler.GetActiveConnections) // 获取活跃连接
			monitoring.GET("/connections/history", sambaHandler.GetConnectionHistory) // 获取连接历史
			monitoring.GET("/audit/logs", sambaHandler.GetAuditLogs)        // 获取审计日志
			monitoring.GET("/multichannel/status", sambaHandler.GetMultiChannelStatus) // 获取多通道状态
		}
	}
}

// RegisterSambaRoutes 注册Samba路由到主路由器
func RegisterSambaRoutes(r *gin.Engine, db *gorm.DB) {
	api := r.Group("/api/v1")
	SetupSambaRoutes(api, db)
}