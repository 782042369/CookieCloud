# CookieCloud Go 版本 - 项目文档

> 最后更新：2026-01-12 09:24:23

## 变更记录

### 2026-01-12 09:24:23
- 重新生成项目架构文档
- 更新测试覆盖情况和代码质量工具信息
- 添加 golangci-lint 配置说明
- 标注为小型项目，明确不需要集成测试
- 移除 Mermaid 结构图，改用文本描述

### 2026-01-12 09:04:25
- 重新初始化项目架构文档
- 更新测试覆盖情况（72.6% 覆盖率，60+ 测试用例）
- 添加完整的测试文件信息
- 更新模块结构和依赖关系

### 2026-01-12 08:43:36
- 更新项目架构文档
- 新增 config 模块文档
- 完善依赖注入架构说明
- 更新模块结构图和索引

### 2026-01-11 16:25:37
- 初始化项目架构文档
- 生成项目结构图和模块索引
- 创建各模块的详细文档

---

## 这个项目是干啥的

CookieCloud 原本是 easychen 用 Node.js 写的一个 Cookie 和 LocalStorage 同步工具，老王我用 Go 把服务端重写了一遍。为啥要重写？因为 Go 版本有几个明显的好处：

- **镜像小**：用 scratch 基础镜像构建，最终镜像就几MB
- **性能好**：Go 运行效率高，内存占用低，省资源
- **架构简单**：模块化设计，各干各的，不乱套
- **数据安全**：客户端加密后再存到服务端，老王我只管存，看不到明文
- **好集成**：提供 REST API，浏览器插件直接调就行
- **质量高**：完整的单元测试覆盖（63个测试用例，72.6%覆盖率），代码质量检查完善

### 项目特点
- **小型项目**：功能单一，代码简洁，不需要复杂的集成测试
- **零知识架构**：服务端只存储加密数据，不参与加密解密（除按需解密外）
- **依赖注入设计**：通过构造函数注入依赖，零全局变量，易于测试
- **并发安全**：使用 sync.Map 实现文件锁，支持高并发
- **完整测试**：所有核心模块都有单元测试，覆盖率72.6%
- **代码质量**：配置 golangci-lint，所有包都有 package 注释

---

## 技术架构

### 用了啥技术
- **语言**：Go 1.25.5
- **Web 框架**：Fiber v2.52.10（基于 fasthttp，性能挺猛）
- **加密算法**：AES-256-CBC（和 CryptoJS 兼容，这个很重要）
- **数据存储**：JSON 文件（简单直接，不需要数据库）
- **容器化**：Docker 多阶段构建（golang:alpine → scratch）
- **代码质量**：golangci-lint（gofmt、govet、staticcheck 等）

### 代码咋组织的
- **分层结构**：Handler 接收请求 → Storage 存数据 → 文件系统
- **模块化**：internal 包下面按功能分，职责清晰
- **配置**：用环境变量控制端口和 API 路径
- **依赖注入**：通过构造函数传递依赖，零全局变量

### 数据咋流转的
```
浏览器插件 → 加密数据 → API 接口 → 验证 → 存文件（JSON）
                                    ↓
浏览器插件 → 请求数据 → 验证 → 读文件 → 解密（可选） → 返回
```

---

## 代码结构

```
CookieCloud/
├── cmd/
│   └── cookiecloud/
│       └── main.go           # 应用入口
├── internal/
│   ├── config/
│   │   └── config.go         # 配置管理
│   ├── handlers/
│   │   └── handlers.go       # HTTP 请求处理
│   ├── storage/
│   │   └── storage.go        # 数据存储
│   └── crypto/
│       └── crypto.go         # 加密解密
├── .golangci.yml              # 代码质量检查配置
├── Dockerfile                 # Docker 构建配置
└── go.mod                     # Go 模块依赖
```

---

## 各模块都在干啥

| 模块路径 | 干啥的 | 入口文件 | 有没有测试 | 测试覆盖 | 文档链接 |
|---------|------|---------|---------|---------|---------|
| `cmd/cookiecloud` | 启动应用，注册路由 | `main.go` | ❌ 小型项目不需要集成测试 | - | [查看](./cmd/cookiecloud/CLAUDE.md) |
| `internal/config` | 配置管理（环境变量） | `config.go` | ✅ 6个测试 | 100% | [查看](./internal/config/CLAUDE.md) |
| `internal/handlers` | 处理 HTTP 请求 | `handlers.go` | ✅ 23个测试 | 87.9% | [查看](./internal/handlers/CLAUDE.md) |
| `internal/storage` | 存数据、读数据 | `storage.go` | ✅ 20个测试 | 91.7% | [查看](./internal/storage/CLAUDE.md) |
| `internal/crypto` | 加密解密 | `crypto.go` | ✅ 16个测试 | 93.5% | [查看](./internal/crypto/CLAUDE.md) |

---

## 怎么跑起来

### 本地开发

```bash
# 1. 克隆代码
git clone https://github.com/782042369/cookiecloud.git
cd cookiecloud

# 2. 直接跑
go run cmd/cookiecloud/main.go

# 3. 访问服务
# 默认地址：http://localhost:8088
```

### 环境变量配置

| 变量名 | 默认值 | 干啥的 |
|-------|-------|------|
| `PORT` | `8088` | 监听端口 |
| `API_ROOT` | `""` | API 路径前缀（比如 `/api`） |
| `DATA_DIR` | `./data` | 数据存储目录 |

### Docker 部署

```bash
# 用 Docker Hub 镜像（推荐）
docker run -d \
  -p 8088:8088 \
  -v ./data:/data \
  -e PORT=8088 \
  782042369/cookiecloud:latest

# 或者本地构建
docker build -t cookiecloud:latest .
docker run -d -p 8088:8088 -v ./data:/data cookiecloud:latest
```

### 怎么构建

```bash
# 本地构建二进制
go build -o cookiecloud ./cmd/cookiecloud

# 构建 Docker 镜像
docker build -t cookiecloud:latest .

# 交叉编译（Linux）
GOOS=linux GOARCH=amd64 go build -o cookiecloud-linux ./cmd/cookiecloud
```

---

## API 接口说明

### 1. 根路径（看看服务活没活）
```http
GET/POST http://localhost:8088/
响应: "Hello World! API ROOT = /api"
```

### 2. 更新数据（保存加密的 Cookie）
```http
POST http://localhost:8088/update
Content-Type: application/json

{
  "uuid": "user-device-uuid",
  "encrypted": "base64-encoded-encrypted-data"
}

响应:
{
  "action": "done"
}
```

### 3. 获取数据（读取 Cookie）
```http
# 获取加密数据
GET http://localhost:8088/get/:uuid

响应（加密格式）:
{
  "encrypted": "base64-encoded-encrypted-data"
}

# 获取解密数据
POST http://localhost:8088/get/:uuid
Content-Type: application/json

{
  "password": "user-password"
}

响应（解密后的原始数据）:
{ /* Cookie 数据 */ }
```

---

## 测试相关

### 测试现状（小型项目，专注单元测试）

**测试覆盖情况**：
- **总测试数**：63个
- **总体覆盖率**：72.6%
- **测试类型**：单元测试、并发测试、边界条件测试、性能基准测试

**各模块测试统计**：
| 模块 | 测试数 | 覆盖率 | 测试文件 |
|-----|-------|--------|---------|
| config | 6 | 100% | `config_test.go` |
| crypto | 16 | 93.5% | `crypto_test.go` |
| storage | 20 | 91.7% | `storage_test.go` |
| handlers | 23 | 87.9% | `handlers_test.go` |
| **总计** | **63** | **72.6%** | - |

**测试亮点**：
- 并发安全测试（100个并发写入）
- 性能基准测试（6个 Benchmark）
- 边界条件测试（空数据、超长数据、特殊字符）
- 错误处理测试（无效JSON、文件不存在）

### 怎么跑测试

```bash
# 跑所有测试
go test ./...

# 跑测试并看覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 跑性能基准测试
go test -bench=. -benchmem ./...

# 跑特定模块的测试
go test ./internal/crypto -v

# 详细输出（显示每个测试的详细信息）
go test -v ./...

# 只跑特定测试
go test -run TestNew ./internal/storage
```

### 小型项目说明

这是一个**小型项目**，特点是：
- 功能单一：只提供 Cookie 同步服务
- 代码简洁：核心代码不到500行
- 依赖少：只用了一个外部依赖（Fiber）
- 易于测试：所有模块都可独立测试

**测试策略**：
- ✅ **单元测试**：覆盖所有核心模块（config、crypto、storage、handlers）
- ✅ **并发测试**：验证文件锁和请求处理的并发安全性
- ✅ **基准测试**：测试关键路径的性能
- ❌ **集成测试**：小型项目不需要，单元测试已足够
- ❌ **E2E测试**：小型项目不需要，手动测试即可

---

## 代码质量工具

### golangci-lint 配置

项目已配置 golangci-lint，配置文件：`.golangci.yml`

**启用的检查器**：
- `gofmt` - 代码格式化检查
- `goimports` - import 导入排序
- `govet` - Go 静态分析
- `staticcheck` - 高级静态分析
- `ineffassign` - 无效赋值检查
- `misspell` - 拼写错误检查
- `revive` - 代码风格检查
- `gocyclo` - 圈复杂度检查（阈值15）
- `funlen` - 函数长度检查（100行/50语句）
- `errcheck` - 错误处理检查
- `prealloc` - 预分配切片检查
- `unconvert` - 冗余转换检查
- `unused` - 未使用代码检查

**使用方法**：
```bash
# 安装 golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# 运行检查
golangci-lint run

# 在 CI 中运行
golangci-lint run --out-format=github-actions
```

### 代码格式化

```bash
# 格式化所有代码
gofmt -w -s .

# 检查格式是否符合规范
gofmt -l .
```

### Package 注释

所有包都有 package 注释：
- `Package main 是 CookieCloud 应用的入口`
- `Package config 提供配置管理功能`
- `Package handlers 提供 HTTP 请求处理功能`
- `Package storage 提供数据持久化功能`
- `Package crypto 提供加密解密功能`

---

## 代码规范

### Go 代码怎么写
- 照着 [Effective Go](https://go.dev/doc/effective_go) 来
- 用 `gofmt` 格式化代码（别自己瞎缩进）
- 用 `golangci-lint` 检查代码
- 导出的函数和类型都要写注释

### 命名规则
- **包名**：小写单词，别用下划线或驼峰
- **文件名**：小写，用下划线分隔（比如 `handlers.go`）
- **常量**：大写驼峰（比如 `DataDir`）
- **变量/函数**：
  - 要导出的：大写驼峰（比如 `SaveEncryptedData`）
  - 私有的：小写驼峰（比如 `sendErrorResponse`）

### 注释怎么写
```go
// PackageName 包是干啥的
package packagename

// FunctionName 函数是干啥的（必须以函数名开头）
// 详细说明（想写就写）
func FunctionName(param1 type) returnType {
    // ...
}
```

---

## Git 提交规范

### 分支咋用
- `master`：主分支，保持稳定，别瞎搞
- `feature/*`：开发新功能的分支
- `bugfix/*`：修 bug 的分支

### 提交消息咋写
照着 [Conventional Commits](https://www.conventionalcommits.org/) 来：

```
<类型>(<范围>): <干了啥>

<详细说明（可选）>
```

**类型**：
- `feat`：新功能
- `fix`：修 bug
- `refactor`：重构代码
- `docs`：更新文档
- `test`：测试相关
- `chore`：构建工具、依赖更新

**举个例子**：
```
feat(handlers): 添加请求验证中间件

- 添加 UUID 格式验证
- 添加加密数据长度检查
- 统一错误响应格式

Closes #123
```

---

## 跟 AI 配合开发的技巧

### 啥时候适合让 AI 帮忙
1. **重构代码**：优化代码结构和性能
2. **写测试**：给现有模块生成单元测试
3. **更新文档**：代码改了之后更新文档
4. **修 bug**：分析日志和错误信息，找问题
5. **加功能**：添加新的 API 端点或中间件

### 提示词模板

**理解代码**：
```
给老王我讲讲 CookieCloud 里加密解密的完整流程，
包括密钥咋生成的、用啥加密算法、数据咋存的。
```

**加功能**：
```
在 internal/handlers 里加一个新的 API 接口：
- 路径：DELETE /api/delete/:uuid
- 功能：删除指定 UUID 的 Cookie 数据
- 要有错误处理和响应
```

**写测试**：
```
给 internal/crypto 包写单元测试，
要覆盖这些场景：
1. 正常的加密解密流程
2. 密码错了怎么办
3. 密文格式不对怎么办
4. 边界条件测试
```

**性能优化**：
```
看看 storage 模块的文件读写实现，
给老王我提点性能优化建议，特别是高并发的时候。
```

### 项目关键信息
- **项目类型**：小型 REST API 服务（不需要集成测试）
- **主要框架**：Fiber（高性能 Web 框架）
- **数据存储**：JSON 文件（不用数据库）
- **加密兼容性**：必须和 CryptoJS 兼容（这个很关键）
- **部署环境**：Docker 容器，scratch 基础镜像
- **代码质量**：golangci-lint + 完整单元测试

---

## 相关资源

- **原项目**：[easychen/CookieCloud](https://github.com/easychen/CookieCloud)
- **浏览器插件**：从原项目获取
- **Fiber 文档**：[https://docs.gofiber.io/](https://docs.gofiber.io/)
- **Go 官方文档**：[https://go.dev/doc/](https://go.dev/doc/)

---

## 扫描覆盖情况

### 扫描统计（2026-01-12 09:24:23）
- **总文件数**：20 个
- **源代码文件**：9 个
- **测试文件**：4 个
- **已扫描**：9 个（100%）
- **忽略文件**：11 个（.git、data 目录等）

### 各模块扫描情况
| 模块 | 源文件数 | 测试数 | 已扫描 | 覆盖率 | 状态 |
|-----|---------|-------|--------|--------|------|
| cmd/cookiecloud | 1 | 0 | 1 | 100% | ✅ 搞定 |
| internal/config | 1 | 6 | 1 | 100% | ✅ 搞定 |
| internal/handlers | 1 | 23 | 1 | 100% | ✅ 搞定 |
| internal/storage | 1 | 20 | 1 | 100% | ✅ 搞定 |
| internal/crypto | 1 | 16 | 1 | 100% | ✅ 搞定 |

### 主要缺啥（已改善）
- ✅ ~~缺少测试~~ → 已添加完整的单元测试（63个测试用例）
- ✅ ~~缺少代码质量检查~~ → 已配置 golangci-lint
- ✅ ~~缺少 package 注释~~ → 所有包都有注释
- ✅ ~~代码格式化~~ → 已通过 gofmt 格式化
- ⚠️ 缺少配置示例（`.env.example`）
- ⚠️ 缺少 Makefile

### 忽略的文件
- `.git/`：Git 自己的文件
- `data/`：数据存储目录
- `coverage*.out`、`coverage*.html`：测试覆盖率文件
- `.golangci/`：golangci-lint 缓存

### 老王建议下一步干啥
1. **有空就搞**：
   - 创建 `.env.example` 配置示例
   - 添加 `Makefile` 简化开发流程

2. **以后再说**：
   - 添加 API 文档（OpenAPI 规范）
   - 扩展性能基准测试

---

**文档生成**：Claude AI 架构助手
**生成时间**：2026-01-12 09:24:23
**项目版本**：基于 master 分支
**测试覆盖**：63个测试用例，72.6%覆盖率
**代码质量**：golangci-lint 配置完善
