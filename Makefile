.PHONY: all build build-web build-server run clean train test vet fmt help

# 默认目标
all: build

# 帮助信息
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build         Build both frontend and backend"
	@echo "  build-web     Build frontend (React/Vite)"
	@echo "  build-server  Build backend (Go)"
	@echo "  run           Run the server (Go)"
	@echo "  train         Train HMM model"
	@echo "  test          Run Go tests"
	@echo "  vet           Run Go vet"
	@echo "  fmt           Format Go code"
	@echo "  clean         Clean build artifacts"

# 构建所有 (前端 + 后端)
build: build-web build-server

# 构建前端
build-web:
	@echo "Building frontend..."
	cd web && npm install && npm run build

# 构建后端
build-server:
	@echo "Building backend..."
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/server cmd/server/main.go

# 运行服务 (会自动使用构建好的前端静态文件)
run:
	@echo "Running server..."
	go run cmd/server/main.go

# 训练 HMM 模型 (使用示例语料)
train:
	@echo "Training HMM model..."
	go run cmd/train_hmm/main.go -corpus data/corpus.txt

# 代码检查与测试
vet:
	go vet ./...

test:
	go test -v ./...

fmt:
	go fmt ./...

# 清理构建产物
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf web/dist/
	# 谨慎清理 node_modules，避免重复下载
	# rm -rf web/node_modules/
	go clean
