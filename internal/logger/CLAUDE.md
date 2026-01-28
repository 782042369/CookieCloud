# internal/logger

[根目录](../../CLAUDE.md) > **internal/logger**

## 模块快照

**职责**：结构化日志记录，支持 INFO/WARN/ERROR 三个级别

**入口文件**：`logger.go`（62 行）

**测试文件**：`logger_test.go`

## 日志级别

### 1. Info

```go
logger.Info("服务启动", "port", "8088", "api_root", "/api")
```

输出：
```
[INFO] 2025/01/28 12:44:21 main.go:27: 服务启动 | port=8088 | api_root=/api
```

### 2. Warn

```go
logger.Warn("速率限制触发", "ip", c.IP(), "path", c.Path())
```

输出：
```
[WARN] 2025/01/28 12:44:21 main.go:68: 速率限制触发 | ip=127.0.0.1 | path=/api/update
```

### 3. Error

```go
logger.Error("文件写入失败", "uuid", uuid, "error", err)
```

输出：
```
[ERROR] 2025/01/28 12:44:21 handlers.go:83: 文件写入失败 | uuid=test-uuid | error=...
```

### 4. RequestError（专用 HTTP 错误）

```go
logger.RequestError(c.Path(), c.Method(), c.IP(), "JSON解析失败", err)
```

自动包含请求上下文（路径、方法、IP）。

## 日志格式

**结构化 key-value 格式**：
```
[LEVEL] YYYY/MM/DD HH:MM:SS file:line: message | key1=value1 | key2=value2
```

**特点**：
- 标准库 `log` 包（无外部依赖）
- 文件名和行号（`log.Lshortfile`）
- 时间戳（`log.LstdFlags`）
- 键值对便于解析

## 输出目标

- **INFO/WARN**：标准输出（`os.Stdout`）
- **ERROR**：标准错误（`os.Stderr`）

## 参考文档

- 完整实现：@internal/logger/logger.go
- 测试文件：@internal/logger/logger_test.go
- 使用示例：@internal/handlers/handlers.go
