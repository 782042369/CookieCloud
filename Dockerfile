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
    -a -ldflags="-w -s -extldflags -static" \
    -trimpath \
    -o main ./cmd/cookiecloud

# 最终生产阶段 - 使用最小的scratch基础镜像
FROM scratch

# 设置工作目录
WORKDIR /

# 从构建阶段复制二进制文件
COPY --from=builder /app/main /main

# 创建数据目录
RUN mkdir -p /data

# 环境变量
ENV PORT=8088

# 暴露端口
EXPOSE 8088


# 使用exec形式启动应用
CMD ["/main"]
