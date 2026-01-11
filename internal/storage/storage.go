package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	dataDir = "./data" // 默认数据目录
	fileLocks sync.Map // 文件锁映射
)

// CookieData Cookie数据结构
type CookieData struct {
	Encrypted string `json:"encrypted"`
}

// InitDataDir 初始化数据目录
func InitDataDir(dir string) error {
	dataDir = dir
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return os.Mkdir(dataDir, 0755)
	}
	return nil
}

// getFileLock 获取指定UUID的文件锁
func getFileLock(uuid string) *sync.Mutex {
	lock, _ := fileLocks.LoadOrStore(uuid, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// SaveEncryptedData 保存加密数据到指定UUID的文件中
func SaveEncryptedData(uuid, encrypted string) error {
	// 获取文件锁
	lock := getFileLock(uuid)
	lock.Lock()
	defer lock.Unlock()

	filePath := filepath.Join(dataDir, uuid+".json")

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
	filePath := filepath.Join(dataDir, uuid+".json")

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
