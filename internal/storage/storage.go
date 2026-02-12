// Package storage 提供数据持久化功能
// 使用 JSON 文件存储加密数据，支持并发安全的文件读写
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// CookieData Cookie数据结构
type CookieData struct {
	Encrypted string `json:"encrypted"`
}

// fileLocks 全局文件锁（按 UUID 维度隔离，减少不同文件之间的锁竞争）
var fileLocks sync.Map

// getFileLock 获取指定UUID的文件锁
func getFileLock(uuid string) *sync.RWMutex {
	lock, _ := fileLocks.LoadOrStore(uuid, &sync.RWMutex{})
	return lock.(*sync.RWMutex)
}

// Storage 数据存储管理器
type Storage struct {
	dataDir string // 数据目录路径
}

// New 创建一个新的 Storage 实例
func New(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("无法创建数据目录 %s: %w", dataDir, err)
	}
	return &Storage{dataDir: dataDir}, nil
}

// checkContext 检查 context 是否已取消
func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// resolveFilePath 解析并校验 UUID 对应的存储文件路径，避免路径穿越
func (s *Storage) resolveFilePath(uuid string) (string, error) {
	if uuid == "" || uuid == "." || uuid == ".." {
		return "", fmt.Errorf("invalid uuid: %q", uuid)
	}
	// 检查路径分隔符，防止路径穿越（跨平台：/ 和 \）
	if strings.ContainsAny(uuid, "/\\") {
		return "", fmt.Errorf("invalid uuid contains path separator: %q", uuid)
	}

	fullPath := filepath.Join(s.dataDir, uuid+".json")
	cleanPath := filepath.Clean(fullPath)
	// 确保清理后的路径仍在 dataDir 内
	if !strings.HasPrefix(cleanPath, filepath.Clean(s.dataDir)+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal detected: uuid %q resolves outside data directory", uuid)
	}
	return cleanPath, nil
}

// SaveEncryptedData 保存加密数据到指定 UUID 的文件中
func (s *Storage) SaveEncryptedData(ctx context.Context, uuid, encrypted string) error {
	// 优先检查 context，避免不必要的锁竞争
	if err := checkContext(ctx); err != nil {
		return err
	}

	filePath, err := s.resolveFilePath(uuid)
	if err != nil {
		return err
	}

	lock := getFileLock(uuid)
	lock.Lock()
	defer lock.Unlock()

	// 获取锁后再次检查 context
	if err := checkContext(ctx); err != nil {
		return err
	}

	content, err := json.Marshal(CookieData{Encrypted: encrypted})
	if err != nil {
		return fmt.Errorf("marshal cookie data: %w", err)
	}

	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// LoadEncryptedData 从指定 UUID 的文件中加载加密数据
func (s *Storage) LoadEncryptedData(ctx context.Context, uuid string) (*CookieData, error) {
	// 优先检查 context，避免不必要的文件操作
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	filePath, err := s.resolveFilePath(uuid)
	if err != nil {
		return nil, err
	}

	lock := getFileLock(uuid)
	lock.RLock()
	defer lock.RUnlock()

	// 获取锁后再次检查 context
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

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
