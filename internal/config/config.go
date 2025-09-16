package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
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
