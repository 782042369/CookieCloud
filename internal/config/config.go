// Package config 提供配置管理功能
// 从环境变量读取配置，支持默认值和类型安全访问
package config

import (
	"os"
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

// getEnv 获取环境变量，没有就返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
