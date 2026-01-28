# internal/config

[根目录](../../CLAUDE.md) > **internal/config**

## 模块快照

**职责**：环境变量配置管理，支持默认值

**入口文件**：`config.go`（33 行）

**测试文件**：`config_test.go`

## 配置项

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `PORT` | `8088` | HTTP 服务端口 |
| `API_ROOT` | `` | API 路径前缀（自动去除尾部斜杠） |
| `DATA_DIR` | `./data` | 数据存储目录 |

## 数据结构

```go
type Config struct {
    Port    string
    APIRoot string
    DataDir string
}
```

## 使用示例

```go
cfg := config.Load()
logger.Info("服务启动",
    "port", cfg.Port,
    "api_root", cfg.APIRoot,
    "data_dir", cfg.DataDir)
```

## 参考文档

- 完整实现：@internal/config/config.go
- 测试文件：@internal/config/config_test.go
