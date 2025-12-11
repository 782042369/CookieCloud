package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DataDir 数据存储目录
	DataDir = "./data"
)

// CookieData Cookie数据结构
type CookieData struct {
	Encrypted string `json:"encrypted"`
}

// InitDataDir 初始化数据目录
func InitDataDir() error {
	if _, err := os.Stat(DataDir); os.IsNotExist(err) {
		return os.Mkdir(DataDir, 0755)
	}
	return nil
}

// SaveEncryptedData 保存加密数据到指定UUID的文件中
func SaveEncryptedData(uuid, encrypted string) error {
	filePath := filepath.Join(DataDir, uuid+".json")
	
	// 创建CookieData结构体实例
	cookieData := CookieData{
		Encrypted: encrypted,
	}
	
	// 序列化为JSON格式
	content, err := json.Marshal(cookieData)
	if err != nil {
		return fmt.Errorf("failed to marshal cookie data: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadEncryptedData 从指定UUID的文件中加载加密数据
func LoadEncryptedData(uuid string) (*CookieData, error) {
	filePath := filepath.Join(DataDir, uuid+".json")
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 解析JSON数据
	var cookieData CookieData
	if err := json.Unmarshal(data, &cookieData); err != nil {
		return nil, err
	}

	return &cookieData, nil
}