package config

import (
	"os"
	"testing"
)

// TestValidateValidConfig 测试有效配置验证
func TestValidateValidConfig(t *testing.T) {
	cfg := &Config{
		Port:    "8088",
		APIRoot: "/api",
		DataDir: "./data",
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("有效配置验证失败: %v", err)
	}
}

// TestValidatePort 测试端口验证（格式 + 范围）
func TestValidatePort(t *testing.T) {
	testCases := []struct {
		port     string
		expected bool // true表示应该通过验证
	}{
		// 有效端口
		{"1", true},
		{"80", true},
		{"8080", true},
		{"8088", true},
		{"65535", true},
		// 非数字格式
		{"abc", false},
		{"8088a", false},
		{"", false},
		// 超出范围
		{"0", false},
		{"-1", false},
		{"65536", false},
		{"99999", false},
	}

	for _, tc := range testCases {
		cfg := &Config{
			Port:    tc.port,
			APIRoot: "/api",
			DataDir: "./data",
		}

		err := cfg.Validate()
		if tc.expected && err != nil {
			t.Errorf("端口号 %q 应该验证通过，但失败了: %v", tc.port, err)
		}
		if !tc.expected && err == nil {
			t.Errorf("端口号 %q 应该验证失败，但通过了", tc.port)
		}
	}
}

// TestValidateDataDir 测试数据目录验证
func TestValidateDataDir(t *testing.T) {
	testCases := []struct {
		name    string
		dataDir string
		valid   bool
	}{
		{"有效目录", "./data", true},
		{"绝对路径", "/var/lib/cookiecloud", true},
		{"空目录", "", false},
		{"包含空字节", "./data\x00", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				Port:    "8088",
				APIRoot: "/api",
				DataDir: tc.dataDir,
			}

			err := cfg.Validate()
			if tc.valid && err != nil {
				t.Errorf("数据目录 %q 应该验证通过，但失败了: %v", tc.dataDir, err)
			}
			if !tc.valid && err == nil {
				t.Errorf("数据目录 %q 应该验证失败，但通过了", tc.dataDir)
			}
		})
	}
}

// TestValidateLoadWithDefaults 测试使用默认值加载配置
func TestValidateLoadWithDefaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("DATA_DIR")

	cfg := Load()
	if err := cfg.Validate(); err != nil {
		t.Errorf("默认配置验证失败: %v", err)
	}

	// 验证默认值
	if cfg.Port != "8088" {
		t.Errorf("默认端口应该是 8088，实际是 %s", cfg.Port)
	}
	if cfg.DataDir != "./data" {
		t.Errorf("默认数据目录应该是 ./data，实际是 %s", cfg.DataDir)
	}
}
