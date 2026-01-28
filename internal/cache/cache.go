// Package cache 提供内存缓存功能
// 使用 sync.Map 实现线程安全的内存缓存，减少磁盘IO
package cache

import (
	"sync"
	"time"
)

// item 缓存项（内部使用）
type item struct {
	data      string
	expiresAt time.Time
}

// Cache 内存缓存（使用 sync.Map 实现线程安全）
type Cache struct {
	items sync.Map
	ttl   time.Duration // 缓存过期时间
}

// New 创建一个新的缓存实例
func New(ttl time.Duration) *Cache {
	return &Cache{
		ttl: ttl,
	}
}

// Set 设置缓存
func (c *Cache) Set(uuid, encrypted string) {
	c.items.Store(uuid, &item{
		data:      encrypted,
		expiresAt: time.Now().Add(c.ttl),
	})
}

// Get 获取缓存
func (c *Cache) Get(uuid string) (string, bool) {
	value, ok := c.items.Load(uuid)
	if !ok {
		return "", false
	}

	it := value.(*item)

	// 检查是否过期
	if time.Now().After(it.expiresAt) {
		c.items.Delete(uuid)
		return "", false
	}

	return it.data, true
}

// Delete 删除缓存
func (c *Cache) Delete(uuid string) {
	c.items.Delete(uuid)
}

// Clear 清空所有缓存（遍历删除每个key）
func (c *Cache) Clear() {
	c.items.Range(func(key, _ interface{}) bool {
		c.items.Delete(key)
		return true
	})
}

// CleanExpired 清理过期的缓存项
func (c *Cache) CleanExpired() {
	now := time.Now()
	c.items.Range(func(key, value interface{}) bool {
		it := value.(*item)
		if now.After(it.expiresAt) {
			c.items.Delete(key)
		}
		return true
	})
}

// Size 返回缓存项数量
func (c *Cache) Size() int {
	count := 0
	c.items.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
