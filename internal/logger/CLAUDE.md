# 日志记录模块 (internal/logger)

[根目录](../../CLAUDE.md) > [internal](../) > **logger**

> 最后更新：2026-01-13

## 变更记录

### 2026-01-13
- 初始化模块文档

---

## 这个模块干啥的

这个模块负责结构化的错误日志记录，主要干这些事：

1. **错误日志**：只记录错误级别的日志
2. **结构化格式**：使用 key-value 格式，便于解析和分析
3. **上下文信息**：自动包含文件名和行号
4. **请求日志**：专门处理 HTTP 请求错误日志

---

## 对外接口

### Error
记录错误日志（结构化 key-value 格式）

```go
func Error(msg string, keyvals ...interface{})
```

**参数**：
- `msg`：错误消息
- `keyvals`：可选的 key-value 对（必须是偶数个）

**输出格式**：
```
[ERROR] 2026/01/13 10:30:45 filename.go:123: 错误消息 | key1=value1 | key2=value2
```

**使用示例**：
```go
// 无额外信息
logger.Error("保存数据失败")

// 带 key-value 对
logger.Error("保存数据失败", "uuid", uuid, "error", err)

// HTTP 请求错误
logger.Error("请求处理失败", "path", "/update", "method", "POST", "ip", "127.0.0.1")
```

---

### RequestError
记录 HTTP 请求错误（包含请求上下文）

```go
func RequestError(path, method, ip, msg string, err error)
```

**参数**：
- `path`：请求路径
- `method`：HTTP 方法
- `ip`：客户端 IP
- `msg`：错误消息
- `err`：错误对象

**输出格式**：
```
[ERROR] 2026/01/13 10:30:45 filename.go:123: 错误消息 | path=/update | method=POST | ip=127.0.0.1 | error=错误详情
```

**使用示例**：
```go
if err := c.BodyParser(&req); err != nil {
    logger.RequestError(c.Path(), c.Method(), c.IP(), "JSON 解析失败", err)
    return sendError(c, fiber.StatusBadRequest, "Bad Request")
}
```

---

## 日志格式

### 结构化设计
日志采用结构化格式，所有字段都用 `key=value` 表示，用 ` | ` 分隔：

```
[ERROR] 日期时间 文件:行号: 消息 | key1=value1 | key2=value2 | key3=value3
```

### 自动包含信息
- **时间戳**：自动包含（通过 `log.LstdFlags`）
- **文件名和行号**：自动包含（通过 `log.Lshortfile`）
- **日志级别**：固定为 `[ERROR]`

### 日志输出位置
所有日志输出到 `stderr`（标准错误流），便于：
- Docker 日志收集
- 日志聚合工具（如 ELK）
- 错误监控（如 Sentry）

---

## 代码结构

### 文件组织
```
internal/logger/
├── logger.go         # 日志记录功能
└── logger_test.go    # 测试文件（未实现测试）
```

### 变量列表

| 变量名 | 类型 | 说明 |
|-------|------|------|
| `errorLogger` | `*log.Logger` | 错误日志记录器（写入 stderr） |

### 函数列表

| 函数名 | 可见性 | 说明 |
|-------|-------|------|
| `Error` | ✅ 公开 | 记录错误日志（结构化 key-value 格式） |
| `RequestError` | ✅ 公开 | 记录 HTTP 请求错误（包含请求上下文） |

---

## 设计决策

### 为什么只记录错误日志？
- **简化日志管理**：避免日志过多
- **聚焦关键问题**：错误是最重要的信息
- **性能考虑**：不记录 INFO/DEBUG 级别日志
- **符合 12-Factor**：日志作为事件流

### 为什么使用结构化格式？
- **易于解析**：key-value 格式便于机器解析
- **易于搜索**：可以按 key 搜索
- **易于分析**：日志聚合工具可以直接使用
- **易于调试**：包含完整的上下文信息

### 为什么输出到 stderr？
- **标准实践**：错误日志应该输出到 stderr
- **Docker 友好**：Docker 自动收集 stderr
- **便于重定向**：可以单独重定向 stdout 和 stderr
- **符合惯例**：Unix 工具链的传统

---

## 使用场景

### 在 handlers 中使用
```go
func (h *Handlers) FiberUpdateHandler(c *fiber.Ctx) error {
    var req UpdateRequest
    if err := c.BodyParser(&req); err != nil {
        logger.RequestError(c.Path(), c.Method(), c.IP(), "JSON 解析失败", err)
        return sendError(c, fiber.StatusBadRequest, "Bad Request")
    }

    if err := h.store.SaveEncryptedData(req.UUID, req.Encrypted); err != nil {
        logger.RequestError(c.Path(), c.Method(), c.IP(), "文件写入失败", err)
        return sendError(c, fiber.StatusInternalServerError, "Internal Server Error")
    }

    return c.JSON(fiber.Map{"action": "done"})
}
```

### 在 crypto 中使用
```go
func Decrypt(uuid, encrypted, password string) []byte {
    key := md5String(uuid+"-"+password)[:16]

    decrypted, err := decryptCryptoJsAesMsg(key, encrypted)
    if err != nil {
        logger.Error("解密失败", "uuid", uuid, "error", err)
        return []byte("{}")
    }
    return decrypted
}
```

### 在 storage 中使用
```go
func (s *Storage) SaveEncryptedData(uuid, encrypted string) error {
    // ... 错误处理
    if err != nil {
        logger.Error("保存数据失败", "uuid", uuid, "error", err)
        return err
    }
    return nil
}
```

---

## 测试情况

### 现在啥情况
- ✅ 有测试文件 `logger_test.go`
- ❌ 但没有实现任何测试

### 老王建议这么测
1. **输出验证**：
   - 测试日志格式是否正确
   - 测试 key-value 是否正确分隔
   - 测试时间戳和文件信息是否包含

2. **边界条件**：
   - 测试无 key-value 对的情况
   - 测试奇数个 keyvals（自动补空字符串）
   - 测试特殊字符处理

3. **集成测试**：
   - 测试与 handlers 的集成
   - 测试与 crypto 的集成
   - 测试与 storage 的集成

### 测试示例
```go
func TestError(t *testing.T) {
    // 重定向 stderr 到缓冲区
    old := os.Stderr
    r, w, _ := os.Pipe()
    os.Stderr = w

    logger.Error("测试错误", "key1", "value1", "key2", "value2")

    w.Close()
    os.Stderr = old

    var buf bytes.Buffer
    buf.ReadFrom(r)
    output := buf.String()

    assert.Contains(t, output, "[ERROR]")
    assert.Contains(t, output, "测试错误")
    assert.Contains(t, output, "key1=value1")
    assert.Contains(t, output, "key2=value2")
}
```

---

## 常见问题

### Q1: 为什么不使用成熟的日志库（如 zap、logrus）？
**A**: 小型项目不需要复杂的日志库：
- 标准库的 `log` 包已经够用
- 避免不必要的依赖
- 代码更简单，更容易理解
- 性能已经足够好

### Q2: 如何添加更多日志级别？
**A**: 如果需要，可以添加：
```go
var infoLogger = log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lshortfile)
var warnLogger = log.New(os.Stdout, "[WARN] ", log.LstdFlags|log.Lshortfile)

func Info(msg string, keyvals ...interface{}) {
    // 类似 Error 的实现
}

func Warn(msg string, keyvals ...interface{}) {
    // 类似 Error 的实现
}
```

### Q3: 如何支持日志轮转？
**A**: 使用外部工具：
- **Docker**：Docker 自动处理日志轮转
- **systemd**：systemd 的 journald 自动轮转
- **logrotate**：Linux 标准的日志轮转工具

### Q4: 如何集成日志聚合工具？
**A**: 由于使用结构化格式，可以直接集成：
- **ELK Stack**：Filebeat → Logstash → Elasticsearch
- **Grafana Loki**：Promtail → Loki
- **CloudWatch**：awslogs 驱动
- **Sentry**：sentry-go SDK

---

## 依赖和配置

### 外部依赖
```go
import (
    "bytes"       // 字节缓冲
    "fmt"         // 格式化输出
    "log"         // 标准日志库
    "os"          // 标准输出
)
```

### 配置
- 无需配置文件
- 使用标准库默认配置
- 输出到 stderr

---

## 相关文件清单

### 源代码文件
- `internal/logger/logger.go` - 日志记录功能

### 依赖模块
- `internal/handlers/handlers.go` - 调用 RequestError
- `internal/crypto/crypto.go` - 调用 Error

### 相关文档
- [项目根文档](../../CLAUDE.md) - 日志策略说明

---

## 日志示例

### handlers 模块的日志
```
[ERROR] 2026/01/13 10:30:45 handlers.go:46: JSON 解析失败 | path=/update | method=POST | ip=127.0.0.1 | error=invalid character '<'
[ERROR] 2026/01/13 10:31:20 handlers.go:56: 文件写入失败 | path=/update | method=POST | ip=127.0.0.1 | error=permission denied
```

### crypto 模块的日志
```
[ERROR] 2026/01/13 10:32:10 crypto.go:32: 解密失败 | uuid=user-123 | error=pkcs7: invalid padding
```

### storage 模块的日志
```
[ERROR] 2026/01/13 10:33:05 storage.go:53: 保存数据失败 | uuid=device-abc | error=no space left on device
```

---

**模块维护者**：782042369
**最后审核**：2026-01-13
**文档版本**：1.0.0
