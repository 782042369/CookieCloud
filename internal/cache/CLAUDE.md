# internal/cache

[根目录](../../CLAUDE.md) > **internal/cache**

## 模块快照

**职责**：内存缓存，使用 sync.Map 实现线程安全，支持 TTL 过期

**入口文件**：`cache.go`（86 行）

**测试文件**：`cache_test.go`

## 核心功能

### 1. 创建缓存实例

```go
cache := cache.New(5 * time.Minute) // TTL 5 分钟
```

### 2. 设置缓存

```go
cache.Set(uuid, encrypted)
```

- 自动记录过期时间（当前时间 + TTL）

### 3. 获取缓存

```go
encrypted, found := cache.Get(uuid)
if !found {
    // 缓存未命中，从存储加载
}
```

- 自动检查过期，过期返回 `found=false`

### 4. 删除缓存

```go
cache.Delete(uuid)
```

### 5. 清空所有缓存

```go
cache.Clear()
```

### 6. 清理过期缓存

```go
cache.CleanExpired() // 主动清理所有过期项
```

### 7. 统计缓存数量

```go
size := cache.Size()
```

## 数据结构

```go
type item struct {
    data      string
    expiresAt time.Time
}

type Cache struct {
    items sync.Map
    ttl   time.Duration
}
```

## 并发安全

使用 `sync.Map` 实现无锁并发读写：
- 多个 goroutine 可同时读写不同 UUID
- 无需显式加锁

## 性能优化

- 缓存命中：0 次磁盘 I/O
- TTL 过期：惰性删除（读取时检查）+ 主动清理
- 适用于读多写少场景

## 参考文档

- 完整实现：@internal/cache/cache.go
- 测试文件：@internal/cache/cache_test.go
- 使用示例：@internal/handlers/handlers.go
