# 数据存储模块 (internal/storage)

[根目录](../../CLAUDE.md) > [internal](../) > **storage**

> 最后更新：2026-01-11 16:25:37

## 变更记录

### 2026-01-11 16:25:37
- 初始化模块文档

---

## 这个模块干啥的

这个模块负责数据的持久化存储，主要干这些事：

1. **管理数据目录**：创建和管理数据存储目录
2. **保存数据**：把加密数据保存到 JSON 文件
3. **读取数据**：从 JSON 文件读取加密数据
4. **数据结构**：定义 Cookie 数据的存储格式
5. **并发控制**：用文件锁防止并发写入冲突

---

## 对外接口

### InitDataDir
初始化数据存储目录

```go
func InitDataDir(dir string) error
```

**功能**：检查数据目录是否存在，不存在就创建

**返回**：
- `nil`：目录已存在或创建成功
- `error`：目录创建失败

**调用时机**：应用启动时（`main.go` 中调用）

---

### SaveEncryptedData
保存加密数据到文件

```go
func SaveEncryptedData(uuid, encrypted string) error
```

**参数**：
- `uuid`：用户设备的唯一标识符
- `encrypted`：Base64 编码的加密数据

**返回**：
- `nil`：保存成功
- `error`：JSON 序列化失败或文件写入失败

**文件路径**：`./data/{uuid}.json`

**文件格式**：
```json
{
  "encrypted": "base64-encoded-data"
}
```

**并发安全**：使用文件锁，同一个 UUID 的并发写入不会冲突

---

### LoadEncryptedData
从文件加载加密数据

```go
func LoadEncryptedData(uuid string) (*CookieData, error)
```

**参数**：
- `uuid`：用户设备的唯一标识符

**返回**：
- `*CookieData`：加密数据结构体指针
- `error`：文件不存在、读取失败或 JSON 解析失败

**文件路径**：`./data/{uuid}.json`

---

## 数据模型

### CookieData
存储在文件中的数据结构

```go
type CookieData struct {
    Encrypted string `json:"encrypted"`
}
```

| 字段 | 类型 | JSON 字段 | 说明 |
|-----|------|----------|------|
| `Encrypted` | string | `encrypted` | Base64 编码的加密数据 |

### 存储示例

**文件名**：`user-device-123.json`

**文件内容**：
```json
{
  "encrypted": "U2FsdGVkX1+vupppZksvRf5pq5g5XjFRlipRkwB0K1Y96Qsv2Lm+31cmzaAILwytJHoXyYkvN2..."
}
```

---

## 存储架构

### 目录结构
```
CookieCloud/
├── data/                    # 数据存储目录
│   ├── user-001.json       # 用户 1 的数据
│   ├── user-002.json       # 用户 2 的数据
│   └── device-abc.json     # 设备的数据
├── cmd/
└── internal/
```

### 存储策略
- **一 UUID 一文件**：每个 UUID 对应一个独立的 JSON 文件
- **覆盖式写入**：每次更新都会覆盖整个文件内容
- **无版本控制**：不保留历史版本
- **文件锁保护**：同一个 UUID 的并发写入有锁保护

### 数据流
```mermaid
graph LR
    A[Handler] -->|保存| B[SaveEncryptedData]
    B -->|序列化| C[JSON Marshal]
    C -->|写入| D[{uuid}.json]

    A -->|加载| E[LoadEncryptedData]
    E -->|读取| D
    D -->|反序列化| F[JSON Unmarshal]
    F -->|返回| A
```

---

## 关键依赖与配置

### 外部依赖
```go
import (
    "encoding/json"  // JSON 序列化/反序列化
    "os"             # 文件系统操作
    "path/filepath"  # 路径处理
    "sync"           # 并发控制
)
```

### 配置变量
```go
var (
    dataDir = "./data"  // 数据存储目录
    fileLocks sync.Map  // 文件锁映射
)
```

### 文件权限
- 目录权限：`0755`（rwxr-xr-x）
- 文件权限：`0644`（rw-r--r--）

---

## 代码结构

### 文件组织
```
internal/storage/
└── storage.go       # 数据存储功能
```

### 常量列表
（无常量，使用变量）

### 类型列表

| 类型名 | 字段 | 说明 |
|-------|------|------|
| `CookieData` | `Encrypted string` | Cookie 数据结构 |

### 函数列表

| 函数名 | 参数 | 返回值 | 说明 |
|-------|------|--------|------|
| `InitDataDir` | `dir string` | `error` | 初始化数据目录 |
| `SaveEncryptedData` | `uuid, encrypted string` | `error` | 保存加密数据 |
| `LoadEncryptedData` | `uuid string` | `(*CookieData, error)` | 加载加密数据 |
| `getFileLock` | `uuid string` | `*sync.Mutex` | 获取文件锁（私有） |

---

## 常见问题

### Q1: 数据存储在磁盘上的格式是什么？
**A**: 每个用户的数据存储为独立的 JSON 文件，文件名为 `{uuid}.json`，内容为：
```json
{"encrypted":"base64-encoded-data"}
```

### Q2: 如何备份数据？
**A**: 直接复制 `./data` 目录即可：
```bash
cp -r data data.backup
```

### Q3: 如何迁移到数据库？
**A**: 可以添加数据库实现层，保持接口一致：
```go
type Storage interface {
    SaveEncryptedData(uuid, encrypted string) error
    LoadEncryptedData(uuid string) (*CookieData, error)
}

// JSON 文件实现
type JSONFileStorage struct{}

// SQLite 实现
type SQLiteStorage struct{}
```

### Q4: 支持删除数据吗？
**A**: 当前未实现删除功能。可以添加：
```go
func DeleteEncryptedData(uuid string) error {
    filePath := filepath.Join(dataDir, uuid+".json")
    return os.Remove(filePath)
}
```

### Q5: 文件并发写入会有问题吗？
**A**: 不会。已经实现了文件锁，同一个 UUID 的并发写入会排队执行，不会冲突。

---

## 并发安全

### 文件锁机制
```go
var fileLocks sync.Map

func getFileLock(uuid string) *sync.Mutex {
    lock, _ := fileLocks.LoadOrStore(uuid, &sync.Mutex{})
    return lock.(*sync.Mutex)
}

func SaveEncryptedData(uuid, encrypted string) error {
    lock := getFileLock(uuid)
    lock.Lock()
    defer lock.Unlock()
    // 原有逻辑...
}
```

**工作原理**：
- 每个 UUID 有独立的锁
- 同一个 UUID 的并发写入会按顺序执行
- 不同 UUID 的写入互不影响

---

## 相关文件清单

### 源代码文件
- `internal/storage/storage.go` - 数据存储功能

### 依赖模块
- `cmd/cookiecloud/main.go` - 应用启动时调用 InitDataDir
- `internal/handlers/handlers.go` - 请求处理时调用保存/加载函数

### 数据目录
- `./data/*.json` - 用户数据文件（运行时生成）

---

**模块维护者**：782042369
**最后审核**：2026-01-11
**文档版本**：1.0.0
