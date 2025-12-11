package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"cookiecloud/internal/handlers"
	"cookiecloud/internal/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// 创建数据目录（如果不存在）
	if err := storage.InitDataDir(); err != nil {
		log.Fatalf("无法初始化数据目录: %v", err)
	}

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
	// 根路径处理器
	app.Get(apiRoot+"/", handlers.FiberRootHandler(apiRoot))
	app.Post(apiRoot+"/", handlers.FiberRootHandler(apiRoot))

	// 更新数据处理器
	app.Post(apiRoot+"/update", handlers.FiberUpdateHandler)

	// 获取数据处理器
	app.Get(apiRoot+"/get/:uuid", handlers.FiberGetHandler)
	app.Post(apiRoot+"/get/:uuid", handlers.FiberGetHandler)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	fmt.Printf("服务器启动于 http://localhost:%s%s\n", port, apiRoot)
	log.Fatal(app.Listen(":" + port))
}
