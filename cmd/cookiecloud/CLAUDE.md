# cmd/cookiecloud

[根目录](../../CLAUDE.md) > **cmd/cookiecloud**

## 模块快照

**职责**：应用入口，初始化 Web 服务器、注册路由、启动 HTTP 服务

**入口文件**：`main.go`（115 行）

**关键依赖**：
- `internal/config` - 配置加载
- `internal/handlers` - HTTP 处理器
- `internal/storage` - 数据存储
- `internal/cache` - 内存缓存
- `internal/logger` - 日志记录

## 核心功能

### 服务器配置

- 端口：默认 8088（环境变量 `PORT`）
- API 根路径：默认 `/`（环境变量 `API_ROOT`）
- 请求体大小限制：11MB
- 读写超时：30 秒

### 中间件

1. **CORS**：允许所有来源（小型项目）
2. **速率限制**：60 次/分钟/IP
3. **压缩**：Best Speed 级别

### 路由注册

```
GET/POST  {API_ROOT}/           -> 欢迎接口
POST      {API_ROOT}/update     -> 更新数据
GET/POST  {API_ROOT}/get/:uuid  -> 获取数据
```

### 优雅关闭

- 监听 `SIGINT` 和 `SIGTERM` 信号
- 30 秒超时关闭服务器
- 关闭存储连接

## 参考文档

- 完整实现：@cmd/cookiecloud/main.go
- 配置说明：@internal/config/CLAUDE.md
- 处理器文档：@internal/handlers/CLAUDE.md
