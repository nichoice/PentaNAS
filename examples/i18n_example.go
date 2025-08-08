package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"pnas/internal/i18n"
	"pnas/internal/middlewares"
	"pnas/internal/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 演示如何在实际项目中使用国际化系统
func main() {
	// 初始化日志器
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 初始化国际化系统
	i18nManager := i18n.NewI18n(i18n.LocaleZhCN, logger)
	if err := i18nManager.LoadMessages("locales"); err != nil {
		logger.Fatal("加载国际化文件失败", zap.Error(err))
	}

	// 创建Gin引擎
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// 添加国际化中间件
	r.Use(middlewares.I18nMiddleware(i18nManager))

	// 模拟用户管理API
	setupUserRoutes(r)

	fmt.Println("🌐 国际化系统演示")
	fmt.Println("==================")

	// 演示各种场景
	demonstrateI18n(r)
}

func setupUserRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		// 获取语言信息
		api.GET("/language", func(c *gin.Context) {
			response.Success(c, response.GetLanguageInfo(c))
		})

		// 用户登录
		api.POST("/auth/login", func(c *gin.Context) {
			var req map[string]string
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ValidationError(c, err)
				return
			}

			username := req["username"]
			password := req["password"]

			// 模拟登录验证
			if username == "" || password == "" {
				response.BadRequest(c, "validation.required", nil)
				return
			}

			if username != "admin" || password != "123456" {
				response.Unauthorized(c, "auth.login.invalid_credentials", nil)
				return
			}

			// 登录成功
			loginData := map[string]interface{}{
				"token":    "mock-jwt-token",
				"username": username,
				"expires":  "2025-08-08T23:59:59Z",
			}
			response.SuccessWithMessage(c, "auth.login.success", loginData)
		})

		// 创建用户
		api.POST("/users", func(c *gin.Context) {
			var req map[string]interface{}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ValidationError(c, err)
				return
			}

			username, ok := req["username"].(string)
			if !ok || username == "" {
				response.BadRequest(c, "validation.required", nil)
				return
			}

			// 模拟用户名已存在
			if username == "existing_user" {
				response.BadRequest(c, "user.username.exists", nil)
				return
			}

			// 创建成功
			userData := map[string]interface{}{
				"id":       123,
				"username": username,
				"status":   "active",
			}
			response.SuccessWithMessage(c, "user.create.success", userData)
		})

		// 获取用户列表
		api.GET("/users", func(c *gin.Context) {
			users := []map[string]interface{}{
				{"id": 1, "username": "admin", "status": "active"},
				{"id": 2, "username": "user1", "status": "active"},
			}

			response.Pagination(c, "user.list.success", users, 2, 1, 10)
		})

		// 用户不存在的情况
		api.GET("/users/999", func(c *gin.Context) {
			response.NotFound(c, "user.not_found", nil)
		})

		// 服务器错误演示
		api.GET("/error", func(c *gin.Context) {
			response.InternalServerError(c, "server.internal_error", fmt.Errorf("模拟数据库连接失败"))
		})
	}
}

func demonstrateI18n(r *gin.Engine) {
	scenarios := []struct {
		name         string
		method       string
		url          string
		body         string
		acceptLang   string
		description  string
	}{
		{
			name:        "中文-获取语言信息",
			method:      "GET",
			url:         "/api/v1/language",
			acceptLang:  "zh-CN",
			description: "获取当前语言设置信息",
		},
		{
			name:        "英文-获取语言信息",
			method:      "GET",
			url:         "/api/v1/language",
			acceptLang:  "en-US",
			description: "Get current language settings",
		},
		{
			name:        "中文-登录成功",
			method:      "POST",
			url:         "/api/v1/auth/login",
			body:        `{"username":"admin","password":"123456"}`,
			acceptLang:  "zh-CN",
			description: "用户登录成功场景",
		},
		{
			name:        "英文-登录失败",
			method:      "POST",
			url:         "/api/v1/auth/login",
			body:        `{"username":"admin","password":"wrong"}`,
			acceptLang:  "en-US",
			description: "Login failure scenario",
		},
		{
			name:        "中文-创建用户成功",
			method:      "POST",
			url:         "/api/v1/users",
			body:        `{"username":"newuser","password":"123456"}`,
			acceptLang:  "zh-CN",
			description: "创建新用户成功",
		},
		{
			name:        "英文-用户名已存在",
			method:      "POST",
			url:         "/api/v1/users",
			body:        `{"username":"existing_user","password":"123456"}`,
			acceptLang:  "en-US",
			description: "Username already exists error",
		},
		{
			name:        "中文-获取用户列表",
			method:      "GET",
			url:         "/api/v1/users",
			acceptLang:  "zh-CN",
			description: "分页获取用户列表",
		},
		{
			name:        "英文-用户不存在",
			method:      "GET",
			url:         "/api/v1/users/999",
			acceptLang:  "en-US",
			description: "User not found error",
		},
		{
			name:        "查询参数覆盖-强制英文",
			method:      "GET",
			url:         "/api/v1/language?lang=en",
			acceptLang:  "zh-CN",
			description: "Query parameter overrides Accept-Language header",
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n%d. %s\n", i+1, scenario.name)
		fmt.Printf("   描述: %s\n", scenario.description)
		fmt.Printf("   请求: %s %s\n", scenario.method, scenario.url)
		if scenario.acceptLang != "" {
			fmt.Printf("   语言: %s\n", scenario.acceptLang)
		}

		// 发送请求
		var req *http.Request
		if scenario.body != "" {
			req = httptest.NewRequest(scenario.method, scenario.url, bytes.NewBufferString(scenario.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(scenario.method, scenario.url, nil)
		}

		if scenario.acceptLang != "" {
			req.Header.Set("Accept-Language", scenario.acceptLang)
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 解析响应
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)

		fmt.Printf("   状态: %d\n", w.Code)
		fmt.Printf("   消息: %s\n", result["message"])

		if data, ok := result["data"]; ok && data != nil {
			fmt.Printf("   数据: %v\n", data)
		}

		if errorMsg, ok := result["error"]; ok && errorMsg != nil {
			fmt.Printf("   错误: %s\n", errorMsg)
		}
	}

	fmt.Println("\n🎉 国际化系统演示完成！")
	fmt.Println("\n💡 使用提示:")
	fmt.Println("   1. 通过 Accept-Language 头部自动检测语言")
	fmt.Println("   2. 通过 ?lang=zh|en 查询参数手动指定语言")
	fmt.Println("   3. 所有API响应都支持中英文切换")
	fmt.Println("   4. 错误消息也会根据语言自动翻译")
}