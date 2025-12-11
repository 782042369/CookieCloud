package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"cookiecloud/internal/handlers"
	"cookiecloud/internal/storage"

	"github.com/gorilla/mux"
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

	// 创建路由器
	r := mux.NewRouter()

	// 注册路由
	// 根路径处理器
	r.HandleFunc(apiRoot+"/", handlers.RootHandler(apiRoot)).Methods("GET", "POST")

	// 更新数据处理器
	r.HandleFunc(apiRoot+"/update", handlers.UpdateHandler).Methods("POST")

	// 获取数据处理器
	r.HandleFunc(apiRoot+"/get/{uuid}", handlers.GetHandler).Methods("GET", "POST")

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	fmt.Printf("服务器启动于 http://localhost:%s%s\n", port, apiRoot)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
