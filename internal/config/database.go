package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig 数据库配置结构体
type DatabaseConfig struct {
	Database struct {
		Development struct {
			Driver          string `yaml:"driver"`
			DSN             string `yaml:"dsn"`
			MaxIdleConns    int    `yaml:"max_idle_conns"`
			MaxOpenConns    int    `yaml:"max_open_conns"`
			ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
			LogLevel        string `yaml:"log_level"`
		} `yaml:"development"`
		Production struct {
			Driver          string `yaml:"driver"`
			Host            string `yaml:"host"`
			Port            int    `yaml:"port"`
			User            string `yaml:"user"`
			Password        string `yaml:"password"`
			DBName          string `yaml:"dbname"`
			SSLMode         string `yaml:"sslmode"`
			TimeZone        string `yaml:"timezone"`
			MaxIdleConns    int    `yaml:"max_idle_conns"`
			MaxOpenConns    int    `yaml:"max_open_conns"`
			ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
			LogLevel        string `yaml:"log_level"`
		} `yaml:"production"`
		Test struct {
			Driver          string `yaml:"driver"`
			DSN             string `yaml:"dsn"`
			MaxIdleConns    int    `yaml:"max_idle_conns"`
			MaxOpenConns    int    `yaml:"max_open_conns"`
			ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
			LogLevel        string `yaml:"log_level"`
		} `yaml:"test"`
	} `yaml:"database"`
}

// DBConfig 单个数据库配置
type DBConfig struct {
	Driver          string
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
	LogLevel        string
}

// InitDatabase 初始化数据库连接
func InitDatabase(configPath string, env string, zapLogger *zap.Logger) (*gorm.DB, error) {
	// 读取配置文件
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取数据库配置文件失败: %w", err)
	}

	// 解析配置
	var config DatabaseConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("解析数据库配置失败: %w", err)
	}

	// 根据环境选择配置
	var dbConfig DBConfig
	switch env {
	case "production":
		dbConfig = DBConfig{
			Driver:          config.Database.Production.Driver,
			DSN:             buildPostgresDSN(config.Database.Production),
			MaxIdleConns:    config.Database.Production.MaxIdleConns,
			MaxOpenConns:    config.Database.Production.MaxOpenConns,
			ConnMaxLifetime: config.Database.Production.ConnMaxLifetime,
			LogLevel:        config.Database.Production.LogLevel,
		}
	case "test":
		dbConfig = DBConfig{
			Driver:          config.Database.Test.Driver,
			DSN:             config.Database.Test.DSN,
			MaxIdleConns:    config.Database.Test.MaxIdleConns,
			MaxOpenConns:    config.Database.Test.MaxOpenConns,
			ConnMaxLifetime: config.Database.Test.ConnMaxLifetime,
			LogLevel:        config.Database.Test.LogLevel,
		}
	default: // development
		dbConfig = DBConfig{
			Driver:          config.Database.Development.Driver,
			DSN:             config.Database.Development.DSN,
			MaxIdleConns:    config.Database.Development.MaxIdleConns,
			MaxOpenConns:    config.Database.Development.MaxOpenConns,
			ConnMaxLifetime: config.Database.Development.ConnMaxLifetime,
			LogLevel:        config.Database.Development.LogLevel,
		}
	}

	// 创建数据库连接
	db, err := connectDatabase(dbConfig, zapLogger)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(dbConfig.ConnMaxLifetime) * time.Second)

	zapLogger.Info("数据库连接成功", 
		zap.String("driver", dbConfig.Driver),
		zap.String("environment", env),
		zap.Int("max_idle_conns", dbConfig.MaxIdleConns),
		zap.Int("max_open_conns", dbConfig.MaxOpenConns),
	)

	return db, nil
}

// connectDatabase 连接数据库
func connectDatabase(config DBConfig, zapLogger *zap.Logger) (*gorm.DB, error) {
	// 配置 GORM 日志
	gormLogger := logger.New(
		&GormLoggerWriter{zapLogger: zapLogger},
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  parseGormLogLevel(config.LogLevel),
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	var dialector gorm.Dialector

	switch config.Driver {
	case "sqlite":
		// 确保 SQLite 数据库目录存在
		if config.DSN != ":memory:" {
			dir := filepath.Dir(config.DSN)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("创建数据库目录失败: %w", err)
			}
		}
		dialector = sqlite.Open(config.DSN)
	case "postgres":
		dialector = postgres.Open(config.DSN)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", config.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// buildPostgresDSN 构建 PostgreSQL DSN
func buildPostgresDSN(prodConfig struct {
	Driver          string `yaml:"driver"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	DBName          string `yaml:"dbname"`
	SSLMode         string `yaml:"sslmode"`
	TimeZone        string `yaml:"timezone"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	LogLevel        string `yaml:"log_level"`
}) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		prodConfig.Host, prodConfig.User, prodConfig.Password, prodConfig.DBName, prodConfig.Port, prodConfig.SSLMode, prodConfig.TimeZone)
}

// parseGormLogLevel 解析 GORM 日志级别
func parseGormLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}

// GormLoggerWriter GORM 日志写入器，集成 zap
type GormLoggerWriter struct {
	zapLogger *zap.Logger
}

func (w *GormLoggerWriter) Printf(format string, args ...interface{}) {
	w.zapLogger.Info(fmt.Sprintf(format, args...))
}