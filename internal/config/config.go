// Package config 提供配置管理功能
// 从环境变量读取配置，支持默认值和类型安全访问
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config 应用配置
type Config struct {
	Port    string
	APIRoot string
	DataDir string
}

// Load 加载配置（从环境变量读取，没有就用默认值）
func Load() *Config {
	return &Config{
		Port:    getEnv("PORT", "8088"),
		APIRoot: strings.TrimSuffix(getEnv("API_ROOT", ""), "/"),
		DataDir: getEnv("DATA_DIR", "./data"),
	}
}

// Validate 验证配置的合法性，返回验证错误或 nil
func (c *Config) Validate() error {
	port, err := strconv.Atoi(c.Port)
	if err != nil {
		return fmt.Errorf("invalid PORT: %q is not a valid number", c.Port)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid PORT: %d must be between 1 and 65535", port)
	}

	if c.DataDir == "" {
		return fmt.Errorf("DATA_DIR cannot be empty")
	}
	if strings.ContainsAny(c.DataDir, "\x00") {
		return fmt.Errorf("DATA_DIR contains invalid characters")
	}

	return nil
}

// getEnv 获取环境变量，没有就返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
