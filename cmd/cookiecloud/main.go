// Package main 是 CookieCloud 应用的入口
// 负责初始化 Web 服务器、注册路由和启动 HTTP 服务
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cookiecloud/internal/cache"
	"cookiecloud/internal/config"
	"cookiecloud/internal/handlers"
	"cookiecloud/internal/logger"
	"cookiecloud/internal/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {
	cfg := config.Load()

	logger.Info("CookieCloud 服务启动", "port", cfg.Port, "api_root", cfg.APIRoot, "data_dir", cfg.DataDir)

	store, err := storage.New(cfg.DataDir)
	if err != nil {
		logger.Error("无法初始化存储", "error", err)
		os.Exit(1)
	}

	// 初始化缓存，TTL设置为5分钟
	dataCache := cache.New(5 * time.Minute)
	logger.Info("缓存已启用", "ttl", "5分钟")

	h := handlers.New(store, dataCache)

	app := fiber.New(fiber.Config{
		// 限制请求体大小为11MB（比handlers层验证稍大一些）
		BodyLimit:             11 * 1024 * 1024,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		DisableStartupMessage: false,
		EnablePrintRoutes:     false,
	})

	// CORS配置（限制允许的来源）
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // 小型项目，暂时允许所有来源，生产环境建议配置具体域名
		AllowMethods:     "GET,POST,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization",
		AllowCredentials: false,
		MaxAge:           86400,
		ExposeHeaders:    "Content-Length",
	}))

	// 速率限制（每分钟60次请求，防止DDoS）
	app.Use(limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.Warn("速率限制触发", "ip", c.IP(), "path", c.Path())
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}))

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	registerRoutes(app, h, cfg.APIRoot)

	logger.Info("服务器监听", "address", "http://localhost:"+cfg.Port+cfg.APIRoot)

	// 启动服务器（在独立的 goroutine 中）
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- app.Listen(":" + cfg.Port)
	}()

	// 信号监听
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 等待服务器错误或信号
	select {
	case err := <-serverErr:
		// 服务器启动失败或运行出错
		logger.Error("服务器异常退出", "error", err)
		os.Exit(1)
	case sig := <-sigChan:
		// 收到关闭信号
		logger.Info("收到关闭信号", "signal", sig)
		gracefulShutdown(app, store)
	}
}

func gracefulShutdown(app *fiber.App, store *storage.Storage) {
	logger.Info("正在优雅关闭...")

	// 关闭存储，忽略错误（程序即将退出）
	_ = store.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭服务器，忽略错误（程序即将退出）
	_ = app.ShutdownWithContext(ctx)

	logger.Info("服务器已关闭")
}

func registerRoutes(app *fiber.App, h *handlers.Handlers, apiRoot string) {
	app.Get(apiRoot+"/", handlers.FiberRootHandler(apiRoot))
	app.Post(apiRoot+"/", handlers.FiberRootHandler(apiRoot))
	app.Post(apiRoot+"/update", h.FiberUpdateHandler)
	app.Get(apiRoot+"/get/:uuid", h.FiberGetHandler)
	app.Post(apiRoot+"/get/:uuid", h.FiberGetHandler)
}
