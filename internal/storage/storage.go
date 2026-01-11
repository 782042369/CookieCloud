package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// CookieData Cookie数据结构
type CookieData struct {
	Encrypted string `json:"encrypted"`
}

// Storage 数据存储管理器，持有配置和状态
type Storage struct {
	dataDir   string       // 数据目录路径
	fileLocks sync.Map     // 文件锁映射（每个UUID一个锁）
}

// New 创建一个新的 Storage 实例（依赖注入配置）
func New(dataDir string) (*Storage, error) {
	// 检查并创建数据目录
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.Mkdir(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("无法创建数据目录 %s: %w", dataDir, err)
		}
	}

	return &Storage{
		dataDir: dataDir,
	}, nil
}

// getFileLock 获取指定UUID的文件锁
func (s *Storage) getFileLock(uuid string) *sync.Mutex {
	lock, _ := s.fileLocks.LoadOrStore(uuid, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// SaveEncryptedData 保存加密数据到指定UUID的文件中
func (s *Storage) SaveEncryptedData(uuid, encrypted string) error {
	// 获取文件锁
	lock := s.getFileLock(uuid)
	lock.Lock()
	defer lock.Unlock()

	filePath := filepath.Join(s.dataDir, uuid+".json")

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
func (s *Storage) LoadEncryptedData(uuid string) (*CookieData, error) {
	filePath := filepath.Join(s.dataDir, uuid+".json")

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
