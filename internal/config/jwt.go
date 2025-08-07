package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// JWTConfig JWT配置结构体
type JWTConfig struct {
	JWT struct {
		SecretKey            string `yaml:"secret_key"`
		ExpiresHours         int    `yaml:"expires_hours"`
		Issuer               string `yaml:"issuer"`
		RefreshExpiresHours  int    `yaml:"refresh_expires_hours"`
	} `yaml:"jwt"`
}

// LoadJWTConfig 加载JWT配置
func LoadJWTConfig(configPath string) (*JWTConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config JWTConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetExpirationTime 获取Token过期时间
func (c *JWTConfig) GetExpirationTime() time.Duration {
	return time.Duration(c.JWT.ExpiresHours) * time.Hour
}

// GetRefreshExpirationTime 获取刷新Token过期时间
func (c *JWTConfig) GetRefreshExpirationTime() time.Duration {
	return time.Duration(c.JWT.RefreshExpiresHours) * time.Hour
}