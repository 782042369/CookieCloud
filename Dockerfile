# 构建阶段：使用多阶段构建最小化生产镜像
FROM golang:1.25-alpine as builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖（利用Docker缓存层）
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# 复制源代码
COPY . .

# 构建优化的Go应用
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -o main ./cmd/cookiecloud

# 最终生产阶段
FROM alpine:latest

# 安装最小运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder --chown=appuser:appgroup /app/main ./main

# 创建数据目录并设置权限
RUN mkdir -p ./data && chown appuser:appgroup ./data

# 设置非特权用户
USER appuser

# 环境变量
ENV PORT=8088 \
    GIN_MODE=release

# 暴露端口
EXPOSE 8088

# 健康检查（使
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8088/ || exit 1

# 使用exec形式启动应用
CMD ["./main"]
