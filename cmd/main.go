package main

import (
	"pnas/internal/routes"
	_ "pnas/cmd/docs" // 导入生成的 docs
	"github.com/gin-gonic/gin"
)

// @title PNAS API
// @version 1.0
// @description PNAS 项目 API 文档
// @host localhost:8080
// @BasePath /
func main() {
	gin.SetMode(gin.DebugMode)
	gin.ForceConsoleColor()
	
	r := gin.Default()
	
	// 设置路由
	routes.SetupRoutes(r)

	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}
