# 构建阶段：使用多阶段构建最小化生产镜像
# 阶段一：构建
FROM node:24-alpine as builder
WORKDIR /app

# 安装 pnpm 并配置缓存
RUN npm i -g pnpm@10 && \
    pnpm config set store-dir /root/.pnpm-store

# 优先复制包管理文件以利用构建缓存
COPY package.json pnpm-lock.yaml ./

# 安装所有依赖（包括devDependencies）
RUN  pnpm install --frozen-lockfile

COPY . .

# 执行构建
RUN rm -rf node_modules && \
    pnpm install --prod --frozen-lockfile && \
    pnpm add @vercel/nft fs-extra --save-prod

# 生产阶段：仅安装生产依赖
FROM node:24-alpine as production-deps

WORKDIR /app

COPY --from=builder /app /app

RUN export PROJECT_ROOT=/app/ && \
    node /app/scripts/minify-docker.cjs && \
    rm -rf /app/node_modules /app/scripts && \
    mv /app/app-minimal/node_modules /app/ && \
    rm -rf /app/app-minimal

# 最终生产阶段
FROM node:24-alpine
WORKDIR /app

# 环境变量
ENV NODE_ENV=production \
# 设置时区
  TZ="Asia/Shanghai"

RUN
# 从各阶段复制必要文件
COPY --from=builder /app/app.js ./
COPY --from=production-deps /app/node_modules ./node_modules/


CMD ["node", "app.js"]

# 暴露端口
EXPOSE 8088
