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
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 打印启动信息
	log.Println("[INFO] CookieCloud 服务启动")
	log.Printf("[INFO] 监听端口: %s\n", cfg.Port)
	log.Printf("[INFO] API 路径: %s\n", cfg.APIRoot)
	log.Printf("[INFO] 数据目录: %s\n", cfg.DataDir)

	// 创建数据目录
	if err := storage.InitDataDir(cfg.DataDir); err != nil {
		log.Fatalf("[ERROR] 无法初始化数据目录: %v\n", err)
	}

	// 创建Fiber应用
	app := fiber.New()

	// 添加CORS中间件
	app.Use(cors.New())

	// 注册路由
	registerRoutes(app, cfg.APIRoot)

	// 启动服务器
	log.Printf("[INFO] 服务器监听: http://localhost:%s%s\n", cfg.Port, cfg.APIRoot)

	// 设置优雅关闭
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("\n[INFO] 收到关闭信号，正在优雅关闭...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Printf("[ERROR] 关闭失败: %v\n", err)
		}

		log.Println("[INFO] 服务器已关闭")
		os.Exit(0)
	}()

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("[ERROR] 启动失败: %v\n", err)
	}
}

// registerRoutes 注册所有路由
func registerRoutes(app *fiber.App, apiRoot string) {
	// 根路径处理器
	app.Get(apiRoot+"/", handlers.FiberRootHandler(apiRoot))
	app.Post(apiRoot+"/", handlers.FiberRootHandler(apiRoot))

	// 更新数据处理器
	app.Post(apiRoot+"/update", handlers.FiberUpdateHandler)

	// 获取数据处理器
	app.Get(apiRoot+"/get/:uuid", handlers.FiberGetHandler)
	app.Post(apiRoot+"/get/:uuid", handlers.FiberGetHandler)
}
