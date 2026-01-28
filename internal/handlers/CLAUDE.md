# internal/handlers

[根目录](../../CLAUDE.md) > **internal/handlers**

## 模块快照

**职责**：HTTP 路由处理、请求验证、响应格式化

**入口文件**：`handlers.go`（147 行）

**测试文件**：`handlers_test.go`（722 行，包含基准测试）

## 核心处理器

### 1. RootHandler

```go
func FiberRootHandler(apiRoot string) fiber.Handler
```

返回欢迎信息：`Hello World! API ROOT = {apiRoot}`

### 2. UpdateHandler

```go
func (h *Handlers) FiberUpdateHandler(c *fiber.Ctx) error
```

**请求体**：
```json
{
  "uuid": "user-uuid",
  "encrypted": "base64-encoded-data"
}
```

**验证规则**：
- UUID 和 encrypted 字段必填
- UUID 长度 ≤ 256 字符（防止超长字符串攻击）
- encrypted 长度 ≤ 10MB

**副作用**：更新内存缓存

**响应**：`{"action": "done"}`

### 3. GetHandler

```go
func (h *Handlers) FiberGetHandler(c *fiber.Ctx) error
```

**GET 请求**：返回加密数据
```json
{
  "encrypted": "base64-encoded-data"
}
```

**POST 请求**（可选解密）：
```json
{
  "password": "user-password"
}
```

响应：解密后的 JSON 数据或 `{}`（解密失败）

## 数据模型

```go
type UpdateRequest struct {
    Encrypted string `json:"encrypted"`
    UUID      string `json:"uuid"`
}

type DecryptRequest struct {
    Password string `json:"password"`
}
```

## 安全特性

- UUID 长度限制（UTF-8 字符计数）
- 请求体大小限制（10MB）
- 统一错误响应格式：`{"action": "error", "reason": "..."}`
- 结构化日志记录（包含 IP、路径、方法）

## 性能测试

测试文件包含两个基准测试：
- `BenchmarkUpdateHandler` - 更新性能
- `BenchmarkGetHandler` - 获取性能

## 参考文档

- 完整实现：@internal/handlers/handlers.go
- 测试文件：@internal/handlers/handlers_test.go
- 存储层：@internal/storage/CLAUDE.md
- 加密层：@internal/crypto/CLAUDE.md
