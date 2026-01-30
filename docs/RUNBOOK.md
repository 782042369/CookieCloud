# CookieCloud - 运维手册

> 最后更新：2026-01-30（添加 .env.example 配置示例）

本文档面向 CookieCloud Go 版本的运维工程师，介绍部署流程、监控告警、故障排查等核心内容。

## 目录

- [部署指南](#部署指南)
- [配置管理](#配置管理)
- [监控与告警](#监控与告警)
- [日志管理](#日志管理)
- [故障排查](#故障排查)
- [性能调优](#性能调优)
- [安全加固](#安全加固)
- [备份与恢复](#备份与恢复)

---

## 部署指南

### 系统要求

| 资源 | 最低配置 | 推荐配置 |
|------|---------|---------|
| CPU | 1 核 | 2 核+ |
| 内存 | 128MB | 512MB+ |
| 磁盘 | 100MB | 1GB+（根据数据量） |
| 网络 | 1Mbps | 10Mbps+ |
| 操作系统 | Linux (amd64) | Ubuntu 20.04+ / CentOS 8+ |

### Docker 部署（推荐）

#### 1. 使用 Docker Compose

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  cookiecloud:
    image: 782042369/cookiecloud:latest
    container_name: cookiecloud-app
    restart: always

    # 环境变量
    environment:
      - PORT=8088
      - API_ROOT=/api
      - TZ=Asia/Shanghai

    # 端口映射
    ports:
      - "8088:8088"

    # 数据卷（持久化存储）
    volumes:
      - ./data:/data

    # 资源限制（小型个人项目）
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 64M

    # 健康检查
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8088/api/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

    # 日志配置
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

启动服务：

```bash
# 启动
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止
docker-compose down

# 重启
docker-compose restart
```

#### 2. 使用 Docker 命令

```bash
# 拉取镜像
docker pull 782042369/cookiecloud:latest

# 运行容器
docker run -d \
  --name cookiecloud-app \
  --restart always \
  -p 8088:8088 \
  -v $(pwd)/data:/data \
  -e PORT=8088 \
  -e TZ=Asia/Shanghai \
  782042369/cookiecloud:latest

# 查看容器状态
docker ps | grep cookiecloud

# 查看日志
docker logs -f cookiecloud-app

# 进入容器（注意：scratch 镜像无 shell）
# 如需调试，请使用 alpine 基础镜像重新构建
```

### 二进制部署

#### 1. 下载预编译二进制

```bash
# 下载最新版本（从 GitHub Releases）
wget https://github.com/782042369/CookieCloud/releases/latest/download/cookiecloud-linux-amd64

# 赋予执行权限
chmod +x cookiecloud-linux-amd64

# 移动到系统路径
sudo mv cookiecloud-linux-amd64 /usr/local/bin/cookiecloud
```

#### 2. 编译安装

```bash
# 克隆仓库
git clone https://github.com/782042369/CookieCloud.git
cd CookieCloud

# 编译（静态链接）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-s -w" -o cookiecloud ./cmd/cookiecloud

# 安装
sudo install -m 755 cookiecloud /usr/local/bin/
```

#### 3. 创建 systemd 服务

创建 `/etc/systemd/system/cookiecloud.service`：

```ini
[Unit]
Description=CookieCloud Service
After=network.target

[Service]
Type=simple
User=cookiecloud
Group=cookiecloud
WorkingDirectory=/opt/cookiecloud
ExecStart=/usr/local/bin/cookiecloud
Restart=always
RestartSec=5

# 环境变量
Environment="PORT=8088"
Environment="API_ROOT=/api"
Environment="DATA_DIR=/var/lib/cookiecloud"

# 资源限制
LimitNOFILE=65536
LimitNPROC=4096

# 安全加固
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/cookiecloud

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
# 创建用户和目录
sudo useradd -r -s /bin/false cookiecloud
sudo mkdir -p /var/lib/cookiecloud
sudo chown -R cookiecloud:cookiecloud /var/lib/cookiecloud

# 重载配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start cookiecloud

# 开机自启
sudo systemctl enable cookiecloud

# 查看状态
sudo systemctl status cookiecloud

# 查看日志
sudo journalctl -u cookiecloud -f
```

### Kubernetes 部署

创建 `k8s-deployment.yaml`：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cookiecloud
  labels:
    app: cookiecloud
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cookiecloud
  template:
    metadata:
      labels:
        app: cookiecloud
    spec:
      containers:
      - name: cookiecloud
        image: 782042369/cookiecloud:latest
        ports:
        - containerPort: 8088
        env:
        - name: PORT
          value: "8088"
        - name: API_ROOT
          value: "/api"
        - name: TZ
          value: "Asia/Shanghai"
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "256Mi"
            cpu: "250m"
        volumeMounts:
        - name: data
          mountPath: /data
        livenessProbe:
          httpGet:
            path: /api/
            port: 8088
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /api/
            port: 8088
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: cookiecloud-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: cookiecloud
spec:
  selector:
    app: cookiecloud
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8088
  type: LoadBalancer
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cookiecloud-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```

部署：

```bash
# 部署
kubectl apply -f k8s-deployment.yaml

# 查看状态
kubectl get pods -l app=cookiecloud
kubectl get svc cookiecloud

# 查看日志
kubectl logs -f deployment/cookiecloud

# 扩缩容
kubectl scale deployment cookiecloud --replicas=3
```

---

## 配置管理

### 环境变量说明

**完整的环境变量配置示例**请参考：@.env.example

| 变量名 | 默认值 | 说明 | 示例 |
|--------|--------|------|------|
| `PORT` | `8088` | HTTP 服务端口 | `8088` |
| `API_ROOT` | `` | API 路径前缀（自动去除尾部斜杠） | `/api` |
| `DATA_DIR` | `./data` | 数据存储目录 | `/var/lib/cookiecloud` |
| `TZ` | `UTC` | 时区（Docker 镜像默认 `Asia/Shanghai`） | `Asia/Shanghai` |

### 配置文件示例

#### Docker Compose

```yaml
environment:
  - PORT=8088
  - API_ROOT=/api
  - TZ=Asia/Shanghai
```

#### systemd

```ini
[Service]
Environment="PORT=8088"
Environment="API_ROOT=/api"
Environment="DATA_DIR=/var/lib/cookiecloud"
```

#### Kubernetes

```yaml
env:
- name: PORT
  value: "8088"
- name: API_ROOT
  value: "/api"
- name: TZ
  value: "Asia/Shanghai"
```

### Nginx 反向代理

```nginx
upstream cookiecloud {
    server 127.0.0.1:8088;
    keepalive 32;
}

server {
    listen 80;
    server_name cookie.example.com;

    # 强制 HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name cookie.example.com;

    # SSL 证书
    ssl_certificate /etc/letsencrypt/live/cookie.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/cookie.example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # 日志
    access_log /var/log/nginx/cookiecloud-access.log;
    error_log /var/log/nginx/cookiecloud-error.log;

    # 反向代理
    location /api/ {
        proxy_pass http://cookiecloud/api/;
        proxy_http_version 1.1;

        # 请求头
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 超时设置
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;

        # 缓冲设置
        proxy_buffering off;
        proxy_request_buffering off;
    }

    # 健康检查
    location /health {
        proxy_pass http://cookiecloud/api/;
        access_log off;
    }
}
```

---

## 监控与告警

### 健康检查

```bash
# 简单健康检查
curl http://localhost:8088/api/

# 预期输出
Hello World! API ROOT = /api
```

### Prometheus 监控

由于项目未集成 Prometheus 端点，建议通过 **Blackbox Exporter** 进行监控：

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'cookiecloud'
    metrics_path: /probe
    params:
      module:
      - http_2xx
    static_configs:
      - targets:
        - http://cookiecloud:8088/api/
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: blackbox-exporter:9115
```

### Grafana 仪表盘

推荐监控指标：

| 指标 | 说明 | 告警阈值 |
|------|------|---------|
| 服务可用性 | HTTP 状态码 200 | < 99.9% |
| 响应时间 | 请求延迟 | > 1s |
| 磁盘空间 | DATA_DIR 使用率 | > 80% |
| CPU 使用率 | 容器/进程 CPU | > 80% |
| 内存使用率 | 容器/进程内存 | > 80% |
| 速率限制 | 触发次数/分钟 | > 100 |

### 告警规则（Prometheus）

```yaml
groups:
  - name: cookiecloud
    rules:
      - alert: CookieCloudServiceDown
        expr: up{job="cookiecloud"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "CookieCloud 服务宕机"
          description: "实例 {{ $labels.instance }} 已宕机超过 1 分钟"

      - alert: CookieCloudHighLatency
        expr: http_request_duration_seconds{quantile="0.95"} > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "CookieCloud 响应缓慢"
          description: "P95 延迟超过 1 秒"

      - alert: CookieCloudDiskSpaceHigh
        expr: node_filesystem_avail_bytes{mountpoint="/data"} / node_filesystem_size_bytes{mountpoint="/data"} < 0.2
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "CookieCloud 磁盘空间不足"
          description: "数据目录剩余空间不足 20%"
```

---

## 日志管理

### 日志格式

```
[LEVEL] YYYY/MM/DD HH:MM:SS file:line: message | key1=value1 | key2=value2
```

**示例**：

```
[INFO] 2025/01/28 12:44:21 main.go:27: 服务启动 | port=8088 | api_root=/api
[WARN] 2025/01/28 12:44:21 main.go:68: 速率限制触发 | ip=127.0.0.1 | path=/api/update
[ERROR] 2025/01/28 12:44:21 handlers.go:83: 文件写入失败 | uuid=test-uuid | error=...
```

### 日志级别

| 级别 | 用途 | 输出目标 |
|------|------|---------|
| `INFO` | 正常运行信息 | 标准输出 (stdout) |
| `WARN` | 警告信息（速率限制等） | 标准输出 (stdout) |
| `ERROR` | 错误信息 | 标准错误 (stderr) |

### 日志收集（ELK）

#### Docker 日志驱动

```yaml
services:
  cookiecloud:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
        labels: "service"
```

#### Filebeat 收集

`filebeat.yml`：

```yaml
filebeat.inputs:
  - type: container
    paths:
      - '/var/lib/docker/containers/*/*.log'
    processors:
      - add_docker_metadata:
          host: "unix:///var/run/docker.sock"

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  indices:
    - index: "cookiecloud-%{+yyyy.MM.dd}"
```

### 日志分析查询（Elasticsearch）

```bash
# 查询最近 1 小时的错误日志
curl -X GET "elasticsearch:9200/cookiecloud-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        { "match": { "level": "ERROR" } },
        { "range": { "@timestamp": { "gte": "now-1h" } } }
      ]
    }
  }
}
'

# 统计速率限制触发次数
curl -X GET "elasticsearch:9200/cookiecloud-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": { "message": "速率限制触发" }
  },
  "aggs": {
    "by_ip": {
      "terms": { "field": "ip.keyword" }
    }
  }
}
'
```

---

## 故障排查

### 常见问题（2026-01-30 更新）

#### 1. 服务无法启动

**症状**：容器/进程立即退出

**排查步骤**：

```bash
# 查看日志
docker logs cookiecloud-app
# 或
journalctl -u cookiecloud -n 50

# 检查端口占用
sudo netstat -tulpn | grep 8088
# 或
sudo lsof -i :8088

# 检查数据目录权限
ls -la ./data
sudo chown -R cookiecloud:cookiecloud ./data

# 检查磁盘空间
df -h ./data
```

**常见原因**：
- 端口被占用
- 数据目录无写权限
- 磁盘空间不足
- 环境变量配置错误

#### 2. 请求返回 500 错误

**症状**：API 请求失败，HTTP 状态码 500

**排查步骤**：

```bash
# 查看错误日志
docker logs cookiecloud-app 2>&1 | grep ERROR

# 检查磁盘 I/O
iostat -x 1

# 检查文件完整性
ls -lh ./data/*.json

# 手动测试 API
curl -v http://localhost:8088/api/get/test-uuid
```

**常见原因**：
- 磁盘 I/O 异常
- 数据文件损坏
- 内存不足

#### 3. 速率限制触发

**症状**：HTTP 状态码 429

**排查步骤**：

```bash
# 查看触发日志
docker logs cookiecloud-app 2>&1 | grep "速率限制触发"

# 识别异常 IP
docker logs cookiecloud-app 2>&1 | grep "速率限制触发" | awk '{print $NF}' | sort | uniq -c | sort -rn
```

**解决方案**：
- 临时封禁异常 IP（防火墙/Nginx）
- 调整速率限制参数（修改代码）
- 验证是否为 DDoS 攻击

#### 4. 内存泄漏

**症状**：内存使用率持续增长

**排查步骤**：

```bash
# 监控内存使用
docker stats cookiecloud-app

# 生成内存 profile
go tool pprof http://localhost:8088/debug/pprof/heap
```

**解决方案**：
- 重启服务（临时）
- 检查代码中的内存泄漏（goroutine 泄漏、未关闭的连接等）
- 调整缓存 TTL

### 故障排查命令清单

```bash
# 服务状态
systemctl status cookiecloud
docker ps | grep cookiecloud
kubectl get pods -l app=cookiecloud

# 日志查看
journalctl -u cookiecloud -f
docker logs -f cookiecloud-app
kubectl logs -f deployment/cookiecloud

# 端口监听
sudo netstat -tulpn | grep 8088
sudo ss -tulpn | grep 8088

# 资源使用
docker stats cookiecloud-app
top -p $(pgrep cookiecloud)

# 磁盘空间
df -h ./data
du -sh ./data/*

# 网络连通性
curl -v http://localhost:8088/api/
ping -c 3 example.com
```

---

## 性能调优

### 1. 调整速率限制

修改 `cmd/cookiecloud/main.go`：

```go
app.Use(limiter.New(limiter.Config{
    Max:        120,           // 增加：60 -> 120
    Expiration: 30 * time.Second,  // 缩短：1分钟 -> 30秒
    KeyGenerator: func(c *fiber.Ctx) string {
        return c.IP()
    },
}))
```

### 2. 调整缓存 TTL

修改 `cmd/cookiecloud/main.go`：

```go
// 从 5 分钟调整为 10 分钟
dataCache := cache.New(10 * time.Minute)
```

### 3. 调整请求体大小限制

修改 `cmd/cookiecloud/main.go`：

```go
app := fiber.New(fiber.Config{
    BodyLimit: 20 * 1024 * 1024,  // 增加：11MB -> 20MB
    // ...
})
```

### 4. 数据目录优化

```bash
# 使用 SSD 存储
sudo mount -t ext4 /dev/sdb1 /var/lib/cookiecloud

# 调整挂载选项（noatime 减少写入）
/dev/sdb1 /var/lib/cookiecloud ext4 defaults,noatime 0 2

# 定期清理过期数据（Cron）
0 3 * * * find /var/lib/cookiecloud -name "*.json" -mtime +30 -delete
```

---

## 安全加固

### 1. 最小权限原则

```bash
# 创建专用用户
sudo useradd -r -s /bin/false cookiecloud

# 设置数据目录权限
sudo chown -R cookiecloud:cookiecloud /var/lib/cookiecloud
sudo chmod 750 /var/lib/cookiecloud
```

### 2. 防火墙配置

```bash
# UFW (Ubuntu)
sudo ufw allow 8088/tcp
sudo ufw enable

# firewalld (CentOS)
sudo firewall-cmd --permanent --add-port=8088/tcp
sudo firewall-cmd --reload

# iptables
sudo iptables -A INPUT -p tcp --dport 8088 -j ACCEPT
sudo iptables -A INPUT -j DROP
```

### 3. 启用 HTTPS

使用 Let's Encrypt：

```bash
# 安装 certbot
sudo apt install certbot

# 生成证书
sudo certbot certonly --standalone -d cookie.example.com

# 自动续期（Cron）
0 0 * * * certbot renew --quiet
```

### 4. 限制 CORS（生产环境）

修改 `cmd/cookiecloud/main.go`：

```go
app.Use(cors.New(cors.Config{
    AllowOrigins:     "https://yourdomain.com",  // 替换 *
    AllowMethods:     "GET,POST,OPTIONS",
    AllowHeaders:     "Content-Type,Authorization",
    AllowCredentials: false,
    MaxAge:           86400,
}))
```

### 5. 安全响应头

```go
app.Use(func(c *fiber.Ctx) error {
    c.Set("X-Content-Type-Options", "nosniff")
    c.Set("X-Frame-Options", "DENY")
    c.Set("X-XSS-Protection", "1; mode=block")
    c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    return c.Next()
})
```

---

## 备份与恢复

### 数据备份

#### 手动备份

```bash
# 创建备份目录
BACKUP_DIR="/backup/cookiecloud-$(date +%Y%m%d)"
mkdir -p "$BACKUP_DIR"

# 复制数据
cp -r ./data "$BACKUP_DIR/"

# 压缩
tar czf "$BACKUP_DIR.tar.gz" "$BACKUP_DIR"
rm -rf "$BACKUP_DIR"

# 上传到远程存储（示例：AWS S3）
aws s3 cp "$BACKUP_DIR.tar.gz" s3://my-bucket/cookiecloud/
```

#### 自动备份（Cron）

```bash
# /etc/cron.d/cookiecloud-backup
0 2 * * * root /opt/scripts/backup-cookiecloud.sh
```

`backup-cookiecloud.sh`：

```bash
#!/bin/bash
BACKUP_DIR="/backup/cookiecloud-$(date +%Y%m%d)"
mkdir -p "$BACKUP_DIR"
cp -r /var/lib/cookiecloud "$BACKUP_DIR/"
tar czf "$BACKUP_DIR.tar.gz" "$BACKUP_DIR"
rm -rf "$BACKUP_DIR"

# 保留最近 7 天的备份
find /backup -name "cookiecloud-*.tar.gz" -mtime +7 -delete

# 上传到远程
aws s3 cp "$BACKUP_DIR.tar.gz" s3://my-bucket/cookiecloud/
```

### 数据恢复

```bash
# 停止服务
systemctl stop cookiecloud
# 或
docker-compose stop

# 恢复数据
tar xzf /backup/cookiecloud-20250128.tar.gz
cp -r cookiecloud-20250128/data/* ./data/

# 调整权限
chown -R cookiecloud:cookiecloud ./data

# 启动服务
systemctl start cookiecloud
# 或
docker-compose start
```

### 灾难恢复

#### 1. 服务器完全损坏

```bash
# 在新服务器上安装环境
# （参考部署指南）

# 从远程存储下载最新备份
aws s3 cp s3://my-bucket/cookiecloud/cookiecloud-latest.tar.gz ./

# 恢复数据
tar xzf cookiecloud-latest.tar.gz
sudo cp -r cookiecloud-data/* /var/lib/cookiecloud/

# 启动服务
sudo systemctl start cookiecloud
```

#### 2. 数据损坏

```bash
# 识别损坏的文件
find ./data -name "*.json" -exec sh -c 'jq empty {} && echo "OK: {}" || echo "CORRUPTED: {}"' \;

# 从备份恢复特定文件
tar xzf /backup/cookiecloud-20250128.tar.gz
cp cookiecloud-20250128/data/uuid-xxx.json ./data/
```

---

## 版本升级

### Docker 升级

```bash
# 拉取最新镜像
docker pull 782042369/cookiecloud:latest

# 停止并删除旧容器
docker-compose down

# 启动新容器
docker-compose up -d

# 验证
curl http://localhost:8088/api/
```

### 二进制升级

```bash
# 下载新版本
wget https://github.com/782042369/CookieCloud/releases/latest/download/cookiecloud-linux-amd64

# 备份旧版本
sudo cp /usr/local/bin/cookiecloud /usr/local/bin/cookiecloud.bak

# 安装新版本
sudo install -m 755 cookiecloud-linux-amd64 /usr/local/bin/cookiecloud

# 重启服务
sudo systemctl restart cookiecloud

# 验证
curl http://localhost:8088/api/

# 如有问题，回滚
sudo systemctl stop cookiecloud
sudo cp /usr/local/bin/cookiecloud.bak /usr/local/bin/cookiecloud
sudo systemctl start cookiecloud
```

### Kubernetes 滚动升级

```bash
# 更新镜像版本
kubectl set image deployment/cookiecloud cookiecloud=782042369/cookiecloud:v1.2.3

# 查看滚动升级状态
kubectl rollout status deployment/cookiecloud

# 如有问题，回滚
kubectl rollout undo deployment/cookiecloud
```

---

## 参考资源

- [Docker 官方文档](https://docs.docker.com/)
- [Kubernetes 官方文档](https://kubernetes.io/docs/)
- [Fiber 框架文档](https://docs.gofiber.io/)
- [Prometheus 监控最佳实践](https://prometheus.io/docs/practices/)

---

**遇到问题？** 请提交 [Issue](https://github.com/782042369/CookieCloud/issues)
