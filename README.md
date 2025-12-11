# CookieCloud Go 版本

fork 自 https://github.com/easychen/CookieCloud

CookieCloud 是一个轻量级的 Cookie 和 LocalStorage 同步工具，专为在您的设备之间同步浏览器数据而设计。此仓库提供了原 Node.js 版本服务端的 Go 语言重写，具有更小的镜像体积和更高的运行效率。

## 功能特性

- 📦 **轻量级部署**：Go 版本相比 Node.js 版本具有更小的镜像体积和更低的资源占用

## 项目优势

相比于原版 Node.js 实现，Go 版本具有以下优势：

- 更小的镜像尺寸（基于 Alpine Linux 构建）
- 更高的执行效率和更低的内存占用
- 更简单的部署方式，无需依赖复杂的 Node.js 环境

## 快速开始

使用 Docker compose 快速部署：

```bash yaml
services:
  cookiecloud:
    image: 782042369/cookiecloud:latest
    container_name: cookiecloud-app
    restart: always
    volumes:
      - ./data:/data/api/data
    ports:
      - 8088:8088
```

## 浏览器插件

请前往原项目获取浏览器插件：
https://github.com/easychen/CookieCloud

## 许可证

本项目基于 MIT 许可证开源。
