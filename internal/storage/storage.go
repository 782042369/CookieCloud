// Package storage 提供数据持久化功能
// 使用 JSON 文件存储加密数据，支持并发安全的文件读写
package storage

import (
	"context"
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

// fileLocks 全局文件锁（使用 sync.Map，Go 会自动清理未使用的条目）
var fileLocks sync.Map

// getFileLock 获取指定UUID的文件锁
func getFileLock(uuid string) *sync.Mutex {
	lock, _ := fileLocks.LoadOrStore(uuid, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// Storage 数据存储管理器
type Storage struct {
	dataDir string // 数据目录路径
}

// New 创建一个新的 Storage 实例
func New(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建数据目录 %s: %w", dataDir, err)
	}
	return &Storage{dataDir: dataDir}, nil
}

// SaveEncryptedData 保存加密数据到指定 UUID 的文件中
// 支持 context 取消信号，在文件操作前检查 context 状态
func (s *Storage) SaveEncryptedData(ctx context.Context, uuid, encrypted string) error {
	// 检查 context 是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err() // 返回 context 取消错误
	default:
		// 继续执行
	}

	lock := getFileLock(uuid)
	lock.Lock()
	defer lock.Unlock()

	filePath := filepath.Join(s.dataDir, uuid+".json")
	content, err := json.Marshal(CookieData{Encrypted: encrypted})
	if err != nil {
		return fmt.Errorf("marshal cookie data: %w", err)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// LoadEncryptedData 从指定 UUID 的文件中加载加密数据
// 支持 context 取消信号，在文件操作前检查 context 状态
func (s *Storage) LoadEncryptedData(ctx context.Context, uuid string) (*CookieData, error) {
	// 检查 context 是否已取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err() // 返回 context 取消错误
	default:
		// 继续执行
	}

	filePath := filepath.Join(s.dataDir, uuid+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cookieData CookieData
	if err := json.Unmarshal(data, &cookieData); err != nil {
		return nil, fmt.Errorf("unmarshal cookie data: %w", err)
	}

	return &cookieData, nil
}

// Close 关闭存储管理器（空实现，保持接口兼容）
func (s *Storage) Close() error {
	return nil
}
