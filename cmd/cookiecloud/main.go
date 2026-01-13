// Package main 是 CookieCloud 应用的入口
// 负责初始化 Web 服务器、注册路由和启动 HTTP 服务
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cookiecloud/internal/config"
	"cookiecloud/internal/handlers"
	"cookiecloud/internal/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	cfg := config.Load()

	log.Println("[INFO] CookieCloud 服务启动")
	log.Printf("[INFO] 监听端口: %s, API路径: %s, 数据目录: %s\n", cfg.Port, cfg.APIRoot, cfg.DataDir)

	store, err := storage.New(cfg.DataDir)
	if err != nil {
		log.Fatalf("[ERROR] 无法初始化存储: %v\n", err)
	}

	h := handlers.New(store)

	app := fiber.New(fiber.Config{
		BodyLimit:    0,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,OPTIONS",
		AllowHeaders:     "Content-Type",
		AllowCredentials: false,
		MaxAge:           86400,
	}))

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	registerRoutes(app, h, cfg.APIRoot)

	log.Printf("[INFO] 服务器监听: http://localhost:%s%s", cfg.Port, cfg.APIRoot)

	go setupGracefulShutdown(app, store)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("[ERROR] 启动失败: %v", err)
	}
}

func setupGracefulShutdown(app *fiber.App, store *storage.Storage) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("[INFO] 收到关闭信号，正在优雅关闭...")

	_ = store.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_ = app.ShutdownWithContext(ctx)

	log.Println("[INFO] 服务器已关闭")
	os.Exit(0)
}

func registerRoutes(app *fiber.App, h *handlers.Handlers, apiRoot string) {
	app.Get(apiRoot+"/", handlers.FiberRootHandler(apiRoot))
	app.Post(apiRoot+"/", handlers.FiberRootHandler(apiRoot))
	app.Post(apiRoot+"/update", h.FiberUpdateHandler)
	app.Get(apiRoot+"/get/:uuid", h.FiberGetHandler)
	app.Post(apiRoot+"/get/:uuid", h.FiberGetHandler)
}
