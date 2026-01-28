# internal/storage

[根目录](../../CLAUDE.md) > **internal/storage**

## 模块快照

**职责**：JSON 文件存储，支持并发安全和 context.Context 取消信号

**入口文件**：`storage.go`（99 行）

**测试文件**：`storage_test.go`

## 数据模型

```go
type CookieData struct {
    Encrypted string `json:"encrypted"`
}
```

## 核心功能

### 1. 创建存储实例

```go
store, err := storage.New("./data")
// 自动创建数据目录（权限 0755）
```

### 2. 保存数据

```go
err := store.SaveEncryptedData(ctx, uuid, encrypted)
```

- **文件路径**：`{DATA_DIR}/{uuid}.json`
- **并发安全**：每个 UUID 使用独立的 Mutex（sync.Map）
- **取消支持**：检查 `ctx.Done()`，文件操作前可中断

### 3. 加载数据

```go
data, err := store.LoadEncryptedData(ctx, uuid)
```

- 返回 `*CookieData` 或错误
- 支持 context 取消

### 4. 关闭存储

```go
store.Close() // 空实现，保持接口兼容
```

## 并发安全

使用全局 `sync.Map` 存储文件锁：

```go
var fileLocks sync.Map

func getFileLock(uuid string) *sync.Mutex {
    lock, _ := fileLocks.LoadOrStore(uuid, &sync.Mutex{})
    return lock.(*sync.Mutex)
}
```

每个 UUID 的文件读写互斥，不同 UUID 并行无冲突。

## Context 取消示例

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := store.SaveEncryptedData(ctx, uuid, data)
if errors.Is(err, context.Canceled) {
    // 操作被取消
}
```

## 参考文档

- 完整实现：@internal/storage/storage.go
- 测试文件：@internal/storage/storage_test.go
