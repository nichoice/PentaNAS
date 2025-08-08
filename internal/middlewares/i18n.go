package middlewares

import (
	"pnas/internal/i18n"

	"github.com/gin-gonic/gin"
)

// I18nMiddleware 国际化中间件
func I18nMiddleware(i18nManager *i18n.I18n) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取语言设置
		acceptLanguage := c.GetHeader("Accept-Language")
		locale := i18n.GetLocaleFromHeader(acceptLanguage)
		
		// 也可以从查询参数获取语言设置（优先级更高）
		if lang := c.Query("lang"); lang != "" {
			switch lang {
			case "zh", "zh-CN", "zh-cn":
				locale = i18n.LocaleZhCN
			case "en", "en-US", "en-us":
				locale = i18n.LocaleEnUS
			}
		}
		
		// 将语言设置和翻译函数存储到上下文中
		c.Set("locale", locale)
		c.Set("i18n", i18nManager)
		c.Set("t", func(key string, args ...interface{}) string {
			return i18nManager.T(locale, key, args...)
		})
		
		c.Next()
	}
}

// GetT 从上下文获取翻译函数
func GetT(c *gin.Context) func(string, ...interface{}) string {
	if t, exists := c.Get("t"); exists {
		return t.(func(string, ...interface{}) string)
	}
	// 如果没有找到翻译函数，返回一个默认的
	return func(key string, args ...interface{}) string {
		return key
	}
}

// GetLocale 从上下文获取当前语言
func GetLocale(c *gin.Context) i18n.Locale {
	if locale, exists := c.Get("locale"); exists {
		return locale.(i18n.Locale)
	}
	return i18n.LocaleZhCN // 默认中文
}