package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/teatak/seg/pkg/api"
	"github.com/teatak/seg/pkg/engine"
)

func main() {
	var config engine.Config
	var port int
	var dataDir string
	var enableFrontend bool

	// 解析命令行参数
	flag.IntVar(&port, "port", 8080, "HTTP server port")
	flag.StringVar(&dataDir, "data", "./data", "Data directory")
	flag.BoolVar(&enableFrontend, "web", true, "Enable web frontend")
	flag.Parse()

	config.DataDir = dataDir
	config.BaseDict = filepath.Join(dataDir, "dict/base.txt")
	config.UserDict = filepath.Join(dataDir, "dict/user.txt")
	config.StagingDict = filepath.Join(dataDir, "dict/staging.txt")
	config.FeedbackDB = filepath.Join(dataDir, "feedback.json")
	config.RequestLog = filepath.Join(dataDir, "requests.json")

	// 初始化引擎
	e, err := engine.NewEngine(config)
	if err != nil {
		log.Fatalf("Failed to initialize engine: %v", err)
	}

	// 初始化 API 处理器
	handler := api.NewHandler(e)

	// 获取默认 ServeMux
	mux := http.DefaultServeMux

	// 注册 API 路由
	handler.RegisterRoutes(mux, "/api")

	// 静态文件服务 (可选)
	if enableFrontend {
		// 注册前端路由到 /console
		handler.RegisterFrontend(mux, "./console/dist", "/api", "/console")

		// 根路径重定向到 /console
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.Redirect(w, r, "/console/", http.StatusFound)
			} else {
				http.NotFound(w, r)
			}
		})
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "API Server Running (Web frontend disabled).")
		})
	}

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server starting on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
