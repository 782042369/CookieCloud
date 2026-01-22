# ============================================
# 极简版 Dockerfile - Scratch 基础镜像
# 目标镜像大小：4-5MB
# ============================================
# 优化措施：
# 1. Go 二进制：UPX 压缩（减少60%）
# 2. 基础镜像：Scratch 空镜像
# ============================================

# 阶段一：构建 service (Go版本)
FROM golang:1.25-alpine AS service-builder
WORKDIR /app

LABEL stage="service-builder"

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN echo "📦 下载 Go 依赖..." && \
    go mod download && \
    echo "✅ 验证依赖完成" && \
    go mod verify

# 复制源代码
COPY . .

# 安装 UPX 压缩工具
RUN echo "🔧 安装 UPX 压缩工具..." && \
    apk add --no-cache upx

# 构建完全静态的Go应用（极致优化）
RUN echo "🔨 开始构建 Go 应用..." && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
    -ldflags="-s -w -extldflags '-static' -buildid=" \
    -trimpath \
    -o main ./cmd/cookiecloud && \
    chmod +x main && \
    echo "✅ Go 应用构建完成"

# 使用 UPX 压缩二进制文件（减少50-70%体积）
RUN echo "🗜️  使用 UPX 压缩二进制文件..." && \
    upx --best --lzma main && \
    echo "✅ UPX 压缩完成"

# ============================================
# 最终生产阶段：使用 Scratch（空镜像）
# ============================================
FROM scratch

# 设置工作目录
WORKDIR /app

LABEL stage="production"

# 从 service-builder 阶段复制 Go 二进制
COPY --from=service-builder /app/main ./main

# 设置环境变量（时区默认为中国）
ENV PORT=8088
ENV TZ=Asia/Shanghai

# 声明端口
EXPOSE 8088

# ============================================
# 注意：Scratch 镜像不包含 shell，因此：
# - 无法使用 HEALTHCHECK（没有 wget/curl）
# - 无法进入容器调试（没有 sh/bash）
# - 推荐使用外部健康检查（如 Kubernetes livenessProbe）
# ============================================

CMD ["./main"]
