package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"cookiecloud/internal/handlers"
	"cookiecloud/internal/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// 创建数据目录和初始化数据库（如果不存在）
	if err := storage.InitDataDir(); err != nil {
		log.Fatalf("无法初始化数据目录和数据库: %v", err)
	}

	// 注册信号处理以优雅关闭
	go handleSignals()

	// 从环境变量获取API根路径
	apiRoot := os.Getenv("API_ROOT")
	if apiRoot != "" {
		apiRoot = strings.TrimSuffix(apiRoot, "/")
	}

	// 创建Fiber应用
	app := fiber.New()

	// 添加CORS中间件
	app.Use(cors.New())

	// 注册路由
	registerRoutes(app, apiRoot)

	// 启动服务器
	port := getPort()
	
	fmt.Printf("服务器启动于 http://localhost:%s%s\n", port, apiRoot)
	log.Fatal(app.Listen(":" + port))
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

// getPort 获取端口号，优先从环境变量获取
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8088"
	}
	return port
}

// handleSignals 处理系统信号以优雅关闭应用
func handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n接收到终止信号，正在关闭...")
	storage.CloseDB()
	os.Exit(0)
}