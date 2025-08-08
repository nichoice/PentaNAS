package i18n

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// Locale 语言类型
type Locale string

const (
	LocaleZhCN Locale = "zh-CN" // 简体中文
	LocaleEnUS Locale = "en-US" // 美式英语
)

// I18n 国际化管理器
type I18n struct {
	defaultLocale Locale
	messages      map[Locale]map[string]string
	mutex         sync.RWMutex
	logger        *zap.Logger
}

// NewI18n 创建国际化管理器实例
func NewI18n(defaultLocale Locale, logger *zap.Logger) *I18n {
	return &I18n{
		defaultLocale: defaultLocale,
		messages:      make(map[Locale]map[string]string),
		logger:        logger,
	}
}

// LoadMessages 从目录加载语言文件
func (i *I18n) LoadMessages(dir string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// 支持的语言文件
	locales := []Locale{LocaleZhCN, LocaleEnUS}
	
	for _, locale := range locales {
		filename := filepath.Join(dir, fmt.Sprintf("%s.json", locale))
		if err := i.loadMessageFile(locale, filename); err != nil {
			i.logger.Warn("加载语言文件失败", 
				zap.String("locale", string(locale)),
				zap.String("file", filename),
				zap.Error(err),
			)
			// 继续加载其他语言文件，不中断
			continue
		}
		i.logger.Debug("语言文件加载成功", 
			zap.String("locale", string(locale)),
			zap.String("file", filename),
		)
	}

	return nil
}

// loadMessageFile 加载单个语言文件
func (i *I18n) loadMessageFile(locale Locale, filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		return err
	}

	i.messages[locale] = messages
	return nil
}

// T 翻译函数
func (i *I18n) T(locale Locale, key string, args ...interface{}) string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	// 尝试获取指定语言的翻译
	if messages, exists := i.messages[locale]; exists {
		if message, exists := messages[key]; exists {
			if len(args) > 0 {
				return fmt.Sprintf(message, args...)
			}
			return message
		}
	}

	// 如果指定语言没有找到，尝试默认语言
	if locale != i.defaultLocale {
		if messages, exists := i.messages[i.defaultLocale]; exists {
			if message, exists := messages[key]; exists {
				if len(args) > 0 {
					return fmt.Sprintf(message, args...)
				}
				return message
			}
		}
	}

	// 如果都没找到，返回key本身
	i.logger.Warn("翻译键未找到", 
		zap.String("locale", string(locale)),
		zap.String("key", key),
	)
	return key
}

// GetLocaleFromHeader 从HTTP头部获取语言设置
func GetLocaleFromHeader(acceptLanguage string) Locale {
	if acceptLanguage == "" {
		return LocaleZhCN // 默认中文
	}

	// 解析Accept-Language头部
	languages := strings.Split(acceptLanguage, ",")
	for _, lang := range languages {
		// 去除权重信息 (如 zh-CN;q=0.9)
		lang = strings.TrimSpace(strings.Split(lang, ";")[0])
		
		switch {
		case strings.HasPrefix(lang, "zh"):
			return LocaleZhCN
		case strings.HasPrefix(lang, "en"):
			return LocaleEnUS
		}
	}

	return LocaleZhCN // 默认中文
}

// GetSupportedLocales 获取支持的语言列表
func (i *I18n) GetSupportedLocales() []Locale {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	locales := make([]Locale, 0, len(i.messages))
	for locale := range i.messages {
		locales = append(locales, locale)
	}
	return locales
}