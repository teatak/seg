package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/teatak/cart"
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

	// 配置路径
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

	// 初始化 Cart App
	app := cart.Default()

	// 适配器: http.Handler -> cart.Handler
	// Cart 的 Use/ANY 需要 func(*Context, Next)
	wrap := func(h http.Handler) cart.Handler {
		return func(c *cart.Context, next cart.Next) {
			h.ServeHTTP(c.Response, c.Request)
			next()
		}
	}

	// 1. 注册 API
	apiMux := http.NewServeMux()
	handler.RegisterRoutes(apiMux, "/api")

	app.Route("/api/*path", func(r *cart.Router) {
		r.ANY(wrap(apiMux))
	})

	// 2. 注册前端
	if enableFrontend {
		webMux := http.NewServeMux()
		handler.RegisterFrontend(webMux, "./console/dist", "/api", "/console")

		app.Route("/console/*path", func(r *cart.Router) {
			r.ANY(wrap(webMux))
		})
		// 根路径重定向
		app.Route("/", func(r *cart.Router) {
			r.GET(func(c *cart.Context) {
				c.Redirect(http.StatusFound, "/console/")
			})
		})
	} else {
		app.Route("/", func(r *cart.Router) {
			r.GET(func(c *cart.Context) {
				c.Response.Write([]byte("API Server Running (Cart version)."))
			})
		})
	}

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting Cart server on http://localhost%s", addr)

	if _, err := app.Run(addr); err != nil {
		log.Fatal(err)
	}
}
