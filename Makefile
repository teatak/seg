.PHONY: all build build-web build-server run clean train

# 默认目标
all: build

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
	go build -o bin/server cmd/server/main.go

# 运行服务 (会自动使用构建好的前端静态文件)
run:
	@echo "Running server..."
	go run cmd/server/main.go

# 训练 HMM 模型 (使用示例语料)
train:
	@echo "Training HMM model..."
	go run cmd/train_hmm/main.go -corpus data/corpus.txt

# 清理构建产物
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf web/dist/
	rm -rf web/node_modules/
