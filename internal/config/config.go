package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Audit    AuditConfig    `mapstructure:"audit"`
}

// ServerConfig holds the server configuration
type ServerConfig struct {
	Port        string `mapstructure:"port"`
	Debug       bool   `mapstructure:"debug"`
	Version     string `mapstructure:"version"`
	Title       string `mapstructure:"title"`
	Description string `mapstructure:"description"`
}

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// JWTConfig holds the JWT configuration
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
}

// AuditConfig holds the audit configuration
type AuditConfig struct {
	Enabled           bool     `mapstructure:"enabled"`
	WatchPaths        []string `mapstructure:"watch_paths"`
	ExcludePatterns   []string `mapstructure:"exclude_patterns"`
	BatchSize         int      `mapstructure:"batch_size"`
	FlushInterval     int      `mapstructure:"flush_interval"` // seconds
	RecursiveWatch    bool     `mapstructure:"recursive_watch"`
	EnableAPI         bool     `mapstructure:"enable_api"`
	EnableFilesystem  bool     `mapstructure:"enable_filesystem"`
	RetentionDays     int      `mapstructure:"retention_days"`
}

var AppConfig *Config

// LoadConfig loads the configuration from file
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Enable reading environment variables
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.debug", false)
	viper.SetDefault("database.path", "pnas.db")
	viper.SetDefault("jwt.secret", "pnas-secret-key")
	viper.SetDefault("jwt.expire", 24)

	// Audit default values
	viper.SetDefault("audit.enabled", true)
	viper.SetDefault("audit.watch_paths", []string{"/wuzhou"})
	viper.SetDefault("audit.exclude_patterns", []string{
		".DS_Store", ".Spotlight-V100", ".Trashes", ".fseventsd",
		".TemporaryItems", "Thumbs.db", "desktop.ini", "~$*",
		".tmp", ".temp", ".swp", ".~", "#*",
	})
	viper.SetDefault("audit.batch_size", 100)
	viper.SetDefault("audit.flush_interval", 5)
	viper.SetDefault("audit.recursive_watch", true)
	viper.SetDefault("audit.enable_api", true)
	viper.SetDefault("audit.enable_filesystem", true)
	viper.SetDefault("audit.retention_days", 90)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using defaults")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	AppConfig = &config
	return AppConfig, nil
}

// GenerateDefaultConfig creates a default configuration file
func GenerateDefaultConfig(configPath string) error {
	// Create default config
	defaultConfig := Config{
		Server: ServerConfig{
			Port:        "8080",
			Debug:       false,
			Version:     "1.0.0",
			Title:       "PNAS - Personal Network Attached Storage",
			Description: "个人网络附加存储系统",
		},
		Database: DatabaseConfig{
			Path: "pnas.db",
		},
		JWT: JWTConfig{
			Secret: "your-secret-key-change-this",
			Expire: 24,
		},
		Audit: AuditConfig{
			Enabled:    true,
			WatchPaths: []string{"/wuzhou"},
			ExcludePatterns: []string{
				".DS_Store", ".Spotlight-V100", ".Trashes", ".fseventsd",
				".TemporaryItems", "Thumbs.db", "desktop.ini", "~$*",
				".tmp", ".temp", ".swp", ".~", "#*", "*.log", "*.lock",
			},
			BatchSize:        100,
			FlushInterval:    5,
			RecursiveWatch:   true,
			EnableAPI:        true,
			EnableFilesystem: true,
			RetentionDays:    90,
		},
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Printf("Generated default configuration file: %s", configPath)
	return nil
}

// EnsureConfigExists checks if config file exists, if not, creates a default one
func EnsureConfigExists(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Configuration file %s not found, generating default configuration", configPath)
		return GenerateDefaultConfig(configPath)
	}
	return nil
}

// EnsureDirectoryExists ensures the specified directory exists
func EnsureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Printf("Directory %s not found, creating it", dirPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
		log.Printf("Created directory: %s", dirPath)
	}
	return nil
}
