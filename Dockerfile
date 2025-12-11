# 构建阶段：使用多阶段构建最小化生产镜像
# 阶段一：构建 service (Go版本)
FROM golang:1.25-alpine as service-builder
WORKDIR /app

# 复制Go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY cmd ./cmd
COPY internal ./internal

# 构建优化的Go应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -trimpath -o main ./cmd/cookiecloud


# 最终生产阶段
FROM alpine:latest
WORKDIR /app

# 安装ca-certificates以支持HTTPS请求
RUN apk --no-cache add ca-certificates

# 创建非特权用户
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 从 service-builder 阶段复制所有必要文件
COPY --from=service-builder --chown=appuser:appgroup /app/main ./main

# 创建 data 目录并设置正确的所有权
RUN mkdir -p ./data && chown appuser:appgroup ./data

# 设置用户权限
USER appuser

# 声明端口
ENV PORT=8088

EXPOSE 8088

# 添加健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8088/ || exit 1

CMD ["./main"]
