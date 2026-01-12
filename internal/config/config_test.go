package config

import (
	"os"
	"testing"
)

// TestLoadWithDefaults 测试使用默认配置加载
func TestLoadWithDefaults(t *testing.T) {
	// 清理环境变量
	os.Unsetenv("PORT")
	os.Unsetenv("API_ROOT")
	os.Unsetenv("DATA_DIR")

	cfg := Load()

	if cfg.Port != "8088" {
		t.Errorf("期望端口为 '8088'，实际得到 '%s'", cfg.Port)
	}

	if cfg.APIRoot != "" {
		t.Errorf("期望 API_ROOT 为空字符串，实际得到 '%s'", cfg.APIRoot)
	}

	if cfg.DataDir != "./data" {
		t.Errorf("期望数据目录为 './data'，实际得到 '%s'", cfg.DataDir)
	}
}

// TestLoadWithEnvVars 测试使用环境变量覆盖默认值
func TestLoadWithEnvVars(t *testing.T) {
	// 设置环境变量
	os.Setenv("PORT", "9000")
	os.Setenv("API_ROOT", "/api/v1")
	os.Setenv("DATA_DIR", "/var/lib/cookiecloud")

	// 测试结束后清理环境变量
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("API_ROOT")
		os.Unsetenv("DATA_DIR")
	}()

	cfg := Load()

	if cfg.Port != "9000" {
		t.Errorf("期望端口为 '9000'，实际得到 '%s'", cfg.Port)
	}

	if cfg.APIRoot != "/api/v1" {
		t.Errorf("期望 API_ROOT 为 '/api/v1'，实际得到 '%s'", cfg.APIRoot)
	}

	if cfg.DataDir != "/var/lib/cookiecloud" {
		t.Errorf("期望数据目录为 '/var/lib/cookiecloud'，实际得到 '%s'", cfg.DataDir)
	}
}

// TestLoadWithTrailingSlash 测试自动移除 API_ROOT 尾部斜杠
func TestLoadWithTrailingSlash(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"/api/", "/api"},
		{"/api/v1/", "/api/v1"},
		{"/api/v1", "/api/v1"},
		{"", ""},
		{"/", ""},
	}

	for _, tc := range testCases {
		os.Setenv("API_ROOT", tc.input)

		cfg := Load()

		if cfg.APIRoot != tc.expected {
			t.Errorf("输入: '%s'，期望: '%s'，实际得到: '%s'", tc.input, tc.expected, cfg.APIRoot)
		}

		os.Unsetenv("API_ROOT")
	}
}

// TestLoadWithPartialEnvVars 测试部分环境变量设置
func TestLoadWithPartialEnvVars(t *testing.T) {
	// 只设置 PORT
	os.Setenv("PORT", "8080")
	defer os.Unsetenv("PORT")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("期望端口为 '8080'，实际得到 '%s'", cfg.Port)
	}

	// 其他应该使用默认值
	if cfg.APIRoot != "" {
		t.Errorf("期望 API_ROOT 为空字符串，实际得到 '%s'", cfg.APIRoot)
	}

	if cfg.DataDir != "./data" {
		t.Errorf("期望数据目录为 './data'，实际得到 '%s'", cfg.DataDir)
	}
}

// TestGetEnv 测试获取环境变量的辅助函数
func TestGetEnv(t *testing.T) {
	// 测试已设置的环境变量
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default_value")
	if result != "test_value" {
		t.Errorf("期望 'test_value'，实际得到 '%s'", result)
	}

	// 测试未设置的环境变量
	result = getEnv("NON_EXISTENT_VAR", "default_value")
	if result != "default_value" {
		t.Errorf("期望 'default_value'，实际得到 '%s'", result)
	}
}

// TestConfigStruct 测试配置结构体不为 nil
func TestConfigStruct(t *testing.T) {
	cfg := Load()

	if cfg == nil {
		t.Fatal("Load() 不应该返回 nil")
	}

	// 验证所有字段都已设置
	if cfg.Port == "" {
		t.Error("Port 字段不应为空")
	}
	if cfg.DataDir == "" {
		t.Error("DataDir 字段不应为空")
	}
	// APIRoot 可以为空，这是正常情况
}
