# 配置管理模块 (internal/config)

[根目录](../../CLAUDE.md) > [internal](../) > **config**

> 最后更新：2026-01-12 08:43:36

## 变更记录

### 2026-01-12 08:43:36
- 初始化模块文档

---

## 这个模块干啥的

这个模块负责应用的配置管理，主要干这些事：

1. **加载环境变量**：从系统环境变量读取配置
2. **提供默认值**：当环境变量未设置时使用合理的默认值
3. **统一配置入口**：为整个应用提供统一的配置访问点
4. **类型安全**：通过结构体提供类型安全的配置访问

---

## 对外接口

### Config
配置结构体，包含所有应用级别的配置项

```go
type Config struct {
    Port    string
    APIRoot string
    DataDir string
}
```

| 字段 | 类型 | 默认值 | 说明 |
|-----|------|-------|------|
| `Port` | string | `8088` | HTTP 服务监听端口 |
| `APIRoot` | string | `""` | API 路径前缀（会自动去掉尾部斜杠） |
| `DataDir` | string | `./data` | 数据存储目录 |

---

### Load
加载配置（从环境变量读取，没有就用默认值）

```go
func Load() *Config
```

**返回**：`*Config` - 配置结构体指针

**环境变量映射**：

| 环境变量 | 字段 | 默认值 | 备注 |
|---------|------|-------|------|
| `PORT` | `Port` | `8088` | HTTP 监听端口 |
| `API_ROOT` | `APIRoot` | `""` | API 路径前缀 |
| `DATA_DIR` | `DataDir` | `./data` | 数据目录路径 |

**使用示例**：
```go
cfg := config.Load()
fmt.Println(cfg.Port)      // "8088" 或环境变量 PORT 的值
fmt.Println(cfg.APIRoot)   // "" 或环境变量 API_ROOT 的值（已去尾部斜杠）
fmt.Println(cfg.DataDir)   // "./data" 或环境变量 DATA_DIR 的值
```

---

## 配置示例

### 默认配置
```bash
# 不设置任何环境变量，使用默认值
go run cmd/cookiecloud/main.go
# Port: 8088
# APIRoot: ""
# DataDir: ./data
```

### 自定义配置
```bash
# 设置环境变量
export PORT=9000
export API_ROOT=/api/v1
export DATA_DIR=/var/lib/cookiecloud

go run cmd/cookiecloud/main.go
# Port: 9000
# APIRoot: /api/v1
# DataDir: /var/lib/cookiecloud
```

### Docker 部署配置
```bash
docker run -d \
  -p 8088:8088 \
  -v ./data:/data \
  -e PORT=8088 \
  -e API_ROOT=/api \
  782042369/cookiecloud:latest
```

---

## 关键特性

### 自动处理尾部斜杠
`API_ROOT` 会自动去掉尾部斜杠，避免路由重复斜杠问题：

```bash
API_ROOT=/api/     # 实际使用：/api
API_ROOT=/api      # 实际使用：/api
```

### 相对路径支持
`DataDir` 支持相对路径，相对于应用启动时的工作目录：

```bash
DATA_DIR=./data        # 当前目录下的 data 文件夹
DATA_DIR=../data       # 上级目录下的 data 文件夹
DATA_DIR=/var/data     # 绝对路径
```

---

## 代码结构

### 文件组织
```
internal/config/
└── config.go         # 配置管理功能
```

### 类型列表

| 类型名 | 字段 | 说明 |
|-------|------|------|
| `Config` | `Port, APIRoot, DataDir` | 应用配置结构体 |

### 函数列表

| 函数名 | 参数 | 返回值 | 说明 |
|-------|------|--------|------|
| `Load` | 无 | `*Config` | 加载配置（环境变量 + 默认值） |
| `getEnv` | `key, defaultValue string` | `string` | 获取环境变量（私有函数） |

---

## 使用场景

### 在 main.go 中使用
```go
func main() {
    // 加载配置
    cfg := config.Load()

    // 创建 storage 实例
    store, err := storage.New(cfg.DataDir)

    // 启动服务器
    app := fiber.New()
    registerRoutes(app, h, cfg.APIRoot)
    app.Listen(":" + cfg.Port)
}
```

### 在其他模块中使用
通过依赖注入传递配置，而不是直接导入 config 包：

```go
// ❌ 不推荐：直接导入
import "cookiecloud/internal/config"

func SomeFunction() {
    cfg := config.Load()  // 每次都重新加载
}

// ✅ 推荐：依赖注入
func New(dataDir string) *Storage {
    return &Storage{dataDir: dataDir}
}
```

---

## 常见问题

### Q1: 为什么不使用配置文件（如 config.yaml）？
**A**:
- 简单性：环境变量是容器化应用的最佳实践
- 12-Factor App：符合 [The Twelve-Factor App](https://12factor.net/config) 规范
- Docker 友好：Docker 原生支持环境变量注入
- 安全性：敏感信息通过环境变量传递，不写入文件

### Q2: 如何支持更多配置项？
**A**: 在 `Config` 结构体中添加字段，并在 `Load()` 函数中读取：
```go
type Config struct {
    Port      string
    APIRoot   string
    DataDir   string
    Timeout   int    // 新增：超时时间（秒）
    Debug     bool   // 新增：调试模式
}

func Load() *Config {
    return &Config{
        Port:      getEnv("PORT", "8088"),
        APIRoot:   strings.TrimSuffix(getEnv("API_ROOT", ""), "/"),
        DataDir:   getEnv("DATA_DIR", "./data"),
        Timeout:   atoi(getEnv("TIMEOUT", "30")),    // 新增
        Debug:     getEnv("DEBUG", "false") == "true", // 新增
    }
}
```

### Q3: 如何验证配置的有效性？
**A**: 添加验证方法：
```go
func (c *Config) Validate() error {
    if port, err := strconv.Atoi(c.Port); err != nil || port < 1 || port > 65535 {
        return fmt.Errorf("invalid PORT: %s", c.Port)
    }
    if c.DataDir == "" {
        return fmt.Errorf("DATA_DIR cannot be empty")
    }
    return nil
}

// 使用
cfg := config.Load()
if err := cfg.Validate(); err != nil {
    log.Fatalf("配置错误: %v", err)
}
```

### Q4: 如何支持配置热重载？
**A**: 当前实现不支持热重载。如需支持，可以：
1. 监听信号（如 SIGHUP）
2. 重新调用 `config.Load()`
3. 更新全局配置或通知各模块重新加载

---

## 相关文件清单

### 源代码文件
- `internal/config/config.go` - 配置管理功能

### 依赖模块
- `cmd/cookiecloud/main.go` - 应用启动时加载配置

### 相关文档
- [项目根文档](../../CLAUDE.md) - 环境变量配置说明

---

**模块维护者**：782042369
**最后审核**：2026-01-12
**文档版本**：1.0.0
